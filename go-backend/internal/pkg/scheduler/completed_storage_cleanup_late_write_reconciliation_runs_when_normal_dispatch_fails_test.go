package scheduler

import (
	"context"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestOriginalStorageCleanupReconciliationSchedulerContinuesWhenNormalDispatchLosesOwnership(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&outbox.Event{}))
	now := time.Now().UTC()
	cleanup := outbox.Event{EventKey: "completed-original-cleanup", EventType: outbox.EventTypeObjectStorageCleanup, AggregateType: "media_asset", AggregateID: "1", Status: "processed", Attempts: 1, LastAttemptAt: &now, ProcessedAt: &now}
	require.NoError(t, db.Create(&cleanup).Error)
	require.NoError(t, db.Create(&outbox.Event{EventKey: "normal-dispatch-failure", EventType: outbox.EventTypeOrderPaid, AggregateType: "order", AggregateID: "1", AvailableAt: now.Add(-time.Minute)}).Error)
	worker := service.NewOutboxService(repository.NewOutboxRepository(db))
	worker.RegisterHandler(outbox.EventTypeOrderPaid, func(context.Context, outbox.Event) error { return repository.ErrOutboxOwnershipLost })
	reconciled := 0
	worker.RegisterHandler(outbox.EventTypeObjectStorageCleanup, func(context.Context, outbox.Event) error { reconciled++; return nil })
	scheduler := NewOutboxDispatchScheduler(worker, config.WorkerConfig{})
	scheduler.dispatchOnce(t.Context())
	require.Equal(t, 1, reconciled)
	var saved outbox.Event
	require.NoError(t, db.First(&saved, cleanup.ID).Error)
	require.Equal(t, "processed", saved.Status)
	require.Equal(t, 1, saved.Attempts)
	require.NotNil(t, saved.StorageCleanupLastReconciledAt)
	require.NotNil(t, saved.ReconcileAfter)
}
