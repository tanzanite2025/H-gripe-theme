package repository

import (
	"regexp"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestClaimReadyEventsUsesDatabaseClockForPostgresLeaseExpiry(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	now := time.Date(2026, time.September, 16, 8, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta("NOW()")+`.*`+regexp.QuoteMeta("INTERVAL '1 second'")).
		WithArgs("pending", "failed", sqlmock.AnyArg(), "processing", 300.0, 100).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectCommit()

	claimed, err := NewOutboxRepository(db).ClaimReadyEvents(now, "worker-a", 0, 5*time.Minute)
	require.NoError(t, err)
	require.Empty(t, claimed)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestOutboxFailureOperationsFilterAndTransition(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&outbox.Event{}))
	now := time.Now().UTC()
	require.NoError(t, db.Create(&outbox.Event{
		EventKey: "failed:1", EventType: "order.payment_succeeded", AggregateType: "order", AggregateID: "1",
		Payload: []byte(`{"recipient_email":"buyer@example.test"}`), Status: outbox.EventStatusFailed,
		Attempts: 3, MaxAttempts: 3, AvailableAt: now, LastError: "smtp timeout",
	}).Error)
	require.NoError(t, db.Create(&outbox.Event{
		EventKey: "dead:2", EventType: "after_sales.status_changed", AggregateType: "after_sales_case", AggregateID: "2",
		Payload: []byte(`{}`), Status: outbox.EventStatusDeadLetter, Attempts: 10, MaxAttempts: 10, AvailableAt: now,
	}).Error)
	repo := NewOutboxRepository(db)
	events, err := repo.FindFailedEvents(OutboxEventListOptions{EventType: "order.payment_succeeded"})
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.NoError(t, repo.RetryFailedEvent(events[0].ID, "smtp recovered", now.Add(time.Minute)))
	var saved outbox.Event
	require.NoError(t, db.First(&saved, events[0].ID).Error)
	require.Equal(t, outbox.EventStatusPending, saved.Status)
	require.Zero(t, saved.Attempts)

	var dead outbox.Event
	require.NoError(t, db.Where("event_key = ?", "dead:2").First(&dead).Error)
	require.NoError(t, repo.IgnoreFailedEvent(dead.ID, "permanent invalid recipient", now.Add(2*time.Minute)))
	saved = outbox.Event{}
	require.NoError(t, db.First(&saved, dead.ID).Error)
	require.Equal(t, outbox.EventStatusIgnored, saved.Status)
}
