package repository

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const suggestionAndPayPalAttachmentLifecycleNamespace = "suggestion-and-paypal-attachment-uploads"

var ErrSuggestionOrPayPalAttachmentReceiptUnavailable = errors.New("attachment upload is unavailable, expired, claimed for cleanup or belongs to another owner")

func validateSuggestionOrPayPalAttachmentReceiptIdentity(event *outbox.Event, payload *outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload) error {
	if event == nil || payload == nil || len(payload.ObjectKeys) != 1 {
		return ErrSuggestionOrPayPalAttachmentReceiptUnavailable
	}
	namespace, ownerID, id, valid := storage.SuggestionOrPayPalAttachmentObjectIdentity(payload.ObjectKeys[0])
	digest, digestErr := hex.DecodeString(payload.ContentSHA256)
	if !valid || ownerID != payload.OwnerID || payload.ResourceID != id || event.EventKey != outbox.SuggestionAndPayPalAttachmentReceiptEventKeyPrefix+id || event.EventType != outbox.EventTypeObjectStorageCleanup || payload.ResourceType != outbox.SuggestionAndPayPalAttachmentCleanupResourceType || payload.Size <= 0 || digestErr != nil || len(digest) != 32 || payload.RequestedAt.IsZero() || payload.Name == "" {
		return ErrSuggestionOrPayPalAttachmentReceiptUnavailable
	}
	if namespace == storage.PayPalDisputeInvoiceAttachmentNamespace {
		if payload.MimeType != "application/pdf" || payload.Size > 5<<20 {
			return ErrSuggestionOrPayPalAttachmentReceiptUnavailable
		}
	} else {
		switch payload.MimeType {
		case "image/png", "image/jpeg", "image/webp", "image/gif":
		default:
			return ErrSuggestionOrPayPalAttachmentReceiptUnavailable
		}
		if payload.Size > 5<<20 {
			return ErrSuggestionOrPayPalAttachmentReceiptUnavailable
		}
	}
	return nil
}

func (r *OutboxRepository) RunSuggestionOrPayPalAttachmentUploadWithPrecommittedReceipt(ctx context.Context, event *outbox.Event, payload *outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload, write func(*gorm.DB) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, nested := r.db.Statement.ConnPool.(gorm.TxCommitter); nested {
		return fmt.Errorf("attachment upload requires an independent committed receipt transaction")
	}
	if err := validateSuggestionOrPayPalAttachmentReceiptIdentity(event, payload); err != nil {
		return err
	}
	if event.ID != 0 || write == nil {
		return ErrSuggestionOrPayPalAttachmentReceiptUnavailable
	}
	if ctx == nil {
		ctx = context.Background()
	}
	writeContext, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	deadline, _ := writeContext.Deadline()
	payload.WriteDeadline, payload.UploadState, payload.BindingExpiresAt = deadline.UTC(), "uploading", nil
	event.AvailableAt = payload.WriteDeadline
	encoded, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	event.Payload = encoded
	if err := r.db.WithContext(writeContext).Create(event).Error; err != nil {
		return err
	}
	return r.WithSuggestionOrPayPalAttachmentLifecycleLockUntilProviderReturns(writeContext, func(tx *gorm.DB) error {
		stored, current, err := NewOutboxRepository(tx).FindSuggestionOrPayPalAttachmentReceiptForUpdate(payload.ObjectKeys[0])
		if err != nil {
			return err
		}
		if stored.ID != event.ID || !isUnclaimedSuggestionOrPayPalAttachmentReceipt(stored) || !reflect.DeepEqual(current, payload) || !time.Now().Before(event.AvailableAt) || !time.Now().Before(stored.AvailableAt) {
			return ErrSuggestionOrPayPalAttachmentReceiptUnavailable
		}
		if err := write(tx); err != nil {
			return err
		}
		if err := writeContext.Err(); err != nil {
			return err
		}
		expires := time.Now().UTC().Add(24 * time.Hour)
		current.UploadState, current.BindingExpiresAt = "ready", &expires
		encoded, err := json.Marshal(current)
		if err != nil {
			return err
		}
		result := tx.Model(stored).Updates(map[string]any{"payload": encoded, "available_at": expires})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return ErrSuggestionOrPayPalAttachmentReceiptUnavailable
		}
		*payload = *current
		return nil
	})
}

func isUnclaimedSuggestionOrPayPalAttachmentReceipt(event *outbox.Event) bool {
	return event.Status == outbox.EventStatusPending && event.Attempts == 0 && event.LastAttemptAt == nil
}

func (r *OutboxRepository) WithSuggestionOrPayPalAttachmentLifecycleLockUntilProviderReturns(ctx context.Context, operation func(*gorm.DB) error) error {
	if r == nil {
		return gorm.ErrInvalidDB
	}
	return runDedicatedPublicObjectPhysicalCleanupTransactionHoldingNamespaceLockUntilProviderReturns(ctx, r.db, suggestionAndPayPalAttachmentLifecycleNamespace, operation)
}

func (r *OutboxRepository) FindSuggestionOrPayPalAttachmentReceiptForUpdate(key string) (*outbox.Event, *outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload, error) {
	_, _, id, valid := storage.SuggestionOrPayPalAttachmentObjectIdentity(key)
	if !valid {
		return nil, nil, ErrSuggestionOrPayPalAttachmentReceiptUnavailable
	}
	var event outbox.Event
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("event_key = ?", outbox.SuggestionAndPayPalAttachmentReceiptEventKeyPrefix+id).First(&event).Error; err != nil {
		return nil, nil, errors.Join(ErrSuggestionOrPayPalAttachmentReceiptUnavailable, err)
	}
	var payload outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload
	if json.Unmarshal(event.Payload, &payload) != nil {
		return nil, nil, ErrSuggestionOrPayPalAttachmentReceiptUnavailable
	}
	if err := validateSuggestionOrPayPalAttachmentReceiptIdentity(&event, &payload); err != nil {
		return nil, nil, err
	}
	if payload.ObjectKeys[0] != key {
		return nil, nil, ErrSuggestionOrPayPalAttachmentReceiptUnavailable
	}
	return &event, &payload, nil
}

func (r *OutboxRepository) ConsumeReadySuggestionOrPayPalAttachmentReceiptWithinOwnerTransaction(key string, ownerID uint, verify func(*outbox.SuggestionAndPayPalAttachmentUploadReceiptPayload) error) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if _, active := r.db.Statement.ConnPool.(gorm.TxCommitter); !active {
		return fmt.Errorf("attachment receipt exchange requires an active owner transaction")
	}
	if err := acquireDedicatedPublicObjectNamespaceLockWithinTransaction(r.db, suggestionAndPayPalAttachmentLifecycleNamespace); err != nil {
		return err
	}
	event, payload, err := r.FindSuggestionOrPayPalAttachmentReceiptForUpdate(key)
	if err != nil {
		return err
	}
	if ownerID == 0 || payload.OwnerID != ownerID || !isUnclaimedSuggestionOrPayPalAttachmentReceipt(event) || payload.UploadState != "ready" || payload.BindingExpiresAt == nil || !time.Now().Before(*payload.BindingExpiresAt) || !time.Now().Before(event.AvailableAt) || verify == nil {
		return ErrSuggestionOrPayPalAttachmentReceiptUnavailable
	}
	if err := verify(payload); err != nil {
		return err
	}
	result := r.db.Where("id = ? AND event_key = ?", event.ID, event.EventKey).Delete(&outbox.Event{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrSuggestionOrPayPalAttachmentReceiptUnavailable
	}
	return nil
}
