package product

import "strings"

func IsRetiredWheelsetSpecificationTemplateSlug(slug string) bool {
	return strings.EqualFold(strings.TrimSpace(slug), ProductTypeWheelset)
}
