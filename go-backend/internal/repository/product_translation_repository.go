package repository

import (
	"fmt"
	"strings"

	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
)

type productTranslationRouteRecord struct {
	ID       uint
	ParentID *uint
	Locale   string
	Slug     string
}

type productTranslationGroupRecord struct {
	ID       uint
	ParentID *uint
	Locale   string
	Name     string
	Slug     string
	SKU      string
	Status   string
}

func (r *ProductRepository) FindTranslationParent(id uint) (*product.Product, error) {
	var parent product.Product
	if err := r.db.
		Select("id, parent_id, locale, slug").
		Where("id = ?", id).
		First(&parent).Error; err != nil {
		return nil, err
	}
	return &parent, nil
}

func (r *ProductRepository) TranslationLocaleExists(parentID uint, locale string, excludeID uint) (bool, error) {
	query := r.db.Model(&product.Product{}).
		Where("parent_id = ? AND locale = ?", parentID, locale)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProductRepository) ProductSlugExists(slug, locale string) (bool, error) {
	var count int64
	if err := r.db.Model(&product.Product{}).
		Where("slug = ? AND locale = ?", slug, locale).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProductRepository) FindTranslationGroupMembers(rootIDs []uint) (map[uint][]product.ProductTranslation, error) {
	result := make(map[uint][]product.ProductTranslation)
	if len(rootIDs) == 0 {
		return result, nil
	}

	var records []productTranslationGroupRecord
	if err := r.db.Model(&product.Product{}).
		Select("id, parent_id, locale, name, slug, sku, status").
		Where("id IN ? OR parent_id IN ?", rootIDs, rootIDs).
		Order("locale ASC, id ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}

	for _, record := range records {
		rootID := record.ID
		isRoot := record.ParentID == nil || *record.ParentID == 0
		if !isRoot {
			rootID = *record.ParentID
		}

		result[rootID] = append(result[rootID], product.ProductTranslation{
			ID:       record.ID,
			ParentID: record.ParentID,
			Locale:   strings.TrimSpace(record.Locale),
			Name:     strings.TrimSpace(record.Name),
			Slug:     strings.TrimSpace(record.Slug),
			SKU:      strings.TrimSpace(record.SKU),
			Status:   strings.TrimSpace(record.Status),
			IsRoot:   isRoot,
		})
	}

	return result, nil
}

func (r *ProductRepository) CreateTranslatedCopy(source, target *product.Product, variantSKUs map[uint]string) error {
	if source == nil || target == nil {
		return fmt.Errorf("source and target products are required")
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		variants := make([]product.ProductVariant, 0, len(source.Variants))
		for _, sourceVariant := range source.Variants {
			sku := strings.TrimSpace(variantSKUs[sourceVariant.ID])
			if sku == "" {
				return fmt.Errorf("missing translated SKU for variant %d", sourceVariant.ID)
			}

			clonedVariant := sourceVariant
			clonedVariant.ID = 0
			clonedVariant.ProductID = 0
			clonedVariant.DeletedAt = gorm.DeletedAt{}
			clonedVariant.SKU = sku
			variants = append(variants, clonedVariant)
		}

		if len(variants) > 0 {
			syncProductSummaryFromVariants(target, variants)
		}
		if err := tx.Create(target).Error; err != nil {
			return err
		}

		if len(source.SpecValues) > 0 {
			specValues := make([]product.ProductSpecValue, 0, len(source.SpecValues))
			for _, sourceSpec := range source.SpecValues {
				clonedSpec := sourceSpec
				clonedSpec.ID = 0
				clonedSpec.ProductID = target.ID
				clonedSpec.SpecDefinition = nil
				specValues = append(specValues, clonedSpec)
			}
			if err := tx.Create(&specValues).Error; err != nil {
				return err
			}
		}

		optionValueIDs := make(map[uint]uint, len(source.VariantOptionValues))
		if len(source.VariantOptionValues) > 0 {
			optionValues := make([]product.ProductVariantOptionValue, 0, len(source.VariantOptionValues))
			for _, sourceOptionValue := range source.VariantOptionValues {
				clonedOptionValue := sourceOptionValue
				clonedOptionValue.ID = 0
				clonedOptionValue.ProductID = target.ID
				optionValues = append(optionValues, clonedOptionValue)
			}
			if err := tx.Create(&optionValues).Error; err != nil {
				return err
			}
			for index, sourceOptionValue := range source.VariantOptionValues {
				optionValueIDs[sourceOptionValue.ID] = optionValues[index].ID
			}
		}

		variantIDs := make(map[uint]uint, len(source.Variants))
		if len(variants) > 0 {
			for index := range variants {
				variants[index].ProductID = target.ID
			}
			if err := tx.Create(&variants).Error; err != nil {
				return err
			}
			for index, sourceVariant := range source.Variants {
				variantIDs[sourceVariant.ID] = variants[index].ID
			}
		}

		if len(source.Media) > 0 {
			mediaItems := make([]product.ProductMedia, 0, len(source.Media))
			for _, sourceMedia := range source.Media {
				clonedMedia := sourceMedia
				clonedMedia.ID = 0
				clonedMedia.ProductID = target.ID
				clonedMedia.DeletedAt = gorm.DeletedAt{}

				if sourceMedia.VariantID != nil {
					if translatedVariantID, ok := variantIDs[*sourceMedia.VariantID]; ok {
						clonedMedia.VariantID = &translatedVariantID
					} else {
						clonedMedia.VariantID = nil
					}
				}
				if sourceMedia.VariantOptionValueID != nil {
					if translatedOptionValueID, ok := optionValueIDs[*sourceMedia.VariantOptionValueID]; ok {
						clonedMedia.VariantOptionValueID = &translatedOptionValueID
					} else {
						clonedMedia.VariantOptionValueID = nil
					}
				}
				mediaItems = append(mediaItems, clonedMedia)
			}

			if err := ensureProductMediaReferencesBelongToProduct(tx, target.ID, mediaItems); err != nil {
				return err
			}
			if err := tx.Create(&mediaItems).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// FindPublicTranslationRoutes returns active localized routes in the current
// product's translation group. Product translations are direct children of a
// root product; the service layer must enforce that relationship for writes.
func (r *ProductRepository) FindPublicTranslationRoutes(productID uint) ([]product.ProductTranslationRoute, error) {
	var current productTranslationRouteRecord
	if err := r.db.Model(&product.Product{}).
		Select("id, parent_id").
		Where("id = ?", productID).
		First(&current).Error; err != nil {
		return nil, err
	}

	rootID := current.ID
	if current.ParentID != nil && *current.ParentID > 0 {
		rootID = *current.ParentID
	}

	var records []productTranslationRouteRecord
	if err := r.db.Model(&product.Product{}).
		Select("id, locale, slug").
		Where("status = ?", "active").
		Where("(id = ? OR parent_id = ?)", rootID, rootID).
		Order("id ASC").
		Find(&records).Error; err != nil {
		return nil, err
	}

	routes := make([]product.ProductTranslationRoute, 0, len(records))
	seenLocales := make(map[string]struct{}, len(records))
	for _, record := range records {
		locale := strings.TrimSpace(record.Locale)
		slug := strings.TrimSpace(record.Slug)
		if locale == "" || slug == "" {
			continue
		}
		if _, exists := seenLocales[locale]; exists {
			continue
		}
		seenLocales[locale] = struct{}{}
		routes = append(routes, product.ProductTranslationRoute{
			Locale: locale,
			Slug:   slug,
		})
	}

	return routes, nil
}
