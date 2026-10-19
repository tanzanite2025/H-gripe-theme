package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrVisualShowcaseDraftUploadReceiptUnavailable = errors.New("visual showcase draft upload is unavailable, expired or already claimed for cleanup")

func (r *OutboxRepository) RunVisualShowcaseDraftUploadWithPrecommittedExactKeyCleanupReceipt(ctx context.Context, receipt *outbox.Event, payload *homevisualtile.VisualShowcaseDraftUploadCleanupReceiptPayload, writeObject func(*gorm.DB) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, nested := r.db.Statement.ConnPool.(gorm.TxCommitter); nested {
		return fmt.Errorf("visual showcase draft upload requires an independent committed transaction")
	}
	if receipt == nil || payload == nil || receipt.ID != 0 || len(payload.ObjectKeys) != 1 || writeObject == nil {
		return ErrVisualShowcaseDraftUploadReceiptUnavailable
	}
	key := payload.ObjectKeys[0]
	id, valid := storage.VisualShowcaseDraftUploadReceiptIDFromObjectKey(key)
	if !valid || receipt.EventKey != homevisualtile.VisualShowcaseDraftUploadReceiptEventKeyPrefix+id || payload.ResourceID != id || payload.ResourceType != "home_visual_tile" || receipt.EventType != outbox.EventTypeObjectStorageCleanup {
		return ErrVisualShowcaseDraftUploadReceiptUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	writeContext, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	deadline, _ := writeContext.Deadline()
	payload.WriteDeadline = deadline.UTC()
	payload.UploadState, payload.DraftExpiresAt = "uploading", nil
	receipt.AvailableAt = payload.WriteDeadline
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	receipt.Payload = encoded
	if err := r.db.WithContext(writeContext).Create(receipt).Error; err != nil {
		return err
	}
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(writeContext, r.db, dedicatedHomeVisualObjectNamespace, func(tx *gorm.DB) error {
		stored, storedPayload, err := findVisualShowcaseDraftUploadReceiptForUpdateWithinNamespaceTransaction(tx, key)
		if err != nil {
			return errors.Join(ErrVisualShowcaseDraftUploadReceiptUnavailable, err)
		}
		var storedPayloadDocument, expectedPayloadDocument interface{}
		if json.Unmarshal(stored.Payload, &storedPayloadDocument) != nil || json.Unmarshal(receipt.Payload, &expectedPayloadDocument) != nil || stored.ID != receipt.ID || !isUnclaimedVisualShowcaseDraftCleanupReceipt(stored) || storedPayload.UploadState != "uploading" || !storedPayload.WriteDeadline.After(time.Now()) || !stored.AvailableAt.After(time.Now()) || !reflect.DeepEqual(storedPayloadDocument, expectedPayloadDocument) {
			return ErrVisualShowcaseDraftUploadReceiptUnavailable
		}
		if err := writeObject(tx); err != nil {
			return err
		}
		if err := writeContext.Err(); err != nil {
			return err
		}
		expires := time.Now().UTC().Add(24 * time.Hour)
		storedPayload.UploadState, storedPayload.DraftExpiresAt = "ready", &expires
		encoded, err := json.Marshal(storedPayload)
		if err != nil {
			return err
		}
		result := tx.Model(stored).Updates(map[string]interface{}{"payload": encoded, "available_at": expires})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrVisualShowcaseDraftUploadReceiptUnavailable
		}
		*payload = *storedPayload
		return nil
	})
}

func findVisualShowcaseDraftUploadReceiptForUpdateWithinNamespaceTransaction(tx *gorm.DB, key string) (*outbox.Event, *homevisualtile.VisualShowcaseDraftUploadCleanupReceiptPayload, error) {
	id, valid := storage.VisualShowcaseDraftUploadReceiptIDFromObjectKey(key)
	if !valid {
		return nil, nil, ErrVisualShowcaseDraftUploadReceiptUnavailable
	}
	var receipt outbox.Event
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("event_key = ?", homevisualtile.VisualShowcaseDraftUploadReceiptEventKeyPrefix+id).First(&receipt).Error; err != nil {
		return nil, nil, err
	}
	var payload homevisualtile.VisualShowcaseDraftUploadCleanupReceiptPayload
	if json.Unmarshal(receipt.Payload, &payload) != nil || receipt.EventType != outbox.EventTypeObjectStorageCleanup || payload.ResourceType != "home_visual_tile" || payload.ResourceID != id || len(payload.ObjectKeys) != 1 || payload.ObjectKeys[0] != key {
		return nil, nil, ErrVisualShowcaseDraftUploadReceiptUnavailable
	}
	return &receipt, &payload, nil
}

func isUnclaimedVisualShowcaseDraftCleanupReceipt(receipt *outbox.Event) bool {
	return receipt.Status == outbox.EventStatusPending && receipt.Attempts == 0 && receipt.LastAttemptAt == nil
}

// Called under the owner transaction's namespace lock. Receipt removal rolls
// back with the placement. Previously saved owner rows authorize later reuse;
// neither a missing Outbox row nor mere readable bytes prove a completed upload.
func (r *HomeVisualTileRepository) BindVisualShowcaseDraftUploadReceiptsWithinOwnerTransaction(keys []string, resolveReference func(string) string) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, active := r.db.Statement.ConnPool.(gorm.TxCommitter); !active {
		return fmt.Errorf("visual showcase draft binding requires an active owner transaction")
	}
	if err := acquireDedicatedPublicObjectNamespaceLockWithinTransaction(r.db, dedicatedHomeVisualObjectNamespace); err != nil {
		return err
	}
	unique := make(map[string]bool)
	for _, key := range keys {
		if _, draft := storage.VisualShowcaseDraftUploadReceiptIDFromObjectKey(key); draft {
			unique[key] = true
		}
	}
	sorted := make([]string, 0, len(unique))
	for key := range unique {
		sorted = append(sorted, key)
	}
	sort.Strings(sorted)
	missing := make(map[string]bool)
	for _, key := range sorted {
		receipt, payload, err := findVisualShowcaseDraftUploadReceiptForUpdateWithinNamespaceTransaction(r.db, key)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			missing[key] = true
			continue
		}
		if err != nil {
			return err
		}
		if !isUnclaimedVisualShowcaseDraftCleanupReceipt(receipt) || payload.UploadState != "ready" || payload.DraftExpiresAt == nil || !payload.DraftExpiresAt.After(time.Now()) {
			return ErrVisualShowcaseDraftUploadReceiptUnavailable
		}
		result := r.db.Unscoped().Where("id = ? AND event_key = ?", receipt.ID, receipt.EventKey).Delete(&outbox.Event{})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrVisualShowcaseDraftUploadReceiptUnavailable
		}
	}
	if len(missing) == 0 {
		return nil
	}
	if resolveReference == nil {
		return ErrVisualShowcaseDraftUploadReceiptUnavailable
	}
	for afterID := uint(0); ; {
		var owners []homevisualtile.Tile
		if err := r.db.Unscoped().Select("id", "storage_key", "image_url", "thumbnail_url").Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&owners).Error; err != nil {
			return err
		}
		for _, owner := range owners {
			for _, key := range []string{owner.StorageKey, resolveReference(owner.ImageURL), resolveReference(owner.ThumbnailURL)} {
				delete(missing, key)
			}
		}
		if len(missing) == 0 {
			return nil
		}
		if len(owners) < 256 {
			return ErrVisualShowcaseDraftUploadReceiptUnavailable
		}
		afterID = owners[len(owners)-1].ID
	}
}
