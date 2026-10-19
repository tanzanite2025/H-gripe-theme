package repository

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
)

const PermanentMediaOriginalUploadReceiptEventKeyPrefix = "permanent-media-original-upload:"

var ErrPermanentMediaOriginalUploadReceiptNoLongerWritable = errors.New("permanent media original receipt is unavailable, changed, expired or already claimed")

// The receipt is committed on the asset repository's connection before any IO.
// Quota -> exact object -> receipt is the lock order; worker claims skip locked
// receipts. Transaction cancellation cannot release locks before provider return.
func (r *MediaRepository) RunPermanentMediaOriginalUploadWithPrecommittedCleanupReceiptAndAtomicAssetExchange(ctx context.Context, identity media.MediaUploadActorIdentity, receipt *outbox.Event, writeAndRegister func(*MediaRepository, *gorm.DB) error) error {
	if err := r.RequireIndependentMediaProviderOperationTransaction(); err != nil {
		return err
	}
	if receipt == nil || receipt.ID != 0 || writeAndRegister == nil || !identity.IsValidMediaUploadActorIdentity() || receipt.EventType != outbox.EventTypeObjectStorageCleanup {
		return ErrPermanentMediaOriginalUploadReceiptNoLongerWritable
	}
	attempt := strings.TrimPrefix(receipt.EventKey, PermanentMediaOriginalUploadReceiptEventKeyPrefix)
	id, err := uuid.Parse(attempt)
	if err != nil || id == uuid.Nil || id.String() != attempt || receipt.EventKey != PermanentMediaOriginalUploadReceiptEventKeyPrefix+attempt {
		return ErrPermanentMediaOriginalUploadReceiptNoLongerWritable
	}
	var payload outbox.ObjectStorageCleanupPayload
	if json.Unmarshal(receipt.Payload, &payload) != nil || payload.ResourceType != "unregistered_media_upload" || len(payload.ObjectKeys) != 1 || !storage.IsPreallocatedPermanentMediaOriginalObjectKey(payload.ObjectKeys[0]) {
		return ErrPermanentMediaOriginalUploadReceiptNoLongerWritable
	}
	key := payload.ObjectKeys[0]
	if payload.ResourceID != fmt.Sprintf("%x", sha256.Sum256([]byte(key))) {
		return ErrPermanentMediaOriginalUploadReceiptNoLongerWritable
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
	return r.WithinMediaAssetTransactionWithRequestContext(context.WithoutCancel(writeContext), func(repo *MediaRepository, tx *gorm.DB) error {
		if err := writeContext.Err(); err != nil {
			return err
		}
		tx = tx.WithContext(writeContext)
		repo = repo.WithTx(tx)
		if err := repo.lockMediaUploadActorQuotaWithinTransaction(identity); err != nil {
			return err
		}
		if err := repo.LockExactMediaObjectBeforePhysicalCleanupWithinTransaction(key); err != nil {
			return err
		}
		var stored outbox.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND event_key = ?", receipt.ID, receipt.EventKey).First(&stored).Error; err != nil {
			return errors.Join(ErrPermanentMediaOriginalUploadReceiptNoLongerWritable, err)
		}
		var expectedDocument, storedDocument interface{}
		if json.Unmarshal(receipt.Payload, &expectedDocument) != nil || json.Unmarshal(stored.Payload, &storedDocument) != nil || !reflect.DeepEqual(expectedDocument, storedDocument) || stored.EventType != receipt.EventType || stored.Status != outbox.EventStatusPending || stored.Attempts != 0 || stored.LastAttemptAt != nil || !time.Now().Before(stored.AvailableAt) || !time.Now().Before(receipt.AvailableAt) {
			return ErrPermanentMediaOriginalUploadReceiptNoLongerWritable
		}
		if err := writeAndRegister(repo, tx); err != nil {
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
			return ErrPermanentMediaOriginalUploadReceiptNoLongerWritable
		}
		return nil
	})
}
