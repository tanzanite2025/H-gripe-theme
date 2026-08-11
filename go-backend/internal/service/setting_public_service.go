package service

import (
	"commerce-platform/internal/domain/setting"
	"fmt"
	"strconv"
	"strings"
)

// GetSiteSettings 获取站点设置
func (s *SettingService) GetSiteSettings(locale string) (*setting.SiteSettings, error) {
	cacheKey := settingsStructuredCacheKey("site", locale)

	// 尝试从缓存获取
	var siteSettings setting.SiteSettings
	if s.cache != nil && s.cache.Get(cacheKey, &siteSettings) == nil {
		return &siteSettings, nil
	}

	settings, err := s.getPublicByGroupWithLocaleFallback("site", locale)
	if err != nil {
		return nil, err
	}

	// 转换为结构体
	siteSettings = setting.SiteSettings{}
	for _, st := range settings {
		switch st.Key {
		case "site_name":
			siteSettings.SiteName = st.Value
		case "brand_title":
			siteSettings.BrandTitle = st.Value
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
		case "admin_brand_name":
			siteSettings.AdminBrandName = st.Value
		case "admin_brand_initial":
			siteSettings.AdminBrandInitial = st.Value
		case "admin_panel_label":
			siteSettings.AdminPanelLabel = st.Value
		case "admin_login_title":
			siteSettings.AdminLoginTitle = st.Value
		case "admin_footer_text":
			siteSettings.AdminFooterText = st.Value
		case "admin_html_title":
			siteSettings.AdminHTMLTitle = st.Value
		}
	}
	if siteSettings.BrandTitle != "" {
		siteSettings.SiteName = siteSettings.BrandTitle
	} else {
		siteSettings.BrandTitle = siteSettings.SiteName
	}

	// 写入缓存
	if s.cache != nil {
		_ = s.cache.Set(cacheKey, &siteSettings, s.cacheTTL)
	}

	return &siteSettings, nil
}

func (s *SettingService) getPublicByGroupWithLocaleFallback(group, locale string) ([]setting.Setting, error) {
	merged := make([]setting.Setting, 0)
	seenKeys := make(map[string]struct{})

	for _, candidateLocale := range publicSettingFallbackLocales(locale) {
		settings, err := s.GetPublicByGroup(group, candidateLocale)
		if err != nil {
			return nil, err
		}
		for _, st := range settings {
			if _, ok := seenKeys[st.Key]; ok {
				continue
			}
			seenKeys[st.Key] = struct{}{}
			merged = append(merged, st)
		}
	}

	return merged, nil
}

func publicSettingFallbackLocales(locale string) []string {
	locales := make([]string, 0, 4)
	seen := make(map[string]struct{})
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		locales = append(locales, value)
	}

	add(locale)
	add(strings.ToLower(strings.ReplaceAll(locale, "-", "_")))
	add("en")
	add("global")

	return locales
}

// GetEmailSettings 获取邮件设置
func (s *SettingService) GetEmailSettings(locale string) (*setting.EmailSettings, error) {
	cacheKey := settingsStructuredCacheKey("email", locale)

	// 尝试从缓存获取
	var emailSettings setting.EmailSettings
	if s.cache != nil && s.cache.Get(cacheKey, &emailSettings) == nil {
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
	if s.cache != nil {
		_ = s.cache.Set(cacheKey, &emailSettings, s.cacheTTL)
	}

	return &emailSettings, nil
}

// GetSocialSettings 获取社交媒体设置
func (s *SettingService) GetSocialSettings(locale string) (*setting.SocialSettings, error) {
	cacheKey := settingsStructuredCacheKey("social", locale)

	// 尝试从缓存获取
	var socialSettings setting.SocialSettings
	if s.cache != nil && s.cache.Get(cacheKey, &socialSettings) == nil {
		return &socialSettings, nil
	}

	// 从数据库获取
	settings, err := s.GetPublicByGroup("social", locale)
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
	if s.cache != nil {
		_ = s.cache.Set(cacheKey, &socialSettings, s.cacheTTL)
	}

	return &socialSettings, nil
}
