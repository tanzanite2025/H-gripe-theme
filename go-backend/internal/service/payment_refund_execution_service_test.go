package service

import (
	"context"
	"errors"
	"testing"
	"time"

	currencydomain "commerce-platform/internal/domain/currency"
	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestPaymentServiceExecutePendingRefundCompletesLocalRefund(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 120)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 80)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_gateway_1",
			PaymentID: transaction.TransactionID,
			Amount:    80,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	completedRefund, execution, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.NoError(t, err)

	require.Equal(t, "completed", completedRefund.Status)
	require.NotNil(t, completedRefund.RefundID)
	require.Equal(t, "re_gateway_1", *completedRefund.RefundID)
	require.NotNil(t, completedRefund.CompletedAt)
	require.Contains(t, completedRefund.GatewayResponse, "re_gateway_1")
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, execution.Status)
	require.Equal(t, "re_gateway_1", execution.ProviderRefundID)
	require.Equal(t, "succeeded", execution.ProviderStatus)
	require.Equal(t, refundExecutionIdempotencyKey(refund.ID), gateway.options.IdempotencyKey)
	require.Equal(t, transaction.Currency, gateway.options.Currency)
	require.Equal(t, transaction.Amount, gateway.options.OriginalAmount)
	require.Equal(t, orderRecord.OrderNumber, gateway.options.MerchantOrderNumber)
	require.Equal(t, transaction.TransactionID, gateway.options.ProviderTransactionID)
	require.Equal(t, transaction.TransactionID, gateway.paymentID)
	require.Equal(t, 80.0, gateway.amount)
	require.Equal(t, orderRecord.OrderNumber, execution.MerchantOrderNumber)
	require.Equal(t, transaction.TransactionID, execution.ProviderTransactionID)

	updatedTransaction, err := repository.NewPaymentRepository(db).FindTransactionByID(transaction.ID)
	require.NoError(t, err)
	require.Equal(t, "completed", updatedTransaction.Status)
}

func TestPaymentServiceExecutePendingRefundRecordsGatewayFailure(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)
	orderRecord, transaction := createRefundRecommendationPaidOrder(t, db, 90)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 40)
	gateway := &recordingRefundGateway{err: errors.New("gateway unavailable")}

	_, execution, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "gateway unavailable")
	require.NotNil(t, execution)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusFailed, execution.Status)
	require.Contains(t, execution.ErrorMessage, "gateway unavailable")

	storedRefund, err := repository.NewPaymentRepository(db).FindRefundByID(refund.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", storedRefund.Status)
	require.Nil(t, storedRefund.RefundID)
	require.Nil(t, storedRefund.CompletedAt)
}

func TestPaymentServiceExecutePendingRefundPersistsHistoricalFXSnapshot(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)

	fetchedAt := time.Date(2026, time.August, 6, 9, 0, 0, 0, time.UTC)
	capturedAt := time.Date(2026, time.August, 6, 10, 0, 0, 0, time.UTC)
	fxSnapshot := currencydomain.OrderFXSnapshot{
		Version:         currencydomain.OrderFXSnapshotVersion,
		BaseCurrency:    "USD",
		OrderCurrency:   "EUR",
		BaseToOrderRate: 0.9,
		Source:          "historical_order_quote",
		CapturedAt:      capturedAt,
		RateFetchedAt:   &fetchedAt,
	}

	orderRecord := orderdomain.Order{
		OrderNumber:     "ORD-RFX-1",
		UserID:          11,
		Status:          "paid",
		PaymentStatus:   "paid",
		TotalAmount:     90,
		Currency:        "EUR",
		FXSnapshotData:  currencydomain.OrderFXSnapshotJSON(fxSnapshot),
		ShippingAddress: orderdomain.Address{Country: "DE"},
		BillingAddress:  orderdomain.Address{Country: "DE"},
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	transaction := paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "pi_rfx_1",
		PaymentMethod: "stripe",
		Amount:        90,
		Currency:      "EUR",
		Status:        "completed",
	}
	require.NoError(t, db.Create(&transaction).Error)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 30)
	gateway := &recordingRefundGateway{
		response: &pgateway.RefundResponse{
			ID:        "re_rfx_1",
			PaymentID: transaction.TransactionID,
			Amount:    30,
			Status:    "succeeded",
			CreatedAt: time.Now().UTC(),
		},
	}

	completedRefund, execution, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.NoError(t, err)
	require.Equal(t, paymentdomain.PaymentRefundExecutionStatusSucceeded, execution.Status)
	require.Equal(t, "completed", completedRefund.Status)

	storedRefund, err := repository.NewPaymentRepository(db).FindRefundByID(refund.ID)
	require.NoError(t, err)
	persistedSnapshot, err := currencydomain.ParseOrderFXSnapshot(storedRefund.FXSnapshotData)
	require.NoError(t, err)
	require.Equal(t, fxSnapshot.BaseCurrency, persistedSnapshot.BaseCurrency)
	require.Equal(t, fxSnapshot.OrderCurrency, persistedSnapshot.OrderCurrency)
	require.InDelta(t, fxSnapshot.BaseToOrderRate, persistedSnapshot.BaseToOrderRate, 0.0001)
	require.Equal(t, fxSnapshot.Source, persistedSnapshot.Source)
	require.NotNil(t, persistedSnapshot.RateFetchedAt)
	require.True(t, fetchedAt.Equal(*persistedSnapshot.RateFetchedAt))
}

func TestPaymentServiceExecutePendingRefundRejectsMissingHistoricalFXSnapshot(t *testing.T) {
	db := newPaymentRefundRecommendationTestDB(t)
	service := newPaymentServiceWithRefundExecution(db)

	orderRecord := orderdomain.Order{
		OrderNumber:   "ORD-RFX-MISSING",
		UserID:        11,
		Status:        "paid",
		PaymentStatus: "paid",
		TotalAmount:   90,
		Currency:      "EUR",
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	transaction := paymentdomain.Transaction{
		OrderID:       orderRecord.ID,
		TransactionID: "pi_rfx_missing",
		PaymentMethod: "stripe",
		Amount:        90,
		Currency:      "EUR",
		Status:        "completed",
	}
	require.NoError(t, db.Create(&transaction).Error)
	refund := createPendingRefundRecord(t, db, orderRecord.ID, transaction.ID, 30)
	gateway := &recordingRefundGateway{}

	_, _, err := service.ExecutePendingRefund(context.Background(), ExecutePendingRefundInput{
		RefundID: refund.ID,
		AdminID:  12,
		Provider: "stripe",
		Gateway:  gateway,
	})
	require.ErrorIs(t, err, ErrHistoricalRefundFXSnapshotMissing)
	require.Empty(t, gateway.paymentID)

	storedRefund, err := repository.NewPaymentRepository(db).FindRefundByID(refund.ID)
	require.NoError(t, err)
	require.Equal(t, "pending", storedRefund.Status)
	require.Equal(t, "{}", string(storedRefund.FXSnapshotData))
}

func newPaymentServiceWithRefundExecution(db *gorm.DB) *PaymentService {
	paymentRepo := repository.NewPaymentRepository(db)
	txManager := repository.NewTxManager(
		db,
		repository.NewOrderRepository(db),
		repository.NewProductRepository(db),
		repository.NewCouponRepository(db),
		repository.NewLoyaltyRepository(db),
		paymentRepo,
	)
	txManager.ConfigurePaymentRefundExecutionRepository(repository.NewPaymentRefundExecutionRepository(db))
	return NewPaymentService(txManager, paymentRepo)
}

func createPendingRefundRecord(t *testing.T, db *gorm.DB, orderID uint, transactionID uint, amount float64) paymentdomain.Refund {
	t.Helper()

	refund := paymentdomain.Refund{
		OrderID:         orderID,
		TransactionID:   transactionID,
		Amount:          amount,
		RequestedAmount: amount,
		Status:          "pending",
		Reason:          "manual pending refund",
		RefundedBy:      7,
	}
	require.NoError(t, db.Create(&refund).Error)
	return refund
}

type recordingRefundGateway struct {
	paymentID string
	amount    float64
	options   pgateway.RefundOptions
	response  *pgateway.RefundResponse
	err       error
}

func (g *recordingRefundGateway) CreatePayment(ctx context.Context, req *pgateway.PaymentRequest) (*pgateway.PaymentResponse, error) {
	return nil, errors.New("not implemented")
}

func (g *recordingRefundGateway) CapturePayment(ctx context.Context, paymentID string) (*pgateway.PaymentResponse, error) {
	return nil, errors.New("not implemented")
}

func (g *recordingRefundGateway) RefundPayment(ctx context.Context, paymentID string, amount float64) (*pgateway.RefundResponse, error) {
	return g.RefundPaymentWithOptions(ctx, paymentID, amount, pgateway.RefundOptions{})
}

func (g *recordingRefundGateway) RefundPaymentWithOptions(ctx context.Context, paymentID string, amount float64, options pgateway.RefundOptions) (*pgateway.RefundResponse, error) {
	g.paymentID = paymentID
	g.amount = amount
	g.options = options
	if g.err != nil {
		return nil, g.err
	}
	return g.response, nil
}

func (g *recordingRefundGateway) GetPayment(ctx context.Context, paymentID string) (*pgateway.PaymentResponse, error) {
	return nil, errors.New("not implemented")
}

func (g *recordingRefundGateway) VerifyWebhook(payload []byte, signature string) (bool, error) {
	return false, errors.New("not implemented")
}
