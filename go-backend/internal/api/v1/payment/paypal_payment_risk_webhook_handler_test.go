package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/orderevidence"
	paymentdomain "commerce-platform/internal/domain/payment"
	shippingdomain "commerce-platform/internal/domain/shipping"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	paypalapi "github.com/plutov/paypal/v4"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPayPalDisputeHelpersReadTransactionAmountAndTimestamp(t *testing.T) {
	resource := map[string]interface{}{
		"create_time": "2026-07-31T10:15:00Z",
		"dispute_amount": map[string]interface{}{
			"value":         "249.90",
			"currency_code": "USD",
		},
		"disputed_transactions": []interface{}{
			map[string]interface{}{
				"seller_transaction_id": "7AB12345CD678901E",
				"invoice_id":            "ORD-PAYPAL-1",
			},
		},
	}

	amount, currency := paypalDisputeAmount(resource)

	require.InDelta(t, 249.90, amount, 0.000001)
	require.Equal(t, "USD", currency)
	require.Equal(t, "7AB12345CD678901E", paypalDisputePaymentID(resource))
	require.Equal(t, "ORD-PAYPAL-1", paypalDisputeOrderReference(resource))
	require.Equal(
		t,
		time.Date(2026, time.July, 31, 10, 15, 0, 0, time.UTC),
		paypalRiskOccurredAt(resource),
	)
}

func TestRecordPayPalDisputeRiskEventOnlyRecordsAndQueuesManualEvidenceReview(t *testing.T) {
	db, handler, submitter := newPayPalDisputeWebhookHarness(t)
	orderRecord := seedPayPalDisputeWebhookOrder(t, db, "ORD-PAYPAL-WEBHOOK-1", 249.90, "DHL999")
	seedPayPalDisputeWebhookTransaction(t, db, orderRecord.ID, "PAYPAL-CAPTURE-WEBHOOK-1", 249.90)
	seedPayPalDisputeWebhookTracking(t, db, orderRecord.ID, "DHL999")

	resource := map[string]interface{}{
		"dispute_id":               "PP-D-WEBHOOK-1",
		"reason":                   "MERCHANDISE_OR_SERVICE_NOT_RECEIVED",
		"status":                   "WAITING_FOR_SELLER_RESPONSE",
		"dispute_state":            "REQUIRED_ACTION",
		"dispute_life_cycle_stage": "CHARGEBACK",
		"create_time":              "2026-07-31T10:15:00Z",
		"dispute_amount": map[string]interface{}{
			"value":         "249.90",
			"currency_code": "USD",
		},
		"disputed_transactions": []interface{}{
			map[string]interface{}{
				"seller_transaction_id": "PAYPAL-CAPTURE-WEBHOOK-1",
				"invoice_id":            orderRecord.OrderNumber,
			},
		},
	}
	resourceBytes, err := json.Marshal(resource)
	require.NoError(t, err)
	event := pgateway.PayPalWebhookEvent{
		ID:        "WH-PAYPAL-DISPUTE-1",
		EventType: "CUSTOMER.DISPUTE.CREATED",
		Resource:  resourceBytes,
	}
	payload, err := json.Marshal(event)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/v1/payment/webhooks/paypal", nil)

	handled, err := handler.recordPayPalDisputeRiskEvent(context, event, payload)

	require.NoError(t, err)
	require.True(t, handled)
	require.Empty(t, submitter.disputeID)
	require.Nil(t, submitter.params)

	submitted, exists := context.Get("paypal_dispute_evidence_submitted")
	require.True(t, exists)
	require.Equal(t, false, submitted)
	pendingReview, exists := context.Get("paypal_dispute_evidence_pending_review")
	require.True(t, exists)
	require.Equal(t, true, pendingReview)
	orderID, exists := context.Get("paypal_dispute_order_id")
	require.True(t, exists)
	require.Equal(t, orderRecord.ID, orderID)

	var dispute paymentdomain.PayPalDispute
	require.NoError(t, db.Where("paypal_dispute_id = ?", "PP-D-WEBHOOK-1").First(&dispute).Error)
	require.NotNil(t, dispute.OrderID)
	require.Equal(t, orderRecord.ID, *dispute.OrderID)
	require.NotNil(t, dispute.TransactionID)
	require.Nil(t, dispute.EvidenceSubmittedAt)
	require.Empty(t, dispute.EvidenceSubmissionError)
	require.Empty(t, dispute.EvidenceSubmissionPayload)

	var snapshotCount int64
	require.NoError(t, db.Model(&orderevidence.OrderEvidenceSubmissionSnapshot{}).
		Where("provider = ? AND dispute_id = ?", "paypal", dispute.ID).
		Count(&snapshotCount).Error)
	require.Equal(t, int64(0), snapshotCount)
}

type fakePayPalWebhookDisputeEvidenceSubmitter struct {
	disputeID string
	params    *paypalapi.DisputeProvideEvidenceParams
}

func (f *fakePayPalWebhookDisputeEvidenceSubmitter) ProvideEvidence(_ context.Context, disputeID string, params *paypalapi.DisputeProvideEvidenceParams) error {
	f.disputeID = disputeID
	f.params = params
	return nil
}

func newPayPalDisputeWebhookHarness(t *testing.T) (*gorm.DB, *Handler, *fakePayPalWebhookDisputeEvidenceSubmitter) {
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
	require.NoError(t, db.AutoMigrate(
		&orderdomain.Order{},
		&orderdomain.OrderItem{},
		&paymentdomain.Transaction{},
		&paymentdomain.Refund{},
		&paymentdomain.RefundLineItem{},
		&paymentdomain.PayPalDispute{},
		&orderevidence.OrderEvidenceSubmissionSnapshot{},
		&shippingdomain.TrackingProviderConfig{},
		&shippingdomain.TrackingShipment{},
		&shippingdomain.TrackingEvent{},
	))

	orderRepo := repository.NewOrderRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	shippingRepo := repository.NewShippingRepository(db)
	orderEvidenceSubmissionRepo := repository.NewOrderEvidenceSubmissionSnapshotRepository(db)
	paymentService := service.NewPaymentService(nil, paymentRepo)
	paymentService.ConfigureEvidenceSources(orderRepo, nil)
	paymentService.ConfigureOrderEvidenceSubmissionSnapshotRepository(orderEvidenceSubmissionRepo)
	paymentService.ConfigureOrderEvidenceAssembler(
		service.NewOrderEvidencePackageAssembler(orderRepo, nil, shippingRepo),
	)
	submitter := &fakePayPalWebhookDisputeEvidenceSubmitter{}
	paymentService.ConfigurePayPalDisputeEvidenceSubmitter(submitter)
	handler := NewHandler(paymentService, nil, nil, nil, nil, nil, nil, nil, nil)
	return db, handler, submitter
}

func seedPayPalDisputeWebhookOrder(t *testing.T, db *gorm.DB, orderNumber string, total float64, trackingNumber string) orderdomain.Order {
	t.Helper()
	paidAt := time.Date(2026, time.July, 29, 9, 0, 0, 0, time.UTC)
	shippedAt := time.Date(2026, time.July, 30, 9, 0, 0, 0, time.UTC)
	variantID := uint(1)
	orderRecord := orderdomain.Order{
		OrderNumber:         orderNumber,
		UserID:              7,
		Status:              "shipped",
		PaymentMethod:       "paypal",
		PaymentStatus:       "paid",
		ShippingMethod:      "standard",
		ShippingStatus:      "delivered",
		TrackingNumber:      trackingNumber,
		ProviderCarrierCode: "DHL",
		ProviderCarrierName: "DHL",
		SubtotalAmount:      total,
		TotalAmount:         total,
		Currency:            "USD",
		PaidAt:              &paidAt,
		ShippedAt:           &shippedAt,
		ShippingAddress: orderdomain.Address{
			FirstName:  "Ada",
			LastName:   "Lovelace",
			Address1:   "1 Carbon Road",
			City:       "Los Angeles",
			State:      "CA",
			PostalCode: "90001",
			Country:    "US",
			Email:      "ada@example.test",
		},
		BillingAddress: orderdomain.Address{
			FirstName:  "Ada",
			LastName:   "Lovelace",
			Address1:   "1 Carbon Road",
			City:       "Los Angeles",
			State:      "CA",
			PostalCode: "90001",
			Country:    "US",
			Email:      "ada@example.test",
		},
		Items: []orderdomain.OrderItem{
			{
				ProductID:   1,
				VariantID:   &variantID,
				ProductName: "Carbon wheelset",
				SKU:         "C50-DT240",
				Quantity:    1,
				Price:       total,
				Subtotal:    total,
				Total:       total,
			},
		},
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	return orderRecord
}

func seedPayPalDisputeWebhookTransaction(t *testing.T, db *gorm.DB, orderID uint, transactionID string, amount float64) {
	t.Helper()
	completedAt := time.Date(2026, time.July, 29, 9, 1, 0, 0, time.UTC)
	require.NoError(t, db.Create(&paymentdomain.Transaction{
		OrderID:       orderID,
		TransactionID: transactionID,
		PaymentMethod: "paypal",
		Amount:        amount,
		Currency:      "USD",
		Status:        "completed",
		CompletedAt:   &completedAt,
	}).Error)
}

func seedPayPalDisputeWebhookTracking(t *testing.T, db *gorm.DB, orderID uint, trackingNumber string) {
	t.Helper()
	provider := shippingdomain.TrackingProviderConfig{
		ProviderCode: "mock",
		ProviderName: "Mock Tracking",
		Enabled:      true,
	}
	require.NoError(t, db.Create(&provider).Error)
	deliveredAt := time.Date(2026, time.July, 31, 13, 0, 0, 0, time.UTC)
	require.NoError(t, db.Create(&shippingdomain.TrackingShipment{
		OrderID:             orderID,
		TrackingProviderID:  provider.ID,
		TrackingNumber:      trackingNumber,
		ProviderCarrierCode: "DHL",
		RegistrationStatus:  "registered",
		SyncStatus:          "synced",
		EventCount:          1,
		LastEventAt:         &deliveredAt,
		Enabled:             true,
	}).Error)
	require.NoError(t, db.Create(&shippingdomain.TrackingEvent{
		OrderID:             orderID,
		TrackingNumber:      trackingNumber,
		ProviderCarrierCode: "DHL",
		Status:              "Delivered",
		Location:            "Los Angeles, US",
		Description:         "Delivered and signed by recipient",
		EventTime:           deliveredAt,
	}).Error)
}
