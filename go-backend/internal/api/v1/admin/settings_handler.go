package admin

import (
	"net/http"
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/domain/user"
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

// SettingsHandler 系统设置处理器
type SettingsHandler struct {
	settingRepo *repository.SettingRepository
	userRepo    *repository.UserRepository
}

// NewSettingsHandler 创建系统设置处理器
func NewSettingsHandler(settingRepo *repository.SettingRepository, userRepo *repository.UserRepository) *SettingsHandler {
	return &SettingsHandler{
		settingRepo: settingRepo,
		userRepo:    userRepo,
	}
}

// ListPublicChatAgents 获取 public chat 客服 profile 概览
func (h *SettingsHandler) ListPublicChatAgents(c *gin.Context) {
	agents, err := h.userRepo.FindAllCustomerServiceAgentProfiles(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取客服 profile 失败"})
		return
	}

	items := make([]gin.H, 0, len(agents))
	exposedAgents := 0
	for _, agent := range agents {
		exposed := agent.UserID != nil && agent.Status == "active" && agent.User != nil && agent.User.Status == "active" && auth.IsCustomerServiceAgentRole(agent.User.Role)
		if exposed {
			exposedAgents++
		}
		var userID *uint
		rawRole := ""
		normalizedRole := auth.RoleUser
		userStatus := ""
		if agent.UserID != nil {
			value := *agent.UserID
			userID = &value
		}
		if agent.User != nil {
			rawRole = agent.User.Role
			normalizedRole = auth.NormalizeRole(agent.User.Role)
			userStatus = agent.User.Status
		}
		items = append(items, gin.H{
			"id":              agent.ID,
			"agent_id":        agent.AgentID,
			"user_id":         userID,
			"username":        usernameFromProfile(agent),
			"email":           agent.PublicEmail(),
			"display_name":    agent.DisplayName(),
			"raw_role":        rawRole,
			"normalized_role": normalizedRole,
			"user_status":     userStatus,
			"profile_status":  agent.Status,
			"online_status":   agent.OnlineStatus,
			"avatar":          agent.Avatar,
			"whatsapp":        agent.WhatsApp,
			"exposed":         exposed,
		})
	}

	warnings := []string{}
	if len(agents) == 0 {
		warnings = append(warnings, "当前未配置 public chat 客服 profile")
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"profile_count":  len(agents),
			"exposed_agents": exposedAgents,
		},
		"agents":   items,
		"warnings": warnings,
	})
}

func usernameFromProfile(agent user.AgentProfile) string {
	if agent.User == nil {
		return ""
	}
	return agent.User.Username
}

// GetAllSettings 获取所有设置
func (h *SettingsHandler) GetAllSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	group := c.Query("group")

	var settings []setting.Setting
	var err error

	if group != "" {
		settings, err = h.settingRepo.GetByGroup(group, locale)
	} else {
		settings, err = h.settingRepo.GetAll(locale)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// GetSetting 获取单个设置
func (h *SettingsHandler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	locale := c.DefaultQuery("locale", "en")

	s, err := h.settingRepo.Get(key, locale)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "设置不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"setting": s})
}

// UpdateSetting 更新设置
func (h *SettingsHandler) UpdateSetting(c *gin.Context) {
	var req setting.UpdateSettingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Locale == "" {
		req.Locale = "en"
	}
	if req.Type == "" {
		req.Type = "string"
	}

	s := &setting.Setting{
		Key:         req.Key,
		Value:       req.Value,
		Type:        req.Type,
		Group:       req.Group,
		Locale:      req.Locale,
		IsPublic:    req.IsPublic,
		Description: req.Description,
	}

	if err := h.settingRepo.Set(s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"setting": s})
}

// BatchUpdateSettings 批量更新设置
func (h *SettingsHandler) BatchUpdateSettings(c *gin.Context) {
	var req setting.BatchUpdateSettingsRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings := make([]setting.Setting, 0, len(req.Settings))
	for _, s := range req.Settings {
		if s.Locale == "" {
			s.Locale = "en"
		}
		if s.Type == "" {
			s.Type = "string"
		}

		settings = append(settings, setting.Setting{
			Key:         s.Key,
			Value:       s.Value,
			Type:        s.Type,
			Group:       s.Group,
			Locale:      s.Locale,
			IsPublic:    s.IsPublic,
			Description: s.Description,
		})
	}

	if err := h.settingRepo.BatchSet(settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "批量更新设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "批量更新成功", "count": len(settings)})
}

// DeleteSetting 删除设置
func (h *SettingsHandler) DeleteSetting(c *gin.Context) {
	key := c.Param("key")
	locale := c.DefaultQuery("locale", "en")

	if err := h.settingRepo.Delete(key, locale); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// GetGroups 获取所有设置分组
func (h *SettingsHandler) GetGroups(c *gin.Context) {
	groups, err := h.settingRepo.GetGroups()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取分组失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

// GetSiteSettings 获取站点设置
func (h *SettingsHandler) GetSiteSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	settings, err := h.settingRepo.GetByGroup("site", locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取站点设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// GetEmailSettings 获取邮件设置
func (h *SettingsHandler) GetEmailSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	settings, err := h.settingRepo.GetByGroup("email", locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取邮件设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// GetSEOSettings 获取 SEO 设置
func (h *SettingsHandler) GetSEOSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	settings, err := h.settingRepo.GetByGroup("seo", locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取 SEO 设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// GetSocialSettings 获取社交媒体设置
func (h *SettingsHandler) GetSocialSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	settings, err := h.settingRepo.GetByGroup("social", locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取社交媒体设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}

// GetPaymentSettings 获取支付设置
func (h *SettingsHandler) GetPaymentSettings(c *gin.Context) {
	locale := c.DefaultQuery("locale", "en")
	settings, err := h.settingRepo.GetByGroup("payment", locale)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取支付设置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"settings": settings})
}
