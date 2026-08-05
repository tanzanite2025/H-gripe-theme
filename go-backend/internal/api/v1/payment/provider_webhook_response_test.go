package payment

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
