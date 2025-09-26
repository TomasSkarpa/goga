package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type Config struct {
	AIAPIKey string `json:"aiApiKey"`
}

type ConfigHandler struct {
	configPath string
	config     *Config
}

func NewConfigHandler() *ConfigHandler {
	return &ConfigHandler{
		configPath: "./config.json",
		config:     &Config{},
	}
}

func (h *ConfigHandler) LoadConfig() error {
	data, err := os.ReadFile(h.configPath)
	if err != nil {
		return nil // Use defaults if config doesn't exist
	}
	
	if err := json.Unmarshal(data, h.config); err != nil {
		return err
	}
	
	// Decrypt API key after loading
	if h.config.AIAPIKey != "" {
		decrypted, err := h.decrypt(h.config.AIAPIKey)
		if err == nil {
			h.config.AIAPIKey = decrypted
		}
	}
	
	return nil
}

func (h *ConfigHandler) SaveConfig() error {
	// Encrypt API key before saving
	encryptedConfig := *h.config
	if h.config.AIAPIKey != "" {
		encrypted, err := h.encrypt(h.config.AIAPIKey)
		if err != nil {
			return err
		}
		encryptedConfig.AIAPIKey = encrypted
	}
	
	data, err := json.Marshal(encryptedConfig)
	if err != nil {
		return err
	}
	return os.WriteFile(h.configPath, data, 0600) // More restrictive permissions
}



func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	var newConfig Config
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if newConfig.AIAPIKey != "" {
		h.config.AIAPIKey = newConfig.AIAPIKey
	}

	if err := h.SaveConfig(); err != nil {
		log.Printf("Failed to save config: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration saved"})
}



func (h *ConfigHandler) GetConfig(c *gin.Context) {
	// Don't return actual API key, just indicate if it's set
	response := map[string]interface{}{
		"aiApiKey": "", // Never return the actual key
		"hasApiKey": h.config.AIAPIKey != "",
	}
	c.JSON(http.StatusOK, response)
}

func (h *ConfigHandler) encrypt(text string) (string, error) {
	key := h.getEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	
	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (h *ConfigHandler) decrypt(encryptedText string) (string, error) {
	key := h.getEncryptionKey()
	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}
	
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", err
	}
	
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	
	return string(plaintext), nil
}

func (h *ConfigHandler) getEncryptionKey() []byte {
	// Generate key from machine-specific data
	hostname, _ := os.Hostname()
	wd, _ := os.Getwd()
	seed := "goga-" + hostname + "-" + wd
	hash := sha256.Sum256([]byte(seed))
	return hash[:]
}