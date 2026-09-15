package repository

import (
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/payment"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMarkPaymentExpiredReportsWhetherPendingUnpaidOrderWasClaimed(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&order.Order{}, &payment.PaymentReview{}, &payment.StripeDispute{}, &payment.PayPalDispute{}))

	repo := NewOrderRepository(db)
	orderRecord := order.Order{
		OrderNumber:   "ORD-MARK-PAYMENT-EXPIRED",
		UserID:        42,
		Status:        "pending",
		PaymentStatus: "unpaid",
		TotalAmount:   100,
		Currency:      "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	expiredAt := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	claimed, err := repo.MarkPaymentExpired(orderRecord.ID, expiredAt)
	require.NoError(t, err)
	require.True(t, claimed)

	claimed, err = repo.MarkPaymentExpired(orderRecord.ID, expiredAt.Add(time.Minute))
	require.NoError(t, err)
	require.False(t, claimed)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.Equal(t, "payment_expired", savedOrder.Status)
	require.Equal(t, "expired", savedOrder.PaymentStatus)
	require.Equal(t, expiredAt, savedOrder.CancelledAt.UTC())
}

func TestUpdateStatusUsesExpectedStatusAsAtomicCAS(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&order.Order{}))

	repo := NewOrderRepository(db)
	orderRecord := order.Order{
		OrderNumber:   "ORD-STATUS-CAS",
		UserID:        42,
		Status:        "pending",
		PaymentStatus: "unpaid",
		TotalAmount:   100,
		Currency:      "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	require.NoError(t, repo.UpdateStatus(orderRecord.ID, "pending", "processing"))
	err = repo.UpdateStatus(orderRecord.ID, "pending", "completed")
	require.ErrorIs(t, err, ErrOrderStatusConflict)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.Equal(t, "processing", savedOrder.Status)
}

func TestReleasePaymentLiabilityReviewHoldRequiresNoActiveReviewOrDispute(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&order.Order{}, &payment.PaymentReview{}, &payment.StripeDispute{}, &payment.PayPalDispute{}))

	repo := NewOrderRepository(db)
	orderRecord := order.Order{
		OrderNumber:     "ORD-RELEASE-LIABILITY-HOLD",
		UserID:          42,
		Status:          "needs_review",
		PaymentStatus:   "paid",
		FulfillmentHold: true,
		TotalAmount:     2500,
		Currency:        "USD",
	}
	require.NoError(t, db.Create(&orderRecord).Error)

	review := payment.PaymentReview{
		OrderID: &orderRecord.ID,
		Status:  "approved",
		Reason:  "high_value_liability_shift_not_transferred",
		Source:  "gateway_verification",
	}
	require.NoError(t, db.Create(&review).Error)

	activeReview := payment.PaymentReview{
		OrderID: &orderRecord.ID,
		Status:  "pending",
		Reason:  "manual_payment_review",
		Source:  "operator",
	}
	require.NoError(t, db.Create(&activeReview).Error)

	require.NoError(t, repo.ReleasePaymentLiabilityReviewHold(orderRecord.ID))

	var held order.Order
	require.NoError(t, db.First(&held, orderRecord.ID).Error)
	require.Equal(t, "needs_review", held.Status)
	require.True(t, held.FulfillmentHold)

	activeReview.Status = "approved"
	require.NoError(t, db.Save(&activeReview).Error)
	require.NoError(t, repo.ReleasePaymentLiabilityReviewHold(orderRecord.ID))

	var released order.Order
	require.NoError(t, db.First(&released, orderRecord.ID).Error)
	require.Equal(t, "processing", released.Status)
	require.False(t, released.FulfillmentHold)
}

func TestSoftDeleteUnpaidCancelledOrPaymentExpiredOrderRecordGuardsFinancialOrders(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&order.Order{}))

	repo := NewOrderRepository(db)
	testCases := []struct {
		name          string
		status        string
		paymentStatus string
		canHide       bool
	}{
		{name: "cancelled unpaid", status: "cancelled", paymentStatus: "unpaid", canHide: true},
		{name: "payment expired", status: "payment_expired", paymentStatus: "expired", canHide: true},
		{name: "paid", status: "paid", paymentStatus: "paid", canHide: false},
		{name: "refunded", status: "refunded", paymentStatus: "refunded", canHide: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			orderRecord := order.Order{
				OrderNumber:   "ORD-REPOSITORY-HIDE-" + strings.ReplaceAll(testCase.name, " ", "-"),
				UserID:        42,
				Status:        testCase.status,
				PaymentStatus: testCase.paymentStatus,
				TotalAmount:   100,
				Currency:      "USD",
			}
			require.NoError(t, db.Create(&orderRecord).Error)

			hidden, err := repo.SoftDeleteUnpaidCancelledOrPaymentExpiredOrderRecord(orderRecord.ID)
			require.NoError(t, err)
			require.Equal(t, testCase.canHide, hidden)

			var retainedOrder order.Order
			require.NoError(t, db.Unscoped().First(&retainedOrder, orderRecord.ID).Error)
			require.Equal(t, testCase.canHide, retainedOrder.DeletedAt.Valid)
		})
	}
}
