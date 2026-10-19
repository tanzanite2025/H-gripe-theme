package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/pkg/storage"
)

type galleryTransactionalSaveOptions struct {
	CreateNewGallery                   bool
	ProductIDs                         *[]uint
	NewImages                          []gallery.GalleryImage
	MediaMutation                      *GalleryMediaFieldMutation
	SkipPreviouslyPlacedImportedImages bool
}

// SaveImportedGalleryAndAppendPreviouslyUnplacedImages applies one imported
// gallery atomically. Existing image metadata and product links are not imported.
func (r *GalleryRepository) SaveImportedGalleryAndAppendPreviouslyUnplacedImages(item *gallery.Gallery, images []gallery.GalleryImage) (int, int, error) {
	return r.saveGalleryWithOptionalProductLinksAndAppendedImagesInOneTransaction(item, galleryTransactionalSaveOptions{CreateNewGallery: item.ID == 0, NewImages: images, SkipPreviouslyPlacedImportedImages: true})
}

func (r *GalleryRepository) saveGalleryWithOptionalProductLinksAndAppendedImagesInOneTransaction(item *gallery.Gallery, options galleryTransactionalSaveOptions) (int, int, error) {
	added, skipped := 0, 0
	err := r.db.Transaction(func(tx *gorm.DB) error {
		previous := gallery.Gallery{}
		if !options.CreateNewGallery {
			if err := lockActiveGalleryParentWithinMutation(tx, item.ID, &previous); err != nil {
				return err
			}
			if options.MediaMutation != nil && !options.MediaMutation.UpdateCoverImage {
				item.CoverImage = previous.CoverImage
			}
		}
		var bindings []publicBusinessMediaReferenceBinding
		if item.CoverImage != previous.CoverImage {
			bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &item.CoverImage, ExpectedMediaType: "image"})
		}
		var previousImages []gallery.GalleryImage
		if options.SkipPreviouslyPlacedImportedImages && !options.CreateNewGallery {
			// Include tombstones: rerunning an import must not resurrect a removed image.
			if err := tx.Unscoped().Where("gallery_id = ?", item.ID).Find(&previousImages).Error; err != nil {
				return err
			}
			if err := r.resolveExistingGalleryImageIdentitiesForImportDeduplicationWithoutRewritingRecords(tx, previousImages); err != nil {
				return err
			}
		}
		images := make([]gallery.GalleryImage, 0, len(options.NewImages))
		for _, candidate := range options.NewImages {
			if candidate.ID != 0 {
				return gorm.ErrInvalidData
			}
			if options.SkipPreviouslyPlacedImportedImages && galleryImagePrimaryIdentityAlreadyPlaced(candidate, previousImages) {
				skipped++
				continue
			}
			images = append(images, candidate)
		}
		for index := range images {
			var err error
			bindings, err = appendGalleryImageMediaBindingsForChangedPlacements(bindings, &images[index], &gallery.GalleryImage{})
			if err != nil {
				return err
			}
		}
		if err := r.validateAndCanonicalizeGalleryMediaBindingsWithOneSortedAssetLockPlan(tx, bindings); err != nil {
			return err
		}
		if options.CreateNewGallery {
			if err := tx.Omit(clause.Associations).Create(item).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&gallery.Gallery{}).Where("id = ?", item.ID).Select("*").Omit(clause.Associations, "id", "created_at", "deleted_at").Updates(item).Error; err != nil {
				return err
			}
		}
		if options.ProductIDs != nil {
			if err := r.replaceGalleryProductLinksTx(tx, item.ID, *options.ProductIDs); err != nil {
				return err
			}
		}
		accepted := make([]gallery.GalleryImage, 0, len(images))
		for _, image := range images {
			if options.SkipPreviouslyPlacedImportedImages && (galleryImagePrimaryIdentityAlreadyPlaced(image, previousImages) || galleryImagePrimaryIdentityAlreadyPlaced(image, accepted)) {
				skipped++
				continue
			}
			image.GalleryID = item.ID
			if image.Thumbnail == "" {
				image.Thumbnail = image.URL
			}
			accepted = append(accepted, image)
		}
		if len(accepted) > 0 {
			if err := tx.Omit(clause.Associations).Create(&accepted).Error; err != nil {
				return err
			}
		}
		added = len(accepted)
		if options.CreateNewGallery {
			item.Images = accepted
		}
		return nil
	})
	if err != nil {
		return 0, 0, err
	}
	return added, skipped, nil
}

func (r *GalleryRepository) resolveExistingGalleryImageIdentitiesForImportDeduplicationWithoutRewritingRecords(tx *gorm.DB, images []gallery.GalleryImage) error {
	repo := NewMediaRepository(tx)
	hasAssets := repo.hasTable("media_assets")
	for index := range images {
		if images[index].MediaAssetID != nil {
			continue
		}
		reference := images[index].URL
		key, err := storage.TrustedBusinessMediaReferenceObjectKey(reference, nil, "")
		if err == nil && r.businessMediaReferenceResolver != nil {
			key, err = r.businessMediaReferenceResolver.ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference)
		}
		if err != nil {
			continue
		}
		identity, err := repo.resolveRegisteredBusinessMediaObjectIdentity(reference, key, hasAssets, hasAssets && repo.hasAssetDerivativesTable())
		if err != nil {
			return mapPublicBusinessMediaValidationErrorToGalleryMediaValidationError(mapRegisteredMediaIdentityResolutionErrorToInvalidProductMediaReference(err))
		}
		if identity.AssetID != 0 {
			id := identity.AssetID
			images[index].MediaAssetID = &id
		}
	}
	return nil
}

func galleryImagePrimaryIdentityAlreadyPlaced(candidate gallery.GalleryImage, placed []gallery.GalleryImage) bool {
	for _, existing := range placed {
		if candidate.MediaAssetID != nil && existing.MediaAssetID != nil && *candidate.MediaAssetID == *existing.MediaAssetID {
			return true
		}
		if candidate.URL != "" && candidate.URL == existing.URL {
			return true
		}
	}
	return false
}

func lockActiveGalleryParentWithinMutation(tx *gorm.DB, id uint, item *gallery.Gallery) error {
	if id == 0 {
		return gorm.ErrRecordNotFound
	}
	return tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(item, id).Error
}
