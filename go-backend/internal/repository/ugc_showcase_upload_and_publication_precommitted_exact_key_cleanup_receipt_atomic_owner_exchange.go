package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const UGCShowcasePrecommittedObjectWriteReceiptEventKeyPrefix = "ugc-showcase-precommitted-object-write:"

var ErrUGCShowcaseObjectWriteReceiptNoLongerWritable = errors.New("UGC showcase object write receipt is unavailable, expired or already claimed")

func (r *OutboxRepository) RequireIndependentUGCShowcaseObjectWriteReceiptTransaction() error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, nested := r.db.Statement.ConnPool.(gorm.TxCommitter); nested {
		return fmt.Errorf("UGC write receipt requires an independent committed transaction")
	}
	return nil
}

// Every destination, including not-yet-attempted files, is durable before IO.
// A successful owner transaction consumes only this temporary workflow receipt.
func (r *OutboxRepository) RunUGCShowcaseObjectWritesWithPrecommittedCleanupReceiptAndAtomicOwnerExchange(ctx context.Context, receipt *outbox.Event, writeAndBind func(*gorm.DB) error) error {
	if err := r.RequireIndependentUGCShowcaseObjectWriteReceiptTransaction(); err != nil {
		return err
	}
	if receipt == nil || receipt.ID != 0 || writeAndBind == nil || !strings.HasPrefix(receipt.EventKey, UGCShowcasePrecommittedObjectWriteReceiptEventKeyPrefix) || receipt.EventType != outbox.EventTypeObjectStorageCleanup {
		return ErrUGCShowcaseObjectWriteReceiptNoLongerWritable
	}
	var payload outbox.ObjectStorageCleanupPayload
	if json.Unmarshal(receipt.Payload, &payload) != nil || payload.ResourceType != "ugc_showcase_failed_upload" || len(payload.ObjectKeys) == 0 {
		return ErrUGCShowcaseObjectWriteReceiptNoLongerWritable
	}
	seen := make(map[string]bool)
	for _, key := range payload.ObjectKeys {
		normalized, valid := storage.NormalizeObjectKey(key)
		if !valid || normalized != key || seen[key] || !(strings.HasPrefix(key, "showcase/pending/") || strings.HasPrefix(key, "showcase/approved/")) {
			return ErrUGCShowcaseObjectWriteReceiptNoLongerWritable
		}
		seen[key] = true
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
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(writeContext, r.db, dedicatedUGCShowcaseLifecycleNamespace, func(tx *gorm.DB) error {
		var stored outbox.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND event_key = ?", receipt.ID, receipt.EventKey).First(&stored).Error; err != nil {
			return errors.Join(ErrUGCShowcaseObjectWriteReceiptNoLongerWritable, err)
		}
		var expectedDocument, storedDocument interface{}
		if json.Unmarshal(receipt.Payload, &expectedDocument) != nil || json.Unmarshal(stored.Payload, &storedDocument) != nil || !reflect.DeepEqual(expectedDocument, storedDocument) || stored.EventType != receipt.EventType || stored.Status != outbox.EventStatusPending || stored.Attempts != 0 || stored.LastAttemptAt != nil || !time.Now().Before(stored.AvailableAt) || !time.Now().Before(receipt.AvailableAt) {
			return ErrUGCShowcaseObjectWriteReceiptNoLongerWritable
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
			return ErrUGCShowcaseObjectWriteReceiptNoLongerWritable
		}
		return nil
	})
}
