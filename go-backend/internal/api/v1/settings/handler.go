package settings

import (
	"commerce-platform/internal/api/middleware"
	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	settingService            *service.SettingService
	websiteProfileService     *service.WebsiteProfileService
	mediaService              *service.MediaService
	siteLogoService           *service.SiteLogoService
	refundReturnPolicyService *service.RefundReturnPolicyService
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

func (h *Handler) ConfigureMediaService(mediaService *service.MediaService) {
	if h == nil {
		return
	}
	h.mediaService = mediaService
}

func (h *Handler) ConfigureSiteLogoService(siteLogoService *service.SiteLogoService) {
	if h == nil {
		return
	}
	h.siteLogoService = siteLogoService
}

func (h *Handler) ConfigureRefundReturnPolicyService(policyService *service.RefundReturnPolicyService) {
	if h == nil {
		return
	}
	h.refundReturnPolicyService = policyService
}

func (h *Handler) GetSiteSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", middleware.GetLocale(c))

	settings, err := h.settingService.GetSiteSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, h.publicSiteSettings(settings))
}

func (h *Handler) GetAllPublicSettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetAllPublic(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	settings = service.FilterDomainManagedSettings(settings)
	settings = h.publicSettings(settings)

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
	settings = h.publicSettings(settings)

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

	c.JSON(http.StatusOK, h.publicSetting(setting))
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

func (h *Handler) publicSiteSettings(settings *settingdomain.SiteSettings) *settingdomain.SiteSettings {
	if settings == nil {
		return nil
	}
	publicSettings := *settings
	publicSettings.SiteLogoWidth = 0
	publicSettings.SiteLogoHeight = 0
	if logoURL := h.currentSiteLogoURL(); logoURL != "" {
		publicSettings.SiteLogo = logoURL
	}
	if width, height, ok := h.publicSiteLogoDimensions(publicSettings.SiteLogo); ok {
		publicSettings.SiteLogoWidth = width
		publicSettings.SiteLogoHeight = height
	} else if width, height, ok := h.publicMediaDimensions(publicSettings.SiteLogo); ok {
		publicSettings.SiteLogoWidth = width
		publicSettings.SiteLogoHeight = height
	}
	publicSettings.SiteLogo = h.publicSiteLogoURL(publicSettings.SiteLogo)
	publicSettings.SiteFavicon = h.publicMediaURL(publicSettings.SiteFavicon)
	return &publicSettings
}

func (h *Handler) publicWebsiteProfile(settings *settingdomain.WebsiteProfileSettings) *settingdomain.WebsiteProfileSettings {
	if settings == nil {
		return nil
	}
	publicSettings := *settings
	publicSettings.AvatarURL = h.publicMediaURL(publicSettings.AvatarURL)
	publicSettings.FactoryImageURL = h.publicMediaURL(publicSettings.FactoryImageURL)
	return &publicSettings
}

func (h *Handler) publicSettings(settings []settingdomain.Setting) []settingdomain.Setting {
	if len(settings) == 0 {
		return settings
	}
	publicSettings := make([]settingdomain.Setting, 0, len(settings))
	for _, item := range settings {
		publicSettings = append(publicSettings, *h.publicSetting(&item))
	}
	return publicSettings
}

func (h *Handler) publicSetting(item *settingdomain.Setting) *settingdomain.Setting {
	if item == nil {
		return nil
	}
	publicItem := *item
	switch publicItem.Key {
	case "site_logo":
		publicItem.Value = h.publicSiteLogoURL(publicItem.Value)
	case "site_favicon",
		settingdomain.WebsiteProfileKeyAvatarURL,
		settingdomain.WebsiteProfileKeyFactoryImageURL:
		publicItem.Value = h.publicMediaURL(publicItem.Value)
	}
	return &publicItem
}

func (h *Handler) currentSiteLogoURL() string {
	if h == nil || h.siteLogoService == nil {
		return ""
	}
	return h.siteLogoService.CurrentPublicURL()
}

func (h *Handler) publicSiteLogoURL(value string) string {
	if h == nil || h.siteLogoService == nil {
		return h.publicMediaURL(value)
	}
	canonical := h.siteLogoService.CanonicalPublicURL(value)
	if canonical != value {
		return canonical
	}
	return h.publicMediaURL(value)
}

func (h *Handler) publicSiteLogoDimensions(value string) (int, int, bool) {
	if h == nil || h.siteLogoService == nil {
		return 0, 0, false
	}
	return h.siteLogoService.PublicDimensions(value)
}

func (h *Handler) publicMediaURL(value string) string {
	if h == nil || h.mediaService == nil {
		return value
	}
	return h.mediaService.CanonicalPublicMediaURL(value)
}

func (h *Handler) publicMediaDimensions(value string) (int, int, bool) {
	if h == nil || h.mediaService == nil {
		return 0, 0, false
	}
	return h.mediaService.PublicMediaDimensions(value)
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

	c.JSON(http.StatusOK, h.publicWebsiteProfile(settings))
}
