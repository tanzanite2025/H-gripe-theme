package service

import (
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
	cacheKey := settingValueCacheKey(key, locale)

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
