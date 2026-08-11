package service

import (
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/cache"
	"commerce-platform/internal/repository"
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
	cacheKey := settingValueCacheKey(key, locale)

	// 尝试从缓存获取
	var st setting.Setting
	if s.cache != nil && s.cache.Get(cacheKey, &st) == nil {
		return &st, nil
	}

	// 从数据库获取
	result, err := s.settingRepo.Get(key, locale)
	if err != nil {
		return nil, err
	}

	// 写入缓存
	if s.cache != nil {
		_ = s.cache.Set(cacheKey, result, s.cacheTTL)
	}

	return result, nil
}

func (s *SettingService) GetPublic(key, locale string) (*setting.Setting, error) {
	cacheKey := settingPublicValueCacheKey(key, locale)

	var st setting.Setting
	if s.cache != nil && s.cache.Get(cacheKey, &st) == nil {
		return &st, nil
	}

	result, err := s.settingRepo.GetPublic(key, locale)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, result, s.cacheTTL)
	}

	return result, nil
}

func (s *SettingService) GetPublicByGroup(group, locale string) ([]setting.Setting, error) {
	cacheKey := settingsPublicGroupCacheKey(group, locale)

	var settings []setting.Setting
	if s.cache != nil && s.cache.Get(cacheKey, &settings) == nil {
		return settings, nil
	}

	settings, err := s.settingRepo.GetPublicByGroup(group, locale)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, settings, s.cacheTTL)
	}

	return settings, nil
}

func (s *SettingService) GetPublicGroups() ([]string, error) {
	cacheKey := settingsPublicGroupsCacheKey()

	var groups []string
	if s.cache != nil && s.cache.Get(cacheKey, &groups) == nil {
		return groups, nil
	}

	groups, err := s.settingRepo.GetPublicGroups()
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, groups, s.cacheTTL*10)
	}

	return groups, nil
}
