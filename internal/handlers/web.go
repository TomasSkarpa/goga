package handlers

import (
	"goga/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type WebHandler struct {
	repo *repository.ImageRepository
}

func NewWebHandler(repo *repository.ImageRepository) *WebHandler {
	return &WebHandler{
		repo: repo,
	}
}

func (h *WebHandler) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Goga - Photo Gallery",
	})
}

func (h *WebHandler) ImageDetail(c *gin.Context) {
	id := c.Param("id")
	
	image, err := h.repo.GetByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Image not found",
		})
		return
	}
	
	c.HTML(http.StatusOK, "image-detail.html", gin.H{
		"title": image.OriginalName + " - Goga",
		"image": image,
		"timestamp": time.Now().Unix(),
	})
}