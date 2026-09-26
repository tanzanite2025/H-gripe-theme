package service

import (
	"commerce-platform/internal/domain/setting"
	"encoding/json"
	"fmt"
	"gorm.io/gorm"
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
	if s.cache != nil && s.cache.Get(cacheKey, &settings) == nil {
		return settings, nil
	}

	settings, err := s.settingRepo.GetAll(locale)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, settings, s.cacheTTL)
	}

	return settings, nil
}

func (s *SettingService) GetAllPublic(locale string) ([]setting.Setting, error) {
	cacheKey := settingsPublicCacheKey(locale)

	var settings []setting.Setting
	if s.cache != nil && s.cache.Get(cacheKey, &settings) == nil {
		return settings, nil
	}

	settings, err := s.settingRepo.GetAllPublic(locale)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, settings, s.cacheTTL)
	}

	return settings, nil
}

func (s *SettingService) GetByGroup(group, locale string) ([]setting.Setting, error) {
	cacheKey := settingsGroupCacheKey(group, locale)

	var settings []setting.Setting
	if s.cache != nil && s.cache.Get(cacheKey, &settings) == nil {
		return settings, nil
	}

	settings, err := s.settingRepo.GetByGroup(group, locale)
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, settings, s.cacheTTL)
	}

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

// BatchSetTx persists settings inside an existing transaction. Domain
// services use this when a setting is the durable public reference for a
// resource whose metadata is being changed in the same transaction.
func (s *SettingService) BatchSetTx(tx *gorm.DB, settings []setting.Setting) error {
	if s == nil || s.settingRepo == nil || tx == nil {
		return fmt.Errorf("setting transaction is unavailable")
	}
	if err := s.settingRepo.WithTx(tx).BatchSet(settings); err != nil {
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
	if s.cache != nil && s.cache.Get(cacheKey, &groups) == nil {
		return groups, nil
	}

	groups, err := s.settingRepo.GetGroups()
	if err != nil {
		return nil, err
	}

	if s.cache != nil {
		_ = s.cache.Set(cacheKey, groups, s.cacheTTL*10)
	}

	return groups, nil
}
