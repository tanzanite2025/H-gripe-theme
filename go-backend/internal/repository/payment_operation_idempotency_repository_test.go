package repository

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"commerce-platform/internal/domain/payment"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPaymentOperationIdempotencyRepositoryReclaimFencesPreviousOwner(t *testing.T) {
	db := newPaymentOperationIdempotencyRepositoryTestDB(t)
	repo := NewPaymentOperationIdempotencyRepository(db)
	now := time.Now().UTC()
	expiredAt := now.Add(-time.Minute)
	record := &payment.PaymentOperationIdempotency{
		UserID:         9,
		Scope:          "wechat_confirm",
		IdempotencyKey: "confirm-1",
		RequestHash:    "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Status:         payment.PaymentOperationIdempotencyPending,
		ClaimToken:     "previous-owner",
		LeaseExpiresAt: &expiredAt,
	}
	require.NoError(t, db.Create(record).Error)

	claimed, err := repo.ReclaimExpiredQuery(
		record.ID,
		record.RequestHash,
		"new-owner",
		now,
		now.Add(time.Minute),
	)
	require.NoError(t, err)
	require.True(t, claimed)

	err = repo.Complete(record.ID, "previous-owner", http.StatusOK, "application/json", `{"stale":true}`)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrPaymentOperationIdempotencyClaimLost))
	require.NoError(t, repo.Complete(record.ID, "new-owner", http.StatusOK, "application/json", `{"paid":true}`))

	var saved payment.PaymentOperationIdempotency
	require.NoError(t, db.First(&saved, record.ID).Error)
	require.Equal(t, payment.PaymentOperationIdempotencyCompleted, saved.Status)
	require.Equal(t, `{"paid":true}`, saved.ResponseBody)
	require.Empty(t, saved.ClaimToken)
	require.Nil(t, saved.LeaseExpiresAt)
}

func TestPaymentOperationIdempotencyRepositoryClaimsExpiredMutationForReconciliation(t *testing.T) {
	db := newPaymentOperationIdempotencyRepositoryTestDB(t)
	repo := NewPaymentOperationIdempotencyRepository(db)
	now := time.Now().UTC()
	expiredAt := now.Add(-time.Minute)
	record := &payment.PaymentOperationIdempotency{
		UserID:         9,
		Scope:          "paypal_capture",
		IdempotencyKey: "capture-1",
		RequestHash:    "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Status:         payment.PaymentOperationIdempotencyPending,
		ClaimToken:     "dead-owner",
		LeaseExpiresAt: &expiredAt,
	}
	require.NoError(t, db.Create(record).Error)

	claimed, err := repo.ClaimReconciliation(
		record.ID,
		record.RequestHash,
		"reconciler-1",
		now,
		now.Add(time.Minute),
	)
	require.NoError(t, err)
	require.True(t, claimed)

	var reconciling payment.PaymentOperationIdempotency
	require.NoError(t, db.First(&reconciling, record.ID).Error)
	require.Equal(t, payment.PaymentOperationIdempotencyReconciling, reconciling.Status)
	require.Equal(t, "reconciler-1", reconciling.ClaimToken)
	require.NotNil(t, reconciling.ReconciliationStartedAt)

	require.NoError(t, repo.ReleaseClaim(record.ID, "reconciler-1", now.Add(time.Second)))
	claimed, err = repo.ClaimReconciliation(
		record.ID,
		record.RequestHash,
		"reconciler-2",
		now.Add(2*time.Second),
		now.Add(2*time.Minute),
	)
	require.NoError(t, err)
	require.True(t, claimed)

	var reclaimed payment.PaymentOperationIdempotency
	require.NoError(t, db.First(&reclaimed, record.ID).Error)
	require.Equal(t, payment.PaymentOperationIdempotencyReconciling, reclaimed.Status)
	require.Equal(t, "reconciler-2", reclaimed.ClaimToken)
}

func newPaymentOperationIdempotencyRepositoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&payment.PaymentOperationIdempotency{}))
	return db
}
