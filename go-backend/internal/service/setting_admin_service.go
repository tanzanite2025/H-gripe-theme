package service

import (
	"encoding/json"
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

	s.invalidateSettingCaches(key, group, locale)
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
	cacheKey := settingsAllCacheKey(locale)

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
	cacheKey := settingsPublicCacheKey(locale)

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
	cacheKey := settingsGroupCacheKey(group, locale)

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
		s.invalidateSettingCaches(st.Key, st.Group, st.Locale)
	}

	return nil
}

func (s *SettingService) Delete(key, locale string) error {
	st, _ := s.settingRepo.Get(key, locale)

	if err := s.settingRepo.Delete(key, locale); err != nil {
		return err
	}

	group := ""
	if st != nil {
		group = st.Group
	}

	s.invalidateSettingCaches(key, group, locale)
	return nil
}

func (s *SettingService) GetGroups() ([]string, error) {
	cacheKey := settingsGroupsCacheKey()

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
