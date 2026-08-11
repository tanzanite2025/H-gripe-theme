package admin

import (
	"errors"
	"net/http"

	analyticsdomain "commerce-platform/internal/domain/analytics"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type AnalyticsHandler struct {
	analyticsService *service.AnalyticsService
}

func NewAnalyticsHandler(analyticsService *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsService: analyticsService}
}

func (h *AnalyticsHandler) Get(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	settings, err := h.analyticsService.Get(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (h *AnalyticsHandler) Update(c *gin.Context) {
	var request analyticsdomain.UpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if request.Locale == "" {
		request.Locale = c.DefaultQuery("locale", "en")
	}

	settings, err := h.analyticsService.Update(request)
	if err != nil {
		if errors.Is(err, service.ErrInvalidAnalyticsSettings) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}
