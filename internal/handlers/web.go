package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebHandler struct{}

func NewWebHandler() *WebHandler {
	return &WebHandler{}
}

func (h *WebHandler) Dashboard(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Goga - Photo Gallery",
	})
}

func (h *WebHandler) ImageDetail(c *gin.Context) {
	id := c.Param("id")
	c.HTML(http.StatusOK, "image-detail.html", gin.H{
		"title":   "Image Details",
		"imageID": id,
	})
}