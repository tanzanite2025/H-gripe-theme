package admin

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"time"

	"commerce-platform/internal/pkg/apierror"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

const paymentCallbackProbeTimeout = 5 * time.Second

type paymentCallbackProbeHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type paymentCallbackCheckResult struct {
	Provider                 pgateway.GatewayType `json:"provider"`
	CallbackURL              string               `json:"callback_url"`
	Method                   string               `json:"method"`
	CheckedAt                time.Time            `json:"checked_at"`
	DurationMS               int64                `json:"duration_ms"`
	Reachable                bool                 `json:"reachable"`
	TransportReachable       bool                 `json:"transport_reachable"`
	RouteReachable           bool                 `json:"route_reachable"`
	ExpectedSignatureFailure bool                 `json:"expected_signature_failure"`
	StatusCode               int                  `json:"status_code,omitempty"`
	Status                   string               `json:"status,omitempty"`
	Error                    string               `json:"error,omitempty"`
}

func (h *PaymentHandler) CheckGatewayCallback(c *gin.Context) {
	provider, err := pgateway.ParseGatewayType(c.Param("provider"))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	startedAt := paymentAuditStartedAt()

	baseURL := ""
	if h != nil {
		baseURL = h.publicBaseURL
	}
	baseURL = pgateway.NormalizePublicBaseURL(baseURL)
	if baseURL == "" {
		h.recordPaymentAdminAudit(c, paymentAdminAuditEvent{
			StartedAt:    startedAt,
			Action:       paymentAuditActionProbe,
			Resource:     paymentAuditResourceGatewayCallback,
			Status:       paymentAuditStatusFailed,
			ErrorMessage: "SERVER_BASE_URL is required before probing payment callback reachability",
			Changes: map[string]interface{}{
				"provider":            string(provider),
				"base_url_configured": false,
			},
		})
		apierror.RespondError(c, http.StatusServiceUnavailable, "payment_callback_base_url_missing", "SERVER_BASE_URL is required before probing payment callback reachability")
		return
	}

	result := h.probeGatewayCallback(c.Request.Context(), provider, pgateway.GatewayWebhookURL(baseURL, provider))
	h.recordPaymentGatewayCallbackAudit(c, startedAt, result)
	response.Success(c, result)
}

func (h *PaymentHandler) probeGatewayCallback(ctx context.Context, provider pgateway.GatewayType, callbackURL string) paymentCallbackCheckResult {
	startedAt := time.Now().UTC()
	result := paymentCallbackCheckResult{
		Provider:    provider,
		CallbackURL: callbackURL,
		Method:      http.MethodPost,
		CheckedAt:   startedAt,
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, callbackURL, bytes.NewReader(callbackProbePayload(provider)))
	if err != nil {
		result.Error = err.Error()
		return result
	}
	for key, value := range callbackProbeHeaders(provider) {
		req.Header.Set(key, value)
	}

	resp, err := h.gatewayCallbackProbeClient().Do(req)
	result.DurationMS = time.Since(startedAt).Milliseconds()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	if resp == nil {
		result.Error = "empty callback probe response"
		return result
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))

	result.TransportReachable = true
	result.StatusCode = resp.StatusCode
	result.Status = resp.Status
	result.RouteReachable = resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusMethodNotAllowed
	result.ExpectedSignatureFailure = isExpectedPaymentCallbackProbeStatus(resp.StatusCode)
	result.Reachable = result.TransportReachable && result.RouteReachable
	return result
}

func (h *PaymentHandler) gatewayCallbackProbeClient() paymentCallbackProbeHTTPClient {
	if h != nil && h.callbackProbeClient != nil {
		return h.callbackProbeClient
	}
	return &http.Client{Timeout: paymentCallbackProbeTimeout}
}

func callbackProbePayload(provider pgateway.GatewayType) []byte {
	if provider == pgateway.GatewayAlipay {
		return []byte("codex_admin_probe=1")
	}
	return []byte(`{"codex_admin_probe":true}`)
}

func callbackProbeHeaders(provider pgateway.GatewayType) map[string]string {
	headers := map[string]string{
		"User-Agent": "PaymentCallbackReachabilityProbe/1.0",
	}
	if provider == pgateway.GatewayAlipay {
		headers["Content-Type"] = "application/x-www-form-urlencoded"
		return headers
	}
	headers["Content-Type"] = "application/json"
	return headers
}

func isExpectedPaymentCallbackProbeStatus(status int) bool {
	switch status {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusUnprocessableEntity:
		return true
	default:
		return false
	}
}
