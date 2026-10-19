package main

import (
	"errors"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
)

func importSettingsWithAtomicRegisteredMediaValidation(settings *repository.SettingRepository, input []SettingImport) (map[string]int, error) {
	stats := map[string]int{"total": len(input), "created": 0, "updated": 0, "skipped": 0}
	records := make([]setting.Setting, 0, len(input))
	for _, item := range input {
		locale := item.Locale
		if locale == "" {
			locale = "en"
		}
		existing, err := settings.Get(item.Key, locale)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			stats["created"]++
		case existing.Value == item.Value:
			stats["skipped"]++
		default:
			stats["updated"]++
		}
		records = append(records, setting.Setting{Key: item.Key, Value: item.Value, Type: item.Type, Group: item.Group, Locale: locale, IsPublic: item.IsPublic, Description: item.Description})
	}
	if err := settings.SetMultipleValuesPreservingExistingVisibilityAndDescription(records); err != nil {
		return nil, err
	}
	// Statistics describe the input's value changes; repository locks and identity
	// checks, not this pre-read, decide the write's correctness.
	return stats, nil
}
