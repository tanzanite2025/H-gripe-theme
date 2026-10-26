package repository

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/outbox"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrOutboxDispositionAuditRequired = errors.New("outbox disposition requires an authenticated audit actor")
	ErrOutboxDispositionAuditWrite    = errors.New("outbox disposition audit could not be saved")
	ErrOutboxDispositionInputInvalid  = errors.New("outbox disposition operation or note is invalid")
)

// ReconcileFailedEventWithTransactionalAudit owns both writes on the outbox
// database. It never falls back to a separately configured best-effort audit
// recorder. A successful result is returned only after Transaction commits.
func (r *OutboxRepository) ReconcileFailedEventWithTransactionalAudit(id uint, operation, note string, changedAt time.Time, actorAudit *audit.AuditLog) (outbox.FailureDispositionResult, error) {
	var result outbox.FailureDispositionResult
	if r == nil || r.db == nil {
		return result, gorm.ErrInvalidDB
	}
	if actorAudit == nil || actorAudit.UserID == 0 {
		return result, ErrOutboxDispositionAuditRequired
	}
	note = strings.TrimSpace(note)
	if id == 0 || (operation != "ignore" && operation != "retry") || utf8.RuneCountInString(note) < 3 || utf8.RuneCountInString(note) > 2000 {
		return result, ErrOutboxDispositionInputInvalid
	}
	if changedAt.IsZero() {
		changedAt = time.Now().UTC()
	} else {
		changedAt = changedAt.UTC()
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// PostgreSQL renders timestamptz values in to_jsonb using the session
		// timezone. Canonical UTC makes the receipt verifiable from another
		// connection without changing any persisted timestamp.
		if tx.Name() == "postgres" {
			if err := tx.Exec("SET LOCAL TIME ZONE 'UTC'").Error; err != nil {
				return err
			}
		}
		var before outbox.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&before, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrOutboxEventNotFound
			}
			return err
		}
		if (before.Status != outbox.EventStatusFailed && before.Status != outbox.EventStatusDeadLetter) ||
			before.LockedAt != nil || before.LockedBy != "" || before.UncertainAt != nil || before.ReconcileAfter != nil || before.ProcessedAt != nil || changedAt.Before(before.UpdatedAt) {
			return ErrOutboxInvalidTransition
		}
		previous, err := captureFailedOutboxEventDispositionAuditSnapshot(tx, before)
		if err != nil {
			return err
		}
		// Ignoring a failure preserves its original error and payload. Its note
		// belongs to the atomic receipt, not a destructive last_error replacement.
		updates := map[string]interface{}{"status": outbox.EventStatusIgnored, "updated_at": changedAt}
		if operation == "retry" {
			updates["status"] = outbox.EventStatusPending
			updates["attempts"] = 0
			updates["available_at"] = changedAt
			updates["last_error"] = "manual retry: " + note
		}
		updated := tx.Model(&outbox.Event{}).
			Where("id = ? AND status = ? AND updated_at = ?", id, before.Status, before.UpdatedAt).
			Where("locked_at IS NULL AND COALESCE(locked_by, '') = '' AND uncertain_at IS NULL AND reconcile_after IS NULL AND processed_at IS NULL").
			Updates(updates)
		if updated.Error != nil {
			return updated.Error
		}
		if updated.RowsAffected != 1 {
			return ErrOutboxInvalidTransition
		}
		var after outbox.Event
		if err := tx.First(&after, id).Error; err != nil {
			return err
		}
		current, err := captureFailedOutboxEventDispositionAuditSnapshot(tx, after)
		if err != nil {
			return err
		}
		oldValue, err := json.Marshal(previous)
		if err != nil {
			return err
		}
		newValue, err := json.Marshal(outbox.FailureDispositionAuditReceipt{Contract: outbox.FailureDispositionAuditContract, Operation: operation, Note: note, Event: current})
		if err != nil {
			return err
		}
		// Only request attribution comes from the trusted HTTP audit context.
		// Operation, resource, state and fingerprints come from locked DB rows.
		receipt := *actorAudit
		receipt.ID = 0
		receipt.Action, receipt.Resource, receipt.ResourceID = "reconcile", "outbox_event", id
		receipt.Status, receipt.ErrorMessage, receipt.Changes = "success", "", ""
		receipt.OldValue, receipt.NewValue, receipt.CreatedAt = string(oldValue), string(newValue), time.Now().UTC()
		if err := NewAuditRepository(tx).CreateAuditLog(&receipt); err != nil {
			return fmt.Errorf("%w: %w", ErrOutboxDispositionAuditWrite, err)
		}
		result = outbox.FailureDispositionResult{EventID: id, Status: after.Status, UpdatedAt: after.UpdatedAt, AuditID: receipt.ID}
		return nil
	})
	if err != nil {
		return outbox.FailureDispositionResult{}, err
	}
	return result, nil
}

func captureFailedOutboxEventDispositionAuditSnapshot(tx *gorm.DB, event outbox.Event) (outbox.FailureDispositionAuditSnapshot, error) {
	payloadDigest := sha256.Sum256(event.Payload)
	errorDigest := sha256.Sum256([]byte(event.LastError))
	result := outbox.FailureDispositionAuditSnapshot{
		EventID: event.ID, EventKey: event.EventKey, EventType: event.EventType,
		AggregateType: event.AggregateType, AggregateID: event.AggregateID,
		PayloadSHA256: hex.EncodeToString(payloadDigest[:]), Status: event.Status,
		Attempts: event.Attempts, MaxAttempts: event.MaxAttempts,
		LastErrorSHA256: hex.EncodeToString(errorDigest[:]), LastErrorBytes: len(event.LastError), UpdatedAt: event.UpdatedAt,
	}
	if tx.Name() == "postgres" {
		result.RowFingerprintFormat = "postgresql_complete_row_jsonb_sha256_v1"
		if err := tx.Raw(`SELECT encode(sha256(convert_to(to_jsonb(event)::text,'UTF8')),'hex') FROM outbox_events event WHERE event.id=?`, event.ID).Scan(&result.RowSHA256).Error; err != nil {
			return result, err
		}
	} else {
		result.RowFingerprintFormat = "domain_event_json_sha256_v1"
		encoded, err := json.Marshal(event)
		if err != nil {
			return result, err
		}
		digest := sha256.Sum256(encoded)
		result.RowSHA256 = hex.EncodeToString(digest[:])
	}
	if len(result.RowSHA256) != 64 {
		return result, ErrOutboxInvalidTransition
	}
	return result, nil
}
