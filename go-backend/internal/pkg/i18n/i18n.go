package i18n

import (
	"strings"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/pkg/locales"
)

var supportedLocales map[string]bool

// Init 初始化国际化支持
func Init(cfg config.I18nConfig) {
	supportedLocales = make(map[string]bool)
	for _, locale := range cfg.SupportedLocales {
		if normalized := locales.ResolveSupported(locale); normalized != "" {
			supportedLocales[normalized] = true
		}
	}
}

// IsValidLocale 检查语言代码是否有效
func IsValidLocale(locale string) bool {
	if supportedLocales == nil {
		return locales.IsSupported(locale)
	}
	normalized := locales.ResolveSupported(locale)
	if normalized == "" {
		return false
	}
	return supportedLocales[normalized]
}

// NormalizeLocale 规范化语言代码
func NormalizeLocale(locale string) string {
	return locales.Normalize(locale)
}

// GetLocaleFromPath 从URL路径提取语言代码
// 例如: /fr/products -> fr
func GetLocaleFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 0 {
		locale := NormalizeLocale(parts[0])
		if IsValidLocale(locale) {
			return locale
		}
	}
	return "en"
}
