package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/product"
)

// Configuration supplies storage aliases and server-owned derivative snapshots;
// all database reads made through it must use the product transaction.
type ProductMediaTransactionalAssetBindingResolver interface {
	PublicBusinessMediaReferenceObjectKeyResolver
	ProductMediaImageVariantsWithinProductTransaction(tx *gorm.DB, assetID uint) (map[string]product.ProductMediaImageVariant, string, error)
}

func (r *ProductRepository) ConfigureProductMediaTransactionalAssetBindingResolver(resolver ProductMediaTransactionalAssetBindingResolver) {
	r.mediaAssetBindingResolver = resolver
}

// The caller's transaction retains the same asset locks used by retirement and
// metadata edits until every product placement has committed or rolled back.
func (r *ProductRepository) lockValidateAndCanonicalizeProductMediaAssetBindings(tx *gorm.DB, items, existingItems []product.ProductMedia, additionalBindings ...publicBusinessMediaReferenceBinding) error {
	if len(items) == 0 && len(additionalBindings) == 0 {
		return nil
	}
	existingByID := make(map[uint]product.ProductMedia, len(existingItems))
	for _, existing := range existingItems {
		existingByID[existing.ID] = existing
	}
	var references []string
	var assetIDs []uint
	for _, item := range items {
		references = append(references, item.URL, item.ThumbnailURL, item.PosterURL)
		if item.MediaAssetID != nil {
			assetIDs = append(assetIDs, *item.MediaAssetID)
		}
	}
	for _, binding := range additionalBindings {
		if binding.URL != nil {
			references = append(references, *binding.URL)
		}
		if binding.AssetID != nil && *binding.AssetID != nil {
			assetIDs = append(assetIDs, **binding.AssetID)
		}
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, r.mediaAssetBindingResolver, references, assetIDs)
	if err != nil {
		return err
	}
	identities, keys, lockedAssets := resolved.identities, resolved.keys, resolved.assets
	for index := range items {
		item := &items[index]
		main := identities[item.URL]
		existing, retained := existingByID[item.ID]
		preserveExistingUnresolvedMainURL := false
		if strings.TrimSpace(item.URL) == "" {
			return fmt.Errorf("%w: media URL is required", ErrProductMediaReferenceInvalid)
		}
		if item.MediaAssetID != nil && main.AssetID != *item.MediaAssetID {
			// Old placements can contain an unconfigured host/port alias. Its
			// persisted ID still protects the original asset, but is not evidence
			// authorizing this URL for another placement or a changed address.
			if main.AssetID != 0 || keys[item.URL] != "" || !retained || existing.MediaAssetID == nil || *existing.MediaAssetID != *item.MediaAssetID || existing.URL != item.URL {
				return fmt.Errorf("%w: media asset ID and URL identify different files", ErrProductMediaReferenceInvalid)
			}
			main.AssetID, preserveExistingUnresolvedMainURL = *item.MediaAssetID, true
		}
		for _, field := range []struct{ requested, previous string }{{item.URL, existing.URL}, {item.ThumbnailURL, existing.ThumbnailURL}, {item.PosterURL, existing.PosterURL}} {
			if keys[field.requested] != "" && identities[field.requested].AssetID == 0 && (!retained || field.requested != field.previous) {
				return fmt.Errorf("%w: managed media file is not registered; select or upload an active asset", ErrProductMediaReferenceInvalid)
			}
		}
		// A registered URL-only reference acquires the existing placement's asset
		// ID. Static/external URLs stay unmanaged; old unregistered upload paths
		// may only remain on the same existing placement, in the same field.
		item.ImageVariantData = product.ProductMediaImageVariantsJSON(nil)
		generatedThumbnail := ""
		if item.MediaType == "" {
			item.MediaType = "image"
		}
		if main.AssetID != 0 {
			asset := lockedAssets[main.AssetID]
			if asset.MediaType != item.MediaType {
				return fmt.Errorf("%w: product media type does not match the asset", ErrProductMediaReferenceInvalid)
			}
			assetID := main.AssetID
			item.MediaAssetID = &assetID
			if !preserveExistingUnresolvedMainURL {
				item.URL = asset.URL
			}
			if item.MediaType == "image" && r.mediaAssetBindingResolver != nil {
				variants, thumbnail, err := r.mediaAssetBindingResolver.ProductMediaImageVariantsWithinProductTransaction(tx, assetID)
				if err != nil {
					return err
				}
				item.ImageVariantData = product.ProductMediaImageVariantsJSON(variants)
				generatedThumbnail = thumbnail
				if strings.TrimSpace(item.ThumbnailURL) == "" {
					item.ThumbnailURL = thumbnail
				}
			}
		}
		for _, address := range []*string{&item.ThumbnailURL, &item.PosterURL} {
			identity, exists := identities[*address]
			if !exists || identity.AssetID == 0 {
				continue
			}
			asset := lockedAssets[identity.AssetID]
			if asset.MediaType != "image" {
				return fmt.Errorf("%w: thumbnails and posters must reference image assets", ErrProductMediaReferenceInvalid)
			}
			// Independently selected derivatives use their original image's stable
			// address. Their display derivatives can be rebuilt without breaking a
			// video's poster or a different product image's custom thumbnail.
			if address == &item.ThumbnailURL && identity.AssetID == main.AssetID && identity.URL == generatedThumbnail {
				*address = identity.URL
			} else {
				*address = asset.URL
			}
		}
	}
	for _, binding := range additionalBindings {
		if err := resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(binding); err != nil {
			return err
		}
	}

	return nil
}

func mapRegisteredMediaIdentityResolutionErrorToInvalidProductMediaReference(err error) error {
	if errors.Is(err, ErrRegisteredMediaObjectOwnershipAmbiguous) {
		return fmt.Errorf("%w: %v", ErrProductMediaReferenceInvalid, err)
	}
	return err
}
