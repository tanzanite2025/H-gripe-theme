package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
)

var (
	ErrGalleryMediaReferenceInvalid = errors.New("gallery media reference is invalid")
	ErrGalleryMediaAssetNotFound    = errors.New("gallery image media asset not found")
)

type GalleryMediaFieldMutation struct{ UpdateCoverImage bool }
type GalleryImageMediaFieldMutation struct{ UpdateOriginalImage, UpdateThumbnail bool }

func (r *GalleryRepository) ConfigurePublicBusinessMediaReferenceIdentityResolver(resolver PublicBusinessMediaReferenceObjectKeyResolver) {
	r.businessMediaReferenceResolver = resolver
}

func appendGalleryImageMediaBindingsForChangedPlacements(bindings []publicBusinessMediaReferenceBinding, image, previous *gallery.GalleryImage) ([]publicBusinessMediaReferenceBinding, error) {
	primaryChanged := image.URL != previous.URL || !sameGalleryMediaAssetID(image.MediaAssetID, previous.MediaAssetID) || image.ID == 0
	if primaryChanged {
		if image.URL == "" && (image.MediaAssetID == nil || *image.MediaAssetID == 0) {
			return nil, ErrGalleryMediaReferenceInvalid
		}
		bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &image.URL, AssetID: &image.MediaAssetID, ExpectedMediaType: "image"})
	}
	if image.Thumbnail != previous.Thumbnail || image.ID == 0 {
		bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &image.Thumbnail, ExpectedMediaType: "image"})
	}
	return bindings, nil
}

func sameGalleryMediaAssetID(first, second *uint) bool {
	return first == nil && second == nil || first != nil && second != nil && *first == *second
}

func (r *GalleryRepository) validateAndCanonicalizeGalleryMediaBindingsWithOneSortedAssetLockPlan(tx *gorm.DB, bindings []publicBusinessMediaReferenceBinding) error {
	if len(bindings) == 0 {
		return nil
	}
	references := make([]string, 0, len(bindings))
	var ids []uint
	for _, binding := range bindings {
		references = append(references, *binding.URL)
		if binding.AssetID != nil && *binding.AssetID != nil {
			ids = append(ids, **binding.AssetID)
		}
	}
	// Preserve the existing missing-selection API error. Eligibility and identity
	// are still rechecked under the common ascending locks below.
	if len(ids) > 0 {
		var existing []uint
		if err := tx.Model(&media.MediaAsset{}).Where("id IN ?", ids).Pluck("id", &existing).Error; err != nil {
			return err
		}
		found := make(map[uint]bool, len(existing))
		for _, id := range existing {
			found[id] = true
		}
		for _, id := range ids {
			if id == 0 || !found[id] {
				return ErrGalleryMediaAssetNotFound
			}
		}
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, r.businessMediaReferenceResolver, references, ids)
	if err != nil {
		return mapPublicBusinessMediaValidationErrorToGalleryMediaValidationError(err)
	}
	for _, binding := range bindings {
		if err := resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(binding); err != nil {
			return mapPublicBusinessMediaValidationErrorToGalleryMediaValidationError(err)
		}
	}
	return nil
}

func mapPublicBusinessMediaValidationErrorToGalleryMediaValidationError(err error) error {
	if errors.Is(err, ErrProductMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrGalleryMediaReferenceInvalid, err)
	}
	return err
}
