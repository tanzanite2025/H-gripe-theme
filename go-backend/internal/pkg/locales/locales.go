package locales

import (
	"strings"
)

type Language struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
	Enabled    bool   `json:"enabled"`
}

var SupportedLanguages = []Language{
	{Code: "en", Name: "English", NativeName: "English", Enabled: true},
	{Code: "fr", Name: "French", NativeName: "Français", Enabled: true},
	{Code: "de", Name: "German", NativeName: "Deutsch", Enabled: true},
	{Code: "es", Name: "Spanish", NativeName: "Español", Enabled: true},
	{Code: "ja", Name: "Japanese", NativeName: "日本語", Enabled: true},
	{Code: "ko", Name: "Korean", NativeName: "한국어", Enabled: true},
	{Code: "it", Name: "Italian", NativeName: "Italiano", Enabled: true},
	{Code: "pt", Name: "Portuguese", NativeName: "Português", Enabled: true},
	{Code: "ru", Name: "Russian", NativeName: "Русский", Enabled: true},
	{Code: "ar", Name: "Arabic", NativeName: "العربية", Enabled: true},
	{Code: "fi", Name: "Finnish", NativeName: "Suomi", Enabled: true},
	{Code: "da", Name: "Danish", NativeName: "Dansk", Enabled: true},
	{Code: "th", Name: "Thai", NativeName: "ไทย", Enabled: true},
	{Code: "sv", Name: "Swedish", NativeName: "Svenska", Enabled: true},
	{Code: "id", Name: "Indonesian", NativeName: "Bahasa Indonesia", Enabled: true},
	{Code: "ms", Name: "Malay", NativeName: "Bahasa Melayu", Enabled: true},
	{Code: "be", Name: "Belarusian", NativeName: "Беларуская", Enabled: true},
	{Code: "tr", Name: "Turkish", NativeName: "Türkçe", Enabled: true},
	{Code: "bn", Name: "Bengali", NativeName: "বাংলা", Enabled: true},
	{Code: "fa", Name: "Persian", NativeName: "فارسی", Enabled: true},
	{Code: "nl", Name: "Dutch", NativeName: "Nederlands", Enabled: true},
	{Code: "hi", Name: "Hindi", NativeName: "हिन्दी", Enabled: true},
	{Code: "ur", Name: "Urdu", NativeName: "اردو", Enabled: true},
	{Code: "mr", Name: "Marathi", NativeName: "मराठी", Enabled: true},
	{Code: "pcm", Name: "Nigerian Pidgin", NativeName: "Nigerian Pidgin", Enabled: true},
	{Code: "fil", Name: "Filipino", NativeName: "Filipino", Enabled: true},
	{Code: "te", Name: "Telugu", NativeName: "తెలుగు", Enabled: true},
	{Code: "ha", Name: "Hausa", NativeName: "Hausa", Enabled: true},
	{Code: "ps", Name: "Pashto", NativeName: "پښتو", Enabled: true},
	{Code: "sw", Name: "Swahili", NativeName: "Kiswahili", Enabled: true},
	{Code: "tl", Name: "Tagalog", NativeName: "Tagalog", Enabled: true},
	{Code: "ta", Name: "Tamil", NativeName: "தமிழ்", Enabled: true},
	{Code: "jv", Name: "Javanese", NativeName: "Basa Jawa", Enabled: true},
	{Code: "zh_cn", Name: "Chinese (Simplified)", NativeName: "简体中文", Enabled: true},
}

var supportedLocaleSet = buildSupportedLocaleSet()

func SupportedLocaleCodes() []string {
	codes := make([]string, 0, len(SupportedLanguages))
	for _, language := range SupportedLanguages {
		codes = append(codes, language.Code)
	}
	return codes
}

func EnabledLocaleCodes() []string {
	codes := make([]string, 0, len(SupportedLanguages))
	for _, language := range SupportedLanguages {
		if language.Enabled {
			codes = append(codes, language.Code)
		}
	}
	return codes
}

func Normalize(locale string) string {
	cleaned := clean(locale)
	if cleaned == "" {
		return "en"
	}
	return normalizeCleaned(cleaned)
}

func ResolveSupported(locale string) string {
	cleaned := clean(locale)
	if cleaned == "" {
		return ""
	}

	normalized := normalizeCleaned(cleaned)
	if IsSupportedCode(normalized) {
		return normalized
	}
	return ""
}

func IsSupported(locale string) bool {
	return ResolveSupported(locale) != ""
}

func IsSupportedCode(code string) bool {
	return supportedLocaleSet[code]
}

func buildSupportedLocaleSet() map[string]bool {
	result := make(map[string]bool, len(SupportedLanguages))
	for _, language := range SupportedLanguages {
		if language.Enabled {
			result[language.Code] = true
		}
	}
	return result
}

func clean(locale string) string {
	value := strings.TrimSpace(locale)
	if first, _, ok := strings.Cut(value, ","); ok {
		value = first
	}
	if first, _, ok := strings.Cut(value, ";"); ok {
		value = first
	}
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")

	var result strings.Builder
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func normalizeCleaned(locale string) string {
	switch locale {
	case "zh", "zh-cn", "zh-hans", "zh-sg":
		return "zh_cn"
	}

	if base, _, ok := strings.Cut(locale, "-"); ok && base != "" && base != "zh" {
		return base
	}

	return locale
}
