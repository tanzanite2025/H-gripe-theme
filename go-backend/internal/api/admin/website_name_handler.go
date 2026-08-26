package admin

import (
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WebsiteNameHandler struct {
	websiteNameService *service.WebsiteNameService
}

func NewWebsiteNameHandler(websiteNameService *service.WebsiteNameService) *WebsiteNameHandler {
	return &WebsiteNameHandler{websiteNameService: websiteNameService}
}

func (h *WebsiteNameHandler) Get(c *gin.Context) {
	if h == nil || h.websiteNameService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "website name service unavailable"})
		return
	}

	locale := c.DefaultQuery("locale", "en")
	settings, err := h.websiteNameService.GetAdmin(locale)
	if err != nil {
		if errors.Is(err, service.ErrUnsupportedLocale) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch why this name settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (h *WebsiteNameHandler) Update(c *gin.Context) {
	if h == nil || h.websiteNameService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "website name service unavailable"})
		return
	}

	var request setting.WebsiteNameUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if request.Locale == "" {
		request.Locale = c.DefaultQuery("locale", "en")
	}

	settings, err := h.websiteNameService.Update(request)
	if err != nil {
		if errors.Is(err, service.ErrUnsupportedLocale) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update why this name settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}
