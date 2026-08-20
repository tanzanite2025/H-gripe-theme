package service

import (
	"context"
	"errors"
	"testing"
	"time"

	outboxdomain "commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/resilience"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOutboxServiceProcessesRegisteredHandler(t *testing.T) {
	db, service := newTestOutboxService(t)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:      "order.paid:1:txn_1",
		EventType:     outboxdomain.EventTypeOrderPaid,
		AggregateType: outboxdomain.AggregateTypeOrder,
		AggregateID:   "1",
		Payload:       datatypes.JSON([]byte(`{"order_id":1}`)),
		AvailableAt:   now.Add(-time.Minute),
	}).Error)

	handled := 0
	service.RegisterHandler(outboxdomain.EventTypeOrderPaid, func(_ context.Context, event outboxdomain.Event) error {
		handled++
		assert.Equal(t, "order.paid:1:txn_1", event.EventKey)
		assert.Equal(t, 1, event.Attempts)
		return nil
	})

	result, err := service.ProcessPending(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, handled)
	assert.Equal(t, 1, result.Claimed)
	assert.Equal(t, 1, result.Processed)
	assert.Equal(t, 0, result.Failed)

	var saved outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", "order.paid:1:txn_1").First(&saved).Error)
	assert.Equal(t, outboxdomain.EventStatusProcessed, saved.Status)
	assert.Equal(t, 1, saved.Attempts)
	assert.NotNil(t, saved.ProcessedAt)
}

func TestOutboxServiceRetriesFailedHandler(t *testing.T) {
	db, service := newTestOutboxService(t)
	now := time.Now().UTC()
	originalJitter := outboxRetryJitterInt63n
	t.Cleanup(func() { outboxRetryJitterInt63n = originalJitter })
	outboxRetryJitterInt63n = func(n int64) int64 {
		return n - 1
	}
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:      "order.paid:2:txn_2",
		EventType:     outboxdomain.EventTypeOrderPaid,
		AggregateType: outboxdomain.AggregateTypeOrder,
		AggregateID:   "2",
		Payload:       datatypes.JSON([]byte(`{"order_id":2}`)),
		AvailableAt:   now.Add(-time.Minute),
		MaxAttempts:   3,
	}).Error)

	service.RegisterHandler(outboxdomain.EventTypeOrderPaid, func(_ context.Context, _ outboxdomain.Event) error {
		return errors.New("erp unavailable")
	})

	result, err := service.ProcessPending(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Claimed)
	assert.Equal(t, 0, result.Processed)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 0, result.DeadLetter)

	var saved outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", "order.paid:2:txn_2").First(&saved).Error)
	assert.Equal(t, outboxdomain.EventStatusFailed, saved.Status)
	assert.Equal(t, 1, saved.Attempts)
	assert.Contains(t, saved.LastError, "erp unavailable")
	delay := saved.AvailableAt.Sub(now)
	assert.GreaterOrEqual(t, delay, defaultOutboxRetryBaseDelay)
	assert.Less(t, delay, defaultOutboxRetryBaseDelay+defaultOutboxRetryJitter)
}

func TestOutboxServicePausesUnknownExternalOutcome(t *testing.T) {
	db, service := newTestOutboxService(t)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:      "order.paid:unknown:1",
		EventType:     outboxdomain.EventTypeOrderPaid,
		AggregateType: outboxdomain.AggregateTypeOrder,
		AggregateID:   "unknown-1",
		Payload:       datatypes.JSON([]byte(`{"order_id":1}`)),
		AvailableAt:   now.Add(-time.Minute),
	}).Error)

	service.RegisterHandler(outboxdomain.EventTypeOrderPaid, func(_ context.Context, _ outboxdomain.Event) error {
		return resilience.ErrExternalOutcomeUnknown
	})

	result, err := service.ProcessPending(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Claimed)
	assert.Equal(t, 1, result.Unknown)
	assert.Zero(t, result.Failed)
	assert.Zero(t, result.DeadLetter)

	var saved outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", "order.paid:unknown:1").First(&saved).Error)
	assert.Equal(t, outboxdomain.EventStatusUnknown, saved.Status)
	assert.NotNil(t, saved.UncertainAt)
	assert.NotNil(t, saved.ReconcileAfter)
	assert.NotNil(t, saved.LastAttemptAt)
	assert.Contains(t, saved.LastError, resilience.ErrExternalOutcomeUnknown.Error())
}

func TestOutboxServiceResumesUnknownEventOnlyAfterExplicitReconciliation(t *testing.T) {
	db, service := newTestOutboxService(t)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:       "order.paid:unknown:2",
		EventType:      outboxdomain.EventTypeOrderPaid,
		AggregateType:  outboxdomain.AggregateTypeOrder,
		AggregateID:    "unknown-2",
		Payload:        datatypes.JSON([]byte(`{"order_id":2}`)),
		Status:         outboxdomain.EventStatusUnknown,
		AvailableAt:    now.Add(15 * time.Minute),
		UncertainAt:    ptrTime(now),
		ReconcileAfter: ptrTime(now.Add(15 * time.Minute)),
	}).Error)

	handled := 0
	service.RegisterHandler(outboxdomain.EventTypeOrderPaid, func(_ context.Context, _ outboxdomain.Event) error {
		handled++
		return nil
	})

	result, err := service.ProcessPending(context.Background(), now, 10)
	require.NoError(t, err)
	assert.Zero(t, result.Claimed)
	assert.Zero(t, handled)

	var saved outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", "order.paid:unknown:2").First(&saved).Error)
	require.NoError(t, service.ResumeUnknownEvent(saved.ID, now, "provider query confirmed no side effect", now))

	result, err = service.ProcessPending(context.Background(), now, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Claimed)
	assert.Equal(t, 1, result.Processed)
	assert.Equal(t, 1, handled)
}

func TestOutboxServiceMarksUnknownEventProcessedAfterReconciliation(t *testing.T) {
	db, service := newTestOutboxService(t)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:       "order.paid:unknown:3",
		EventType:      outboxdomain.EventTypeOrderPaid,
		AggregateType:  outboxdomain.AggregateTypeOrder,
		AggregateID:    "unknown-3",
		Payload:        datatypes.JSON([]byte(`{"order_id":3}`)),
		Status:         outboxdomain.EventStatusUnknown,
		AvailableAt:    now,
		UncertainAt:    ptrTime(now),
		ReconcileAfter: ptrTime(now),
	}).Error)

	var event outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", "order.paid:unknown:3").First(&event).Error)
	processedAt := now.Add(time.Minute)
	require.NoError(t, service.MarkUnknownEventProcessed(event.ID, "provider query confirmed event was already delivered", processedAt))

	var saved outboxdomain.Event
	require.NoError(t, db.First(&saved, event.ID).Error)
	assert.Equal(t, outboxdomain.EventStatusProcessed, saved.Status)
	assert.NotNil(t, saved.ProcessedAt)
	assert.Equal(t, processedAt, saved.ProcessedAt.UTC())
	assert.Nil(t, saved.UncertainAt)
	assert.Nil(t, saved.ReconcileAfter)
}

func TestOutboxRetryDelayAddsJitterAndCapsAtMaximum(t *testing.T) {
	originalJitter := outboxRetryJitterInt63n
	t.Cleanup(func() { outboxRetryJitterInt63n = originalJitter })
	outboxRetryJitterInt63n = func(n int64) int64 {
		return n - 1
	}

	assert.Equal(t, defaultOutboxRetryBaseDelay+defaultOutboxRetryJitter-time.Nanosecond, outboxRetryDelay(1))
	assert.Equal(t, 2*defaultOutboxRetryBaseDelay+defaultOutboxRetryJitter-time.Nanosecond, outboxRetryDelay(2))
	assert.Equal(t, defaultOutboxRetryMaxDelay, outboxRetryDelay(12))
}

func TestOutboxServiceDeadLettersUnhandledEventAtMaxAttempts(t *testing.T) {
	db, service := newTestOutboxService(t)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:      "order.paid:3:txn_3",
		EventType:     outboxdomain.EventTypeOrderPaid,
		AggregateType: outboxdomain.AggregateTypeOrder,
		AggregateID:   "3",
		Payload:       datatypes.JSON([]byte(`{"order_id":3}`)),
		AvailableAt:   now.Add(-time.Minute),
		MaxAttempts:   1,
	}).Error)

	result, err := service.ProcessPending(context.Background(), now, 10)

	require.NoError(t, err)
	assert.Equal(t, 1, result.Claimed)
	assert.Equal(t, 1, result.Unhandled)
	assert.Equal(t, 1, result.DeadLetter)

	var saved outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", "order.paid:3:txn_3").First(&saved).Error)
	assert.Equal(t, outboxdomain.EventStatusDeadLetter, saved.Status)
	assert.Equal(t, 1, saved.Attempts)
	assert.Contains(t, saved.LastError, "no handler registered")
}

func TestOutboxRepositoryFinalStateRequiresOwningWorker(t *testing.T) {
	db, _ := newTestOutboxService(t)
	repo := repository.NewOutboxRepository(db)
	now := time.Now().UTC()
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:      "order.paid:ownership:1",
		EventType:     outboxdomain.EventTypeOrderPaid,
		AggregateType: outboxdomain.AggregateTypeOrder,
		AggregateID:   "1",
		Payload:       datatypes.JSON([]byte(`{}`)),
		AvailableAt:   now.Add(-time.Minute),
	}).Error)

	claimed, err := repo.ClaimReadyEvents(now, "worker-a", 1, time.Minute)
	require.NoError(t, err)
	require.Len(t, claimed, 1)

	err = repo.MarkProcessedByWorker(claimed[0].ID, "worker-b", now)
	require.ErrorIs(t, err, repository.ErrOutboxOwnershipLost)

	err = repo.MarkFailedByWorker(
		claimed[0].ID,
		"worker-b",
		outboxdomain.EventStatusFailed,
		"late failure",
		now.Add(time.Minute),
		now,
	)
	require.ErrorIs(t, err, repository.ErrOutboxOwnershipLost)

	err = repo.RefreshProcessingLockByWorker(claimed[0].ID, "worker-b", now.Add(10*time.Second))
	require.ErrorIs(t, err, repository.ErrOutboxOwnershipLost)

	var saved outboxdomain.Event
	require.NoError(t, db.First(&saved, claimed[0].ID).Error)
	assert.Equal(t, outboxdomain.EventStatusProcessing, saved.Status)
	assert.Equal(t, "worker-a", saved.LockedBy)
	assert.Nil(t, saved.ProcessedAt)

	require.NoError(t, repo.MarkProcessedByWorker(claimed[0].ID, "worker-a", now))
	require.NoError(t, db.First(&saved, claimed[0].ID).Error)
	assert.Equal(t, outboxdomain.EventStatusProcessed, saved.Status)
	assert.Empty(t, saved.LockedBy)
}

func TestOutboxServiceDoesNotMarkAfterOwnershipLost(t *testing.T) {
	db, service := newTestOutboxService(t)
	service.workerID = "worker-a"
	now := time.Now().UTC()
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:      "order.paid:ownership:2",
		EventType:     outboxdomain.EventTypeOrderPaid,
		AggregateType: outboxdomain.AggregateTypeOrder,
		AggregateID:   "2",
		Payload:       datatypes.JSON([]byte(`{}`)),
		AvailableAt:   now.Add(-time.Minute),
		MaxAttempts:   3,
	}).Error)

	service.RegisterHandler(outboxdomain.EventTypeOrderPaid, func(_ context.Context, event outboxdomain.Event) error {
		return db.Model(&outboxdomain.Event{}).
			Where("id = ?", event.ID).
			Updates(map[string]interface{}{
				"status":    outboxdomain.EventStatusProcessing,
				"locked_by": "worker-b",
				"locked_at": now.Add(time.Second),
			}).Error
	})

	result, err := service.ProcessPending(context.Background(), now, 10)
	require.ErrorIs(t, err, repository.ErrOutboxOwnershipLost)
	assert.Equal(t, 1, result.Claimed)
	assert.Equal(t, 0, result.Processed)

	var saved outboxdomain.Event
	require.NoError(t, db.Where("event_key = ?", "order.paid:ownership:2").First(&saved).Error)
	assert.Equal(t, outboxdomain.EventStatusProcessing, saved.Status)
	assert.Equal(t, "worker-b", saved.LockedBy)
	assert.Nil(t, saved.ProcessedAt)
	assert.Empty(t, saved.LastError)
}

func TestOutboxLeaseHeartbeatFailureIsUnknownOutcome(t *testing.T) {
	db, service := newTestOutboxService(t)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	service.lockTimeout = 3 * time.Millisecond

	err = service.handleEventWithLeaseHeartbeat(
		context.Background(),
		outboxdomain.Event{ID: 1},
		func(context.Context, outboxdomain.Event) error {
			require.NoError(t, sqlDB.Close())
			time.Sleep(10 * time.Millisecond)
			return nil
		},
	)

	require.ErrorIs(t, err, resilience.ErrExternalOutcomeUnknown)
}

func TestOutboxRepositoryCountsOnlyCustomerServiceRealtimeStatuses(t *testing.T) {
	db, service := newTestOutboxService(t)
	now := time.Now().UTC()
	repo := repository.NewOutboxRepository(db)
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:      "customer-service-pending",
		EventType:     outboxdomain.EventTypeCustomerServiceRealtime,
		AggregateType: outboxdomain.AggregateTypeCustomerServiceConversation,
		AggregateID:   "1",
		Payload:       datatypes.JSON([]byte(`{}`)),
		Status:        outboxdomain.EventStatusPending,
		AvailableAt:   now,
	}).Error)
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:      "customer-service-dead-letter",
		EventType:     outboxdomain.EventTypeCustomerServiceRealtime,
		AggregateType: outboxdomain.AggregateTypeCustomerServiceConversation,
		AggregateID:   "2",
		Payload:       datatypes.JSON([]byte(`{}`)),
		Status:        outboxdomain.EventStatusDeadLetter,
		AvailableAt:   now,
	}).Error)
	require.NoError(t, db.Create(&outboxdomain.Event{
		EventKey:      "order-pending",
		EventType:     outboxdomain.EventTypeOrderPaid,
		AggregateType: outboxdomain.AggregateTypeOrder,
		AggregateID:   "3",
		Payload:       datatypes.JSON([]byte(`{}`)),
		Status:        outboxdomain.EventStatusPending,
		AvailableAt:   now,
	}).Error)

	counts, err := repo.CountEventsByStatus(outboxdomain.EventTypeCustomerServiceRealtime)
	require.NoError(t, err)
	assert.Equal(t, int64(1), counts[outboxdomain.EventStatusPending])
	assert.Equal(t, int64(1), counts[outboxdomain.EventStatusDeadLetter])
	assert.Zero(t, counts[outboxdomain.EventStatusProcessed])

	require.NoError(t, service.RefreshCustomerServiceRealtimeMetrics())
}

func newTestOutboxService(t *testing.T) (*gorm.DB, *OutboxService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(&outboxdomain.Event{}))
	return db, NewOutboxService(repository.NewOutboxRepository(db))
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
