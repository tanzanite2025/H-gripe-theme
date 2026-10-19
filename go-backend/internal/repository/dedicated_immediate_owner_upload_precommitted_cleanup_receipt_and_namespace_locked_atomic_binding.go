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

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const DedicatedImmediateOwnerUploadReceiptEventKeyPrefix = "dedicated-immediate-owner-upload:"

var ErrDedicatedImmediateOwnerUploadReceiptNoLongerWritable = errors.New("dedicated immediate owner upload receipt is no longer writable")

// The existing cleanup event is the write-ahead receipt. Success atomically
// exchanges this temporary receipt for the owner and its normal cleanup path.
// Failure leaves the exact original event for the existing ownership-aware worker.
func (r *OutboxRepository) RunDedicatedImmediateOwnerUploadWithPrecommittedCleanupReceipt(ctx context.Context, namespace string, receipt *outbox.Event, writeAndBind func(*gorm.DB) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, nested := r.db.Statement.ConnPool.(gorm.TxCommitter); nested {
		return fmt.Errorf("dedicated upload receipt requires an independent committed transaction")
	}
	if namespace != dedicatedSiteLogoObjectNamespace && namespace != dedicatedSiteFaviconObjectNamespace && namespace != dedicatedCustomerServiceAvatarObjectNamespace {
		return fmt.Errorf("unsupported dedicated immediate owner upload namespace")
	}
	if receipt == nil || receipt.ID != 0 || !strings.HasPrefix(receipt.EventKey, DedicatedImmediateOwnerUploadReceiptEventKeyPrefix) || writeAndBind == nil {
		return fmt.Errorf("invalid dedicated immediate owner upload cleanup receipt")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	uploadContext, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	deadline, _ := uploadContext.Deadline()
	receipt.AvailableAt = deadline.UTC()
	if err := r.db.WithContext(uploadContext).Create(receipt).Error; err != nil {
		return err
	}
	return r.runDedicatedImmediateOwnerUploadWithLockedUnclaimedReceipt(uploadContext, namespace, receipt, writeAndBind)
}

func (r *OutboxRepository) runDedicatedImmediateOwnerUploadWithLockedUnclaimedReceipt(ctx context.Context, namespace string, receipt *outbox.Event, writeAndBind func(*gorm.DB) error) error {
	// Only the transaction lifetime ignores cancellation. Its namespace and
	// receipt locks must survive until synchronous provider IO has returned.
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(ctx, r.db, namespace, func(tx *gorm.DB) error {
		var stored outbox.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND event_key = ?", receipt.ID, receipt.EventKey).First(&stored).Error; err != nil {
			return errors.Join(ErrDedicatedImmediateOwnerUploadReceiptNoLongerWritable, err)
		}
		// A claim is a permanent fence for this invocation, including failed or
		// manually retried cleanup. Never write after a worker may have deleted.
		var storedPayload, expectedPayload interface{}
		if json.Unmarshal(stored.Payload, &storedPayload) != nil || json.Unmarshal(receipt.Payload, &expectedPayload) != nil ||
			stored.Status != outbox.EventStatusPending || stored.Attempts != 0 || stored.LastAttemptAt != nil ||
			!time.Now().Before(stored.AvailableAt) || stored.EventType != receipt.EventType || !reflect.DeepEqual(storedPayload, expectedPayload) {
			return ErrDedicatedImmediateOwnerUploadReceiptNoLongerWritable
		}
		if err := writeAndBind(tx); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		// Only this newly created workflow receipt is removed. Historical cleanup
		// tasks are neither consumed nor rewritten by a successful new upload.
		result := tx.Unscoped().Where("id = ? AND event_key = ?", stored.ID, stored.EventKey).Delete(&outbox.Event{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrDedicatedImmediateOwnerUploadReceiptNoLongerWritable
		}
		return nil
	})
}
