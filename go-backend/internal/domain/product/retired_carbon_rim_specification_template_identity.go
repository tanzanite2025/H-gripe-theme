package product

import "strings"

// Historical rows remain readable, but these identities cannot define new products.
func IsRetiredCarbonRimSpecificationTemplateSlug(slug string) bool {
	switch strings.ToLower(strings.TrimSpace(slug)) {
	case "rim", "carbon_rim", "carbon-rim":
		return true
	default:
		return false
	}
}
