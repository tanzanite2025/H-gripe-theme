package service

import "fmt"

func settingValueCacheKey(key, locale string) string {
	return fmt.Sprintf("setting:%s:%s", key, locale)
}

func settingsAllCacheKey(locale string) string {
	return fmt.Sprintf("settings:all:%s", locale)
}

func settingsPublicCacheKey(locale string) string {
	return fmt.Sprintf("settings:public:%s", locale)
}

func settingsGroupCacheKey(group, locale string) string {
	return fmt.Sprintf("settings:group:%s:%s", group, locale)
}

func settingsStructuredCacheKey(group, locale string) string {
	return fmt.Sprintf("settings:%s:%s", group, locale)
}

func settingsGroupsCacheKey() string {
	return "settings:groups"
}

func (s *SettingService) invalidateSettingCaches(key, group, locale string) {
	cacheKeys := []string{
		settingValueCacheKey(key, locale),
		settingsAllCacheKey(locale),
		settingsPublicCacheKey(locale),
	}

	if group != "" {
		cacheKeys = append(cacheKeys,
			settingsGroupCacheKey(group, locale),
			settingsStructuredCacheKey(group, locale),
		)
	}

	for _, cacheKey := range cacheKeys {
		_ = s.cache.Delete(cacheKey)
	}
}
