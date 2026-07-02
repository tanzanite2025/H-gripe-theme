package admin

import (
	"net/http"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	settingsService *service.AdminSettingsService
}

func NewSettingsHandler(settingsService *service.AdminSettingsService) *SettingsHandler {
	return &SettingsHandler{settingsService: settingsService}
}

func (h *SettingsHandler) ListPublicChatAgents(c *gin.Context) {
	overview, err := h.settingsService.ListPublicChatAgents(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch public chat agents"})
		return
	}

	c.JSON(http.StatusOK, overview)
}

func (h *SettingsHandler) GetAllSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	group := c.Query("group")

	settings, err := h.settingsService.ListSettings(locale, group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

func (h *SettingsHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	locale := c.DefaultQuery("locale", "en")

	s, err := h.settingsService.GetSetting(key, locale)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "setting not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"setting": s})
}

func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	var req setting.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s, err := h.settingsService.UpdateSetting(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"setting": s})
}

func (h *SettingsHandler) BatchUpdateSettings(c *gin.Context) {
	var req setting.BatchUpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	count, err := h.settingsService.BatchUpdateSettings(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "batch update completed", "count": count})
}

func (h *SettingsHandler) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	locale := c.DefaultQuery("locale", "en")

	if err := h.settingsService.DeleteSetting(key, locale); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "setting deleted successfully"})
}

func (h *SettingsHandler) GetGroups(c *gin.Context) {
	groups, err := h.settingsService.GetGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch setting groups"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *SettingsHandler) GetSiteSettings(c *gin.Context) {
	h.writeSettingsGroup(c, "site", "failed to fetch site settings")
}

func (h *SettingsHandler) GetEmailSettings(c *gin.Context) {
	h.writeSettingsGroup(c, "email", "failed to fetch email settings")
}

func (h *SettingsHandler) GetSEOSettings(c *gin.Context) {
	h.writeSettingsGroup(c, "seo", "failed to fetch SEO settings")
}

func (h *SettingsHandler) GetSocialSettings(c *gin.Context) {
	h.writeSettingsGroup(c, "social", "failed to fetch social settings")
}

func (h *SettingsHandler) GetPaymentSettings(c *gin.Context) {
	h.writeSettingsGroup(c, "payment", "failed to fetch payment settings")
}

func (h *SettingsHandler) writeSettingsGroup(c *gin.Context, group, errorMessage string) {
	locale := c.DefaultQuery("locale", "en")
	settings, err := h.settingsService.GetByGroup(group, locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": errorMessage})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}
