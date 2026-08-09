package service

import (
	"fmt"
	"strings"
	"tanzanite/internal/pkg/locales"
)

func normalizeLocale(locale string) string {
	return locales.Normalize(locale)
}

func requireSupportedLocale(locale string) (string, error) {
	normalized := locales.ResolveSupported(locale)
	if normalized != "" {
		return normalized, nil
	}

	if strings.TrimSpace(locale) == "" {
		return "", fmt.Errorf("%w: locale is required", ErrUnsupportedLocale)
	}
	return "", fmt.Errorf("%w: %s", ErrUnsupportedLocale, strings.TrimSpace(locale))
}

func optionalSupportedLocale(locale string) (string, error) {
	if strings.TrimSpace(locale) == "" {
		return "", nil
	}
	return requireSupportedLocale(locale)
}
