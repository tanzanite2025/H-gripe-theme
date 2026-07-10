package service

import (
	"tanzanite/internal/domain/setting"
)

type AdminSettingsService struct {
	settings *SettingService
}

func NewAdminSettingsService(settings *SettingService) *AdminSettingsService {
	return &AdminSettingsService{
		settings: settings,
	}
}

func (s *AdminSettingsService) ListSettings(locale, group string) ([]setting.Setting, error) {
	if group != "" {
		return s.settings.GetByGroup(group, locale)
	}
	return s.settings.GetAll(locale)
}

func (s *AdminSettingsService) GetSetting(key, locale string) (*setting.Setting, error) {
	return s.settings.Get(key, locale)
}

func (s *AdminSettingsService) UpdateSetting(req setting.UpdateSettingRequest) (*setting.Setting, error) {
	st := normalizeSettingRequest(req)
	if err := s.settings.BatchSet([]setting.Setting{st}); err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *AdminSettingsService) BatchUpdateSettings(req setting.BatchUpdateSettingsRequest) (int, error) {
	settings := make([]setting.Setting, 0, len(req.Settings))
	for _, item := range req.Settings {
		settings = append(settings, normalizeSettingRequest(item))
	}

	if err := s.settings.BatchSet(settings); err != nil {
		return 0, err
	}

	return len(settings), nil
}

func (s *AdminSettingsService) DeleteSetting(key, locale string) error {
	return s.settings.Delete(key, locale)
}

func (s *AdminSettingsService) GetGroups() ([]string, error) {
	return s.settings.GetGroups()
}

func (s *AdminSettingsService) GetByGroup(group, locale string) ([]setting.Setting, error) {
	return s.settings.GetByGroup(group, locale)
}

func normalizeSettingRequest(req setting.UpdateSettingRequest) setting.Setting {
	locale := req.Locale
	if locale == "" {
		locale = "en"
	}

	settingType := req.Type
	if settingType == "" {
		settingType = "string"
	}

	return setting.Setting{
		Key:         req.Key,
		Value:       req.Value,
		Type:        settingType,
		Group:       req.Group,
		Locale:      locale,
		IsPublic:    req.IsPublic,
		Description: req.Description,
	}
}
