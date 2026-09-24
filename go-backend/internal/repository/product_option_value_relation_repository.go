package repository

import (
	"commerce-platform/internal/domain/product"
	"fmt"
	"sort"

	"gorm.io/gorm"
)

func replaceProductOptionValueRelations(tx *gorm.DB, productID uint, relations []product.ProductOptionValueRelation) error {
	normalized, err := normalizeProductOptionValueRelations(tx, productID, relations)
	if err != nil {
		return err
	}

	if err := tx.Where("product_id = ?", productID).Delete(&product.ProductOptionValueRelation{}).Error; err != nil {
		return err
	}
	if len(normalized) == 0 {
		return nil
	}
	return tx.Create(&normalized).Error
}

func normalizeProductOptionValueRelations(tx *gorm.DB, productID uint, relations []product.ProductOptionValueRelation) ([]product.ProductOptionValueRelation, error) {
	if len(relations) == 0 {
		return nil, nil
	}

	ids := make([]uint, 0, len(relations)*2)
	requestedIDs := make(map[uint]struct{}, len(relations)*2)
	for _, relation := range relations {
		if !product.IsValidOptionValueRelationType(relation.RelationType) {
			return nil, fmt.Errorf("%w: unsupported relation type %q", ErrProductOptionValueRelationInvalid, relation.RelationType)
		}
		if relation.SourceOptionValueID == 0 || relation.TargetOptionValueID == 0 {
			return nil, fmt.Errorf("%w: source and target option value IDs are required", ErrProductOptionValueRelationInvalid)
		}
		if relation.SourceOptionValueID == relation.TargetOptionValueID {
			return nil, fmt.Errorf("%w: option value %d cannot relate to itself", ErrProductOptionValueRelationInvalid, relation.SourceOptionValueID)
		}
		requestedIDs[relation.SourceOptionValueID] = struct{}{}
		requestedIDs[relation.TargetOptionValueID] = struct{}{}
	}
	for id := range requestedIDs {
		ids = append(ids, id)
	}

	var ownedCount int64
	if err := tx.Model(&product.ProductVariantOptionValue{}).
		Where("product_id = ? AND id IN ?", productID, ids).
		Count(&ownedCount).Error; err != nil {
		return nil, err
	}
	if ownedCount != int64(len(ids)) {
		return nil, fmt.Errorf("%w: every source and target must belong to product %d", ErrProductOptionValueRelationInvalid, productID)
	}

	normalized := make([]product.ProductOptionValueRelation, 0, len(relations))
	seen := make(map[string]struct{}, len(relations))
	for _, relation := range relations {
		sourceID := relation.SourceOptionValueID
		targetID := relation.TargetOptionValueID
		if relation.RelationType == product.OptionValueRelationConflicts && sourceID > targetID {
			sourceID, targetID = targetID, sourceID
		}
		key := fmt.Sprintf("%s:%d:%d", relation.RelationType, sourceID, targetID)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, product.ProductOptionValueRelation{
			ProductID:           productID,
			SourceOptionValueID: sourceID,
			TargetOptionValueID: targetID,
			RelationType:        relation.RelationType,
		})
	}
	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].RelationType != normalized[j].RelationType {
			return normalized[i].RelationType < normalized[j].RelationType
		}
		if normalized[i].SourceOptionValueID != normalized[j].SourceOptionValueID {
			return normalized[i].SourceOptionValueID < normalized[j].SourceOptionValueID
		}
		return normalized[i].TargetOptionValueID < normalized[j].TargetOptionValueID
	})
	return normalized, nil
}
