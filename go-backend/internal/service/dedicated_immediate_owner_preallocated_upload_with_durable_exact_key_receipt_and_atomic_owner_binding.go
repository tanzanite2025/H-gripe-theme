package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"strconv"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// The provider's returned URL never grants deletion authority. Only the exact
// preallocated key committed in the receipt may be cleaned up after failure.
func uploadDedicatedImmediateOwnerObjectWithWriteAheadCleanupReceiptAndTransactionalBinding(ctx context.Context, repo *repository.OutboxRepository, provider storage.StorageService, namespace string, profileID uint, file *multipart.FileHeader, bind func(*gorm.DB, string, string) error) (string, error) {
	writer, supported := provider.(storage.PreallocatedDedicatedImmediateOwnerObjectWriter)
	if !supported || repo == nil {
		return "", ErrObjectStorageCleanupUnavailable
	}
	key, err := storage.AllocateDedicatedImmediateOwnerUploadObjectKey(namespace, file.Filename)
	if err != nil {
		return "", err
	}
	var receipt *outbox.Event
	attemptID := uuid.NewString()
	if namespace == CustomerServiceAvatarStoragePrefix {
		payload, err := json.Marshal(outbox.CustomerServiceAvatarCleanupPayload{URL: provider.GetURL(key), ExpectedUploadBytes: &file.Size})
		if err != nil {
			return "", err
		}
		receipt = &outbox.Event{EventType: outbox.EventTypeCustomerServiceAvatarCleanup, AggregateType: outbox.AggregateTypeCustomerServiceAgentProfile,
			AggregateID: strconv.FormatUint(uint64(profileID), 10), Payload: payload, MaxAttempts: customerServiceAvatarCleanupMaxAttempts}
	} else {
		resource := objectCleanupResourceSiteLogo
		aggregate := outbox.AggregateTypeSiteLogo
		if namespace == SiteFaviconStoragePrefix {
			resource, aggregate = objectCleanupResourceSiteFavicon, outbox.AggregateTypeSiteFavicon
		}
		receipt, err = newObjectStorageCleanupEvent(resource, attemptID, aggregate, attemptID, []string{key}, map[string]int64{key: file.Size})
		if err != nil {
			return "", err
		}
	}
	receipt.EventKey = repository.DedicatedImmediateOwnerUploadReceiptEventKeyPrefix + attemptID
	expectedURL := provider.GetURL(key)
	resolved, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(provider, namespace, expectedURL)
	if err != nil || resolved != key {
		return "", fmt.Errorf("dedicated upload provider cannot represent its preallocated object key")
	}
	err = repo.RunDedicatedImmediateOwnerUploadWithPrecommittedCleanupReceipt(ctx, namespace, receipt, func(tx *gorm.DB) error {
		reader, err := file.Open()
		if err != nil {
			return err
		}
		defer reader.Close()
		url, err := writer.WriteDedicatedImmediateOwnerObjectAtPreallocatedKey(tx.Statement.Context, reader, key, storage.DedicatedImmediateOwnerUploadCacheControl)
		if err != nil {
			return err
		}
		returnedKey, err := resolveCanonicalDedicatedUploadedObjectKeyFromOwningProvider(provider, namespace, url)
		if err != nil || returnedKey != key {
			return errors.New("dedicated upload provider returned a different object identity")
		}
		return bind(tx, expectedURL, key)
	})
	if err != nil {
		return "", err
	}
	return expectedURL, nil
}
