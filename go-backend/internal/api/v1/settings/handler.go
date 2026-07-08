package settings

import (
	"net/http"
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	settingService *service.SettingService
}

func NewHandler(settingService *service.SettingService) *Handler {
	return &Handler{
		settingService: settingService,
	}
}

func (h *Handler) GetSiteSettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetSiteSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetQuickBuySettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetQuickBuySettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetAllPublicSettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetAllPublic(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
		"total":    len(settings),
	})
}

func (h *Handler) GetSettingsByGroup(c *gin.Context) {
	group := c.Param("group")
	locale := c.DefaultQuery("locale", "en")

	settings, err := h.settingService.GetByGroup(group, locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"group":    group,
		"settings": settings,
		"total":    len(settings),
	})
}

func (h *Handler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	locale := c.DefaultQuery("locale", "en")

	setting, err := h.settingService.Get(key, locale)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func (h *Handler) GetGroups(c *gin.Context) {
	groups, err := h.settingService.GetGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  len(groups),
	})
}

func (h *Handler) GetSEOSettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetSEOSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetSocialSettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetSocialSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
