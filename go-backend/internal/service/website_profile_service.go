package service

import (
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/websitecontent"
	"fmt"
	"strings"
)

const websiteProfileDefaultFactoryImageURL = "/company/ourstory/factory/factory-premoldlayupworkshop6.webp"

type WebsiteProfileService struct {
	settings *SettingService
}

func NewWebsiteProfileService(settings *SettingService) *WebsiteProfileService {
	return &WebsiteProfileService{settings: settings}
}

func (s *WebsiteProfileService) Get(locale string) (*setting.WebsiteProfileSettings, error) {
	return s.resolve(locale, true)
}

func (s *WebsiteProfileService) GetAdmin(locale string) (*setting.WebsiteProfileSettings, error) {
	return s.resolve(locale, false)
}

func (s *WebsiteProfileService) Update(request setting.WebsiteProfileUpdateRequest) (*setting.WebsiteProfileSettings, error) {
	locale := normalizeWebsiteProfileLocale(request.Locale)
	values := request.Settings()
	values.Locale = locale
	normalizedBody, err := websitecontent.NormalizeWebsiteBody(values.StatementBody)
	if err != nil {
		return nil, fmt.Errorf("normalize website profile body: %w", err)
	}
	values.StatementBody = normalizedBody

	records := websiteProfileRecords(values, locale)
	if err := s.settings.BatchSet(records); err != nil {
		return nil, err
	}

	return s.GetAdmin(locale)
}

func (s *WebsiteProfileService) resolve(locale string, publicOnly bool) (*setting.WebsiteProfileSettings, error) {
	normalizedLocale := normalizeWebsiteProfileLocale(locale)
	result := setting.DefaultWebsiteProfileSettings(normalizedLocale)
	if result.FactoryImageURL == "" {
		result.FactoryImageURL = websiteProfileDefaultFactoryImageURL
	}

	seenKeys := make(map[string]struct{})
	statementSourceFound := false
	for _, candidateLocale := range websiteProfileFallbackLocales(normalizedLocale) {
		records, err := s.records(candidateLocale, publicOnly)
		if err != nil {
			return nil, err
		}

		candidateBody := ""
		candidateLegacyParts := map[string]string{}
		for _, record := range records {
			if _, seen := seenKeys[record.Key]; seen {
				continue
			}

			switch record.Key {
			case setting.WebsiteProfileKeyStatementBody:
				if candidateBody != "" || strings.TrimSpace(record.Value) == "" {
					continue
				}
				value, normalizeErr := websitecontent.NormalizeWebsiteBody(record.Value)
				if normalizeErr != nil {
					return nil, fmt.Errorf("normalize website profile body: %w", normalizeErr)
				}
				if strings.TrimSpace(value) == "" {
					continue
				}
				candidateBody = value
				continue
			case setting.WebsiteProfileKeyStatementParagraph1, setting.WebsiteProfileKeyStatementParagraph2:
				if _, seen := candidateLegacyParts[record.Key]; seen || strings.TrimSpace(record.Value) == "" {
					continue
				}
				candidateLegacyParts[record.Key] = record.Value
				continue
			}

			if strings.TrimSpace(record.Value) == "" {
				continue
			}
			seenKeys[record.Key] = struct{}{}
			applyWebsiteProfileRecord(&result, record.Key, record.Value)
		}

		if statementSourceFound {
			continue
		}

		if candidateBody != "" {
			result.StatementBody = candidateBody
			statementSourceFound = true
			continue
		}

		if len(candidateLegacyParts) > 0 {
			legacyBody := strings.TrimSpace(strings.Join([]string{
				candidateLegacyParts[setting.WebsiteProfileKeyStatementParagraph1],
				candidateLegacyParts[setting.WebsiteProfileKeyStatementParagraph2],
			}, "\n\n"))
			normalizedBody, normalizeErr := websitecontent.NormalizeWebsiteBody(legacyBody)
			if normalizeErr != nil {
				return nil, fmt.Errorf("normalize legacy website profile body: %w", normalizeErr)
			}
			result.StatementBody = normalizedBody
			statementSourceFound = true
		}
	}

	normalizedBody, err := websitecontent.NormalizeWebsiteBody(result.StatementBody)
	if err != nil {
		return nil, fmt.Errorf("normalize website profile body: %w", err)
	}
	result.StatementBody = normalizedBody
	result.Locale = normalizedLocale
	if result.FactoryImageURL == "" {
		result.FactoryImageURL = websiteProfileDefaultFactoryImageURL
	}
	return &result, nil
}

func (s *WebsiteProfileService) records(locale string, publicOnly bool) ([]setting.Setting, error) {
	if publicOnly {
		return s.settings.GetPublicByGroup(setting.WebsiteProfileGroup, locale)
	}
	return s.settings.GetByGroup(setting.WebsiteProfileGroup, locale)
}

func websiteProfileRecords(values setting.WebsiteProfileSettings, locale string) []setting.Setting {
	records := make([]setting.Setting, 0, 24)
	add := func(key, value, recordLocale string) {
		records = append(records, setting.Setting{
			Key:         key,
			Value:       strings.TrimSpace(value),
			Type:        "string",
			Group:       setting.WebsiteProfileGroup,
			Locale:      recordLocale,
			IsPublic:    true,
			Description: fmt.Sprintf("Website profile: %s", key),
		})
	}

	add(setting.WebsiteProfileKeyEyebrow, values.Eyebrow, locale)
	add(setting.WebsiteProfileKeyTitle, values.Title, locale)
	add(setting.WebsiteProfileKeyLead, values.Lead, locale)
	add(setting.WebsiteProfileKeyScope, values.Scope, locale)
	add(setting.WebsiteProfileKeyAvatarLabel, values.AvatarLabel, locale)
	add(setting.WebsiteProfileKeyAvatarMark, values.AvatarMark, locale)
	add(setting.WebsiteProfileKeyProfileLabel, values.ProfileLabel, locale)
	add(setting.WebsiteProfileKeyProfileRole, values.ProfileRole, locale)
	add(setting.WebsiteProfileKeyProfileContext, values.ProfileContext, locale)
	add(setting.WebsiteProfileKeyStatementEyebrow, values.StatementEyebrow, locale)
	add(setting.WebsiteProfileKeyStatementTitle, values.StatementTitle, locale)
	add(setting.WebsiteProfileKeyStatementBody, values.StatementBody, locale)
	add(setting.WebsiteProfileKeyStatementParagraph1, "", locale)
	add(setting.WebsiteProfileKeyStatementParagraph2, "", locale)
	add(setting.WebsiteProfileKeyFactoryImageAlt, values.FactoryImageAlt, locale)
	add(setting.WebsiteProfileKeyFactoryImageCaption, values.FactoryImageCaption, locale)
	add(setting.WebsiteProfileKeyFactoryEyebrow, values.FactoryEyebrow, locale)
	add(setting.WebsiteProfileKeyFactoryTitle, values.FactoryTitle, locale)
	add(setting.WebsiteProfileKeyFactoryBody, values.FactoryBody, locale)
	add(setting.WebsiteProfileKeyFactoryCTA, values.FactoryCTA, locale)

	add(setting.WebsiteProfileKeyAvatarURL, values.AvatarURL, "global")
	add(setting.WebsiteProfileKeyFactoryImageURL, values.FactoryImageURL, "global")
	add(setting.WebsiteProfileKeyFactoryLink, values.FactoryLink, "global")

	return records
}

func applyWebsiteProfileRecord(target *setting.WebsiteProfileSettings, key, value string) {
	switch key {
	case setting.WebsiteProfileKeyEyebrow:
		target.Eyebrow = value
	case setting.WebsiteProfileKeyTitle:
		target.Title = value
	case setting.WebsiteProfileKeyLead:
		target.Lead = value
	case setting.WebsiteProfileKeyScope:
		target.Scope = value
	case setting.WebsiteProfileKeyAvatarURL:
		target.AvatarURL = value
	case setting.WebsiteProfileKeyAvatarLabel:
		target.AvatarLabel = value
	case setting.WebsiteProfileKeyAvatarMark:
		target.AvatarMark = value
	case setting.WebsiteProfileKeyProfileLabel:
		target.ProfileLabel = value
	case setting.WebsiteProfileKeyProfileRole:
		target.ProfileRole = value
	case setting.WebsiteProfileKeyProfileContext:
		target.ProfileContext = value
	case setting.WebsiteProfileKeyStatementEyebrow:
		target.StatementEyebrow = value
	case setting.WebsiteProfileKeyStatementTitle:
		target.StatementTitle = value
	case setting.WebsiteProfileKeyStatementBody:
		target.StatementBody = value
	case setting.WebsiteProfileKeyFactoryImageURL:
		target.FactoryImageURL = value
	case setting.WebsiteProfileKeyFactoryImageAlt:
		target.FactoryImageAlt = value
	case setting.WebsiteProfileKeyFactoryImageCaption:
		target.FactoryImageCaption = value
	case setting.WebsiteProfileKeyFactoryEyebrow:
		target.FactoryEyebrow = value
	case setting.WebsiteProfileKeyFactoryTitle:
		target.FactoryTitle = value
	case setting.WebsiteProfileKeyFactoryBody:
		target.FactoryBody = value
	case setting.WebsiteProfileKeyFactoryCTA:
		target.FactoryCTA = value
	case setting.WebsiteProfileKeyFactoryLink:
		target.FactoryLink = value
	}
}

func normalizeWebsiteProfileLocale(locale string) string {
	normalized := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(locale), "-", "_"))
	switch normalized {
	case "", "en":
		return "en"
	case "zh", "zh_cn", "zh_hans", "zh_sg":
		return "zh_cn"
	default:
		return normalized
	}
}

func websiteProfileFallbackLocales(locale string) []string {
	normalized := normalizeWebsiteProfileLocale(locale)
	locales := []string{normalized}
	if normalized != "en" {
		locales = append(locales, "en")
	}
	locales = append(locales, "global")
	return locales
}
