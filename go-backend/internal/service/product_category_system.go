package service

import (
	"fmt"
	"strings"

	"commerce-platform/internal/domain/product"
)

const SystemProductCategoryWheelsetSlug = "wheelset"
const SystemProductCategoryWheelComponentsSlug = "wheel-components"
const SystemProductCategoryTireSlug = "tire"

type systemProductCategoryDefinition struct {
	Slug   string
	Parent string
	Depth  int
}

var systemProductCategoryDefinitions = map[string]systemProductCategoryDefinition{
	SystemProductCategoryWheelsetSlug: {
		Slug: SystemProductCategoryWheelsetSlug,
	},
	SystemProductCategoryWheelComponentsSlug: {
		Slug:  SystemProductCategoryWheelComponentsSlug,
		Depth: 1,
	},
	SystemProductCategoryTireSlug: {
		Slug:   SystemProductCategoryTireSlug,
		Parent: SystemProductCategoryWheelComponentsSlug,
		Depth:  2,
	},
}

func isSystemProductCategorySlug(slug string) bool {
	_, ok := systemProductCategoryDefinitions[strings.ToLower(strings.TrimSpace(slug))]
	return ok
}

func systemProductCategoryDefinitionForSlug(slug string) (systemProductCategoryDefinition, bool) {
	definition, ok := systemProductCategoryDefinitions[strings.ToLower(strings.TrimSpace(slug))]
	return definition, ok
}

func systemProductCategoryParentID(categories []product.ProductCategory, parentSlug string) (uint, error) {
	normalizedSlug := strings.ToLower(strings.TrimSpace(parentSlug))
	for _, category := range categories {
		if strings.ToLower(strings.TrimSpace(category.Slug)) == normalizedSlug {
			return category.ID, nil
		}
	}
	return 0, fmt.Errorf("%w: parent system category %s not found", ErrProductCategorySystemProtected, parentSlug)
}
