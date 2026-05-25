package settings

import (
	"net/http"
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

// GetSiteSettings 获取站点设置
func (h *Handler) GetSiteSettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetSiteSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetQuickBuySettings 获取快速购买设置
func (h *Handler) GetQuickBuySettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetQuickBuySettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetAllSettings 获取所有设置（管理员）
func (h *Handler) GetAllSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")

	settings, err := h.settingService.GetAll(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": settings,
		"total":    len(settings),
	})
}

// GetAllPublicSettings 获取所有公开设置
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

// GetSettingsByGroup 获取分组设置
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

// GetSetting 获取单个设置
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

// UpdateSetting 更新设置（管理员）
func (h *Handler) UpdateSetting(c *gin.Context) {
	var req setting.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Type == "" {
		req.Type = "string"
	}
	if req.Locale == "" {
		req.Locale = "en"
	}

	err := h.settingService.Set(req.Key, req.Value, req.Type, req.Group, req.Locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting updated successfully",
		"key":     req.Key,
	})
}

// BatchUpdateSettings 批量更新设置（管理员）
func (h *Handler) BatchUpdateSettings(c *gin.Context) {
	var req setting.BatchUpdateSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 转换为 Setting 对象
	settings := make([]setting.Setting, len(req.Settings))
	for i, s := range req.Settings {
		if s.Type == "" {
			s.Type = "string"
		}
		if s.Locale == "" {
			s.Locale = "en"
		}

		settings[i] = setting.Setting{
			Key:         s.Key,
			Value:       s.Value,
			Type:        s.Type,
			Group:       s.Group,
			Locale:      s.Locale,
			IsPublic:    s.IsPublic,
			Description: s.Description,
		}
	}

	err := h.settingService.BatchSet(settings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Settings updated successfully",
		"count":   len(settings),
	})
}

// DeleteSetting 删除设置（管理员）
func (h *Handler) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	locale := c.DefaultQuery("locale", "en")

	err := h.settingService.Delete(key, locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting deleted successfully",
		"key":     key,
	})
}

// GetGroups 获取所有分组
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

// GetEmailSettings 获取邮件设置（管理员）
func (h *Handler) GetEmailSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")

	settings, err := h.settingService.GetEmailSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetSEOSettings 获取SEO设置
func (h *Handler) GetSEOSettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetSEOSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// GetSocialSettings 获取社交媒体设置
func (h *Handler) GetSocialSettings(c *gin.Context) {
	locale := middleware.GetLocale(c)

	settings, err := h.settingService.GetSocialSettings(locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}
