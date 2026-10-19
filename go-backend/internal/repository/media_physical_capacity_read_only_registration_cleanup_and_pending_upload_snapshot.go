package repository

import (
	"context"
	"database/sql"
	"errors"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"gorm.io/gorm"
)

var ErrMediaPhysicalCapacitySnapshotLimitExceeded = errors.New("media capacity registration snapshot exceeded its row limit")

type MediaPhysicalCapacityRegistrationSnapshot struct {
	Originals                  []media.MediaAsset
	Derivatives                []media.MediaAssetDerivative
	CleanupReceipts            []outbox.Event
	PendingConversationUploads []media.MediaUploadSession
	BusinessDraftUploads       []media.BusinessDraftMediaUploadReceipt
}

// The transaction ends before provider IO. It takes no ownership locks, never
// skips a locked writer receipt, and includes original/derivative tombstones.
func (r *MediaRepository) ReadOnlyMediaPhysicalCapacityRegistrationSnapshot(ctx context.Context, rowLimit int) (MediaPhysicalCapacityRegistrationSnapshot, error) {
	var snapshot MediaPhysicalCapacityRegistrationSnapshot
	if r == nil || r.db == nil {
		return snapshot, gorm.ErrInvalidDB
	}
	if rowLimit < 1 {
		return snapshot, ErrMediaPhysicalCapacitySnapshotLimitExceeded
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		remaining := rowLimit
		readRows := func(model interface{}, columns string, destination interface{}, condition string, args ...interface{}) error {
			query := tx.Unscoped().Model(model).Select(columns).Order("id").Limit(remaining + 1)
			if condition != "" {
				query = query.Where(condition, args...)
			}
			result := query.Find(destination)
			if result.Error != nil {
				return result.Error
			}
			remaining -= int(result.RowsAffected)
			if remaining < 0 {
				return ErrMediaPhysicalCapacitySnapshotLimitExceeded
			}
			return nil
		}
		if err := readRows(&media.MediaAsset{}, "id, storage_key, url, size, deleted_at", &snapshot.Originals, ""); err != nil {
			return err
		}
		if err := readRows(&media.MediaAssetDerivative{}, "id, media_asset_id, storage_key, url, size, deleted_at", &snapshot.Derivatives, ""); err != nil {
			return err
		}
		if err := readRows(&outbox.Event{}, "id, event_key, event_type, payload, status, attempts, available_at, last_attempt_at", &snapshot.CleanupReceipts,
			"deleted_at IS NULL AND status <> ? AND event_type IN ?", outbox.EventStatusIgnored, []string{outbox.EventTypeObjectStorageCleanup, outbox.EventTypeCustomerServiceAvatarCleanup, outbox.EventTypeCustomerServiceRetentionCleanup}); err != nil {
			return err
		}
		if err := readRows(&media.MediaUploadSession{}, "id, object_key, expected_upload_bytes, status", &snapshot.PendingConversationUploads, "status <> ?", "bound"); err != nil {
			return err
		}
		return readRows(&media.BusinessDraftMediaUploadReceipt{}, "id, original_object_key, object_keys_json, size, status", &snapshot.BusinessDraftUploads, "status <> ?", "bound")
	}, &sql.TxOptions{Isolation: sql.LevelRepeatableRead, ReadOnly: true})
	if err != nil {
		return MediaPhysicalCapacityRegistrationSnapshot{}, err
	}
	return snapshot, nil
}
