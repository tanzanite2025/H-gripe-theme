package service

import (
	"commerce-platform/internal/pkg/locales"
	"fmt"
	"strings"
)

// resolveAutoReplyLocale returns the canonical locale used by automatic
// replies. An empty result means that the request must not trigger a reply.
func resolveAutoReplyLocale(value string) string {
	return locales.ResolveSupported(value)
}

func requireAutoReplyLocale(value string) (string, error) {
	normalized := resolveAutoReplyLocale(value)
	if normalized != "" {
		return normalized, nil
	}

	if strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("%w: locale is required", ErrUnsupportedLocale)
	}
	return "", fmt.Errorf("%w: %s", ErrUnsupportedLocale, strings.TrimSpace(value))
}
