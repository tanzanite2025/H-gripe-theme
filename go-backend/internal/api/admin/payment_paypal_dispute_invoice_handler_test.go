package admin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/pkg/invoice"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPreviewPayPalDisputeCommercialInvoicePDFReturnsInlinePDF(t *testing.T) {
	db, handler := newPayPalDisputeInvoicePreviewHandler(t, true)
	orderRecord := seedAdminPayPalDisputeOrder(t, db, "ORD-ADMIN-PAYPAL-INVOICE-1")
	disputeRecord := seedAdminPayPalDispute(t, db, "PP-D-ADMIN-INVOICE-1", orderRecord.ID)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(disputeRecord.ID), 10)}}
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/payment/paypal-disputes/"+strconv.FormatUint(uint64(disputeRecord.ID), 10)+"/evidence/invoice.pdf", nil)

	handler.PreviewPayPalDisputeCommercialInvoicePDF(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Header().Get("Content-Type"), "application/pdf")
	require.Contains(t, recorder.Header().Get("Content-Disposition"), "inline")
	require.Contains(t, recorder.Header().Get("Content-Disposition"), "commercial-invoice-CI-ORD-ADMIN-PAYPAL-INVOICE-1.pdf")
	require.Equal(t, "private, no-store", recorder.Header().Get("Cache-Control"))
	require.True(t, len(recorder.Body.Bytes()) > 5)
	require.Equal(t, "%PDF-", recorder.Body.String()[:5])
}

func TestPreviewPayPalDisputeCommercialInvoicePDFRequiresSellerProfile(t *testing.T) {
	db, handler := newPayPalDisputeInvoicePreviewHandler(t, false)
	orderRecord := seedAdminPayPalDisputeOrder(t, db, "ORD-ADMIN-PAYPAL-INVOICE-2")
	disputeRecord := seedAdminPayPalDispute(t, db, "PP-D-ADMIN-INVOICE-2", orderRecord.ID)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(disputeRecord.ID), 10)}}
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/payment/paypal-disputes/"+strconv.FormatUint(uint64(disputeRecord.ID), 10)+"/evidence/invoice.pdf", nil)

	handler.PreviewPayPalDisputeCommercialInvoicePDF(context)

	require.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	require.Contains(t, recorder.Body.String(), "seller name and address")
}

func TestPreviewPayPalCommercialInvoicePDFUsesAdHocInput(t *testing.T) {
	_, handler := newPayPalDisputeInvoicePreviewHandler(t, false)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/payment/paypal-invoice-preview.pdf",
		bytes.NewBufferString(`{
			"document_number":"CI-SAMPLE-001",
			"document_date":"2026-08-11",
			"currency":"USD",
			"seller":{"name":"Sample Seller","address":"1 Seller Street\nAustin, TX 78701\nUS"},
			"bill_to":{"name":"Sample Customer","line1":"9 Buyer Avenue","city":"Seattle","state":"WA","postal_code":"98101","country":"US"},
			"ship_to":{"name":"Sample Customer","line1":"9 Buyer Avenue","city":"Seattle","state":"WA","postal_code":"98101","country":"US"},
			"items":[{"description":"Sample product","sku":"SKU-001","quantity":2,"unit_price":50,"total":100}],
			"payment_method":"PayPal",
			"payment_status":"paid",
			"payment_reference":"SAMPLE-CAPTURE",
			"subtotal":100,
			"shipping":10,
			"total":110
		}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.PreviewPayPalCommercialInvoicePDF(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Contains(t, recorder.Header().Get("Content-Type"), "application/pdf")
	require.Contains(t, recorder.Header().Get("Content-Disposition"), "CI-SAMPLE-001")
	require.Equal(t, "%PDF-", recorder.Body.String()[:5])
}

func newPayPalDisputeInvoicePreviewHandler(t *testing.T, configureSeller bool) (*gorm.DB, *PaymentHandler) {
	t.Helper()
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_NAME", "")
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_ADDRESS", "")
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
		&paymentdomain.Refund{},
		&paymentdomain.RefundLineItem{},
		&paymentdomain.PayPalDispute{},
	))

	paymentService := service.NewPaymentService(nil, repository.NewPaymentRepository(db))
	paymentService.ConfigureEvidenceSources(repository.NewOrderRepository(db), nil, nil)
	if configureSeller {
		paymentService.ConfigurePayPalDisputeInvoiceOptions(service.PayPalDisputeInvoiceOptions{
			Seller: invoice.SellerProfile{
				Name:    "Commerce Platform Factory",
				Address: "100 Factory Road\nAustin, TX 78701\nUS",
				Email:   "support@example.test",
			},
		})
	}
	return db, NewPaymentHandler(paymentService, nil)
}

func seedAdminPayPalDisputeOrder(t *testing.T, db *gorm.DB, orderNumber string) orderdomain.Order {
	t.Helper()
	paidAt := time.Date(2026, time.July, 29, 9, 0, 0, 0, time.UTC)
	variantID := uint(1)
	orderRecord := orderdomain.Order{
		OrderNumber:    orderNumber,
		UserID:         7,
		Status:         "shipped",
		PaymentMethod:  "paypal",
		PaymentStatus:  "paid",
		ShippingStatus: "delivered",
		SubtotalAmount: 249.90,
		ShippingFee:    20,
		TotalAmount:    269.90,
		Currency:       "USD",
		PaidAt:         &paidAt,
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
				Price:       249.90,
				Subtotal:    249.90,
				Total:       249.90,
			},
		},
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	return orderRecord
}

func seedAdminPayPalDispute(t *testing.T, db *gorm.DB, paypalID string, orderID uint) paymentdomain.PayPalDispute {
	t.Helper()
	record := paymentdomain.PayPalDispute{
		PayPalDisputeID:   paypalID,
		OrderID:           &orderID,
		ProviderPaymentID: "PAYPAL-CAPTURE-ADMIN-1",
		Amount:            269.90,
		Currency:          "USD",
		Reason:            "MERCHANDISE_OR_SERVICE_NOT_RECEIVED",
		Status:            "WAITING_FOR_SELLER_RESPONSE",
		DisputeState:      "REQUIRED_ACTION",
	}
	require.NoError(t, db.Create(&record).Error)
	return record
}
