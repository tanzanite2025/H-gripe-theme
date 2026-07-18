package setting

import (
	"time"

	"gorm.io/gorm"
)

type Setting struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Key         string         `gorm:"uniqueIndex:idx_setting_key_locale;not null" json:"key"`
	Value       string         `gorm:"type:text" json:"value"`
	Type        string         `gorm:"default:'string'" json:"type"` // string, json, boolean, number
	Locale      string         `gorm:"uniqueIndex:idx_setting_key_locale;default:'en'" json:"locale"`
	Group       string         `gorm:"index" json:"group"`            // site, email, seo, social, quick-buy, etc.
	IsPublic    bool           `gorm:"default:true" json:"is_public"` // 是否公开给前端（非敏感信息）
	Description string         `gorm:"type:text" json:"description"`  // 设置说明
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Setting) TableName() string {
	return "settings"
}

// SiteSettings 站点设置响应
type SiteSettings struct {
	SiteName        string `json:"site_name"`
	SiteDescription string `json:"site_description"`
	SiteLogo        string `json:"site_logo"`
	ContactEmail    string `json:"contact_email"`
	ContactPhone    string `json:"contact_phone"`
	SocialLinks     string `json:"social_links"` // JSON格式
}

// QuickBuySettings 快速购买设置
type QuickBuySettings struct {
	Enabled        bool   `json:"enabled"`
	ButtonText     string `json:"button_text"`
	SuccessMessage string `json:"success_message"`
	RequireLogin   bool   `json:"require_login"`
}

// RedeemSettings 积分兑换配置
type RedeemSettings struct {
	Enabled        bool      `json:"enabled"`
	ExchangeRate   int       `json:"exchange_rate"`
	MinPoints      int       `json:"min_points"`
	MaxValuePerDay float64   `json:"max_value_per_day"`
	CardExpiryDays int       `json:"card_expiry_days"`
	PresetValues   []float64 `json:"preset_values"`
}

// EmailSettings 邮件设置
type EmailSettings struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUsername string `json:"smtp_username"`
	SMTPPassword string `json:"smtp_password"` // 敏感信息，不应公开
	FromEmail    string `json:"from_email"`
	FromName     string `json:"from_name"`
}

// SEOSettings SEO 设置
type SEOSettings struct {
	MetaTitle        string `json:"meta_title"`
	MetaDescription  string `json:"meta_description"`
	MetaKeywords     string `json:"meta_keywords"`
	GoogleAnalytics  string `json:"google_analytics"`
	GoogleTagManager string `json:"google_tag_manager"`
}

// SocialSettings 社交媒体设置
type SocialSettings struct {
	Facebook  string `json:"facebook"`
	Twitter   string `json:"twitter"`
	Instagram string `json:"instagram"`
	LinkedIn  string `json:"linkedin"`
	YouTube   string `json:"youtube"`
	WeChat    string `json:"wechat"`
}

// UpdateSettingRequest 更新设置请求
type UpdateSettingRequest struct {
	Key         string `json:"key" binding:"required"`
	Value       string `json:"value"`
	Type        string `json:"type"`
	Group       string `json:"group"`
	Locale      string `json:"locale"`
	IsPublic    bool   `json:"is_public"`
	Description string `json:"description"`
}

// BatchUpdateSettingsRequest 批量更新设置请求
type BatchUpdateSettingsRequest struct {
	Settings []UpdateSettingRequest `json:"settings" binding:"required"`
}

// SettingResponse 设置响应（用于管理后台）
type SettingResponse struct {
	ID          uint      `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Type        string    `json:"type"`
	Group       string    `json:"group"`
	Locale      string    `json:"locale"`
	IsPublic    bool      `json:"is_public"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
