package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/pkg/storage"

	"github.com/google/uuid"
)

const objectCleanupResourceFailedUGCShowcaseUpload = "ugc_showcase_failed_upload"

func (s *UGCShowcaseService) resolveUGCShowcaseStoredImageObjectKey(reference string) (string, error) {
	if s == nil || s.storage == nil {
		return "", ErrShowcaseStorageUnavailable
	}
	key, err := s.storage.ObjectKey(reference)
	if err != nil {
		key, err = storage.TrustedBusinessMediaReferenceObjectKey(reference, s.storage, s.siteURL)
	}
	normalized, valid := storage.NormalizeObjectKey(key)
	if err != nil || !valid {
		return "", ErrShowcaseImagesInvalid
	}
	return normalized, nil
}

func (s *UGCShowcaseService) validateUGCShowcaseImageBindingsAfterLifecycleLock(ctx context.Context, current, next *ugcshowcase.UGCShowcase) error {
	if next.Status != ugcshowcase.StatusPending && next.Status != ugcshowcase.StatusApproved && next.Status != ugcshowcase.StatusRejected {
		return ErrShowcaseInvalidTransition
	}
	nextReferences, err := decodeShowcaseImageURLs(next.Images)
	if err != nil {
		return err
	}
	oldReferences := map[string]bool{}
	if current != nil {
		if current.Status != next.Status && current.Status != ugcshowcase.StatusPending {
			return ErrShowcaseInvalidTransition
		}
		references, err := decodeShowcaseImageURLs(current.Images)
		if err != nil {
			return err
		}
		for _, reference := range references {
			oldReferences[reference] = true
		}
	} else if next.Status != ugcshowcase.StatusPending {
		return ErrShowcaseInvalidTransition
	}
	for _, reference := range nextReferences {
		if oldReferences[reference] && (current.Status == next.Status || next.Status == ugcshowcase.StatusRejected) {
			continue
		}
		namespace := showcasePendingUploadPrefix
		if next.Status == ugcshowcase.StatusApproved {
			namespace = showcaseApprovedUploadPrefix
		} else if next.Status != ugcshowcase.StatusPending {
			return ErrShowcaseInvalidTransition
		}
		key, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, namespace, reference)
		if err != nil {
			return ErrShowcaseImagesInvalid
		}
		if err := (&MediaService{storage: s.storage}).ValidateMediaObjectExistsBeforeRegisteringOwnership(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *UGCShowcaseService) persistFailedUGCShowcaseImageUploadCompensationBeforeOwnerCheckedCleanup(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	if s == nil || s.storage == nil || s.outbox == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	for _, key := range keys {
		namespace := storage.DedicatedPublicObjectNamespaceForStorageKey(key)
		if namespace != showcasePendingUploadPrefix && namespace != showcaseApprovedUploadPrefix {
			return ErrObjectStorageCleanupUnavailable
		}
		if _, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(s.storage, namespace, key); err != nil {
			return ErrObjectStorageCleanupUnavailable
		}
	}
	if ctx == nil {
		ctx = context.Background()
	}
	compensationContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	attemptID := "failed-upload-" + uuid.NewString()
	event, err := newObjectStorageCleanupEvent(objectCleanupResourceFailedUGCShowcaseUpload, attemptID, outbox.AggregateTypeUGCShowcase, attemptID, keys)
	if err != nil {
		return err
	}
	if err := s.outbox.WithContext(compensationContext).CreateEvent(event); err != nil {
		return fmt.Errorf("%w: persist UGC image compensation: %w", ErrObjectStorageCleanupUnavailable, err)
	}
	_ = NewObjectStorageCleanupOutboxHandler(nil, nil, nil, s).Handle(compensationContext, *event)
	return nil
}

func (s *UGCShowcaseService) pruneDeletedPendingUGCShowcaseImageFromLockedRejectedOwner(item *ugcshowcase.UGCShowcase, key string) (bool, error) {
	references, err := decodeShowcaseImageURLs(item.Images)
	if err != nil {
		return false, err
	}
	remaining := make([]string, 0, len(references))
	for _, reference := range references {
		resolved, err := s.resolveUGCShowcaseStoredImageObjectKey(reference)
		if err == nil && resolved == key {
			continue
		}
		remaining = append(remaining, reference)
	}
	if len(remaining) == len(references) {
		return false, nil
	}
	encoded, err := json.Marshal(remaining)
	if err != nil {
		return false, err
	}
	updated, err := s.repo.UpdateImagesByStatus(item.ID, ugcshowcase.StatusRejected, encoded)
	if updated {
		item.Images = encoded
	}
	return updated, err
}

func (s *UGCShowcaseService) canonicalPendingUGCShowcaseCleanupKeys(references []string) []string {
	var keys []string
	for _, reference := range references {
		key, err := s.resolveUGCShowcaseStoredImageObjectKey(reference)
		if err == nil && strings.HasPrefix(key, showcasePendingUploadPrefix+"/") {
			keys = append(keys, key)
		}
	}
	return normalizeObjectCleanupKeys(keys)
}
