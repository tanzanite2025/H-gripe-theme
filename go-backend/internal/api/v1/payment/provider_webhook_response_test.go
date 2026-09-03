package payment

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestProviderWebhookAcknowledgementsUseProviderExpectedFormats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("alipay success is plain success", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)

		respondAlipayWebhookSuccess(context)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "success", recorder.Body.String())
		require.Contains(t, recorder.Header().Get("Content-Type"), "text/plain")
	})

	t.Run("alipay failure is plain fail", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)

		respondAlipayWebhookFailure(context, http.StatusUnauthorized)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "fail", recorder.Body.String())
		require.Contains(t, recorder.Header().Get("Content-Type"), "text/plain")
	})

	t.Run("wechat success is no content", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)

		respondWechatWebhookSuccess(context)

		require.Equal(t, http.StatusNoContent, recorder.Code)
		require.Empty(t, recorder.Body.String())
	})

	t.Run("wechat failure is api v3 fail body", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)

		respondWechatWebhookFailure(context, http.StatusUnauthorized, "signature failed")

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.JSONEq(t, `{"code":"FAIL","message":"signature failed"}`, recorder.Body.String())
	})
}

func TestProviderWebhookAlreadyPaidAcknowledgementUsesProviderSuccessFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	err := fmt.Errorf("provider duplicate terminal state: %w", service.ErrOrderAlreadyPaid)

	t.Run("paypal success json", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)

		acknowledged := acknowledgeAlreadyPaidProviderWebhook(context, verifiedProviderPayment{
			Provider:      pgateway.GatewayPayPal,
			OrderNumber:   "ORD-ALREADY-PAID",
			TransactionID: "PAYPAL-CAPTURE-LATER",
		}, err)

		require.True(t, acknowledged)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Contains(t, recorder.Body.String(), "Order already paid, webhook acknowledged")
		require.Contains(t, recorder.Body.String(), "PAYPAL-CAPTURE-LATER")
	})

	t.Run("alipay plain success", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)

		acknowledged := acknowledgeAlreadyPaidProviderWebhook(context, verifiedProviderPayment{
			Provider:      pgateway.GatewayAlipay,
			OrderNumber:   "ORD-ALREADY-PAID",
			TransactionID: "ALIPAY-TRADE-LATER",
		}, err)

		require.True(t, acknowledged)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "success", recorder.Body.String())
		require.Contains(t, recorder.Header().Get("Content-Type"), "text/plain")
	})

	t.Run("wechat no content success", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)

		acknowledged := acknowledgeAlreadyPaidProviderWebhook(context, verifiedProviderPayment{
			Provider:      pgateway.GatewayWechat,
			OrderNumber:   "ORD-ALREADY-PAID",
			TransactionID: "WECHAT-TRANSACTION-LATER",
		}, err)

		require.True(t, acknowledged)
		require.Equal(t, http.StatusNoContent, recorder.Code)
		require.Empty(t, recorder.Body.String())
	})

	t.Run("other errors are not acknowledged", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)

		acknowledged := acknowledgeAlreadyPaidProviderWebhook(context, verifiedProviderPayment{
			Provider: pgateway.GatewayPayPal,
		}, fmt.Errorf("amount mismatch"))

		require.False(t, acknowledged)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Empty(t, recorder.Body.String())
	})
}

func TestHandleWebhookRejectsOversizedPayloadBeforeProviderDispatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "provider", Value: "paypal"}}
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/v1/payment/webhook/paypal",
		strings.NewReader(strings.Repeat("x", paymentWebhookMaxBodyBytes+1)),
	)

	handler := &Handler{}
	handler.HandleWebhook(context)

	require.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	require.Contains(t, recorder.Body.String(), "payment_webhook_payload_too_large")
}
