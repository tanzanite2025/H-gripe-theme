package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
)

const (
	completedObjectStorageCleanupReconciliationInterval   = time.Hour
	failedObjectStorageCleanupReconciliationRetryInterval = 15 * time.Minute
)

type ObjectStorageCleanupReconciliationResult struct {
	Claimed    int `json:"claimed"`
	Reconciled int `json:"reconciled"`
	Deferred   int `json:"deferred"`
}

// Absence is only an observation, never proof that a remote writer has stopped.
// Keep checking the exact original authorization indefinitely, including after
// process restarts, database disconnects and unknown cleanup provider outcomes.
func (s *OutboxService) ReconcileOriginalObjectStorageCleanupReceiptsForLateWrites(ctx context.Context, now time.Time, limit int) (ObjectStorageCleanupReconciliationResult, error) {
	result := ObjectStorageCleanupReconciliationResult{}
	if s == nil || s.repo == nil {
		return result, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}
	if now.IsZero() {
		now = time.Now().UTC()
	}
	// A fresh token per pass also fences overlapping passes in one process.
	leaseToken := "storage-reconciliation-" + uuid.NewString()
	events, err := s.repo.WithContext(ctx).ClaimOriginalObjectStorageCleanupReceiptsForDurableReconciliation(now, leaseToken, limit, s.lockTimeout)
	if err != nil {
		return result, err
	}
	result.Claimed = len(events)
	for _, event := range events {
		if err := ctx.Err(); err != nil {
			return result, err
		}
		handler := s.handlers[event.EventType]
		var reconciliationErr error
		if handler == nil {
			reconciliationErr = fmt.Errorf("no cleanup handler registered for %s", event.EventType)
		} else {
			reconciliationErr = s.handleOutboxEventWithLeaseHeartbeatAndRefresh(ctx, event, handler, func(refreshedAt time.Time) error {
				return s.repo.WithContext(ctx).RefreshOriginalObjectStorageCleanupReconciliationLease(event.ID, leaseToken, refreshedAt)
			})
		}
		if errors.Is(reconciliationErr, repository.ErrOutboxOwnershipLost) {
			return result, reconciliationErr
		}
		if err := ctx.Err(); err != nil {
			return result, err
		}
		completedAt := time.Now().UTC()
		if now.After(completedAt) {
			completedAt = now.UTC()
		}
		delay, errorMessage := completedObjectStorageCleanupReconciliationInterval, ""
		if reconciliationErr != nil {
			delay = failedObjectStorageCleanupReconciliationRetryInterval
			errorMessage = truncateOutboxError(reconciliationErr.Error())
			result.Deferred++
		} else {
			result.Reconciled++
		}
		if err := s.repo.WithContext(ctx).CompleteOriginalObjectStorageCleanupReconciliation(event.ID, leaseToken, completedAt, completedAt.Add(delay), errorMessage); err != nil {
			return result, err
		}
	}
	return result, nil
}

// The feedback/PayPal handler re-reads its original receipt under the lifecycle
// lock. A completed receipt may delete only while the supplied lease is current.
func originalObjectStorageCleanupReceiptHasMatchingReconciliationLease(stored, supplied outbox.Event) bool {
	return repository.IsObjectStorageCleanupReconciliationEligible(stored) &&
		stored.Status == supplied.Status && stored.LockedAt != nil && supplied.LockedBy != "" && stored.LockedBy == supplied.LockedBy
}
