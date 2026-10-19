package service

import (
	"context"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

// Whole-asset cleanup hard-deletes registration rows. Its original receipt must
// still be able to remove bytes arriving later, using only its authorized keys.
func (s *MediaService) cleanupLateWrittenMediaObjectsWithinTransactionAfterRegistrationRemovalFromOriginalReceipt(ctx context.Context, repo *repository.MediaRepository, assetID uint, keys []string) error {
	if len(keys) == 0 {
		return fmt.Errorf("retired media original cleanup receipt has no exact object keys; ownership needs review")
	}
	var privateOriginal string
	cleanupSnapshot := &media.MediaAsset{ID: assetID}
	privateDerivativePrefix := fmt.Sprintf("%s/%d/", storage.CustomerServiceConversationAttachmentDerivativesPrefix, assetID)
	for _, key := range keys {
		if normalized, valid := storage.NormalizeObjectKey(key); !valid || normalized != key || storage.DedicatedPublicObjectNamespaceForStorageKey(key) != "" || IsAfterSalesEvidenceStorageKey(key) {
			return fmt.Errorf("retired media receipt cannot clean another object's lifecycle")
		}
		if storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(key) {
			privateOriginal = key
		}
		if storage.IsPrivateObjectKey(key) && !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(key) && !strings.HasPrefix(key, privateDerivativePrefix) {
			return fmt.Errorf("retired media receipt cannot clean another private lifecycle")
		}
		cleanupSnapshot.Derivatives = append(cleanupSnapshot.Derivatives, media.MediaAssetDerivative{StorageKey: key, MediaAssetID: assetID})
	}
	cleanupSnapshot.StorageKey = privateOriginal
	if privateOriginal != "" {
		cleanupSnapshot.Visibility = "private"
	}
	if err := repo.LockRegisteredMediaAssetObjectsBeforePhysicalCleanupWithinTransaction(cleanupSnapshot); err != nil {
		return err
	}
	for _, key := range keys {
		references, err := repo.FindExactMediaObjectReferencesBeforePhysicalCleanup(key, s.storage.GetURL(key), 0)
		if err != nil {
			return err
		}
		if len(references) > 0 {
			return &MediaAssetInUseError{References: references}
		}
	}
	for _, key := range keys {
		if err := s.storage.Delete(ctx, key); err != nil {
			return err
		}
	}
	if privateOriginal != "" {
		return s.cleanupRetiredCustomerServiceAttachmentDerivativeNamespace(ctx, &media.MediaAsset{ID: assetID, StorageKey: privateOriginal, Visibility: "private"}, repo)
	}
	return nil
}
