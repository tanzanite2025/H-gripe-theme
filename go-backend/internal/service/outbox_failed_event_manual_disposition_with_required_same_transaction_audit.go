package service

import (
	"context"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/repository"
)

func (s *OutboxService) ReconcileFailedEventWithTransactionalAudit(ctx context.Context, id uint, operation, note string, changedAt time.Time, actorAudit *audit.AuditLog) (outbox.FailureDispositionResult, error) {
	if s == nil || s.repo == nil {
		return outbox.FailureDispositionResult{}, repository.ErrOutboxEventNotFound
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return s.repo.WithContext(ctx).ReconcileFailedEventWithTransactionalAudit(id, operation, note, changedAt, actorAudit)
}
