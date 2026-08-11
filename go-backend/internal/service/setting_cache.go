package service

import "fmt"

const settingCacheVersion = "v5"

func settingValueCacheKey(key, locale string) string {
	return fmt.Sprintf("setting:%s:%s:%s", settingCacheVersion, key, locale)
}

func settingPublicValueCacheKey(key, locale string) string {
	return fmt.Sprintf("setting:%s:public-value:%s:%s", settingCacheVersion, key, locale)
}

func settingsAllCacheKey(locale string) string {
	return fmt.Sprintf("settings:%s:all:%s", settingCacheVersion, locale)
}

func settingsPublicCacheKey(locale string) string {
	return fmt.Sprintf("settings:%s:public:%s", settingCacheVersion, locale)
}

func settingsGroupCacheKey(group, locale string) string {
	return fmt.Sprintf("settings:%s:group:%s:%s", settingCacheVersion, group, locale)
}

func settingsPublicGroupCacheKey(group, locale string) string {
	return fmt.Sprintf("settings:%s:public-group:%s:%s", settingCacheVersion, group, locale)
}

func settingsStructuredCacheKey(group, locale string) string {
	return fmt.Sprintf("settings:%s:structured:%s:%s", settingCacheVersion, group, locale)
}

func settingsGroupsCacheKey() string {
	return fmt.Sprintf("settings:%s:groups", settingCacheVersion)
}

func settingsPublicGroupsCacheKey() string {
	return fmt.Sprintf("settings:%s:public-groups", settingCacheVersion)
}

func (s *SettingService) invalidateSettingCaches(key, group, locale string) {
	cacheKeys := []string{
		settingValueCacheKey(key, locale),
		settingPublicValueCacheKey(key, locale),
		settingsAllCacheKey(locale),
		settingsPublicCacheKey(locale),
		settingsGroupsCacheKey(),
		settingsPublicGroupsCacheKey(),
	}

	if group != "" {
		cacheKeys = append(cacheKeys,
			settingsGroupCacheKey(group, locale),
			settingsPublicGroupCacheKey(group, locale),
			settingsStructuredCacheKey(group, locale),
		)
		if group == "site" && s.cache != nil {
			_ = s.cache.DeletePattern(fmt.Sprintf("settings:%s:structured:site:*", settingCacheVersion))
		}
	}

	for _, cacheKey := range cacheKeys {
		if s.cache != nil {
			_ = s.cache.Delete(cacheKey)
		}
	}
}
