package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
)

// Only the uploading provider's complete, canonical object identity grants
// compensation authority. A filename or another namespace never does.
func resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(provider storage.StorageService, namespace, reference string) (string, error) {
	if provider == nil {
		return "", ErrObjectStorageCleanupUnavailable
	}
	key, err := provider.ObjectKey(reference)
	normalized, valid := storage.NormalizeObjectKey(key)
	if err != nil || !valid || key != normalized || !strings.HasPrefix(key, namespace+"/") {
		return "", fmt.Errorf("dedicated upload has no trusted exact object identity in its owning namespace")
	}
	return key, nil
}

func validateDedicatedSiteIdentityObjectIdentityAndExistenceBeforeOwnerRegistration(ctx context.Context, provider storage.StorageService, namespace, key, reference string) error {
	resolved, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(provider, namespace, reference)
	if err != nil || key != resolved {
		return fmt.Errorf("dedicated site identity registration has inconsistent object identity")
	}
	return (&MediaService{storage: provider}).ValidateMediaObjectExistsBeforeRegisteringOwnership(ctx, key)
}

func persistFailedDedicatedPublicUploadCompensationBeforeAttemptingOwnerCheckedCleanup(ctx context.Context, repo *repository.OutboxRepository, provider storage.StorageService, namespace, resourceType, aggregateType, reference string, handler *ObjectStorageCleanupOutboxHandler) error {
	key, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(provider, namespace, reference)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrObjectStorageCleanupUnavailable, err)
	}
	if repo == nil || handler == nil {
		return ErrObjectStorageCleanupUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	cleanupContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	// Each failed attempt needs a live retry even if a previous cleanup for the
	// same provider key has already completed. Retries use this original event.
	attemptID := "failed-upload-" + uuid.NewString()
	event, err := newObjectStorageCleanupEvent(resourceType, attemptID, aggregateType, attemptID, []string{key})
	if err != nil {
		return err
	}
	if err := repo.WithContext(cleanupContext).CreateEvent(event); err != nil {
		// Never delete bytes if the durable retry could not be persisted.
		return fmt.Errorf("%w: persist dedicated upload compensation: %w", ErrObjectStorageCleanupUnavailable, err)
	}
	// Ownership, retained history and IO errors keep this same task for the
	// existing worker. The upload still reports its original failure.
	_ = handler.Handle(cleanupContext, *event)
	return nil
}
