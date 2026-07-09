package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/repository"
	"time"
)

type SettingService struct {
	settingRepo *repository.SettingRepository
	cache       *cache.RedisCache
	cacheTTL    time.Duration
}

func NewSettingService(settingRepo *repository.SettingRepository, cache *cache.RedisCache, cacheTTL int) *SettingService {
	return &SettingService{
		settingRepo: settingRepo,
		cache:       cache,
		cacheTTL:    time.Duration(cacheTTL) * time.Second,
	}
}

// Get 获取设置
func (s *SettingService) Get(key, locale string) (*setting.Setting, error) {
	cacheKey := fmt.Sprintf("setting:%s:%s", key, locale)

	// 尝试从缓存获取
	var st setting.Setting
	if err := s.cache.Get(cacheKey, &st); err == nil {
		return &st, nil
	}

	// 从数据库获取
	result, err := s.settingRepo.Get(key, locale)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	_ = s.cache.Set(cacheKey, result, s.cacheTTL)

	return result, nil
}

// GetSiteSettings 获取站点设置
func (s *SettingService) GetSiteSettings(locale string) (*setting.SiteSettings, error) {
	cacheKey := fmt.Sprintf("settings:site:%s", locale)

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
	cacheKey := fmt.Sprintf("settings:quick-buy:%s", locale)

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
	cacheKey := fmt.Sprintf("settings:email:%s", locale)

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
	cacheKey := fmt.Sprintf("settings:seo:%s", locale)

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
	cacheKey := fmt.Sprintf("settings:social:%s", locale)

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

// GetRedeemSettings 获取积分兑换配置
func (s *SettingService) GetRedeemSettings(locale string) (*setting.RedeemSettings, error) {
	cacheKey := fmt.Sprintf("settings:redeem:%s", locale)

	// 尝试从缓存获取
	var redeemSettings setting.RedeemSettings
	if err := s.cache.Get(cacheKey, &redeemSettings); err == nil {
		return &redeemSettings, nil
	}

	// 从数据库获取
	settings, err := s.settingRepo.GetByGroup("redeem", locale)
	if err != nil {
		return nil, err
	}

	// 转换为 map 便于查询与校验
	m := make(map[string]string)
	for _, st := range settings {
		m[st.Key] = st.Value
	}

	// 强制校验必要配置项，实现 Fail Loudly 原则
	requiredKeys := []string{
		"tz_redeem_enabled",
		"tz_redeem_exchange_rate",
		"tz_redeem_min_points",
		"tz_redeem_max_value_per_day",
		"tz_redeem_card_expiry_days",
		"tz_redeem_preset_values",
	}

	for _, k := range requiredKeys {
		if _, ok := m[k]; !ok {
			errMsg := fmt.Sprintf("[CRITICAL] Redeem setting '%s' is missing in settings table", k)
			return nil, errors.New(errMsg)
		}
	}

	// 解析数据并映射至结构体
	redeemSettings.Enabled = m["tz_redeem_enabled"] == "1" || m["tz_redeem_enabled"] == "true"

	exchangeRate, err := strconv.Atoi(m["tz_redeem_exchange_rate"])
	if err != nil {
		return nil, fmt.Errorf("[CRITICAL] Invalid format for tz_redeem_exchange_rate: %v", err)
	}
	redeemSettings.ExchangeRate = exchangeRate

	minPoints, err := strconv.Atoi(m["tz_redeem_min_points"])
	if err != nil {
		return nil, fmt.Errorf("[CRITICAL] Invalid format for tz_redeem_min_points: %v", err)
	}
	redeemSettings.MinPoints = minPoints

	maxValue, err := strconv.ParseFloat(m["tz_redeem_max_value_per_day"], 64)
	if err != nil {
		return nil, fmt.Errorf("[CRITICAL] Invalid format for tz_redeem_max_value_per_day: %v", err)
	}
	redeemSettings.MaxValuePerDay = maxValue

	expiryDays, err := strconv.Atoi(m["tz_redeem_card_expiry_days"])
	if err != nil {
		return nil, fmt.Errorf("[CRITICAL] Invalid format for tz_redeem_card_expiry_days: %v", err)
	}
	redeemSettings.CardExpiryDays = expiryDays

	// 解析面值预设列表
	presetStr := m["tz_redeem_preset_values"]
	parts := strings.Split(presetStr, ",")
	presets := make([]float64, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		val, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return nil, fmt.Errorf("[CRITICAL] Invalid format for tz_redeem_preset_values item '%s': %v", p, err)
		}
		presets = append(presets, val)
	}
	redeemSettings.PresetValues = presets

	// 写入缓存
	_ = s.cache.Set(cacheKey, &redeemSettings, s.cacheTTL)

	return &redeemSettings, nil
}
