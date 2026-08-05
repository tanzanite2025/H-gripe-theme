package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPaymentGatewayProductionConfigRequiresTypedConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PAYMENT_CONFIG_MASTER_KEY", "test-payment-config-master-key")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "provider", Value: "stripe"}}
	context.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/admin/settings/payment-gateways/stripe",
		strings.NewReader(`{"environment":"production","credentials":{"api_key":"sk_live_test"}}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	NewPaymentHandler(nil, nil).UpsertGatewayConfig(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "PRODUCTION")
}

func TestPaymentGatewayDeleteRequiresProviderTypedConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "provider", Value: "paypal"}}
	context.Request = httptest.NewRequest(
		http.MethodDelete,
		"/api/admin/settings/payment-gateways/paypal",
		strings.NewReader(`{"confirmation":"DELETE STRIPE"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	NewPaymentHandler(nil, nil).DeleteGatewayConfig(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "DELETE PAYPAL")
}
