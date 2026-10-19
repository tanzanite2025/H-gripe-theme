package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const PrivateOrderBusinessAttachmentReceiptEventKeyPrefix = "private-order-business-attachment-write:"
const PrivateOrderBusinessAttachmentFailedUploadResourceType = "private_order_business_attachment_failed_upload"
const privateOrderBusinessAttachmentLifecycleNamespace = "private-order-business-attachments"

var ErrPrivateOrderBusinessAttachmentReceiptUnavailable = errors.New("private order business attachment receipt is no longer writable")

func ValidatePrivateOrderBusinessAttachmentUploadAttempt(prefix string, keys []string) error {
	parts := strings.Split(prefix, "/")
	valid := len(parts) == 3 && ((parts[0] == "warranty" && parts[1] == "claims") || (parts[0] == "after-sales" && parts[1] == "pending"))
	if len(parts) == 4 && parts[0] == "warranty" && parts[1] == "shipment" {
		id, err := strconv.ParseUint(parts[2], 10, 64)
		valid = err == nil && id > 0 && strconv.FormatUint(id, 10) == parts[2]
	}
	if !valid || len(keys) == 0 || len(keys) > 21 {
		return ErrPrivateOrderBusinessAttachmentReceiptUnavailable
	}
	attempt := strings.TrimPrefix(parts[len(parts)-1], "owned-upload-")
	parsed, err := uuid.Parse(attempt)
	if err != nil || parsed.String() != attempt || parts[len(parts)-1] != "owned-upload-"+attempt {
		return ErrPrivateOrderBusinessAttachmentReceiptUnavailable
	}
	seen := make(map[string]bool, len(keys))
	for _, key := range keys {
		normalized, ok := storage.NormalizeObjectKey(key)
		if !ok || normalized != key || !storage.IsPrivateObjectKey(key) || !strings.HasPrefix(key, prefix+"/") || seen[key] {
			return ErrPrivateOrderBusinessAttachmentReceiptUnavailable
		}
		seen[key] = true
	}
	return nil
}

func (r *OutboxRepository) RunPrivateOrderBusinessAttachmentBatchWithPrecommittedReceiptAndAtomicOwnerExchange(ctx context.Context, receipt *outbox.Event, writeAndBind func(*gorm.DB) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, nested := r.db.Statement.ConnPool.(gorm.TxCommitter); nested {
		return fmt.Errorf("private attachment receipt requires an independent committed transaction")
	}
	if receipt == nil || receipt.ID != 0 || writeAndBind == nil || !strings.HasPrefix(receipt.EventKey, PrivateOrderBusinessAttachmentReceiptEventKeyPrefix) || receipt.EventType != outbox.EventTypeObjectStorageCleanup {
		return ErrPrivateOrderBusinessAttachmentReceiptUnavailable
	}
	var payload outbox.ObjectStorageCleanupPayload
	if json.Unmarshal(receipt.Payload, &payload) != nil || payload.ResourceType != PrivateOrderBusinessAttachmentFailedUploadResourceType {
		return ErrPrivateOrderBusinessAttachmentReceiptUnavailable
	}
	if err := ValidatePrivateOrderBusinessAttachmentUploadAttempt(payload.ResourceID, payload.ObjectKeys); err != nil {
		return err
	}
	if ctx == nil {
		ctx = context.Background()
	}
	writeContext, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	deadline, _ := writeContext.Deadline()
	receipt.AvailableAt = deadline.UTC()
	if err := r.db.WithContext(writeContext).Create(receipt).Error; err != nil {
		return err
	}
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(writeContext, r.db, privateOrderBusinessAttachmentLifecycleNamespace, func(tx *gorm.DB) error {
		var stored outbox.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND event_key = ?", receipt.ID, receipt.EventKey).First(&stored).Error; err != nil {
			return errors.Join(ErrPrivateOrderBusinessAttachmentReceiptUnavailable, err)
		}
		var expectedDocument, storedDocument any
		if json.Unmarshal(receipt.Payload, &expectedDocument) != nil || json.Unmarshal(stored.Payload, &storedDocument) != nil || !reflect.DeepEqual(expectedDocument, storedDocument) || stored.EventType != receipt.EventType || stored.Status != outbox.EventStatusPending || stored.Attempts != 0 || stored.LastAttemptAt != nil || !time.Now().Before(stored.AvailableAt) || !time.Now().Before(receipt.AvailableAt) {
			return ErrPrivateOrderBusinessAttachmentReceiptUnavailable
		}
		if err := writeAndBind(tx); err != nil {
			return err
		}
		if err := writeContext.Err(); err != nil {
			return err
		}
		result := tx.Unscoped().Where("id = ? AND event_key = ?", stored.ID, stored.EventKey).Delete(&outbox.Event{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrPrivateOrderBusinessAttachmentReceiptUnavailable
		}
		return nil
	})
}

func (r *OutboxRepository) WithPrivateOrderBusinessAttachmentCleanupLockUntilProviderReturns(ctx context.Context, operation func(*gorm.DB) error) error {
	if r == nil {
		return gorm.ErrInvalidDB
	}
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(ctx, r.db, privateOrderBusinessAttachmentLifecycleNamespace, operation)
}

func (r *ShipmentRecordRepository) WithPrivateOrderBusinessAttachmentMutationLock(operation func(*gorm.DB) error) error {
	return runWithDedicatedPublicObjectNamespaceTransactionLock(r.db.Statement.Context, r.db, privateOrderBusinessAttachmentLifecycleNamespace, operation)
}

func (m *EmailChallengeTxManager) WithPrivateOrderBusinessAttachmentTransaction(tx *gorm.DB) *EmailChallengeTxManager {
	if m == nil {
		return nil
	}
	copy := *m
	copy.db = tx
	return &copy
}

func (m *EmailChallengeTxManager) WithinPrivateOrderBusinessAttachmentMutationLock(operation func(*gorm.DB) error) error {
	if m == nil {
		return gorm.ErrInvalidDB
	}
	return runWithDedicatedPublicObjectNamespaceTransactionLock(m.db.Statement.Context, m.db, privateOrderBusinessAttachmentLifecycleNamespace, operation)
}

func (m *TxManager) WithPrivateOrderBusinessAttachmentTransaction(tx *gorm.DB) *TxManager {
	if m == nil {
		return nil
	}
	copy := *m
	copy.db = tx
	return &copy
}

func (m *TxManager) WithinPrivateOrderBusinessAttachmentMutationLock(operation func(*gorm.DB) error) error {
	if m == nil {
		return gorm.ErrInvalidDB
	}
	return runWithDedicatedPublicObjectNamespaceTransactionLock(m.db.Statement.Context, m.db, privateOrderBusinessAttachmentLifecycleNamespace, operation)
}
