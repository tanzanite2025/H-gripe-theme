package product

import (
	"strings"

	"commerce-platform/internal/pkg/locales"
)

// LocalizedDisplayName keeps public product categories and category breadcrumbs
// on the same translation rule while preserving the category's durable identity.
func (category ProductCategory) LocalizedDisplayName(locale string) string {
	normalizedLocale := locales.Normalize(locale)
	for _, translation := range category.Translations {
		if locales.Normalize(translation.Locale) == normalizedLocale {
			if name := strings.TrimSpace(translation.Name); name != "" {
				return name
			}
			break
		}
	}
	return category.Name
}
