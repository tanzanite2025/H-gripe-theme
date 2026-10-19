package repository

import (
	"fmt"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

type mediaAssetReferenceQuery struct {
	AssetID                                uint
	URLs                                   []string
	ObjectKeys                             []string
	IncludeRetainedDeletedPlacements       bool
	RegisteredObjectCleanupNamespacePrefix string
	IncludeAllExactObjectRegistrations     bool
	ExcludedRetiredDerivativeID            uint
}

// FindAssetReferences runs the media reference scanners registered for the
// current application schema. Each business domain owns its scanner so new
// source types do not change deletion orchestration.
func (r *MediaRepository) FindAssetReferences(asset *media.MediaAsset) ([]media.AssetReference, error) {
	query := newMediaAssetReferenceQuery(asset)
	if query.AssetID == 0 {
		return []media.AssetReference{}, nil
	}
	// Whole-asset cleanup must honor the same retained placement boundary as
	// exact-generation cleanup, including tombstoned product variant snapshots.
	query.IncludeRetainedDeletedPlacements = true
	originalKey := asset.StorageKey
	if originalKey == "" {
		originalKey = r.resolveRegisteredMediaObjectKeyForCleanupOwnership(asset.URL)
	}
	if asset.Visibility == "private" && storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(originalKey) {
		// Retirement also lists this asset's unregistered derivative objects.
		// Another registration inside that exact namespace must retain them.
		query.RegisteredObjectCleanupNamespacePrefix = fmt.Sprintf("%s/%d/", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID)
	}
	var err error
	query, err = r.includeRegisteredDerivativeURLsAndObjectKeysInMediaReferenceQuery(query, asset)
	if err != nil {
		return nil, err
	}

	references := make([]media.AssetReference, 0)
	for _, scan := range r.assetReferenceScanners() {
		items, err := scan(query)
		if err != nil {
			return nil, err
		}
		references = append(references, items...)
	}
	return references, nil
}
