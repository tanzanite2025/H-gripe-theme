package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

// The caller owns the asset row lock. Physical objects and quota usage remain
// until the shared cleanup worker successfully removes the tombstone.
func (s *MediaService) retireUnreferencedMediaAssetWithinTransaction(repo *repository.MediaRepository, tx *gorm.DB, asset *media.MediaAsset) error {
	if s.objectCleanupOutbox == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	// Replaced derivatives can retain physical objects and historical identity.
	// The caller already holds the asset lock; include their tombstones too.
	cleanupSnapshot, err := repo.FindAssetByIDUnscoped(asset.ID)
	if err != nil {
		return err
	}
	asset = cleanupSnapshot
	if err := s.validateSharedMediaCleanupDoesNotOwnDedicatedPublicObjects(asset); err != nil {
		return err
	}
	if asset.CustomerServiceDerivativeWriteExpiresAt != nil && asset.CustomerServiceDerivativeWriteExpiresAt.After(time.Now().UTC()) {
		return ErrCustomerServiceAttachmentWritesInProgress
	}
	references, err := repo.FindAssetReferences(asset)
	if err != nil {
		return err
	}
	if len(references) > 0 {
		return &MediaAssetInUseError{References: references}
	}
	objectKeys := make([]string, 0, len(asset.Derivatives)+1)
	key, err := s.mediaAssetOriginalStorageKeyForDurableCleanup(asset)
	if err != nil {
		return err
	}
	objectKeys = append(objectKeys, key)
	for _, derivative := range asset.Derivatives {
		key := derivative.StorageKey
		if key == "" {
			key = storageObjectKey(s.storage, derivative.URL)
		}
		if key != "" {
			objectKeys = append(objectKeys, key)
		}
	}
	assetID := strconv.FormatUint(uint64(asset.ID), 10)
	event, err := newObjectStorageCleanupEvent(objectCleanupResourceMediaAsset, assetID, outbox.AggregateTypeMediaAsset, assetID, objectKeys)
	if err != nil {
		return err
	}
	if !asset.DeletedAt.Valid {
		if err := repo.DeleteAsset(asset.ID); err != nil {
			return fmt.Errorf("mark media asset deleted: %w", err)
		}
	}
	if err := s.objectCleanupOutbox.WithTx(tx).CreateEvent(event); err != nil {
		return err
	}
	return repo.RemoveBusinessDraftMediaUploadReceiptWithinRetirementTransaction(asset)
}

func (s *MediaService) mediaAssetOriginalStorageKeyForDurableCleanup(asset *media.MediaAsset) (string, error) {
	if strings.TrimSpace(asset.StorageKey) != "" {
		key, valid := storage.NormalizeObjectKey(asset.StorageKey)
		if !valid {
			return "", fmt.Errorf("media cleanup original storage key is invalid")
		}
		return key, nil
	}
	// Historical cloud assets may only store their URL. Resolve its full key
	// through the configured provider; a filename alone is not object identity.
	key, err := s.storage.ObjectKey(asset.URL)
	if err != nil {
		return "", err
	}
	normalized, valid := storage.NormalizeObjectKey(key)
	if !valid {
		return "", fmt.Errorf("media cleanup original storage key is unavailable")
	}
	return normalized, nil
}
