package admin

import (
	"net/http"
	"strconv"
	"strings"
	"tanzanite/internal/domain/auth"
	"tanzanite/internal/domain/setting"
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

// GetPublicChatAgentCompatibility 获取 public chat 客服映射兼容检查
func (h *SettingsHandler) GetPublicChatAgentCompatibility(c *gin.Context) {
	agents, err := h.userRepo.FindCustomerServiceAgents(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取客服映射失败"})
		return
	}

	stats, err := h.userRepo.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户统计失败"})
		return
	}

	items := make([]gin.H, 0, len(agents))
	for _, agent := range agents {
		normalizedRole := auth.NormalizeRole(agent.Role)
		items = append(items, gin.H{
			"id":              agent.ID,
			"username":        agent.Username,
			"email":           agent.Email,
			"display_name":    displayName(agent.FirstName, agent.LastName, agent.Username, agent.Email),
			"raw_role":        agent.Role,
			"normalized_role": normalizedRole,
			"status":          agent.Status,
			"public_agent_id": strconv.FormatUint(uint64(agent.ID), 10),
			"wp_user_id":      agent.ID,
			"exposed":         agent.Status == "active" && auth.IsCustomerServiceAgentRole(agent.Role),
		})
	}

	warnings := []string{
		"Go users 表仍不保存 PHP wp_tz_cs_agents.whatsapp/avatar/online_status；精确线上 parity 仍需执行 PHP 源表 preflight。",
	}
	if len(agents) == 0 {
		warnings = append(warnings, "当前 Go users 表未找到 active customer-service agent 候选，public chat 可能无法自动分配客服。")
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": gin.H{
			"compatible_agents":       len(items),
			"go_user_role_buckets":    stats["by_role"],
			"php_preflight_required":  true,
			"missing_profile_columns": []string{"whatsapp", "avatar", "online_status"},
		},
		"agents": items,
		"role_mappings": []gin.H{
			{"source": "administrator", "normalized": "admin", "agent_visible": true},
			{"source": "shop_manager", "normalized": "manager", "agent_visible": true},
			{"source": "agent", "normalized": "support", "agent_visible": true},
			{"source": "customer_service", "normalized": "support", "agent_visible": true},
			{"source": "customer_support", "normalized": "support", "agent_visible": true},
			{"source": "editor/viewer/user/subscriber", "normalized": "user/editor/viewer", "agent_visible": false},
		},
		"preflight_sql": []gin.H{
			{
				"title": "PHP active agents source of truth",
				"sql":   "SELECT agent_id, wp_user_id, name, email, status, online_status FROM wp_tz_cs_agents WHERE status = 'active' ORDER BY created_at ASC;",
			},
			{
				"title": "Go users projection candidates",
				"sql":   "SELECT id, username, email, role, status FROM users WHERE status = 'active' AND (role IN ('admin', 'manager', 'support', 'agent', 'administrator', 'shop_manager', 'customer_service', 'customer_support') OR LOWER(role) LIKE '%administrator%' OR LOWER(role) LIKE '%shop_manager%' OR LOWER(role) LIKE '%customer_service%' OR LOWER(role) LIKE '%customer_support%' OR LOWER(role) LIKE '%support%' OR LOWER(role) LIKE '%agent%') ORDER BY role ASC, created_at ASC;",
			},
			{
				"title": "Linked PHP agents missing Go user rows",
				"sql":   "SELECT a.agent_id, a.wp_user_id, a.name, a.email, a.status, a.online_status FROM wp_tz_cs_agents a LEFT JOIN users u ON u.id = a.wp_user_id WHERE a.status = 'active' AND (a.wp_user_id IS NULL OR u.id IS NULL);",
			},
		},
		"warnings": warnings,
	})
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

func displayName(firstName, lastName, username, email string) string {
	fullName := strings.TrimSpace(strings.TrimSpace(firstName) + " " + strings.TrimSpace(lastName))
	if fullName != "" {
		return fullName
	}
	if strings.TrimSpace(username) != "" {
		return strings.TrimSpace(username)
	}
	return strings.TrimSpace(email)
}
