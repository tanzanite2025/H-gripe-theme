package service

import (
	"testing"

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
