package analytics

import (
	"net/http"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	analyticsService *service.AnalyticsService
}

func NewHandler(analyticsService *service.AnalyticsService) *Handler {
	return &Handler{analyticsService: analyticsService}
}

func (h *Handler) Get(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	settings, err := h.analyticsService.Get(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
