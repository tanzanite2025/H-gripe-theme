package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pgateway "commerce-platform/internal/pkg/payment"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPaymentRuntimeStatusUsesConfiguredPublicBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPaymentHandler(nil, nil)
	handler.ConfigurePublicBaseURL("https://payments.example.com/")

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/settings/payment-runtime", nil)
	context.Request.Host = "spoofed.example.test"
	context.Request.Header.Set("X-Forwarded-Host", "proxy-spoof.example.test")
	context.Request.Header.Set("X-Forwarded-Proto", "http")

	handler.GetGatewayRuntimeStatus(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int                       `json:"code"`
		Data pgateway.RuntimeReadiness `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)

	wechatStatus := findAdminRuntimeStatus(t, body.Data, pgateway.GatewayWechat)
	require.Equal(t, "https://payments.example.com/api/v1/payment/webhook/wechat", wechatStatus.CallbackURL)
	require.NotContains(t, wechatStatus.CallbackURL, "spoofed.example.test")
}

func findAdminRuntimeStatus(t *testing.T, readiness pgateway.RuntimeReadiness, provider pgateway.GatewayType) pgateway.GatewayRuntimeStatus {
	t.Helper()
	for _, status := range readiness.Gateways {
		if status.Provider == provider {
			return status
		}
	}
	t.Fatalf("missing admin runtime status for %s", provider)
	return pgateway.GatewayRuntimeStatus{}
}
