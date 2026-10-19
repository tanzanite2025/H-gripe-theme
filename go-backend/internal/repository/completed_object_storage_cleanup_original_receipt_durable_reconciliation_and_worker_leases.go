package repository

import (
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/outbox"

	"gorm.io/gorm"
)

// Only idempotent, ownership-aware object deletion commands may be replayed.
// Payment, notification and other unknown external outcomes remain manual.
func IsObjectStorageCleanupReconciliationEligible(event outbox.Event) bool {
	if event.EventType != outbox.EventTypeObjectStorageCleanup && event.EventType != outbox.EventTypeCustomerServiceAvatarCleanup {
		return false
	}
	return event.Status == outbox.EventStatusProcessed || event.Status == outbox.EventStatusUnknown || event.Status == outbox.EventStatusDeadLetter
}

func objectStorageCleanupReconciliationEligibleOriginalReceipts(query *gorm.DB) *gorm.DB {
	return query.Where("event_type IN ? AND status IN ?",
		[]string{outbox.EventTypeObjectStorageCleanup, outbox.EventTypeCustomerServiceAvatarCleanup},
		[]string{outbox.EventStatusProcessed, outbox.EventStatusUnknown, outbox.EventStatusDeadLetter})
}

// The original receipt is retained with its original status, attempts and writer
// fence. No new deletion authority is inferred from a URL or a bucket listing.
// NULL reconcile_after enrolls surviving historical originals without rewriting
// their payloads; consumed/soft-deleted receipts are deliberately excluded.
func (r *OutboxRepository) ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now time.Time, leaseToken string, limit int, lockTimeout time.Duration) ([]outbox.Event, error) {
	if strings.TrimSpace(leaseToken) == "" {
		return nil, fmt.Errorf("object storage cleanup reconciliation requires a unique lease token")
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	now = now.UTC()
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if lockTimeout <= 0 {
		lockTimeout = 5 * time.Minute
	}
	var claimed []outbox.Event
	err := r.db.Transaction(func(tx *gorm.DB) error {
		dueTimeExpression := "COALESCE(reconcile_after, processed_at, updated_at, created_at)"
		dueTimeParameter := "?"
		if tx.Name() == "sqlite" {
			// SQLite stores offsets as text; compare instants across historical
			// local-time timestamps and new UTC schedules.
			dueTimeExpression = "julianday(" + dueTimeExpression + ")"
			dueTimeParameter = "julianday(?)"
		}
		query := objectStorageCleanupReconciliationEligibleOriginalReceipts(tx.Model(&outbox.Event{})).
			Where(dueTimeExpression+" <= "+dueTimeParameter, now)
		var leaseTime interface{} = now
		if tx.Name() == "postgres" {
			query = query.Where("locked_at IS NULL OR locked_at <= NOW() - (CAST(? AS double precision) * INTERVAL '1 second')", lockTimeout.Seconds())
			leaseTime = gorm.Expr("clock_timestamp()")
		} else {
			query = query.Where("locked_at IS NULL OR locked_at <= ?", now.Add(-lockTimeout))
		}
		if err := r.lockForClaim(query.Order(dueTimeExpression + " ASC, id ASC").Limit(limit)).Find(&claimed).Error; err != nil {
			return err
		}
		if len(claimed) == 0 {
			return nil
		}
		ids := make([]uint, 0, len(claimed))
		for index := range claimed {
			ids = append(ids, claimed[index].ID)
			claimed[index].LockedBy, claimed[index].LockedAt = leaseToken, &now
		}
		// UpdateColumns avoids changing the original dispatch's updated_at.
		return tx.Model(&outbox.Event{}).Where("id IN ?", ids).UpdateColumns(map[string]interface{}{
			"locked_at": leaseTime, "locked_by": leaseToken,
		}).Error
	})
	return claimed, err
}

func (r *OutboxRepository) RefreshOriginalObjectStorageCleanupReconciliationLease(eventID uint, leaseToken string, now time.Time) error {
	var leaseTime interface{} = now.UTC()
	if r.db.Name() == "postgres" {
		leaseTime = gorm.Expr("clock_timestamp()")
	}
	result := objectStorageCleanupReconciliationEligibleOriginalReceipts(r.db.Model(&outbox.Event{})).
		Where("id = ? AND locked_by = ? AND locked_by <> ''", eventID, leaseToken).
		UpdateColumns(map[string]interface{}{"locked_at": leaseTime})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrOutboxOwnershipLost
	}
	return nil
}

func (r *OutboxRepository) CompleteOriginalObjectStorageCleanupReconciliation(eventID uint, leaseToken string, completedAt, nextReconciliationAt time.Time, reconciliationError string) error {
	if completedAt.IsZero() || !nextReconciliationAt.After(completedAt) {
		return fmt.Errorf("object storage cleanup reconciliation must retain a future retry")
	}
	result := objectStorageCleanupReconciliationEligibleOriginalReceipts(r.db.Model(&outbox.Event{})).
		Where("id = ? AND locked_by = ? AND locked_by <> ''", eventID, leaseToken).
		UpdateColumns(map[string]interface{}{
			"locked_at": nil, "locked_by": "", "reconcile_after": nextReconciliationAt.UTC(),
			"storage_cleanup_last_reconciled_at":   completedAt.UTC(),
			"storage_cleanup_reconciliation_error": reconciliationError,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return ErrOutboxOwnershipLost
	}
	return nil
}
