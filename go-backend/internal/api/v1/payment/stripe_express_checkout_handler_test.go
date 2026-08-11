package payment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGetStripeExpressCheckoutConfigurationReturnsOnlyPublishableKey(t *testing.T) {
	t.Setenv("STRIPE_SECRET_KEY", "stripe-secret")
	t.Setenv("STRIPE_PUBLISHABLE_KEY", "stripe-publishable")
	t.Setenv("STRIPE_WEBHOOK_SECRET", "stripe-webhook")

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/v1/payment/stripe/express-checkout/config", nil)

	handler := NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler.GetStripeExpressCheckoutConfiguration(context)

	require.Equal(t, http.StatusOK, recorder.Code)

	var payload struct {
		Data struct {
			PublishableKey string `json:"publishable_key"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &payload))
	require.Equal(t, "stripe-publishable", payload.Data.PublishableKey)
	require.NotContains(t, recorder.Body.String(), "stripe-secret")
}
