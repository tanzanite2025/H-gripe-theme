package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
)

const objectCleanupResourceCustomerServiceUnregisteredUpload = "customer_service_unregistered_upload"

func (s *MediaService) cleanupRetiredCustomerServiceAttachmentDerivativeNamespace(ctx context.Context, asset *media.MediaAsset, referencesRepository *repository.MediaRepository) error {
	if asset.Visibility != "private" || !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(assetStorageKey(asset)) {
		return nil
	}
	// Validate metadata before deleting anything. A corrupted derivative cannot
	// confer ownership of another asset's namespace or another private domain.
	prefix := fmt.Sprintf("%s/%d/", storage.CustomerServiceConversationAttachmentDerivativesPrefix, asset.ID)
	for _, derivative := range asset.Derivatives {
		key := derivative.StorageKey
		if key == "" {
			key = storageObjectKey(s.storage, derivative.URL)
		}
		normalized, valid := storage.NormalizeObjectKey(key)
		if !valid || normalized != key || derivative.MediaAssetID != asset.ID || !strings.HasPrefix(key, prefix) {
			return fmt.Errorf("retired customer-service derivative does not belong to its asset")
		}
	}
	lister, ok := s.storage.(storage.CustomerServiceAttachmentDerivativeNamespaceCleanupLister)
	if !ok {
		return ErrObjectStorageCleanupUnavailable
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		keys, err := lister.ListCustomerServiceAttachmentDerivativeObjectKeysForAssetCleanup(ctx, asset.ID)
		if err != nil {
			return err
		}
		if len(keys) == 0 {
			return nil
		}
		for _, key := range keys {
			normalized, valid := storage.NormalizeObjectKey(key)
			if !valid || normalized != key || !strings.HasPrefix(key, prefix) {
				return fmt.Errorf("private derivative listing returned a foreign storage key")
			}
			references, err := referencesRepository.FindExactMediaObjectReferencesBeforePhysicalCleanup(key, s.storage.GetURL(key), 0)
			if err != nil {
				return err
			}
			retained := references[:0]
			for _, reference := range references {
				if reference.ResourceType == "media_derivative_registration" && reference.ParentResourceID == asset.ID {
					continue
				}
				retained = append(retained, reference)
			}
			if len(retained) > 0 {
				return &MediaAssetInUseError{References: retained}
			}
			if err := s.storage.Delete(ctx, key); err != nil {
				return err
			}
		}
	}
}

func (s *MediaService) cleanupRetiredUnregisteredCustomerServiceUploadOriginal(ctx context.Context, keys []string) error {
	if s == nil || s.repo == nil || s.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	if len(keys) != 1 {
		return fmt.Errorf("unregistered customer-service cleanup requires its exact original key")
	}
	key, valid := storage.NormalizeObjectKey(keys[0])
	if !valid || key != keys[0] || !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(key) {
		return fmt.Errorf("unregistered customer-service cleanup requires the private original namespace")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return s.withinMediaPhysicalCleanupTransactionHoldingLocksUntilProviderReturns(ctx, func(repo *repository.MediaRepository, _ *gorm.DB) error {
		// No asset row belongs to this receipt. Take only the exact object
		// lock, then read committed claims/references without locking rows.
		if err := repo.LockExactMediaObjectBeforePhysicalCleanupWithinTransaction(key); err != nil {
			return err
		}
		references, err := repo.FindExactMediaObjectReferencesBeforePhysicalCleanup(key, s.storage.GetURL(key), 0)
		if err != nil {
			return err
		}
		if len(references) > 0 {
			return &MediaAssetInUseError{References: references}
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		return s.storage.Delete(ctx, key)
	})
}

func (s *CustomerServiceRetentionService) shouldPreserveRegisteredAttachmentObjectDuringLegacyProjectionCleanup(reference string) (bool, error) {
	if s.mediaService == nil || s.mediaService.repo == nil || s.mediaService.storage == nil {
		return false, nil
	}
	key := storageObjectKey(s.mediaService.storage, reference)
	if key == "" {
		return false, nil
	}
	asset, err := s.mediaService.repo.FindAssetByStorageKey(key)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	// Registered public files belong to the library. Registered private files
	// are retired through the media lifecycle instead of raw URL deletion.
	return asset.Visibility == "public" || s.managedCustomerServiceAttachmentRetirementAvailable(), nil
}
