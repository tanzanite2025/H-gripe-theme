package service

import (
	"encoding/json"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/pkg/cache"
	"tanzanite/internal/repository"
	"fmt"
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

// Set 设置值
func (s *SettingService) Set(key, value, settingType, group, locale string) error {
	st := &setting.Setting{
		Key:    key,
		Value:  value,
		Type:   settingType,
		Group:  group,
		Locale: locale,
	}

	if err := s.settingRepo.Set(st); err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("setting:%s:%s", key, locale)
	_ = s.cache.Delete(cacheKey)

	// 清除分组缓存
	groupCacheKey := fmt.Sprintf("settings:%s:%s", group, locale)
	_ = s.cache.Delete(groupCacheKey)

	return nil
}

// SetJSON 设置JSON值
func (s *SettingService) SetJSON(key string, value interface{}, group, locale string) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.Set(key, string(jsonData), "json", group, locale)
}

// GetAll 获取所有设置
func (s *SettingService) GetAll(locale string) ([]setting.Setting, error) {
	cacheKey := fmt.Sprintf("settings:all:%s", locale)

	// 尝试从缓存获取
	var settings []setting.Setting
	if err := s.cache.Get(cacheKey, &settings); err == nil {
		return settings, nil
	}

	// 从数据库获取
	settings, err := s.settingRepo.GetAll(locale)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	_ = s.cache.Set(cacheKey, settings, s.cacheTTL)

	return settings, nil
}

// GetAllPublic 获取所有公开设置
func (s *SettingService) GetAllPublic(locale string) ([]setting.Setting, error) {
	cacheKey := fmt.Sprintf("settings:public:%s", locale)

	// 尝试从缓存获取
	var settings []setting.Setting
	if err := s.cache.Get(cacheKey, &settings); err == nil {
		return settings, nil
	}

	// 从数据库获取
	settings, err := s.settingRepo.GetAllPublic(locale)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	_ = s.cache.Set(cacheKey, settings, s.cacheTTL)

	return settings, nil
}

// GetByGroup 获取分组设置
func (s *SettingService) GetByGroup(group, locale string) ([]setting.Setting, error) {
	cacheKey := fmt.Sprintf("settings:group:%s:%s", group, locale)

	// 尝试从缓存获取
	var settings []setting.Setting
	if err := s.cache.Get(cacheKey, &settings); err == nil {
		return settings, nil
	}

	// 从数据库获取
	settings, err := s.settingRepo.GetByGroup(group, locale)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	_ = s.cache.Set(cacheKey, settings, s.cacheTTL)

	return settings, nil
}

// BatchSet 批量设置
func (s *SettingService) BatchSet(settings []setting.Setting) error {
	if err := s.settingRepo.BatchSet(settings); err != nil {
		return err
	}

	// 清除所有相关缓存
	for _, st := range settings {
		// 清除单个设置缓存
		cacheKey := fmt.Sprintf("setting:%s:%s", st.Key, st.Locale)
		_ = s.cache.Delete(cacheKey)

		// 清除分组缓存
		groupCacheKey := fmt.Sprintf("settings:group:%s:%s", st.Group, st.Locale)
		_ = s.cache.Delete(groupCacheKey)

		// 清除所有设置缓存
		allCacheKey := fmt.Sprintf("settings:all:%s", st.Locale)
		_ = s.cache.Delete(allCacheKey)

		// 清除公开设置缓存
		publicCacheKey := fmt.Sprintf("settings:public:%s", st.Locale)
		_ = s.cache.Delete(publicCacheKey)
	}

	return nil
}

// Delete 删除设置
func (s *SettingService) Delete(key, locale string) error {
	if err := s.settingRepo.Delete(key, locale); err != nil {
		return err
	}

	// 清除缓存
	cacheKey := fmt.Sprintf("setting:%s:%s", key, locale)
	_ = s.cache.Delete(cacheKey)

	// 清除所有设置缓存
	allCacheKey := fmt.Sprintf("settings:all:%s", locale)
	_ = s.cache.Delete(allCacheKey)

	return nil
}

// GetGroups 获取所有分组
func (s *SettingService) GetGroups() ([]string, error) {
	cacheKey := "settings:groups"

	// 尝试从缓存获取
	var groups []string
	if err := s.cache.Get(cacheKey, &groups); err == nil {
		return groups, nil
	}

	// 从数据库获取
	groups, err := s.settingRepo.GetGroups()
	if err != nil {
		return nil, err
	}

	// 写入缓存（较长的TTL，因为分组不常变化）
	_ = s.cache.Set(cacheKey, groups, s.cacheTTL*10)

	return groups, nil
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
			var port int
			fmt.Sscanf(st.Value, "%d", &port)
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
