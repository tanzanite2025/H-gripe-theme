package settings

import (
	"net/http"
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	settingService        *service.SettingService
	websiteProfileService *service.WebsiteProfileService
}

func NewHandler(settingService *service.SettingService, websiteProfileServices ...*service.WebsiteProfileService) *Handler {
	var websiteProfileService *service.WebsiteProfileService
	if len(websiteProfileServices) > 0 {
		websiteProfileService = websiteProfileServices[0]
	}

	return &Handler{
		settingService:        settingService,
		websiteProfileService: websiteProfileService,
	}
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

func (h *Handler) GetAllPublicSettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetAllPublic(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	settings = service.FilterDomainManagedSettings(settings)

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
		"total":    len(settings),
	})
}

func (h *Handler) GetSettingsByGroup(c *gin.Context) {
	group := c.Param("group")
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))
	if service.IsDomainManagedSettingGroup(group) {
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
	if service.IsDomainManagedSettingKey(key) {
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
	groups = service.FilterDomainManagedSettingGroups(groups)

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  len(groups),
	})
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

func (h *Handler) GetWebsiteProfile(c *gin.Context) {
	if h.websiteProfileService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "website profile service unavailable"})
		return
	}

	locale := c.DefaultQuery("locale", middleware.GetLocale(c))
	settings, err := h.websiteProfileService.Get(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
