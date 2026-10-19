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

const OrderEvidenceAttachmentPrecommittedObjectWriteReceiptEventKeyPrefix = "order-evidence-precommitted-object-write:"

var ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable = errors.New("order evidence attachment object write receipt is unavailable, expired or already claimed")

func (r *OutboxRepository) RequireIndependentOrderEvidenceAttachmentObjectWriteReceiptTransaction() error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, nested := r.db.Statement.ConnPool.(gorm.TxCommitter); nested {
		return fmt.Errorf("order evidence write receipt requires an independent committed transaction")
	}
	return nil
}

// The exact private destination is durable before provider IO.
// A successful owner transaction consumes only this temporary workflow receipt.
func (r *OutboxRepository) RunOrderEvidenceAttachmentObjectWritesWithPrecommittedCleanupReceiptAndAtomicOwnerExchange(ctx context.Context, receipt *outbox.Event, writeAndBind func(*gorm.DB) error) error {
	if err := r.RequireIndependentOrderEvidenceAttachmentObjectWriteReceiptTransaction(); err != nil {
		return err
	}
	if receipt == nil || receipt.ID != 0 || writeAndBind == nil || !strings.HasPrefix(receipt.EventKey, OrderEvidenceAttachmentPrecommittedObjectWriteReceiptEventKeyPrefix) || receipt.EventType != outbox.EventTypeObjectStorageCleanup {
		return ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable
	}
	var payload outbox.ObjectStorageCleanupPayload
	if json.Unmarshal(receipt.Payload, &payload) != nil || payload.ResourceType != "order_evidence_failed_upload" || len(payload.ObjectKeys) == 0 {
		return ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable
	}
	parts := strings.Split(payload.ResourceID, "/")
	if len(parts) != 4 || parts[0] != orderEvidenceAttachmentLifecycleNamespace || len(payload.ObjectKeys) != 1 {
		return ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable
	}
	for _, value := range parts[1:3] {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil || id == 0 || strconv.FormatUint(id, 10) != value {
			return ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable
		}
	}
	attempt := strings.TrimPrefix(parts[3], "owned-upload-")
	parsed, err := uuid.Parse(attempt)
	if err != nil || parsed.String() != attempt || parts[3] != "owned-upload-"+attempt {
		return ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable
	}
	key := payload.ObjectKeys[0]
	normalized, valid := storage.NormalizeObjectKey(key)
	if !valid || normalized != key || !strings.HasPrefix(key, payload.ResourceID+"/") || !storage.IsPrivateObjectKey(key) {
		return ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable
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
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(writeContext, r.db, orderEvidenceAttachmentLifecycleNamespace, func(tx *gorm.DB) error {
		var stored outbox.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND event_key = ?", receipt.ID, receipt.EventKey).First(&stored).Error; err != nil {
			return errors.Join(ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable, err)
		}
		var expectedDocument, storedDocument interface{}
		if json.Unmarshal(receipt.Payload, &expectedDocument) != nil || json.Unmarshal(stored.Payload, &storedDocument) != nil || !reflect.DeepEqual(expectedDocument, storedDocument) || stored.EventType != receipt.EventType || stored.Status != outbox.EventStatusPending || stored.Attempts != 0 || stored.LastAttemptAt != nil || !time.Now().Before(stored.AvailableAt) || !time.Now().Before(receipt.AvailableAt) {
			return ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable
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
			return ErrOrderEvidenceAttachmentObjectWriteReceiptNoLongerWritable
		}
		return nil
	})
}
