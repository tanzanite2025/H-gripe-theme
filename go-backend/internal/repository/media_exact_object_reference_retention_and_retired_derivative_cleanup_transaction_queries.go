package repository

import (
	"fmt"

	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

// Exact generation cleanup must not expand to its parent's original or other
// generations. ID-only consumers resolve the current generation dynamically.
func (r *MediaRepository) FindExactMediaObjectReferencesBeforePhysicalCleanup(key, reference string, excludedRetiredDerivativeID uint) ([]media.AssetReference, error) {
	key, valid := storage.NormalizeObjectKey(key)
	if !valid {
		return nil, fmt.Errorf("exact media cleanup object identity is invalid")
	}
	query := mediaAssetReferenceQuery{
		ObjectKeys: []string{key}, URLs: []string{reference, "/uploads/" + key, "uploads/" + key},
		IncludeRetainedDeletedPlacements: true, IncludeAllExactObjectRegistrations: true,
		ExcludedRetiredDerivativeID: excludedRetiredDerivativeID,
	}
	var references []media.AssetReference
	for _, scan := range r.assetReferenceScanners() {
		items, err := scan(query)
		if err != nil {
			return nil, err
		}
		references = append(references, items...)
	}
	return references, nil
}

func (r *MediaRepository) LockExactMediaObjectBeforePhysicalCleanupWithinTransaction(key string) error {
	return r.lockMediaObjectKeysAndPrivateDerivativeNamespacesWithinTransaction([]string{key}, nil)
}

func (r *MediaRepository) HasPublishedBusinessConsumerForExactMediaObject(key, reference string) (bool, error) {
	query := mediaAssetReferenceQuery{ObjectKeys: []string{key}, URLs: []string{reference, "/uploads/" + key, "uploads/" + key}, IncludeRetainedDeletedPlacements: true}
	for _, scan := range []mediaAssetReferenceScanner{
		r.productCategoryImageReferences, r.productBrandLogoReferences, r.productMediaReferences,
		r.productRichTextMediaAttributeReferences, r.productAndTemplateOptionSwatchReferences,
		r.galleryReferences, r.faqReferences, r.postReferences, r.websiteProfileSettingMediaReferences, r.workbenchFeedImageReferences,
	} {
		items, err := scan(query)
		if err != nil {
			return false, err
		}
		published, err := r.dedicatedPublicObjectReferencesIncludePublishedConsumer(items)
		if err != nil || published {
			return published, err
		}
	}
	return false, nil
}

func (r *MediaRepository) FindMediaDerivativeIncludingRetiredGeneration(id uint, forUpdate bool) (*media.MediaAssetDerivative, error) {
	query := r.db.Unscoped()
	if forUpdate {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var derivative media.MediaAssetDerivative
	if err := query.First(&derivative, id).Error; err != nil {
		return nil, err
	}
	return &derivative, nil
}
