package service

import (
	"testing"
	"time"

	currencydomain "commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	orderdomain "commerce-platform/internal/domain/order"

	"github.com/stretchr/testify/require"
)

func TestCalculateRefundFXGainLossUsesProviderSettlementDeduction(t *testing.T) {
	snapshot := currencydomain.OrderFXSnapshot{
		Version:       currencydomain.OrderFXSnapshotVersion,
		BaseCurrency:  "USD",
		OrderCurrency: "EUR",
		RateDecimal:   "0.9523809523809523809523809524",
		Source:        "test",
		CapturedAt:    time.Now().UTC(),
	}
	refund := domainmoney.MustNew(100_000, "EUR")

	loss, lossCurrency, err := calculateRefundFXGainLoss(snapshot, refund, 115_000, "USD")
	require.NoError(t, err)
	require.Equal(t, int64(10_000), loss)
	require.Equal(t, "USD", lossCurrency)

	gain, gainCurrency, err := calculateRefundFXGainLoss(snapshot, refund, 95_000, "USD")
	require.NoError(t, err)
	require.Equal(t, int64(-10_000), gain)
	require.Equal(t, "USD", gainCurrency)
}

func TestBackfillHistoricalRefundFXSnapshotIsAuditedAndImmutable(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	createdAt := time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC)
	orderRecord := &orderdomain.Order{OrderNumber: "ORD-FX-BACKFILL", Currency: "EUR", Status: "paid", PaymentStatus: "paid", CreatedAt: createdAt}
	require.NoError(t, db.Create(orderRecord).Error)

	result, err := paymentService.BackfillHistoricalRefundFXSnapshot(BackfillHistoricalRefundFXSnapshotInput{
		OrderID: orderRecord.ID, BaseCurrency: "USD", OrderCurrency: "EUR", RateDecimal: "0.9234", Source: "provider_statement", CapturedAt: createdAt,
	})
	require.NoError(t, err)
	require.Equal(t, "0.9234", result.Snapshot.RateDecimal)

	var stored orderdomain.Order
	require.NoError(t, db.First(&stored, orderRecord.ID).Error)
	storedSnapshot, err := currencydomain.ParseOrderFXSnapshot(stored.FXSnapshotData)
	require.NoError(t, err)
	require.Equal(t, "provider_statement", storedSnapshot.Source)

	_, err = paymentService.BackfillHistoricalRefundFXSnapshot(BackfillHistoricalRefundFXSnapshotInput{
		OrderID: orderRecord.ID, BaseCurrency: "USD", OrderCurrency: "EUR", RateDecimal: "0.8", Source: "replacement", CapturedAt: createdAt,
	})
	require.ErrorIs(t, err, ErrHistoricalRefundFXSnapshotAlreadyPresent)
}

func TestBackfillHistoricalRefundFXSnapshotRejectsUSDOrders(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := &orderdomain.Order{OrderNumber: "ORD-FX-USD", Currency: "USD", Status: "paid", PaymentStatus: "paid"}
	require.NoError(t, db.Create(orderRecord).Error)

	_, err := paymentService.BackfillHistoricalRefundFXSnapshot(BackfillHistoricalRefundFXSnapshotInput{
		OrderID: orderRecord.ID, BaseCurrency: "USD", OrderCurrency: "USD", RateDecimal: "1", Source: "provider_statement", CapturedAt: time.Now().UTC(),
	})
	require.ErrorIs(t, err, ErrHistoricalRefundFXSnapshotNotApplicable)
}

func TestBackfillHistoricalRefundFXSnapshotRejectsCurrencyMismatch(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := &orderdomain.Order{OrderNumber: "ORD-FX-MISMATCH", Currency: "EUR", Status: "paid", PaymentStatus: "paid"}
	require.NoError(t, db.Create(orderRecord).Error)

	_, err := paymentService.BackfillHistoricalRefundFXSnapshot(BackfillHistoricalRefundFXSnapshotInput{
		OrderID: orderRecord.ID, BaseCurrency: "USD", OrderCurrency: "GBP", RateDecimal: "0.8", Source: "provider_statement", CapturedAt: time.Now().UTC(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not match")
}

func TestCalculateRefundFXGainLossSkipsIncomparableSettlementCurrency(t *testing.T) {
	snapshot := currencydomain.OrderFXSnapshot{
		Version:       currencydomain.OrderFXSnapshotVersion,
		BaseCurrency:  "USD",
		OrderCurrency: "EUR",
		RateDecimal:   "0.9523809523809523809523809524",
		Source:        "test",
		CapturedAt:    time.Now().UTC(),
	}

	loss, lossCurrency, err := calculateRefundFXGainLoss(snapshot, domainmoney.MustNew(100_000, "EUR"), 115_000, "GBP")
	require.NoError(t, err)
	require.Zero(t, loss)
	require.Empty(t, lossCurrency)
}
