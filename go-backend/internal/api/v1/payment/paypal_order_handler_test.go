package payment

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	orderdomain "tanzanite/internal/domain/order"
	paymentdomain "tanzanite/internal/domain/payment"
	pgateway "tanzanite/internal/pkg/payment"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreatePayPalOrderRecordsPendingAttempt(t *testing.T) {
	db, handler := newPayPalHandlerTestHarness(t, &fakePaymentGateway{
		createResponse: &pgateway.PaymentResponse{
			ID:            "PAYPAL-ORDER-1",
			Status:        "CREATED",
			Amount:        84,
			Currency:      "USD",
			PaymentURL:    "https://paypal.example/approve",
			TransactionID: "PAYPAL-ORDER-1",
			Metadata:      map[string]string{"order_id": "ORD-PAYPAL-1"},
		},
	})
	seedPayPalOrder(t, db, "ORD-PAYPAL-1", 7, 84, "pending", "unpaid")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/paypal/orders",
		bytes.NewBufferString(`{"order_number":"ORD-PAYPAL-1","return_url":"https://shop.example/paypal/return","cancel_url":"https://shop.example/cart"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreatePayPalOrder(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var transaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "PAYPAL-ORDER-1").First(&transaction).Error)
	require.Equal(t, "paypal", transaction.PaymentMethod)
	require.Equal(t, "pending", transaction.Status)
	require.InDelta(t, 84, transaction.Amount, 0.001)
}

func TestCapturePayPalOrderMarksMatchingOrderPaid(t *testing.T) {
	db, handler := newPayPalHandlerTestHarness(t, &fakePaymentGateway{
		captureResponse: &pgateway.PaymentResponse{
			ID:            "PAYPAL-ORDER-2",
			Status:        "COMPLETED",
			Amount:        92,
			Currency:      "USD",
			TransactionID: "PAYPAL-CAPTURE-2",
			Metadata:      map[string]string{"order_id": "ORD-PAYPAL-2", "paypal_order_id": "PAYPAL-ORDER-2"},
		},
	})
	orderRecord := seedPayPalOrder(t, db, "ORD-PAYPAL-2", 7, 92, "pending", "unpaid")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Params = gin.Params{{Key: "paypal_order_id", Value: "PAYPAL-ORDER-2"}}
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/paypal/orders/PAYPAL-ORDER-2/capture",
		bytes.NewBufferString(`{"order_number":"ORD-PAYPAL-2"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CapturePayPalOrder(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.Equal(t, "paid", savedOrder.PaymentStatus)
	require.Equal(t, "processing", savedOrder.Status)

	var transaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "PAYPAL-CAPTURE-2").First(&transaction).Error)
	require.Equal(t, "completed", transaction.Status)
	require.NotNil(t, transaction.CompletedAt)
}

func TestCapturePayPalOrderRejectsMismatchedProviderOrderMetadata(t *testing.T) {
	db, handler := newPayPalHandlerTestHarness(t, &fakePaymentGateway{
		captureResponse: &pgateway.PaymentResponse{
			ID:            "PAYPAL-ORDER-3",
			Status:        "COMPLETED",
			Amount:        60,
			Currency:      "USD",
			TransactionID: "PAYPAL-CAPTURE-3",
			Metadata:      map[string]string{"order_id": "OTHER-ORDER"},
		},
	})
	orderRecord := seedPayPalOrder(t, db, "ORD-PAYPAL-3", 7, 60, "pending", "unpaid")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Params = gin.Params{{Key: "paypal_order_id", Value: "PAYPAL-ORDER-3"}}
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/paypal/orders/PAYPAL-ORDER-3/capture",
		bytes.NewBufferString(`{"order_number":"ORD-PAYPAL-3"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CapturePayPalOrder(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.Equal(t, "unpaid", savedOrder.PaymentStatus)
}

type fakePaymentGateway struct {
	createResponse  *pgateway.PaymentResponse
	captureResponse *pgateway.PaymentResponse
	getResponse     *pgateway.PaymentResponse
	createRequest   *pgateway.PaymentRequest
}

func (g *fakePaymentGateway) CreatePayment(_ context.Context, req *pgateway.PaymentRequest) (*pgateway.PaymentResponse, error) {
	g.createRequest = req
	return g.createResponse, nil
}

func (g *fakePaymentGateway) CapturePayment(context.Context, string) (*pgateway.PaymentResponse, error) {
	return g.captureResponse, nil
}

func (g *fakePaymentGateway) RefundPayment(context.Context, string, float64) (*pgateway.RefundResponse, error) {
	return nil, nil
}

func (g *fakePaymentGateway) RefundPaymentWithOptions(context.Context, string, float64, pgateway.RefundOptions) (*pgateway.RefundResponse, error) {
	return nil, nil
}

func (g *fakePaymentGateway) GetPayment(context.Context, string) (*pgateway.PaymentResponse, error) {
	return g.getResponse, nil
}

func (g *fakePaymentGateway) VerifyWebhook([]byte, string) (bool, error) {
	return false, nil
}

func newPayPalHandlerTestHarness(t *testing.T, gateway pgateway.PaymentGateway) (*gorm.DB, *Handler) {
	t.Helper()
	t.Setenv("PAYPAL_CLIENT_ID", "paypal-client")
	t.Setenv("PAYPAL_SECRET", "paypal-secret")
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&orderdomain.Order{}, &orderdomain.OrderItem{}, &paymentdomain.Transaction{}))

	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	txManager := repository.NewTxManager(
		db,
		orderRepo,
		repository.NewProductRepository(db),
		repository.NewCouponRepository(db),
		repository.NewLoyaltyRepository(db),
		paymentRepo,
	)
	paymentService := service.NewPaymentService(txManager, paymentRepo)
	orderService := service.NewOrderService(nil, orderRepo, nil, nil, nil)
	handler := NewHandler(paymentService, orderService, nil, nil, nil, nil, nil, nil, nil, nil)
	handler.gatewayFactory = func(*pgateway.Config) (pgateway.PaymentGateway, error) {
		return gateway, nil
	}
	return db, handler
}

func seedPayPalOrder(t *testing.T, db *gorm.DB, orderNumber string, userID uint, total float64, status string, paymentStatus string) orderdomain.Order {
	t.Helper()
	orderRecord := orderdomain.Order{
		OrderNumber:    orderNumber,
		UserID:         userID,
		Status:         status,
		PaymentMethod:  "paypal",
		PaymentStatus:  paymentStatus,
		ShippingMethod: "standard",
		ShippingStatus: "pending",
		TotalAmount:    total,
		Currency:       "USD",
		ShippingAddress: orderdomain.Address{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Country:   "US",
			Email:     "ada@example.com",
			Phone:     "+15551234567",
		},
		BillingAddress: orderdomain.Address{
			Country: "US",
		},
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	return orderRecord
}
