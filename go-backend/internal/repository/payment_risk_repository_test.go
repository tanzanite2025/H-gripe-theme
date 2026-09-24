package repository

import (
	"encoding/json"
	"testing"
	"time"

	"commerce-platform/internal/domain/outbox"
	paymentdomain "commerce-platform/internal/domain/payment"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPaymentRiskSnapshotAlertTransitionsAreDeduplicated(t *testing.T) {
	db := newPaymentRiskRepositoryTestDB(t)
	repo := NewPaymentRiskRepository(db)
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)

	normal := newPaymentRiskTestSnapshot(paymentdomain.PaymentRiskLevelNormal, now)
	result, err := repo.CreatePaymentRiskSnapshotWithAlert(normal, true)
	require.NoError(t, err)
	require.False(t, result.LevelChanged)
	require.False(t, result.AlertEventCreated)

	warning := newPaymentRiskTestSnapshot(paymentdomain.PaymentRiskLevelWarning, now.Add(time.Hour))
	warning.ReasonsJSON = `["refund_rate_warning"]`
	warning.RefundCount = 80
	warning.RefundRate = 0.08
	result, err = repo.CreatePaymentRiskSnapshotWithAlert(warning, true)
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRiskLevelNormal, result.PreviousLevel)
	require.True(t, result.LevelChanged)
	require.True(t, result.AlertEventCreated)

	repeatedWarning := newPaymentRiskTestSnapshot(paymentdomain.PaymentRiskLevelWarning, now.Add(2*time.Hour))
	result, err = repo.CreatePaymentRiskSnapshotWithAlert(repeatedWarning, true)
	require.NoError(t, err)
	require.False(t, result.LevelChanged)
	require.False(t, result.AlertEventCreated)

	recovered := newPaymentRiskTestSnapshot(paymentdomain.PaymentRiskLevelNormal, now.Add(3*time.Hour))
	result, err = repo.CreatePaymentRiskSnapshotWithAlert(recovered, true)
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRiskLevelWarning, result.PreviousLevel)
	require.True(t, result.LevelChanged)
	require.True(t, result.AlertEventCreated)

	var events []outbox.Event
	require.NoError(t, db.Order("id ASC").Find(&events).Error)
	require.Len(t, events, 2)
	require.Equal(t, outbox.EventTypePaymentRiskLevelChanged, events[0].EventType)

	var payload PaymentRiskLevelChangedPayload
	require.NoError(t, json.Unmarshal(events[0].Payload, &payload))
	require.Equal(t, "stripe", payload.Provider)
	require.Equal(t, paymentdomain.PaymentRiskLevelNormal, payload.PreviousLevel)
	require.Equal(t, paymentdomain.PaymentRiskLevelWarning, payload.CurrentLevel)
	require.Equal(t, int64(80), payload.RefundCount)
	require.InDelta(t, 0.08, payload.RefundRate, 0.000001)
	require.Equal(t, []string{"refund_rate_warning"}, payload.Reasons)
	require.NotContains(t, string(events[0].Payload), "order_id")
	require.NotContains(t, string(events[0].Payload), "customer")

	var state paymentdomain.PaymentRiskAlertState
	require.NoError(t, db.First(&state, "provider = ?", "stripe").Error)
	require.Equal(t, paymentdomain.PaymentRiskLevelNormal, state.CurrentLevel)
	require.Equal(t, recovered.ID, state.CurrentSnapshotID)
}

func TestPaymentRiskSnapshotDoesNotCreateEventsWhileAlertingIsDisabled(t *testing.T) {
	db := newPaymentRiskRepositoryTestDB(t)
	repo := NewPaymentRiskRepository(db)
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)

	warning := newPaymentRiskTestSnapshot(paymentdomain.PaymentRiskLevelWarning, now)
	result, err := repo.CreatePaymentRiskSnapshotWithAlert(warning, false)
	require.NoError(t, err)
	require.True(t, result.LevelChanged)
	require.False(t, result.AlertEventCreated)

	repeatedWarning := newPaymentRiskTestSnapshot(paymentdomain.PaymentRiskLevelWarning, now.Add(time.Hour))
	result, err = repo.CreatePaymentRiskSnapshotWithAlert(repeatedWarning, true)
	require.NoError(t, err)
	require.False(t, result.LevelChanged)
	require.False(t, result.AlertEventCreated)

	critical := newPaymentRiskTestSnapshot(paymentdomain.PaymentRiskLevelCritical, now.Add(2*time.Hour))
	result, err = repo.CreatePaymentRiskSnapshotWithAlert(critical, true)
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRiskLevelWarning, result.PreviousLevel)
	require.True(t, result.AlertEventCreated)

	var eventCount int64
	require.NoError(t, db.Model(&outbox.Event{}).Count(&eventCount).Error)
	require.Equal(t, int64(1), eventCount)
}

func TestCountPaymentRiskMetricsAggregatesExactMinorAmountsByCurrency(t *testing.T) {
	db := newPaymentRiskRepositoryTestDB(t)
	require.NoError(t, db.AutoMigrate(
		&paymentdomain.Transaction{},
		&paymentdomain.Refund{},
		&paymentdomain.StripeDispute{},
		&paymentdomain.PaymentRiskEvent{},
		&paymentdomain.PaymentRiskCheckoutDecision{},
	))
	repo := NewPaymentRiskRepository(db)
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)
	completedAt := now.Add(-time.Hour)

	transactions := []paymentdomain.Transaction{
		{OrderID: 1, TransactionID: "tx-usd-1", PaymentMethod: "stripe", AmountMinor: 1234, Currency: "USD", Status: "completed", CompletedAt: &completedAt},
		{OrderID: 2, TransactionID: "tx-usd-2", PaymentMethod: "stripe", AmountMinor: 567, Currency: "USD", Status: "completed", CompletedAt: &completedAt},
		{OrderID: 3, TransactionID: "tx-jpy-1", PaymentMethod: "stripe", AmountMinor: 500, Currency: "JPY", Status: "completed", CompletedAt: &completedAt},
		{OrderID: 4, TransactionID: "tx-usd-pending", PaymentMethod: "stripe", AmountMinor: 9999, Currency: "USD", Status: "pending", CreatedAt: completedAt},
	}
	require.NoError(t, db.Create(&transactions).Error)

	refundCompletedAt := now.Add(-30 * time.Minute)
	refunds := []paymentdomain.Refund{
		{OrderID: 1, TransactionID: transactions[0].ID, Currency: "USD", AmountMinor: 234, Status: "completed", CompletedAt: &refundCompletedAt},
		{OrderID: 2, TransactionID: transactions[1].ID, Currency: "USD", AmountMinor: 111, Status: "completed", CompletedAt: &refundCompletedAt},
		{OrderID: 3, TransactionID: transactions[2].ID, Currency: "JPY", AmountMinor: 100, Status: "completed", CompletedAt: &refundCompletedAt},
		{OrderID: 1, TransactionID: transactions[0].ID, Currency: "USD", AmountMinor: 999, Status: "pending", CreatedAt: refundCompletedAt},
	}
	require.NoError(t, db.Create(&refunds).Error)

	disputes := []paymentdomain.StripeDispute{
		{StripeDisputeID: "dp-usd", AmountMinor: 300, Currency: "USD", Status: "needs_response", CreatedAt: completedAt},
		{StripeDisputeID: "dp-jpy", AmountMinor: 40, Currency: "JPY", Status: "needs_response", CreatedAt: completedAt},
	}
	require.NoError(t, db.Create(&disputes).Error)

	counts, err := repo.CountPaymentRiskMetrics("stripe", now.Add(-24*time.Hour), now.Add(time.Hour))
	require.NoError(t, err)
	require.Equal(t, int64(3), counts.SuccessfulPaymentCount)
	require.Equal(t, paymentdomain.PaymentRiskAmountMinorByCurrency{"USD": 1801, "JPY": 500}, counts.SuccessfulPaymentAmountMinorByCurrency)
	require.Equal(t, paymentdomain.PaymentRiskAmountMinorByCurrency{"USD": 300, "JPY": 40}, counts.DisputeAmountMinorByCurrency)
	require.Equal(t, int64(3), counts.RefundCount)
	require.Equal(t, paymentdomain.PaymentRiskAmountMinorByCurrency{"USD": 345, "JPY": 100}, counts.RefundAmountMinorByCurrency)
}

func TestPaymentRiskSnapshotPersistsMinorAmountsByCurrency(t *testing.T) {
	db := newPaymentRiskRepositoryTestDB(t)
	repo := NewPaymentRiskRepository(db)
	now := time.Date(2026, time.July, 31, 12, 0, 0, 0, time.UTC)
	snapshot := newPaymentRiskTestSnapshot(paymentdomain.PaymentRiskLevelNormal, now)
	snapshot.SuccessfulPaymentAmountMinorByCurrency = paymentdomain.PaymentRiskAmountMinorByCurrency{"usd": 1234, "JPY": 500}
	snapshot.DisputeAmountMinorByCurrency = paymentdomain.PaymentRiskAmountMinorByCurrency{"USD": 12}
	snapshot.RefundAmountMinorByCurrency = paymentdomain.PaymentRiskAmountMinorByCurrency{"JPY": 9}
	result, err := repo.CreatePaymentRiskSnapshotWithAlert(snapshot, false)
	require.NoError(t, err)
	require.False(t, result.LevelChanged)

	loaded, err := repo.FindLatestPaymentRiskSnapshot("stripe")
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRiskAmountMinorByCurrency{"USD": 1234, "JPY": 500}, loaded.SuccessfulPaymentAmountMinorByCurrency)
	require.Equal(t, paymentdomain.PaymentRiskAmountMinorByCurrency{"USD": 12}, loaded.DisputeAmountMinorByCurrency)
	require.Equal(t, paymentdomain.PaymentRiskAmountMinorByCurrency{"JPY": 9}, loaded.RefundAmountMinorByCurrency)
}

func newPaymentRiskRepositoryTestDB(t *testing.T) *gorm.DB {
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
	require.NoError(t, db.AutoMigrate(
		&paymentdomain.PaymentRiskSnapshot{},
		&paymentdomain.PaymentRiskAlertState{},
		&outbox.Event{},
	))
	return db
}

func newPaymentRiskTestSnapshot(level paymentdomain.PaymentRiskLevel, computedAt time.Time) *paymentdomain.PaymentRiskSnapshot {
	return &paymentdomain.PaymentRiskSnapshot{
		Provider:               "stripe",
		WindowDays:             30,
		WindowStart:            computedAt.AddDate(0, 0, -30),
		WindowEnd:              computedAt,
		SuccessfulPaymentCount: 1000,
		Level:                  level,
		RecommendedAction:      "continue_monitoring",
		ReasonsJSON:            "[]",
		ComputedAt:             computedAt,
	}
}
