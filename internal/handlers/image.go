package handlers

import (
	"fmt"
	"goga/internal/models"
	"goga/internal/repository"
	"goga/pkg/utils"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImageHandler struct {
	repo      *repository.ImageRepository
	uploadDir string
}

func NewImageHandler(repo *repository.ImageRepository, uploadDir string) *ImageHandler {
	return &ImageHandler{
		repo:      repo,
		uploadDir: uploadDir,
	}
}

func (h *ImageHandler) GetImages(c *gin.Context) {
	images, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}

func (h *ImageHandler) GetImage(c *gin.Context) {
	id := c.Param("id")
	image, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}
	c.JSON(http.StatusOK, image)
}

func (h *ImageHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()
	
	// CRITICAL: Validate file before processing
	if err := utils.ValidateImageUpload(file, header); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate unique filename
	id := uuid.New().String()
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%s%s", id, ext)
	filePath := filepath.Join(h.uploadDir, filename)

	// Ensure upload directory exists
	if err := utils.EnsureDir(h.uploadDir); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload directory"})
		return
	}

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer dst.Close()

	size, err := io.Copy(dst, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Get image dimensions
	width, height, err := utils.GetImageDimensions(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read image"})
		return
	}

	// Create image record
	image := &models.Image{
		ID:           id,
		Filename:     filename,
		OriginalName: header.Filename,
		Path:         filePath,
		Size:         size,
		Width:        width,
		Height:       height,
		Format:       utils.GetImageFormat(header.Filename),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.repo.Create(image); err != nil {
		os.Remove(filePath) // Clean up file on database error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image record"})
		return
	}

	c.JSON(http.StatusCreated, image)
}

func (h *ImageHandler) ConvertImage(c *gin.Context) {
	id := c.Param("id")
	
	var req models.ImageConvertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	image, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Generate new filename with new format
	newFilename := fmt.Sprintf("%s.%s", id, req.Format)
	newPath := filepath.Join(h.uploadDir, newFilename)

	// Convert image
	quality := req.Quality
	if quality == 0 {
		quality = 85 // Default quality
	}

	if err := utils.ProcessImage(image.Path, newPath, req.Format, quality); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to convert image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Image converted successfully",
		"new_path": newPath,
		"format": req.Format,
	})
}

func (h *ImageHandler) DeleteImage(c *gin.Context) {
	id := c.Param("id")
	
	image, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Delete file
	if err := os.Remove(image.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		return
	}

	// Delete from database
	if err := h.repo.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}

func (h *ImageHandler) ServeImage(c *gin.Context) {
	id := c.Param("id")
	
	image, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Add cache-busting headers
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
	
	// Get thumbnail size if requested
	if thumb := c.Query("thumb"); thumb != "" {
		if size, err := strconv.Atoi(thumb); err == nil && size > 0 && size <= 500 {
			h.serveThumbnail(c, image, size)
			return
		}
	}

	c.File(image.Path)
}

func (h *ImageHandler) serveThumbnail(c *gin.Context, image *models.Image, size int) {
	thumbDir := filepath.Join(h.uploadDir, "thumbs")
	utils.EnsureDir(thumbDir)
	
	thumbPath := filepath.Join(thumbDir, fmt.Sprintf("%s_%d.jpg", image.ID, size))
	
	// Check if thumbnail exists
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		// Generate thumbnail
		if err := utils.ProcessImage(image.Path, thumbPath, "jpeg", 80); err != nil {
			c.File(image.Path) // Fallback to original
			return
		}
	}
	
	c.File(thumbPath)
}