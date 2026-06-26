package i18n

import (
	"strings"
	"tanzanite/internal/pkg/config"
)

var supportedLocales map[string]bool

// Init 初始化国际化支持
func Init(cfg config.I18nConfig) {
	supportedLocales = make(map[string]bool)
	for _, locale := range cfg.SupportedLocales {
		supportedLocales[locale] = true
	}
}

// IsValidLocale 检查语言代码是否有效
func IsValidLocale(locale string) bool {
	return supportedLocales[locale]
}

// NormalizeLocale 规范化语言代码
func NormalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))

	// 移除非字母字符
	var result strings.Builder
	for _, r := range locale {
		if (r >= 'a' && r <= 'z') || r == '-' {
			result.WriteRune(r)
		}
	}

	normalized := result.String()
	if normalized == "" {
		return "en"
	}

	return normalized
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
