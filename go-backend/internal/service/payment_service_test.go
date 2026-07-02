package service

import (
	"testing"
	"time"

	"tanzanite/internal/domain/order"
	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRecordVerifiedGatewayPaymentCreatesLedgerAndMarksOrderPaid(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-1", 84, "pending", "unpaid")

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_123",
		Amount:        84,
		Currency:      "USD",
	})

	require.NoError(t, err)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "txn_123").First(&savedTransaction).Error)
	assert.Equal(t, orderRecord.ID, savedTransaction.OrderID)
	assert.Equal(t, "stripe", savedTransaction.PaymentMethod)
	assert.Equal(t, "completed", savedTransaction.Status)
	assert.NotNil(t, savedTransaction.CompletedAt)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "paid", savedOrder.PaymentStatus)
	assert.Equal(t, "processing", savedOrder.Status)
	assert.NotNil(t, savedOrder.PaidAt)

	require.NoError(t, paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_123",
		Amount:        84,
		Currency:      "USD",
	}))

	var transactionCount int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).Where("transaction_id = ?", "txn_123").Count(&transactionCount).Error)
	assert.Equal(t, int64(1), transactionCount)
}

func TestRecordVerifiedGatewayPaymentRejectsAmountMismatch(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-PAY-2", 84, "pending", "unpaid")

	err := paymentService.RecordVerifiedGatewayPayment(VerifiedGatewayPaymentInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: "txn_bad_amount",
		Amount:        83,
		Currency:      "USD",
	})

	require.Error(t, err)

	var transactionCount int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).Where("transaction_id = ?", "txn_bad_amount").Count(&transactionCount).Error)
	assert.Equal(t, int64(0), transactionCount)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "unpaid", savedOrder.PaymentStatus)
	assert.Equal(t, "pending", savedOrder.Status)
}

func TestCreateAdminRefundReservesPendingAmount(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-1", 100, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_1", 100, "USD")

	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        60,
		Reason:        "customer request",
	}
	require.NoError(t, paymentService.CreateAdminRefund(&refund, 7))

	assert.Equal(t, "pending", refund.Status)
	assert.Equal(t, uint(7), refund.RefundedBy)
	assert.Nil(t, refund.RefundID)

	excessRefund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        50,
	}
	require.Error(t, paymentService.CreateAdminRefund(&excessRefund, 7))

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Where("transaction_id = ?", transaction.ID).Count(&refundCount).Error)
	assert.Equal(t, int64(1), refundCount)
}

func TestRecordVerifiedGatewayRefundCompletesPendingRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-2", 84, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_2", 84, "USD")
	refund := paymentdomain.Refund{
		OrderID:       orderRecord.ID,
		TransactionID: transaction.ID,
		Amount:        84,
	}
	require.NoError(t, paymentService.CreateAdminRefund(&refund, 7))

	err := paymentService.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: transaction.TransactionID,
		RefundID:      "rf_123",
		Amount:        84,
		Currency:      "USD",
	})
	require.NoError(t, err)

	var savedRefund paymentdomain.Refund
	require.NoError(t, db.First(&savedRefund, refund.ID).Error)
	assert.Equal(t, "completed", savedRefund.Status)
	require.NotNil(t, savedRefund.RefundID)
	assert.Equal(t, "rf_123", *savedRefund.RefundID)
	assert.NotNil(t, savedRefund.CompletedAt)

	var savedTransaction paymentdomain.Transaction
	require.NoError(t, db.First(&savedTransaction, transaction.ID).Error)
	assert.Equal(t, "refunded", savedTransaction.Status)

	var savedOrder order.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	assert.Equal(t, "refunded", savedOrder.PaymentStatus)
	assert.Equal(t, "refunded", savedOrder.Status)

	require.NoError(t, paymentService.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: transaction.TransactionID,
		RefundID:      "rf_123",
		Amount:        84,
		Currency:      "USD",
	}))

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Where("refund_id = ?", "rf_123").Count(&refundCount).Error)
	assert.Equal(t, int64(1), refundCount)
}

func TestRecordVerifiedGatewayRefundRejectsOverRefund(t *testing.T) {
	db, paymentService := newTestPaymentService(t)
	orderRecord := seedPaymentOrder(t, db, "ORD-REF-3", 100, "processing", "paid")
	transaction := seedCompletedTransaction(t, db, orderRecord.ID, "txn_ref_3", 100, "USD")

	err := paymentService.RecordVerifiedGatewayRefund(VerifiedGatewayRefundInput{
		Provider:      "stripe",
		OrderNumber:   orderRecord.OrderNumber,
		TransactionID: transaction.TransactionID,
		RefundID:      "rf_too_much",
		Amount:        101,
		Currency:      "USD",
	})

	require.Error(t, err)

	var refundCount int64
	require.NoError(t, db.Model(&paymentdomain.Refund{}).Where("refund_id = ?", "rf_too_much").Count(&refundCount).Error)
	assert.Equal(t, int64(0), refundCount)
}

func newTestPaymentService(t *testing.T) (*gorm.DB, *PaymentService) {
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
		&order.Order{},
		&paymentdomain.Transaction{},
		&paymentdomain.Refund{},
	))

	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	txManager := repository.NewTxManager(db, orderRepo, productRepo, couponRepo, loyaltyRepo, paymentRepo)

	return db, NewPaymentService(txManager, paymentRepo)
}

func seedPaymentOrder(t *testing.T, db *gorm.DB, orderNumber string, total float64, status, paymentStatus string) order.Order {
	t.Helper()

	record := order.Order{
		OrderNumber:   orderNumber,
		UserID:        42,
		Status:        status,
		PaymentStatus: paymentStatus,
		TotalAmount:   total,
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}

func seedCompletedTransaction(t *testing.T, db *gorm.DB, orderID uint, transactionID string, amount float64, currency string) paymentdomain.Transaction {
	t.Helper()

	completedAt := time.Now()
	record := paymentdomain.Transaction{
		OrderID:       orderID,
		TransactionID: transactionID,
		PaymentMethod: "stripe",
		Amount:        amount,
		Currency:      currency,
		Status:        "completed",
		CompletedAt:   &completedAt,
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}
