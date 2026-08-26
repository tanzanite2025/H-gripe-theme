package service

import (
	"commerce-platform/internal/domain/setting"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var ErrSettingManagedByDomainService = errors.New("this setting group must be changed through its domain service")
var ErrSettingInvalid = errors.New("setting value is invalid")

const maskedSettingValue = "********"

var supportedSocialSettingKeys = map[string]struct{}{
	"facebook":  {},
	"instagram": {},
	"x":         {},
	"youtube":   {},
	"reddit":    {},
}

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
		return s.GetByGroup(group, locale)
	}
	settings, err := s.settings.GetAll(locale)
	if err != nil {
		return nil, err
	}
	settings = FilterDomainManagedSettings(settings)
	settings = filterSupportedSocialSettings(settings)
	return maskSensitiveSettings(settings), nil
}

func (s *AdminSettingsService) GetSetting(key, locale string) (*setting.Setting, error) {
	if err := rejectDomainManagedSettingKey(key); err != nil {
		return nil, err
	}
	item, err := s.settings.Get(key, locale)
	if err != nil {
		return nil, err
	}
	return maskSensitiveSetting(item), nil
}

func (s *AdminSettingsService) GetDomainManagedSetting(key, locale string) (*setting.Setting, error) {
	return s.settings.Get(key, locale)
}

func (s *AdminSettingsService) UpdateSetting(req setting.UpdateSettingRequest) (*setting.Setting, error) {
	if err := rejectDomainManagedSetting(req); err != nil {
		return nil, err
	}
	st, err := s.normalizeRequest(req)
	if err != nil {
		return nil, err
	}
	if err := s.settings.BatchSet([]setting.Setting{st}); err != nil {
		return nil, err
	}
	return maskSensitiveSetting(&st), nil
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
		st, err := s.normalizeRequest(item)
		if err != nil {
			return 0, err
		}
		settings = append(settings, st)
	}

	if err := s.settings.BatchSet(settings); err != nil {
		return 0, err
	}

	return len(settings), nil
}

func rejectDomainManagedSettingGroup(group string) error {
	if IsDomainManagedSettingGroup(group) {
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
	return FilterDomainManagedSettingGroups(groups), nil
}

func (s *AdminSettingsService) GetByGroup(group, locale string) ([]setting.Setting, error) {
	if err := rejectDomainManagedSettingGroup(group); err != nil {
		return nil, err
	}
	items, err := s.settings.GetByGroup(group, locale)
	if err != nil {
		return nil, err
	}
	if strings.EqualFold(strings.TrimSpace(group), "social") {
		items = filterSupportedSocialSettings(items)
	}
	return maskSensitiveSettings(items), nil
}

func (s *AdminSettingsService) normalizeRequest(req setting.UpdateSettingRequest) (setting.Setting, error) {
	if err := validateSettingRequest(req); err != nil {
		return setting.Setting{}, err
	}
	st := normalizeSettingRequest(req)
	if isSensitiveSettingKey(st.Key) {
		st.IsPublic = false
		value := strings.TrimSpace(st.Value)
		if value == "" || value == maskedSettingValue {
			if existing, err := s.settings.Get(st.Key, st.Locale); err == nil {
				st.Value = existing.Value
			}
		}
	}
	return st, nil
}

func validateSettingRequest(req setting.UpdateSettingRequest) error {
	if strings.EqualFold(strings.TrimSpace(req.Group), "social") {
		key := strings.ToLower(strings.TrimSpace(req.Key))
		if _, ok := supportedSocialSettingKeys[key]; !ok {
			return fmt.Errorf("%w: unsupported social setting %q", ErrSettingInvalid, req.Key)
		}
	}

	if !strings.HasSuffix(strings.TrimSpace(req.Key), "_endpoint") {
		return nil
	}
	key := strings.TrimSpace(req.Key)
	if !strings.HasPrefix(key, "customs_lookup_") {
		return nil
	}
	if err := validateCustomsLookupEndpoint(req.Value); err != nil {
		return fmt.Errorf("%w: %v", ErrSettingInvalid, err)
	}
	return nil
}

func filterSupportedSocialSettings(items []setting.Setting) []setting.Setting {
	filtered := make([]setting.Setting, 0, len(items))
	for _, item := range items {
		if !strings.EqualFold(strings.TrimSpace(item.Group), "social") {
			filtered = append(filtered, item)
			continue
		}
		if _, ok := supportedSocialSettingKeys[strings.ToLower(strings.TrimSpace(item.Key))]; ok {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func validateCustomsLookupEndpoint(value string) error {
	endpoint := strings.TrimSpace(value)
	if endpoint == "" {
		return errors.New("customs lookup endpoint is required")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme != "https" || parsed.Hostname() == "" || parsed.User != nil {
		return errors.New("customs lookup endpoint must be an HTTPS URL without credentials")
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	if host == "localhost" || strings.HasSuffix(host, ".localhost") || host == "metadata.google.internal" {
		return errors.New("customs lookup endpoint host is not allowed")
	}
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedSettingEndpointIP(ip) {
			return errors.New("customs lookup endpoint host is not allowed")
		}
	} else if resolved, lookupErr := net.LookupIP(host); lookupErr == nil {
		for _, resolvedIP := range resolved {
			if isBlockedSettingEndpointIP(resolvedIP) {
				return errors.New("customs lookup endpoint resolves to a private or local address")
			}
		}
	}
	return nil
}

func isBlockedSettingEndpointIP(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsUnspecified() || ip.IsMulticast()
}

func isSensitiveSettingKey(key string) bool {
	key = strings.ToLower(strings.TrimSpace(key))
	return strings.HasSuffix(key, "_api_key") ||
		strings.HasSuffix(key, "_api_secret") ||
		strings.HasSuffix(key, "_password") ||
		strings.HasSuffix(key, "_access_token") ||
		strings.HasSuffix(key, "_refresh_token")
}

func maskSensitiveSetting(item *setting.Setting) *setting.Setting {
	if item == nil {
		return nil
	}
	copy := *item
	if isSensitiveSettingKey(copy.Key) && strings.TrimSpace(copy.Value) != "" {
		copy.Value = maskedSettingValue
	}
	return &copy
}

func maskSensitiveSettings(items []setting.Setting) []setting.Setting {
	for index := range items {
		masked := maskSensitiveSetting(&items[index])
		if masked != nil {
			items[index] = *masked
		}
	}
	return items
}

func rejectDomainManagedSetting(req setting.UpdateSettingRequest) error {
	if err := rejectDomainManagedSettingGroup(req.Group); err != nil {
		return err
	}
	return rejectDomainManagedSettingKey(req.Key)
}

func rejectDomainManagedSettingKey(key string) error {
	if IsDomainManagedSettingKey(key) {
		return ErrSettingManagedByDomainService
	}
	return nil
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
