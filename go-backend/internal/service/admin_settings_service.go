package service

import (
	"errors"
	"strings"
	"tanzanite/internal/domain/setting"
)

var ErrSettingManagedByDomainService = errors.New("this setting group must be changed through its domain service")

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
		if err := rejectDomainManagedSettingGroup(group); err != nil {
			return nil, err
		}
		return s.settings.GetByGroup(group, locale)
	}
	settings, err := s.settings.GetAll(locale)
	if err != nil {
		return nil, err
	}
	return filterDomainManagedSettings(settings), nil
}

func (s *AdminSettingsService) GetSetting(key, locale string) (*setting.Setting, error) {
	if err := rejectDomainManagedSettingKey(key); err != nil {
		return nil, err
	}
	return s.settings.Get(key, locale)
}

func (s *AdminSettingsService) GetDomainManagedSetting(key, locale string) (*setting.Setting, error) {
	return s.settings.Get(key, locale)
}

func (s *AdminSettingsService) UpdateSetting(req setting.UpdateSettingRequest) (*setting.Setting, error) {
	if err := rejectDomainManagedSetting(req); err != nil {
		return nil, err
	}
	st := normalizeSettingRequest(req)
	if err := s.settings.BatchSet([]setting.Setting{st}); err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *AdminSettingsService) UpdateDomainManagedSetting(req setting.UpdateSettingRequest) (*setting.Setting, error) {
	st := normalizeSettingRequest(req)
	if err := s.settings.BatchSet([]setting.Setting{st}); err != nil {
		return nil, err
	}
	return &st, nil
}

func (s *AdminSettingsService) BatchUpdateSettings(req setting.BatchUpdateSettingsRequest) (int, error) {
	settings := make([]setting.Setting, 0, len(req.Settings))
	for _, item := range req.Settings {
		if err := rejectDomainManagedSetting(item); err != nil {
			return 0, err
		}
		settings = append(settings, normalizeSettingRequest(item))
	}

	if err := s.settings.BatchSet(settings); err != nil {
		return 0, err
	}

	return len(settings), nil
}

func rejectDomainManagedSettingGroup(group string) error {
	normalized := strings.ToLower(strings.TrimSpace(group))
	if normalized == "loyalty" || normalized == "redeem" || normalized == "currency" || normalized == "payment_secret" {
		return ErrSettingManagedByDomainService
	}
	return nil
}

func (s *AdminSettingsService) DeleteSetting(key, locale string) error {
	if err := rejectDomainManagedSettingKey(key); err != nil {
		return err
	}
	return s.settings.Delete(key, locale)
}

func (s *AdminSettingsService) DeleteDomainManagedSetting(key, locale string) error {
	return s.settings.Delete(key, locale)
}

func (s *AdminSettingsService) GetGroups() ([]string, error) {
	groups, err := s.settings.GetGroups()
	if err != nil {
		return nil, err
	}
	return filterDomainManagedSettingGroups(groups), nil
}

func (s *AdminSettingsService) GetByGroup(group, locale string) ([]setting.Setting, error) {
	if err := rejectDomainManagedSettingGroup(group); err != nil {
		return nil, err
	}
	return s.settings.GetByGroup(group, locale)
}

func rejectDomainManagedSetting(req setting.UpdateSettingRequest) error {
	if err := rejectDomainManagedSettingGroup(req.Group); err != nil {
		return err
	}
	return rejectDomainManagedSettingKey(req.Key)
}

func rejectDomainManagedSettingKey(key string) error {
	normalized := strings.ToLower(strings.TrimSpace(key))
	if strings.HasPrefix(normalized, "tz_loyalty_") ||
		strings.HasPrefix(normalized, "tz_redeem_") ||
		strings.HasPrefix(normalized, "currency_") ||
		strings.HasPrefix(normalized, "payment_gateway_") {
		return ErrSettingManagedByDomainService
	}
	return nil
}

func filterDomainManagedSettings(settings []setting.Setting) []setting.Setting {
	filtered := make([]setting.Setting, 0, len(settings))
	for _, item := range settings {
		if rejectDomainManagedSettingGroup(item.Group) == nil && rejectDomainManagedSettingKey(item.Key) == nil {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func filterDomainManagedSettingGroups(groups []string) []string {
	filtered := make([]string, 0, len(groups))
	for _, group := range groups {
		if rejectDomainManagedSettingGroup(group) == nil {
			filtered = append(filtered, group)
		}
	}
	return filtered
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
