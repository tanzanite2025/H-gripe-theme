package setting

// SiteSettings 站点设置响应
type SiteSettings struct {
	SiteName          string `json:"site_name"`
	BrandTitle        string `json:"brand_title"`
	SiteDescription   string `json:"site_description"`
	SiteLogo          string `json:"site_logo"`
	SiteFavicon       string `json:"site_favicon"`
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

// RedeemSettings 积分兑换配置
type RedeemSettings struct {
	Enabled        bool      `json:"enabled"`
	ExchangeRate   int       `json:"exchange_rate"`
	MinPoints      int       `json:"min_points"`
	MaxValuePerDay float64   `json:"max_value_per_day"`
	CardExpiryDays int       `json:"card_expiry_days"`
	PresetValues   []float64 `json:"preset_values"`
}

// LoyaltySettings 积分获取规则配置。
type LoyaltySettings struct {
	ReferralReferrerPoints    int `json:"referral_referrer_points"`
	ReferralRefereePoints     int `json:"referral_referee_points"`
	CheckInBasePoints         int `json:"checkin_base_points"`
	CheckInStreakIntervalDays int `json:"checkin_streak_interval_days"`
	CheckInStreakBonusPoints  int `json:"checkin_streak_bonus_points"`
	CheckInMaxPoints          int `json:"checkin_max_points"`
}

// DefaultLoyaltySettings 返回与默认种子数据一致的积分规则。
func DefaultLoyaltySettings() LoyaltySettings {
	return LoyaltySettings{
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
	}
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

// SocialSettings 社交媒体设置
type SocialSettings struct {
	Facebook  string `json:"facebook"`
	Twitter   string `json:"twitter"`
	Instagram string `json:"instagram"`
	LinkedIn  string `json:"linkedin"`
	YouTube   string `json:"youtube"`
	WeChat    string `json:"wechat"`
}
