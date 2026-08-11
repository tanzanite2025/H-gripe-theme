package admin

import (
	"net/http"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type WebsiteProfileHandler struct {
	websiteProfileService *service.WebsiteProfileService
}

func NewWebsiteProfileHandler(websiteProfileService *service.WebsiteProfileService) *WebsiteProfileHandler {
	return &WebsiteProfileHandler{websiteProfileService: websiteProfileService}
}

func (h *WebsiteProfileHandler) Get(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	settings, err := h.websiteProfileService.GetAdmin(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch website profile settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}

func (h *WebsiteProfileHandler) Update(c *gin.Context) {
	var request setting.WebsiteProfileUpdateRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if request.Locale == "" {
		request.Locale = c.DefaultQuery("locale", "en")
	}

	settings, err := h.websiteProfileService.Update(request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update website profile settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": settings})
}
