package service

import (
	"strings"

	"commerce-platform/internal/domain/setting"
)

func loadPublicManagedSettingValues(settings *SettingService, group, locale string) (map[string]string, error) {
	values := make(map[string]string)
	seenKeys := make(map[string]struct{})

	for _, candidateLocale := range publicSettingFallbackLocales(locale) {
		records, err := settings.GetPublicByGroup(group, candidateLocale)
		if err != nil {
			return nil, err
		}
		for _, record := range records {
			if _, exists := seenKeys[record.Key]; exists {
				continue
			}
			seenKeys[record.Key] = struct{}{}
			values[record.Key] = record.Value
		}
	}

	return values, nil
}

func normalizeManagedSettingLocale(locale string) string {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return "en"
	}
	return strings.ToLower(strings.ReplaceAll(locale, "-", "_"))
}

func managedSettingValue(value string) string {
	return strings.TrimSpace(value)
}

func managedSettingRecords(group, locale string, values map[string]string, descriptions map[string]string) []setting.Setting {
	records := make([]setting.Setting, 0, len(values))
	for key, value := range values {
		records = append(records, setting.Setting{
			Key:         key,
			Value:       value,
			Type:        "string",
			Locale:      locale,
			Group:       group,
			IsPublic:    true,
			Description: descriptions[key],
		})
	}
	return records
}
