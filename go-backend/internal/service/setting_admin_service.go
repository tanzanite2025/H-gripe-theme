package service

import (
	"encoding/json"
	"fmt"
	"tanzanite/internal/domain/setting"
)

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

	cacheKey := fmt.Sprintf("setting:%s:%s", key, locale)
	_ = s.cache.Delete(cacheKey)

	groupCacheKey := fmt.Sprintf("settings:%s:%s", group, locale)
	_ = s.cache.Delete(groupCacheKey)

	return nil
}

func (s *SettingService) SetJSON(key string, value interface{}, group, locale string) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.Set(key, string(jsonData), "json", group, locale)
}

func (s *SettingService) GetAll(locale string) ([]setting.Setting, error) {
	cacheKey := fmt.Sprintf("settings:all:%s", locale)

	var settings []setting.Setting
	if err := s.cache.Get(cacheKey, &settings); err == nil {
		return settings, nil
	}

	settings, err := s.settingRepo.GetAll(locale)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(cacheKey, settings, s.cacheTTL)

	return settings, nil
}

func (s *SettingService) GetAllPublic(locale string) ([]setting.Setting, error) {
	cacheKey := fmt.Sprintf("settings:public:%s", locale)

	var settings []setting.Setting
	if err := s.cache.Get(cacheKey, &settings); err == nil {
		return settings, nil
	}

	settings, err := s.settingRepo.GetAllPublic(locale)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(cacheKey, settings, s.cacheTTL)

	return settings, nil
}

func (s *SettingService) GetByGroup(group, locale string) ([]setting.Setting, error) {
	cacheKey := fmt.Sprintf("settings:group:%s:%s", group, locale)

	var settings []setting.Setting
	if err := s.cache.Get(cacheKey, &settings); err == nil {
		return settings, nil
	}

	settings, err := s.settingRepo.GetByGroup(group, locale)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(cacheKey, settings, s.cacheTTL)

	return settings, nil
}

func (s *SettingService) BatchSet(settings []setting.Setting) error {
	if err := s.settingRepo.BatchSet(settings); err != nil {
		return err
	}

	for _, st := range settings {
		cacheKey := fmt.Sprintf("setting:%s:%s", st.Key, st.Locale)
		_ = s.cache.Delete(cacheKey)

		groupCacheKey := fmt.Sprintf("settings:group:%s:%s", st.Group, st.Locale)
		_ = s.cache.Delete(groupCacheKey)

		allCacheKey := fmt.Sprintf("settings:all:%s", st.Locale)
		_ = s.cache.Delete(allCacheKey)

		publicCacheKey := fmt.Sprintf("settings:public:%s", st.Locale)
		_ = s.cache.Delete(publicCacheKey)
	}

	return nil
}

func (s *SettingService) Delete(key, locale string) error {
	if err := s.settingRepo.Delete(key, locale); err != nil {
		return err
	}

	cacheKey := fmt.Sprintf("setting:%s:%s", key, locale)
	_ = s.cache.Delete(cacheKey)

	allCacheKey := fmt.Sprintf("settings:all:%s", locale)
	_ = s.cache.Delete(allCacheKey)

	return nil
}

func (s *SettingService) GetGroups() ([]string, error) {
	cacheKey := "settings:groups"

	var groups []string
	if err := s.cache.Get(cacheKey, &groups); err == nil {
		return groups, nil
	}

	groups, err := s.settingRepo.GetGroups()
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(cacheKey, groups, s.cacheTTL*10)

	return groups, nil
}
