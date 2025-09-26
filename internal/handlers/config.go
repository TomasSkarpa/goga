package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration saved"})
}



func (h *ConfigHandler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, h.config)
}