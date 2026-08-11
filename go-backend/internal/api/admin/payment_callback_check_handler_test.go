package admin

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pgateway "commerce-platform/internal/pkg/payment"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakePaymentCallbackProbeClient struct {
	request *http.Request
	body    string
	status  int
	err     error
}

func (c *fakePaymentCallbackProbeClient) Do(req *http.Request) (*http.Response, error) {
	c.request = req
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		c.body = string(body)
	}
	if c.err != nil {
		return nil, c.err
	}
	status := c.status
	if status == 0 {
		status = http.StatusUnauthorized
	}
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader("probe response")),
	}, nil
}

func TestPaymentGatewayCallbackCheckUsesConfiguredPublicBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	probeClient := &fakePaymentCallbackProbeClient{status: http.StatusUnauthorized}
	handler := &PaymentHandler{
		publicBaseURL:       "https://payments.example.com/",
		callbackProbeClient: probeClient,
	}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "provider", Value: "stripe"}}
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/settings/payment-runtime/stripe/callback-check", nil)
	context.Request.Host = "spoofed.example.test"
	context.Request.Header.Set("X-Forwarded-Host", "proxy-spoof.example.test")

	handler.CheckGatewayCallback(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotNil(t, probeClient.request)
	require.Equal(t, http.MethodPost, probeClient.request.Method)
	require.Equal(t, "https://payments.example.com/api/v1/payment/webhook/stripe", probeClient.request.URL.String())
	require.Equal(t, "PaymentCallbackReachabilityProbe/1.0", probeClient.request.UserAgent())

	var body struct {
		Code int                        `json:"code"`
		Data paymentCallbackCheckResult `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, pgateway.GatewayStripe, body.Data.Provider)
	require.True(t, body.Data.Reachable)
	require.True(t, body.Data.RouteReachable)
	require.True(t, body.Data.ExpectedSignatureFailure)
	require.Equal(t, http.StatusUnauthorized, body.Data.StatusCode)
}

func TestPaymentGatewayCallbackCheckRequiresConfiguredPublicBaseURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	probeClient := &fakePaymentCallbackProbeClient{status: http.StatusUnauthorized}
	handler := &PaymentHandler{callbackProbeClient: probeClient}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "provider", Value: "paypal"}}
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/settings/payment-runtime/paypal/callback-check", nil)
	context.Request.Host = "spoofed.example.test"
	context.Request.Header.Set("X-Forwarded-Host", "proxy-spoof.example.test")

	handler.CheckGatewayCallback(context)

	require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	require.Nil(t, probeClient.request)
	require.Contains(t, recorder.Body.String(), "payment_callback_base_url_missing")
}
