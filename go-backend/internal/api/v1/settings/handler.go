package settings

import (
	"net/http"
	"strings"
	"tanzanite/internal/api/middleware"
	"tanzanite/internal/domain/setting"
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

var publicDomainManagedSettingGroups = map[string]struct{}{
	"loyalty": {},
	"redeem":  {},
}

func (h *Handler) GetSiteSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))

	settings, err := h.settingService.GetSiteSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func (h *Handler) GetQuickBuySettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))

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
	settings = filterPublicSettingsManagedByDomain(settings)

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
		"total":    len(settings),
	})
}

func (h *Handler) GetSettingsByGroup(c *gin.Context) {
	group := c.Param("group")
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))
	if isPublicSettingGroupManagedByDomain(group) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting group is managed by its domain API"})
		return
	}

	settings, err := h.settingService.GetPublicByGroup(group, locale)
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
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))
	if isPublicSettingKeyManagedByDomain(key) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting is managed by its domain API"})
		return
	}

	setting, err := h.settingService.GetPublic(key, locale)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	c.JSON(http.StatusOK, setting)
}

func (h *Handler) GetGroups(c *gin.Context) {
	groups, err := h.settingService.GetPublicGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	groups = filterPublicSettingGroupsManagedByDomain(groups)

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  len(groups),
	})
}

func (h *Handler) GetSEOSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))

	settings, err := h.settingService.GetSEOSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

func isPublicSettingGroupManagedByDomain(group string) bool {
	_, managed := publicDomainManagedSettingGroups[strings.ToLower(strings.TrimSpace(group))]
	return managed
}

func isPublicSettingKeyManagedByDomain(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	return strings.HasPrefix(normalized, "tz_loyalty_") || strings.HasPrefix(normalized, "tz_redeem_")
}

func filterPublicSettingsManagedByDomain(settings []setting.Setting) []setting.Setting {
	filtered := make([]setting.Setting, 0, len(settings))
	for _, item := range settings {
		if isPublicSettingGroupManagedByDomain(item.Group) || isPublicSettingKeyManagedByDomain(item.Key) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func filterPublicSettingGroupsManagedByDomain(groups []string) []string {
	filtered := make([]string, 0, len(groups))
	for _, group := range groups {
		if isPublicSettingGroupManagedByDomain(group) {
			continue
		}
		filtered = append(filtered, group)
	}
	return filtered
}

func (h *Handler) GetSocialSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))

	settings, err := h.settingService.GetSocialSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
