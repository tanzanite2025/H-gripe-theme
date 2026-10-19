package repository

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
)

const PublicMediaDerivativeUploadReceiptEventKeyPrefix = "public-media-derivative-upload:"

var ErrPublicMediaDerivativeUploadReceiptNoLongerWritable = errors.New("public media derivative receipt is missing, changed, expired or already attempted")
var ErrPublicMediaDerivativeUploadReceiptPersistenceUnavailable = errors.New("public media derivative write-ahead receipt could not be committed")

// Commit every exact-key receipt before IO, then hold parent, sorted object and
// receipt locks until synchronous writes and the metadata exchange complete.
func (r *MediaRepository) RunPublicMediaDerivativeBatchWithPrecommittedReceiptsAndAtomicRegistration(ctx context.Context, parent *media.MediaAsset, receipts []*outbox.Event, writeAndRegister func(*MediaRepository, *gorm.DB) error) error {
	if err := r.RequireIndependentMediaProviderOperationTransaction(); err != nil {
		return err
	}
	if parent == nil || parent.ID == 0 || parent.Visibility != "public" || writeAndRegister == nil || len(receipts) == 0 || len(receipts) > 12 {
		return ErrPublicMediaDerivativeUploadReceiptNoLongerWritable
	}
	keys := make([]string, 0, len(receipts))
	seen := make(map[string]bool)
	for _, receipt := range receipts {
		if receipt == nil || receipt.ID != 0 || receipt.EventType != outbox.EventTypeObjectStorageCleanup {
			return ErrPublicMediaDerivativeUploadReceiptNoLongerWritable
		}
		attempt := strings.TrimPrefix(receipt.EventKey, PublicMediaDerivativeUploadReceiptEventKeyPrefix)
		id, err := uuid.Parse(attempt)
		if err != nil || id == uuid.Nil || id.String() != attempt || receipt.EventKey != PublicMediaDerivativeUploadReceiptEventKeyPrefix+attempt {
			return ErrPublicMediaDerivativeUploadReceiptNoLongerWritable
		}
		var payload outbox.ObjectStorageCleanupPayload
		if json.Unmarshal(receipt.Payload, &payload) != nil || payload.ResourceType != "unregistered_media_derivative" || payload.ResourceID != strconv.FormatUint(uint64(parent.ID), 10) || len(payload.ObjectKeys) != 1 {
			return ErrPublicMediaDerivativeUploadReceiptNoLongerWritable
		}
		key := payload.ObjectKeys[0]
		owner, valid := storage.PreallocatedPublicMediaDerivativeParentAssetID(key)
		if !valid || owner != parent.ID || seen[key] {
			return ErrPublicMediaDerivativeUploadReceiptNoLongerWritable
		}
		keys = append(keys, key)
		seen[key] = true
	}
	if ctx == nil {
		ctx = context.Background()
	}
	writeContext, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	deadline, _ := writeContext.Deadline()
	if err := r.db.WithContext(writeContext).Transaction(func(tx *gorm.DB) error {
		for _, receipt := range receipts {
			receipt.AvailableAt = deadline.UTC()
			if err := tx.Create(receipt).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return errors.Join(ErrPublicMediaDerivativeUploadReceiptPersistenceUnavailable, err)
	}
	return r.WithinMediaAssetTransactionWithRequestContext(context.WithoutCancel(writeContext), func(repo *MediaRepository, tx *gorm.DB) error {
		if err := writeContext.Err(); err != nil {
			return err
		}
		tx = tx.WithContext(writeContext)
		repo = repo.WithTx(tx)
		current, err := repo.FindAssetByIDForUpdate(parent.ID)
		if err != nil {
			return err
		}
		if current.Status != "active" || current.Visibility != "public" || current.StorageKey != parent.StorageKey || current.URL != parent.URL || current.MediaType != "image" {
			return ErrPublicMediaDerivativeUploadReceiptNoLongerWritable
		}
		if err := repo.ValidateBusinessDraftMediaDerivativeGenerationIsPermanentlyBound(current); err != nil {
			return err
		}
		sort.Strings(keys)
		for _, key := range keys {
			if err := repo.LockExactMediaObjectBeforePhysicalCleanupWithinTransaction(key); err != nil {
				return err
			}
		}
		ordered := append([]*outbox.Event(nil), receipts...)
		sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
		for _, receipt := range ordered {
			var stored outbox.Event
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND event_key = ?", receipt.ID, receipt.EventKey).First(&stored).Error; err != nil {
				return errors.Join(ErrPublicMediaDerivativeUploadReceiptNoLongerWritable, err)
			}
			var expectedDocument, storedDocument any
			if json.Unmarshal(receipt.Payload, &expectedDocument) != nil || json.Unmarshal(stored.Payload, &storedDocument) != nil || !reflect.DeepEqual(expectedDocument, storedDocument) || stored.EventType != receipt.EventType || stored.Status != outbox.EventStatusPending || stored.Attempts != 0 || stored.LastAttemptAt != nil || !time.Now().Before(stored.AvailableAt) || !time.Now().Before(receipt.AvailableAt) {
				return ErrPublicMediaDerivativeUploadReceiptNoLongerWritable
			}
		}
		if err := writeAndRegister(repo, tx); err != nil {
			return err
		}
		if err := writeContext.Err(); err != nil {
			return err
		}
		for _, receipt := range ordered {
			result := tx.Unscoped().Where("id = ? AND event_key = ?", receipt.ID, receipt.EventKey).Delete(&outbox.Event{})
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected != 1 {
				return ErrPublicMediaDerivativeUploadReceiptNoLongerWritable
			}
		}
		return nil
	})
}
