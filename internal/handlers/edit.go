package handlers

import (
	"goga/internal/repository"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"os"
	"path/filepath"

	"github.com/disintegration/gift"
	"github.com/disintegration/imaging"
	"github.com/gin-gonic/gin"
)

type EditHandler struct {
	repo      *repository.ImageRepository
	uploadDir string
}

type EditRequest struct {
	Brightness float64 `json:"brightness"`
	Contrast   float64 `json:"contrast"`
	Saturation float64 `json:"saturation"`
	Hue        float64 `json:"hue"`
	Gamma      float64 `json:"gamma"`
	Blur       float64 `json:"blur"`
	Sharpen    float64 `json:"sharpen"`
}

func NewEditHandler(repo *repository.ImageRepository, uploadDir string) *EditHandler {
	return &EditHandler{
		repo:      repo,
		uploadDir: uploadDir,
	}
}

func (h *EditHandler) PreviewEdit(c *gin.Context) {
	id := c.Param("id")
	
	var req EditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageRecord, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Open original image
	src, err := imaging.Open(imageRecord.Path, imaging.AutoOrientation(true))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
		return
	}

	// Apply filters using gift library
	g := gift.New()
	
	if req.Brightness != 0 {
		// Convert -100 to 100 range to -30 to 30
		g.Add(gift.Brightness(float32(req.Brightness * 0.3)))
	}
	if req.Contrast != 0 {
		// Convert -100 to 100 range to -50 to 50
		g.Add(gift.Contrast(float32(req.Contrast * 0.5)))
	}
	if req.Saturation != 0 {
		// Convert -100 to 100 range to -100 to 500
		value := float32(100 + req.Saturation*2)
		if value < 0 { value = 0 }
		g.Add(gift.Saturation(value))
	}
	if req.Hue != 0 {
		g.Add(gift.Hue(float32(req.Hue)))
	}
	if req.Gamma != 0 {
		// Convert -100 to 100 range to 0.3 to 3.0
		value := float32(1.0 + req.Gamma*0.02)
		if value < 0.1 { value = 0.1 }
		if value > 5.0 { value = 5.0 }
		g.Add(gift.Gamma(value))
	}
	if req.Blur > 0 {
		g.Add(gift.GaussianBlur(float32(req.Blur)))
	}
	if req.Sharpen > 0 {
		g.Add(gift.UnsharpMask(float32(req.Sharpen), 1.0, 0.05))
	}

	// Create destination image
	dst := image.NewRGBA(g.Bounds(src.Bounds()))
	g.Draw(dst, src)

	// Set response headers
	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "no-cache")

	// Encode and send
	jpeg.Encode(c.Writer, dst, &jpeg.Options{Quality: 85})
}

func (h *EditHandler) ApplyEdit(c *gin.Context) {
	id := c.Param("id")
	
	var req EditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	imageRecord, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Create backup of original
	backupPath := filepath.Join(h.uploadDir, "backups", imageRecord.Filename)
	os.MkdirAll(filepath.Dir(backupPath), 0755)
	
	// Copy original to backup
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		src, _ := os.Open(imageRecord.Path)
		dst, _ := os.Create(backupPath)
		defer src.Close()
		defer dst.Close()
		src.WriteTo(dst)
	}

	// Apply edits and save
	src, err := imaging.Open(imageRecord.Path, imaging.AutoOrientation(true))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image"})
		return
	}

	// Apply filters
	g := gift.New()
	
	if req.Brightness != 0 {
		g.Add(gift.Brightness(float32(req.Brightness * 0.3)))
	}
	if req.Contrast != 0 {
		g.Add(gift.Contrast(float32(req.Contrast * 0.5)))
	}
	if req.Saturation != 0 {
		value := float32(100 + req.Saturation*2)
		if value < 0 { value = 0 }
		g.Add(gift.Saturation(value))
	}
	if req.Hue != 0 {
		g.Add(gift.Hue(float32(req.Hue)))
	}
	if req.Gamma != 0 {
		value := float32(1.0 + req.Gamma*0.02)
		if value < 0.1 { value = 0.1 }
		if value > 5.0 { value = 5.0 }
		g.Add(gift.Gamma(value))
	}
	if req.Blur > 0 {
		g.Add(gift.GaussianBlur(float32(req.Blur)))
	}
	if req.Sharpen > 0 {
		g.Add(gift.UnsharpMask(float32(req.Sharpen), 1.0, 0.05))
	}

	dst := image.NewRGBA(g.Bounds(src.Bounds()))
	g.Draw(dst, src)

	// Save edited image
	file, err := os.Create(imageRecord.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}
	defer file.Close()

	// Determine format and encode
	ext := filepath.Ext(imageRecord.Path)
	switch ext {
	case ".png":
		png.Encode(file, dst)
	default:
		jpeg.Encode(file, dst, &jpeg.Options{Quality: 95})
	}

	// Clear thumbnails cache
	thumbDir := filepath.Join(h.uploadDir, "thumbs")
	if _, err := os.Stat(thumbDir); err == nil {
		filepath.Walk(thumbDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			filename := filepath.Base(path)
			if len(filename) >= len(id) && filename[:len(id)] == id {
				os.Remove(path)
			}
			return nil
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image edited successfully"})
}

func (h *EditHandler) ResetImage(c *gin.Context) {
	id := c.Param("id")
	
	imageRecord, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Restore from backup
	backupPath := filepath.Join(h.uploadDir, "backups", imageRecord.Filename)
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "No backup found"})
		return
	}

	// Copy backup to original
	src, err := os.Open(backupPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open backup"})
		return
	}
	defer src.Close()

	dst, err := os.Create(imageRecord.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore image"})
		return
	}
	defer dst.Close()

	src.WriteTo(dst)

	// Clear thumbnails cache
	thumbDir := filepath.Join(h.uploadDir, "thumbs")
	if _, err := os.Stat(thumbDir); err == nil {
		filepath.Walk(thumbDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			filename := filepath.Base(path)
			if len(filename) >= len(id) && filename[:len(id)] == id {
				os.Remove(path)
			}
			return nil
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image reset to original"})
}