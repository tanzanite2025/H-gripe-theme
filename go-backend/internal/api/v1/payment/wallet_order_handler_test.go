package payment

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	pgateway "commerce-platform/internal/pkg/payment"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestCreateAlipayOrderRecordsPendingAttempt(t *testing.T) {
	db, handler := newPayPalHandlerTestHarness(t, &fakePaymentGateway{
		createResponse: &pgateway.PaymentResponse{
			ID:            "ORD-ALIPAY-1",
			Status:        "WAIT_BUYER_PAY",
			Amount:        128,
			Currency:      "CNY",
			PaymentURL:    "https://alipay.example/checkout",
			TransactionID: "ORD-ALIPAY-1",
			Metadata:      map[string]string{"out_trade_no": "ORD-ALIPAY-1"},
		},
	})
	t.Setenv("ALIPAY_APP_ID", "alipay-app")
	t.Setenv("ALIPAY_PRIVATE_KEY", "alipay-private-key")
	t.Setenv("ALIPAY_PUBLIC_KEY", "alipay-public-key")
	seedWalletOrder(t, db, "ORD-ALIPAY-1", 7, "alipay", 128, "CNY", "pending", "unpaid")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/alipay/orders",
		bytes.NewBufferString(`{"order_number":"ORD-ALIPAY-1","return_url":"https://shop.example/checkout/alipay/return"}`),
	)
	context.Request.Host = "shop.example"
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreateAlipayOrder(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var transaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "ORD-ALIPAY-1").First(&transaction).Error)
	require.Equal(t, "alipay", transaction.PaymentMethod)
	require.Equal(t, "pending", transaction.Status)
	require.InDelta(t, 128, transaction.Amount, 0.001)
}

func TestCreateAlipayOrderRejectsUnsupportedCurrency(t *testing.T) {
	db, handler := newPayPalHandlerTestHarness(t, &fakePaymentGateway{
		createResponse: &pgateway.PaymentResponse{
			ID:            "ORD-ALIPAY-USD",
			Status:        "WAIT_BUYER_PAY",
			Amount:        128,
			Currency:      "USD",
			PaymentURL:    "https://alipay.example/checkout",
			TransactionID: "ORD-ALIPAY-USD",
		},
	})
	t.Setenv("ALIPAY_APP_ID", "alipay-app")
	t.Setenv("ALIPAY_PRIVATE_KEY", "alipay-private-key")
	t.Setenv("ALIPAY_PUBLIC_KEY", "alipay-public-key")
	seedWalletOrder(t, db, "ORD-ALIPAY-USD", 7, "alipay", 128, "USD", "pending", "unpaid")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/alipay/orders",
		bytes.NewBufferString(`{"order_number":"ORD-ALIPAY-USD","return_url":"https://shop.example/checkout/alipay/return"}`),
	)
	context.Request.Host = "shop.example"
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreateAlipayOrder(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var transactionCount int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).Where("order_id IN (SELECT id FROM orders WHERE order_number = ?)", "ORD-ALIPAY-USD").Count(&transactionCount).Error)
	require.Zero(t, transactionCount)
}

func TestCreateWechatOrderRejectsUnsupportedCurrency(t *testing.T) {
	gateway := &fakePaymentGateway{
		createResponse: &pgateway.PaymentResponse{
			ID:            "ORD-WECHAT-USD",
			Status:        "NOTPAY",
			Amount:        236,
			Currency:      "USD",
			PaymentURL:    "weixin://wxpay/bizpayurl?pr=example",
			TransactionID: "ORD-WECHAT-USD",
		},
	}
	db, handler := newPayPalHandlerTestHarness(t, gateway)
	t.Setenv("WECHAT_MCH_ID", "wechat-mch")
	t.Setenv("WECHAT_APP_ID", "wechat-app")
	t.Setenv("WECHAT_PRIVATE_KEY_PATH", "merchant-private-key.pem")
	t.Setenv("WECHAT_MERCHANT_SERIAL", "merchant-serial")
	t.Setenv("WECHAT_API_V3_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("WECHAT_PAY_PLATFORM_PUBLIC_KEY", "wechat-platform-public-key")
	t.Setenv("WECHAT_PAY_PLATFORM_PUBLIC_KEY_ID", "PUB_KEY_ID")
	seedWalletOrder(t, db, "ORD-WECHAT-USD", 7, "wechat", 236, "USD", "pending", "unpaid")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/wechat/orders",
		bytes.NewBufferString(`{"order_number":"ORD-WECHAT-USD"}`),
	)
	context.Request.Host = "shop.example"
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreateWechatOrder(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Nil(t, gateway.createRequest)
	var transactionCount int64
	require.NoError(t, db.Model(&paymentdomain.Transaction{}).Where("order_id IN (SELECT id FROM orders WHERE order_number = ?)", "ORD-WECHAT-USD").Count(&transactionCount).Error)
	require.Zero(t, transactionCount)
}

func TestCreateWechatOrderUsesConfiguredWebhookBaseURL(t *testing.T) {
	gateway := &fakePaymentGateway{
		createResponse: &pgateway.PaymentResponse{
			ID:            "ORD-WECHAT-CNY",
			Status:        "NOTPAY",
			Amount:        236,
			Currency:      "CNY",
			PaymentURL:    "weixin://wxpay/bizpayurl?pr=example",
			TransactionID: "ORD-WECHAT-CNY",
		},
	}
	db, handler := newPayPalHandlerTestHarness(t, gateway)
	handler.ConfigurePublicBaseURL("https://payments.example.com/")
	t.Setenv("WECHAT_MCH_ID", "wechat-mch")
	t.Setenv("WECHAT_APP_ID", "wechat-app")
	t.Setenv("WECHAT_PRIVATE_KEY_PATH", "merchant-private-key.pem")
	t.Setenv("WECHAT_MERCHANT_SERIAL", "merchant-serial")
	t.Setenv("WECHAT_API_V3_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("WECHAT_PAY_PLATFORM_PUBLIC_KEY", "wechat-platform-public-key")
	t.Setenv("WECHAT_PAY_PLATFORM_PUBLIC_KEY_ID", "PUB_KEY_ID")
	seedWalletOrder(t, db, "ORD-WECHAT-CNY", 7, "wechat", 236, "CNY", "pending", "unpaid")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/wechat/orders",
		bytes.NewBufferString(`{"order_number":"ORD-WECHAT-CNY"}`),
	)
	context.Request.Host = "spoofed.example"
	context.Request.Header.Set("X-Forwarded-Host", "attacker.example")
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreateWechatOrder(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotNil(t, gateway.createRequest)
	require.Equal(t, "https://payments.example.com/api/v1/payment/webhook/wechat", gateway.createRequest.NotifyURL)
}

func TestConfirmWechatOrderMarksMatchingOrderPaid(t *testing.T) {
	db, handler := newPayPalHandlerTestHarness(t, &fakePaymentGateway{
		getResponse: &pgateway.PaymentResponse{
			ID:            "ORD-WECHAT-1",
			Status:        "SUCCESS",
			Amount:        236,
			Currency:      "CNY",
			TransactionID: "WX-TXN-1",
			Metadata:      map[string]string{"out_trade_no": "ORD-WECHAT-1"},
		},
	})
	t.Setenv("WECHAT_MCH_ID", "wechat-mch")
	t.Setenv("WECHAT_APP_ID", "wechat-app")
	t.Setenv("WECHAT_PRIVATE_KEY_PATH", "merchant-private-key.pem")
	t.Setenv("WECHAT_MERCHANT_SERIAL", "merchant-serial")
	t.Setenv("WECHAT_API_V3_KEY", "0123456789abcdef0123456789abcdef")
	t.Setenv("WECHAT_PAY_PLATFORM_PUBLIC_KEY", "wechat-platform-public-key")
	t.Setenv("WECHAT_PAY_PLATFORM_PUBLIC_KEY_ID", "PUB_KEY_ID")
	orderRecord := seedWalletOrder(t, db, "ORD-WECHAT-1", 7, "wechat", 236, "CNY", "pending", "unpaid")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Params = gin.Params{{Key: "order_number", Value: "ORD-WECHAT-1"}}
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/wechat/orders/ORD-WECHAT-1/confirm",
		bytes.NewBufferString(`{}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.ConfirmWechatOrder(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var savedOrder orderdomain.Order
	require.NoError(t, db.First(&savedOrder, orderRecord.ID).Error)
	require.Equal(t, "paid", savedOrder.PaymentStatus)
	require.Equal(t, "processing", savedOrder.Status)

	var transaction paymentdomain.Transaction
	require.NoError(t, db.Where("transaction_id = ?", "WX-TXN-1").First(&transaction).Error)
	require.Equal(t, "completed", transaction.Status)
	require.Equal(t, "wechat", transaction.PaymentMethod)
}

func seedWalletOrder(
	t *testing.T,
	db *gorm.DB,
	orderNumber string,
	userID uint,
	paymentMethod string,
	total float64,
	currency string,
	status string,
	paymentStatus string,
) orderdomain.Order {
	t.Helper()
	orderRecord := orderdomain.Order{
		OrderNumber:    orderNumber,
		UserID:         userID,
		Status:         status,
		PaymentMethod:  paymentMethod,
		PaymentStatus:  paymentStatus,
		ShippingMethod: "standard",
		ShippingStatus: "pending",
		TotalAmount:    total,
		Currency:       currency,
		ShippingAddress: orderdomain.Address{
			FirstName: "Ada",
			LastName:  "Lovelace",
			Country:   "CN",
			Email:     "ada@example.com",
			Phone:     "+8613800000000",
		},
		BillingAddress: orderdomain.Address{
			Country: "CN",
		},
	}
	require.NoError(t, db.Create(&orderRecord).Error)
	return orderRecord
}
