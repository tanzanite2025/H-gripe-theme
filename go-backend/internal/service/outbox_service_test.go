package service

import (
	"context"
	"errors"
	"testing"
	"time"

	outboxdomain "commerce-platform/internal/domain/outbox"
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
	assert.True(t, saved.AvailableAt.After(now))
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
