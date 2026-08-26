package service

import (
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/locales"
	"errors"
	"strings"
)

var ErrWebsiteNameServiceUnavailable = errors.New("website name service unavailable")

type WebsiteNameService struct {
	settings *SettingService
}

func NewWebsiteNameService(settings *SettingService) *WebsiteNameService {
	return &WebsiteNameService{settings: settings}
}

func (s *WebsiteNameService) Get(locale string) (*setting.WebsiteNameSettings, error) {
	if s == nil || s.settings == nil {
		return nil, ErrWebsiteNameServiceUnavailable
	}
	return s.resolve(locale, true)
}

func (s *WebsiteNameService) GetAdmin(locale string) (*setting.WebsiteNameSettings, error) {
	if s == nil || s.settings == nil {
		return nil, ErrWebsiteNameServiceUnavailable
	}

	normalizedLocale := "en"
	if strings.TrimSpace(locale) != "" {
		var err error
		normalizedLocale, err = requireSupportedLocale(locale)
		if err != nil {
			return nil, err
		}
	}

	return s.resolve(normalizedLocale, false)
}

func (s *WebsiteNameService) Update(request setting.WebsiteNameUpdateRequest) (*setting.WebsiteNameSettings, error) {
	if s == nil || s.settings == nil {
		return nil, ErrWebsiteNameServiceUnavailable
	}
	locale := "en"
	if strings.TrimSpace(request.Locale) != "" {
		var err error
		locale, err = requireSupportedLocale(request.Locale)
		if err != nil {
			return nil, err
		}
	}
	values := request.Settings()
	values.Locale = locale

	if err := s.settings.BatchSet(websiteNameRecords(values, locale)); err != nil {
		return nil, err
	}

	return s.GetAdmin(locale)
}

func (s *WebsiteNameService) resolve(locale string, publicOnly bool) (*setting.WebsiteNameSettings, error) {
	normalizedLocale := normalizeWebsiteNameLocale(locale)
	result := setting.DefaultWebsiteNameSettings(normalizedLocale)

	seenKeys := make(map[string]struct{})
	for _, candidateLocale := range websiteNameFallbackLocales(normalizedLocale) {
		records, err := s.records(candidateLocale, publicOnly)
		if err != nil {
			return nil, err
		}

		for _, record := range records {
			if _, seen := seenKeys[record.Key]; seen {
				continue
			}
			if strings.TrimSpace(record.Value) == "" {
				continue
			}
			seenKeys[record.Key] = struct{}{}
			applyWebsiteNameRecord(&result, record.Key, record.Value)
		}
	}

	result.Locale = normalizedLocale
	return &result, nil
}

func (s *WebsiteNameService) records(locale string, publicOnly bool) ([]setting.Setting, error) {
	if publicOnly {
		return s.settings.GetPublicByGroup(setting.WebsiteNameGroup, locale)
	}
	return s.settings.GetByGroup(setting.WebsiteNameGroup, locale)
}

func websiteNameRecords(values setting.WebsiteNameSettings, locale string) []setting.Setting {
	records := make([]setting.Setting, 0, 6)
	add := func(key, value string) {
		records = append(records, setting.Setting{
			Key:         key,
			Value:       strings.TrimSpace(value),
			Type:        "string",
			Group:       setting.WebsiteNameGroup,
			Locale:      locale,
			IsPublic:    true,
			Description: setting.WebsiteNameSettingDescription(key),
		})
	}

	add(setting.WebsiteNameKeyStatus, values.Status)
	add(setting.WebsiteNameKeyIntro, values.Intro)
	add(setting.WebsiteNameKeyEyebrow, values.Eyebrow)
	add(setting.WebsiteNameKeyTitle, values.Title)
	add(setting.WebsiteNameKeyBody, values.Body)
	add(setting.WebsiteNameKeyNote, values.Note)

	return records
}

func applyWebsiteNameRecord(target *setting.WebsiteNameSettings, key, value string) {
	switch key {
	case setting.WebsiteNameKeyStatus:
		target.Status = value
	case setting.WebsiteNameKeyIntro:
		target.Intro = value
	case setting.WebsiteNameKeyEyebrow:
		target.Eyebrow = value
	case setting.WebsiteNameKeyTitle:
		target.Title = value
	case setting.WebsiteNameKeyBody:
		target.Body = value
	case setting.WebsiteNameKeyNote:
		target.Note = value
	}
}

func normalizeWebsiteNameLocale(locale string) string {
	if normalized := locales.ResolveSupported(locale); normalized != "" {
		return normalized
	}
	return "en"
}

func websiteNameFallbackLocales(locale string) []string {
	normalized := normalizeWebsiteNameLocale(locale)
	if normalized == "en" {
		return []string{"en"}
	}
	return []string{normalized, "en"}
}
