package home

import (
	"net/http"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	seoService *service.SEOService
}

func NewHandler(seoService *service.SEOService) *Handler {
	return &Handler{seoService: seoService}
}

func (h *Handler) Get(c *gin.Context) {
	settings, err := h.seoService.GetHome(c.DefaultQuery("locale", "en"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
