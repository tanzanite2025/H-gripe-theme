package service

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/wheelsetcatalog"

	"gorm.io/gorm"
)

func spokeRepairKitModelKeys(models []product.SpokeRepairKitModel) []string {
	keys := make([]string, 0, len(models))
	for _, model := range models {
		keys = append(keys, model.BuildSpokeRepairKitModelCompatibilityValueKey())
	}
	return keys
}

// buildSpokeRepairKitModels resolves admin-selected catalog keys into the
// product-owned compatibility snapshot. The product service never persists a
// buyer-provided label or an unverified catalog record.
func (s *ProductService) buildSpokeRepairKitModels(categoryID *uint, keys []string) ([]product.SpokeRepairKitModel, error) {
	if categoryID == nil || *categoryID == 0 {
		if len(keys) > 0 {
			return nil, fmt.Errorf("%w: wheelset models require the spoke-repair-kits category", ErrSpokeRepairKitModelsInvalid)
		}
		return nil, nil
	}
	if s.productCategoryRepo == nil {
		return nil, fmt.Errorf("%w: category repository is not configured", ErrSpokeRepairKitModelsInvalid)
	}
	category, err := s.productCategoryRepo.FindByID(*categoryID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductCategoryNotFound
		}
		return nil, err
	}
	if category.Slug != product.SpokeRepairKitProductCategorySlug {
		if len(keys) > 0 {
			return nil, fmt.Errorf("%w: wheelset models require the spoke-repair-kits category", ErrSpokeRepairKitModelsInvalid)
		}
		return nil, nil
	}
	seen := make(map[string]struct{}, len(keys))
	models := make([]product.SpokeRepairKitModel, 0, len(keys))
	for index, rawKey := range keys {
		key := strings.ToLower(strings.TrimSpace(rawKey))
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		model, found, lookupErr := wheelsetcatalog.FindWheelsetSpokeCatalogModelByValueKey(key)
		if lookupErr != nil {
			return nil, fmt.Errorf("%w: failed to load wheelset catalog: %v", ErrSpokeRepairKitModelsInvalid, lookupErr)
		}
		if !found {
			return nil, fmt.Errorf("%w: wheelset model %q is not in the verified catalog", ErrSpokeRepairKitModelsInvalid, key)
		}
		sourceCheckedAt := ""
		if model.SourceCheckedAt != nil {
			sourceCheckedAt = strings.TrimSpace(*model.SourceCheckedAt)
		}
		models = append(models, product.SpokeRepairKitModel{
			BrandSlug:         model.BrandSlug,
			BrandName:         model.BrandName,
			WheelsetModelSlug: model.Slug,
			WheelsetModelName: model.Model,
			LifecycleStatus:   model.LifecycleStatus,
			SourceCheckedAt:   sourceCheckedAt,
			SortOrder:         index * 10,
		})
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("%w: select at least one wheelset model", ErrSpokeRepairKitModelsInvalid)
	}
	return models, nil
}
