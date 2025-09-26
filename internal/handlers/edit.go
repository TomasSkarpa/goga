package handlers

import (
	"goga/internal/repository"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
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
	// Basic adjustments
	Brightness float64 `json:"brightness"`
	Contrast   float64 `json:"contrast"`
	Saturation float64 `json:"saturation"`
	Hue        float64 `json:"hue"`
	Gamma      float64 `json:"gamma"`
	Blur       float64 `json:"blur"`
	Sharpen    float64 `json:"sharpen"`
	
	// Advanced features
	Shadows     float64 `json:"shadows"`
	Highlights  float64 `json:"highlights"`
	Temperature float64 `json:"temperature"`
	Tint        float64 `json:"tint"`
	Vibrance    float64 `json:"vibrance"`
	Clarity     float64 `json:"clarity"`
	Vignette    float64 `json:"vignette"`
	Noise       float64 `json:"noise"`
	
	// Transform
	Rotate float64 `json:"rotate"`
	CropX  float64 `json:"cropX"`
	CropY  float64 `json:"cropY"`
	CropW  float64 `json:"cropW"`
	CropH  float64 `json:"cropH"`
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

	finalImg := h.processImage(src, req)

	// Set response headers
	c.Header("Content-Type", "image/jpeg")
	c.Header("Cache-Control", "no-cache")

	// Encode and send
	jpeg.Encode(c.Writer, finalImg, &jpeg.Options{Quality: 85})
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

	finalImg := h.processImage(src, req)

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
		png.Encode(file, finalImg)
	default:
		jpeg.Encode(file, finalImg, &jpeg.Options{Quality: 95})
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

func (h *EditHandler) processImage(src image.Image, req EditRequest) image.Image {
	// Apply filters using gift library
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
	
	// Advanced filters
	if req.Shadows != 0 {
		value := float32(1.0 - req.Shadows*0.01)
		if value < 0.1 { value = 0.1 }
		if value > 3.0 { value = 3.0 }
		g.Add(gift.Gamma(value))
	}
	if req.Highlights != 0 {
		g.Add(gift.Contrast(float32(-req.Highlights * 0.3)))
	}
	if req.Temperature != 0 {
		g.Add(gift.Hue(float32(req.Temperature * 0.5)))
	}
	if req.Tint != 0 {
		g.Add(gift.Saturation(float32(100 + req.Tint)))
	}
	if req.Vibrance != 0 {
		value := float32(100 + req.Vibrance*1.5)
		if value < 0 { value = 0 }
		g.Add(gift.Saturation(value))
	}
	if req.Clarity != 0 {
		g.Add(gift.UnsharpMask(float32(req.Clarity*0.1), 2.0, 0.1))
	}
	if req.Noise > 0 {
		g.Add(gift.GaussianBlur(float32(req.Noise * 0.1)))
	}

	// Apply gift filters
	dst := image.NewRGBA(g.Bounds(src.Bounds()))
	g.Draw(dst, src)
	
	// Apply rotation if needed
	var finalImg image.Image = dst
	if req.Rotate != 0 {
		finalImg = imaging.Rotate(finalImg, req.Rotate, color.Transparent)
	}
	
	// Apply crop if specified
	if req.CropW > 0 && req.CropH > 0 {
		bounds := finalImg.Bounds()
		x := int(req.CropX * float64(bounds.Dx()))
		y := int(req.CropY * float64(bounds.Dy()))
		w := int(req.CropW * float64(bounds.Dx()))
		h := int(req.CropH * float64(bounds.Dy()))
		finalImg = imaging.Crop(finalImg, image.Rect(x, y, x+w, y+h))
	}
	
	// Apply vignette effect
	if req.Vignette != 0 {
		bounds := finalImg.Bounds()
		vignetteImg := image.NewRGBA(bounds)
		centerX := bounds.Dx() / 2
		centerY := bounds.Dy() / 2
		maxDist := float64(centerX)
		if centerY > centerX {
			maxDist = float64(centerY)
		}
		
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				dx := float64(x - centerX)
				dy := float64(y - centerY)
				dist := math.Sqrt(dx*dx + dy*dy)
				factor := 1.0 - (dist/maxDist)*(req.Vignette*0.01)
				if factor < 0 { factor = 0 }
				
				c := color.RGBAModel.Convert(finalImg.At(x, y)).(color.RGBA)
				c.R = uint8(float64(c.R) * factor)
				c.G = uint8(float64(c.G) * factor)
				c.B = uint8(float64(c.B) * factor)
				vignetteImg.Set(x, y, c)
			}
		}
		finalImg = vignetteImg
	}

	return finalImg
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