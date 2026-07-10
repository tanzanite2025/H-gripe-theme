package service

import (
	"fmt"
	"strconv"
	"tanzanite/internal/domain/setting"
)

// GetSiteSettings 获取站点设置
func (s *SettingService) GetSiteSettings(locale string) (*setting.SiteSettings, error) {
	cacheKey := settingsStructuredCacheKey("site", locale)

	// 尝试从缓存获取
	var siteSettings setting.SiteSettings
	if err := s.cache.Get(cacheKey, &siteSettings); err == nil {
		return &siteSettings, nil
	}

	// 从数据库获取
	settings, err := s.settingRepo.GetByGroup("site", locale)
	if err != nil {
		return nil, err
	}

	// 转换为结构体
	siteSettings = setting.SiteSettings{}
	for _, st := range settings {
		switch st.Key {
		case "site_name":
			siteSettings.SiteName = st.Value
		case "site_description":
			siteSettings.SiteDescription = st.Value
		case "site_logo":
			siteSettings.SiteLogo = st.Value
		case "contact_email":
			siteSettings.ContactEmail = st.Value
		case "contact_phone":
			siteSettings.ContactPhone = st.Value
		case "social_links":
			siteSettings.SocialLinks = st.Value
		}
	}

	// 写入缓存
	_ = s.cache.Set(cacheKey, &siteSettings, s.cacheTTL)

	return &siteSettings, nil
}

// GetQuickBuySettings 获取快速购买设置
func (s *SettingService) GetQuickBuySettings(locale string) (*setting.QuickBuySettings, error) {
	cacheKey := settingsStructuredCacheKey("quick-buy", locale)

	// 尝试从缓存获取
	var quickBuySettings setting.QuickBuySettings
	if err := s.cache.Get(cacheKey, &quickBuySettings); err == nil {
		return &quickBuySettings, nil
	}

	// 从数据库获取
	settings, err := s.settingRepo.GetByGroup("quick-buy", locale)
	if err != nil {
		return nil, err
	}

	// 转换为结构体
	quickBuySettings = setting.QuickBuySettings{}
	for _, st := range settings {
		switch st.Key {
		case "enabled":
			quickBuySettings.Enabled = st.Value == "true"
		case "button_text":
			quickBuySettings.ButtonText = st.Value
		case "success_message":
			quickBuySettings.SuccessMessage = st.Value
		case "require_login":
			quickBuySettings.RequireLogin = st.Value == "true"
		}
	}

	// 写入缓存
	_ = s.cache.Set(cacheKey, &quickBuySettings, s.cacheTTL)

	return &quickBuySettings, nil
}

// GetEmailSettings 获取邮件设置
func (s *SettingService) GetEmailSettings(locale string) (*setting.EmailSettings, error) {
	cacheKey := settingsStructuredCacheKey("email", locale)

	// 尝试从缓存获取
	var emailSettings setting.EmailSettings
	if err := s.cache.Get(cacheKey, &emailSettings); err == nil {
		return &emailSettings, nil
	}

	// 从数据库获取
	settings, err := s.settingRepo.GetByGroup("email", locale)
	if err != nil {
		return nil, err
	}

	// 转换为结构体
	emailSettings = setting.EmailSettings{}
	for _, st := range settings {
		switch st.Key {
		case "smtp_host":
			emailSettings.SMTPHost = st.Value
		case "smtp_port":
			port, err := strconv.Atoi(st.Value)
			if err != nil {
				return nil, fmt.Errorf("invalid smtp_port %q: %w", st.Value, err)
			}
			emailSettings.SMTPPort = port
		case "smtp_username":
			emailSettings.SMTPUsername = st.Value
		case "smtp_password":
			emailSettings.SMTPPassword = st.Value
		case "from_email":
			emailSettings.FromEmail = st.Value
		case "from_name":
			emailSettings.FromName = st.Value
		}
	}

	// 写入缓存
	_ = s.cache.Set(cacheKey, &emailSettings, s.cacheTTL)

	return &emailSettings, nil
}

// GetSEOSettings 获取SEO设置
func (s *SettingService) GetSEOSettings(locale string) (*setting.SEOSettings, error) {
	cacheKey := settingsStructuredCacheKey("seo", locale)

	// 尝试从缓存获取
	var seoSettings setting.SEOSettings
	if err := s.cache.Get(cacheKey, &seoSettings); err == nil {
		return &seoSettings, nil
	}

	// 从数据库获取
	settings, err := s.settingRepo.GetByGroup("seo", locale)
	if err != nil {
		return nil, err
	}

	// 转换为结构体
	seoSettings = setting.SEOSettings{}
	for _, st := range settings {
		switch st.Key {
		case "meta_title":
			seoSettings.MetaTitle = st.Value
		case "meta_description":
			seoSettings.MetaDescription = st.Value
		case "meta_keywords":
			seoSettings.MetaKeywords = st.Value
		case "google_analytics":
			seoSettings.GoogleAnalytics = st.Value
		case "google_tag_manager":
			seoSettings.GoogleTagManager = st.Value
		}
	}

	// 写入缓存
	_ = s.cache.Set(cacheKey, &seoSettings, s.cacheTTL)

	return &seoSettings, nil
}

// GetSocialSettings 获取社交媒体设置
func (s *SettingService) GetSocialSettings(locale string) (*setting.SocialSettings, error) {
	cacheKey := settingsStructuredCacheKey("social", locale)

	// 尝试从缓存获取
	var socialSettings setting.SocialSettings
	if err := s.cache.Get(cacheKey, &socialSettings); err == nil {
		return &socialSettings, nil
	}

	// 从数据库获取
	settings, err := s.settingRepo.GetByGroup("social", locale)
	if err != nil {
		return nil, err
	}

	// 转换为结构体
	socialSettings = setting.SocialSettings{}
	for _, st := range settings {
		switch st.Key {
		case "facebook":
			socialSettings.Facebook = st.Value
		case "twitter":
			socialSettings.Twitter = st.Value
		case "instagram":
			socialSettings.Instagram = st.Value
		case "linkedin":
			socialSettings.LinkedIn = st.Value
		case "youtube":
			socialSettings.YouTube = st.Value
		case "wechat":
			socialSettings.WeChat = st.Value
		}
	}

	// 写入缓存
	_ = s.cache.Set(cacheKey, &socialSettings, s.cacheTTL)

	return &socialSettings, nil
}
