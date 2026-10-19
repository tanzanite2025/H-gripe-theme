package service

import (
	"context"
	"crypto/sha256"
	"errors"
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

const (
	objectCleanupResourceUnregisteredMediaUpload     = "unregistered_media_upload"
	objectCleanupResourceUnregisteredMediaDerivative = "unregistered_media_derivative"
	objectCleanupResourceRetiredMediaDerivative      = "retired_media_derivative"
)

// Providers allocate ordinary upload keys during IO. Once a failed upload has
// returned its exact address, persist compensation before attempting deletion.
// If persistence fails, retain bytes and report the error; never fall back to
// an unchecked Delete. Pre-return crashes still require a preallocated receipt.
func (s *MediaService) persistFailedMediaUploadCompensationBeforeAttemptingExactObjectCleanup(ctx context.Context, parent *media.MediaAsset, references []string) (compensationError error) {
	defer func() {
		if compensationError != nil {
			compensationError = fmt.Errorf("%w: %w", ErrObjectStorageCleanupUnavailable, compensationError)
		}
	}()
	if len(references) == 0 {
		return nil
	}
	if s == nil || s.repo == nil || s.storage == nil || s.objectCleanupOutbox == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	cleanupContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	var events []*outbox.Event
	for _, reference := range references {
		key, err := s.storage.ObjectKey(reference)
		normalized, valid := storage.NormalizeObjectKey(key)
		if err != nil || !valid || key != normalized {
			return fmt.Errorf("failed media upload has no trusted exact storage identity")
		}
		resource, resourceID := objectCleanupResourceUnregisteredMediaUpload, fmt.Sprintf("%x", sha256.Sum256([]byte(key)))
		if parent != nil {
			resource, resourceID = objectCleanupResourceUnregisteredMediaDerivative, strconv.FormatUint(uint64(parent.ID), 10)
			if err := validateMediaDerivativeObjectCleanupOwnership(parent, key); err != nil {
				return err
			}
		} else if err := validateFailedSharedMediaOriginalCleanupOwnership(key); err != nil {
			return err
		}
		event, err := newObjectStorageCleanupEvent(resource, resourceID, "media_asset", resourceID, []string{key})
		if err != nil {
			return err
		}
		events = append(events, event)
	}
	if err := s.repo.WithinMediaAssetTransactionWithRequestContext(cleanupContext, func(_ *repository.MediaRepository, tx *gorm.DB) error {
		for _, event := range events {
			if err := s.objectCleanupOutbox.WithTx(tx).CreateEvent(event); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return fmt.Errorf("persist failed media upload compensation: %w", err)
	}
	for _, event := range events {
		// The worker retries this same durable event on references or IO errors.
		_ = NewObjectStorageCleanupOutboxHandler(s, nil, nil, nil).Handle(cleanupContext, *event)
	}
	return nil
}

func validateFailedSharedMediaOriginalCleanupOwnership(key string) error {
	if strings.HasPrefix(key, storage.BusinessDraftMediaObjectPrefix+"/") || storage.IsPrivateObjectKey(key) || IsMediaDerivativeStorageKey(key) ||
		storage.DedicatedPublicObjectNamespaceForStorageKey(key) != "" ||
		showcaseStorageKeyIsInShowcaseNamespace(key) || IsCustomerServiceAvatarStorageKey(key) || IsAfterSalesEvidenceStorageKey(key) {
		return fmt.Errorf("failed shared media upload cannot clean another storage lifecycle")
	}
	return nil
}

func validateMediaDerivativeObjectCleanupOwnership(parent *media.MediaAsset, key string) error {
	if parent == nil || parent.ID == 0 {
		return fmt.Errorf("derivative cleanup requires its registered parent")
	}
	if receiptID, draft := storage.BusinessDraftMediaReceiptIDFromObjectKey(parent.StorageKey); draft {
		if derivativeReceiptID, valid := storage.BusinessDraftMediaReceiptIDFromObjectKey(key); valid && derivativeReceiptID == receiptID && key != parent.StorageKey {
			return nil
		}
	}
	prefix := fmt.Sprintf("%s/%d/", mediaDerivativePrefix, parent.ID)
	if parent.Visibility == "private" {
		if !storage.IsCustomerServiceConversationAttachmentOriginalObjectKey(assetStorageKey(parent)) {
			return fmt.Errorf("derivative cleanup cannot enter another private domain")
		}
		prefix = fmt.Sprintf("%s/%d/", storage.CustomerServiceConversationAttachmentDerivativesPrefix, parent.ID)
	}
	if !strings.HasPrefix(key, prefix) || !IsMediaDerivativeStorageKey(key) {
		return fmt.Errorf("derivative cleanup object does not belong to its parent namespace")
	}
	return nil
}

func (s *MediaService) recordRetiredMediaDerivativeCleanupEventsWithinReplacementTransaction(tx *gorm.DB, derivatives []media.MediaAssetDerivative) error {
	if s.objectCleanupOutbox == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	for _, derivative := range derivatives {
		key, err := s.exactRegisteredDerivativeStorageKeyForCleanup(derivative)
		if err != nil {
			return err
		}
		id, owner := strconv.FormatUint(uint64(derivative.ID), 10), strconv.FormatUint(uint64(derivative.MediaAssetID), 10)
		event, err := newObjectStorageCleanupEvent(objectCleanupResourceRetiredMediaDerivative, id, "media_asset", owner, []string{key})
		if err != nil {
			return err
		}
		if err := s.objectCleanupOutbox.WithTx(tx).CreateEvent(event); err != nil {
			return err
		}
	}
	return nil
}

func (s *MediaService) exactRegisteredDerivativeStorageKeyForCleanup(derivative media.MediaAssetDerivative) (string, error) {
	key, valid := storage.NormalizeObjectKey(derivative.StorageKey)
	urlKey, err := s.ResolveRegisteredMediaObjectKeyForCleanupOwnership(derivative.URL)
	if derivative.StorageKey == "" {
		key, valid = storage.NormalizeObjectKey(urlKey)
	}
	if !valid || err != nil || (urlKey != "" && urlKey != key) || storage.DedicatedPublicObjectNamespaceForStorageKey(key) != "" {
		return "", fmt.Errorf("retired derivative cleanup has inconsistent exact object identity")
	}
	return key, nil
}

func (s *MediaService) cleanupUnregisteredOrRetiredMediaObjectFromDurableEvent(ctx context.Context, resource, resourceID string, keys []string) error {
	if s == nil || s.repo == nil || s.storage == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if len(keys) != 1 {
		return fmt.Errorf("media object cleanup requires exactly one object")
	}
	key, valid := storage.NormalizeObjectKey(keys[0])
	if !valid || key != keys[0] {
		return fmt.Errorf("media object cleanup requires a normalized exact key")
	}
	return s.withinMediaPhysicalCleanupTransactionHoldingLocksUntilProviderReturns(ctx, func(repo *repository.MediaRepository, _ *gorm.DB) error {
		var excludedDerivativeID uint
		reference := s.storage.GetURL(key)
		if resource == objectCleanupResourceUnregisteredMediaUpload {
			if err := validateFailedSharedMediaOriginalCleanupOwnership(key); err != nil {
				return err
			}
		} else {
			id, err := strconv.ParseUint(resourceID, 10, 64)
			if err != nil || id == 0 || uint64(uint(id)) != id {
				return fmt.Errorf("media derivative cleanup identity is invalid")
			}
			parentID := uint(id)
			if resource == objectCleanupResourceRetiredMediaDerivative {
				previous, err := repo.FindMediaDerivativeIncludingRetiredGeneration(uint(id), false)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				if err != nil {
					return err
				}
				parentID = previous.MediaAssetID
			}
			// Parent rows precede derivative rows and object locks, matching
			// registration, replacement and whole-asset cleanup lock ordering.
			parent, err := repo.FindAssetByIDUnscopedForUpdate(parentID)
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if resource == objectCleanupResourceRetiredMediaDerivative {
					_, derivativeError := repo.FindMediaDerivativeIncludingRetiredGeneration(uint(id), false)
					if errors.Is(derivativeError, gorm.ErrRecordNotFound) {
						return nil // whole-asset cleanup removed this generation
					}
					return fmt.Errorf("retired derivative cleanup parent is missing; ownership needs review")
				}
				// Failed derivatives can outlive a concurrently cleaned parent.
				// The durable receipt still restricts its exact public/private
				// derivative namespace; all surviving registrations are rechecked.
				publicPrefix := fmt.Sprintf("%s/%d/", mediaDerivativePrefix, parentID)
				privatePrefix := fmt.Sprintf("%s/%d/", storage.CustomerServiceConversationAttachmentDerivativesPrefix, parentID)
				if !strings.HasPrefix(key, publicPrefix) && !strings.HasPrefix(key, privatePrefix) {
					return fmt.Errorf("failed derivative cleanup does not belong to the recorded parent namespace")
				}
			} else if err != nil {
				return err
			}
			if parent != nil {
				if err := validateMediaDerivativeObjectCleanupOwnership(parent, key); err != nil {
					return err
				}
			}
			if resource == objectCleanupResourceRetiredMediaDerivative {
				derivative, err := repo.FindMediaDerivativeIncludingRetiredGeneration(uint(id), true)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return nil
				}
				if err != nil {
					return err
				}
				registeredKey, err := s.exactRegisteredDerivativeStorageKeyForCleanup(*derivative)
				if err != nil || derivative.MediaAssetID != parentID || registeredKey != key || !derivative.DeletedAt.Valid {
					return fmt.Errorf("retired derivative cleanup no longer matches its registration")
				}
				excludedDerivativeID, reference = derivative.ID, derivative.URL
			}
		}
		if err := repo.LockExactMediaObjectBeforePhysicalCleanupWithinTransaction(key); err != nil {
			return err
		}
		references, err := repo.FindExactMediaObjectReferencesBeforePhysicalCleanup(key, reference, excludedDerivativeID)
		if err != nil {
			return err
		}
		if len(references) > 0 {
			return &MediaAssetInUseError{References: references}
		}
		// Retain the tombstone: new bindings must continue rejecting the retired
		// address after deletion, and the worker can safely repeat Delete.
		return s.storage.Delete(ctx, key)
	})
}
