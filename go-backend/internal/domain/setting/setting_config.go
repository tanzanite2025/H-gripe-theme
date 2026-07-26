package setting

// SiteSettings 站点设置响应
type SiteSettings struct {
	SiteName          string `json:"site_name"`
	BrandTitle        string `json:"brand_title"`
	SiteDescription   string `json:"site_description"`
	SiteURL           string `json:"site_url"`
	SiteLogo          string `json:"site_logo"`
	ContactEmail      string `json:"contact_email"`
	ContactPhone      string `json:"contact_phone"`
	SocialLinks       string `json:"social_links"` // JSON格式
	AdminBrandName    string `json:"admin_brand_name"`
	AdminBrandInitial string `json:"admin_brand_initial"`
	AdminPanelLabel   string `json:"admin_panel_label"`
	AdminLoginTitle   string `json:"admin_login_title"`
	AdminFooterText   string `json:"admin_footer_text"`
	AdminHTMLTitle    string `json:"admin_html_title"`
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
