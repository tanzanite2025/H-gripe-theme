package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

const objectCleanupResourceCustomerServiceUnregisteredUpload = "customer_service_unregistered_upload"

func (s *MediaService) cleanupRetiredCustomerServiceAttachmentDerivativeNamespace(ctx context.Context, asset *media.MediaAsset) error {
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
	key := keys[0]
	if !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(key) {
		return fmt.Errorf("unregistered customer-service cleanup requires the private original namespace")
	}
	claimed, err := s.repo.HasCustomerServiceUploadOrRegisteredMediaAssetAtStorageKey(key)
	if err != nil {
		return err
	}
	if claimed {
		return ErrMediaAssetInUse
	}
	return s.storage.Delete(ctx, key)
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
