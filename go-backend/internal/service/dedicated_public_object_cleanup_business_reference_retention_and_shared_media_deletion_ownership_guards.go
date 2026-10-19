package service

import (
	"fmt"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

// The repository is bound to the dedicated owner's cleanup transaction. Resolve
// aliases through that owner's provider, which may differ from shared storage.
func checkDedicatedPublicObjectBusinessReferencesBeforePhysicalCleanup(repo *repository.MediaRepository, provider storage.StorageService, siteURL, key string) error {
	repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(&MediaService{storage: provider, siteURL: siteURL})
	state, err := repo.FindDedicatedPublicObjectBusinessReferenceState(key)
	if err != nil {
		return fmt.Errorf("check retained dedicated object references for %s: %w", key, err)
	}
	if len(state.RetainedReferences) > 0 {
		// Returning an error keeps the original Outbox event eligible for retry.
		// Historical consumers do not enqueue dedicated cleanup when detached.
		return &MediaAssetInUseError{References: state.RetainedReferences}
	}
	return nil
}

// Check every original and derivative before deleting any bytes or retiring
// metadata. A stale shared registration never acquires a dedicated namespace.
func (s *MediaService) validateSharedMediaCleanupDoesNotOwnDedicatedPublicObjects(asset *media.MediaAsset) error {
	originalKey, err := s.mediaAssetOriginalStorageKeyForDurableCleanup(asset)
	if err != nil {
		return err
	}
	keys := []string{asset.StorageKey, originalKey}
	urls := []string{asset.URL}
	for _, derivative := range asset.Derivatives {
		key := derivative.StorageKey
		if key == "" {
			key = storageObjectKey(s.storage, derivative.URL)
		}
		keys = append(keys, key)
		urls = append(urls, derivative.URL)
	}
	for _, reference := range urls {
		key, err := s.ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference)
		if err == nil && key != "" {
			keys = append(keys, key)
		}
	}
	for _, key := range keys {
		namespace := storage.DedicatedPublicObjectNamespaceForStorageKey(key)
		// Delete itself accepts storage URLs as well as keys. Historical values
		// in storage_key must not bypass ownership through that second parsing.
		if namespace == "" {
			if resolved, err := s.storage.ObjectKey(key); err == nil {
				namespace = storage.DedicatedPublicObjectNamespaceForStorageKey(resolved)
			}
		}
		if namespace != "" {
			return fmt.Errorf("%w: dedicated object %s must be managed through its owning lifecycle", ErrMediaAssetInUse, key)
		}
	}
	return nil
}
