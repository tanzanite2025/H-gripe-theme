package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tanzanite/internal/domain/audit"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakePaymentAuditRecorder struct {
	logs []audit.AuditLog
}

func (r *fakePaymentAuditRecorder) CreateAuditLog(log *audit.AuditLog) error {
	if log != nil {
		r.logs = append(r.logs, *log)
	}
	return nil
}

func TestPaymentGatewayConfigAuditDoesNotStoreSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PAYMENT_CONFIG_MASTER_KEY", "test-payment-config-master-key")

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewPaymentHandler(nil, nil)
	handler.ConfigureAuditService(auditRecorder)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "provider", Value: "stripe"}}
	context.Set("user_id", uint(7))
	context.Set("username", "ops-admin")
	context.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/admin/settings/payment-gateways/stripe",
		strings.NewReader(`{"environment":"production","credentials":{"api_key":"sk_live_super_secret"}}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	context.Request.Header.Set("User-Agent", "admin-test-agent")

	handler.UpsertGatewayConfig(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)

	log := auditRecorder.logs[0]
	require.Equal(t, uint(7), log.UserID)
	require.Equal(t, "ops-admin", log.Username)
	require.Equal(t, paymentAuditActionUpdate, log.Action)
	require.Equal(t, paymentAuditResourceGatewayConfig, log.Resource)
	require.Equal(t, paymentAuditStatusFailed, log.Status)
	require.Equal(t, http.MethodPut, log.Method)
	require.Equal(t, "/api/admin/settings/payment-gateways/stripe", log.Path)
	require.Contains(t, log.ErrorMessage, "PRODUCTION")
	require.NotContains(t, log.Changes, "sk_live_super_secret")

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, "stripe", changes["provider"])
	require.Equal(t, "production", changes["environment"])
	require.Equal(t, "confirmation_mismatch", changes["failure_stage"])
	require.Equal(t, false, changes["confirmation_matched"])
	require.Contains(t, changes["submitted_fields"], "api_key")
}

func TestPaymentCallbackProbeAuditRecordsReachabilityResult(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &fakePaymentAuditRecorder{}
	probeClient := &fakePaymentCallbackProbeClient{status: http.StatusUnauthorized}
	handler := &PaymentHandler{
		publicBaseURL:       "https://payments.example.com",
		callbackProbeClient: probeClient,
	}
	handler.ConfigureAuditService(auditRecorder)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "provider", Value: "paypal"}}
	context.Set("user_id", uint(9))
	context.Set("email", "finance@example.com")
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/settings/payment-runtime/paypal/callback-check", nil)

	handler.CheckGatewayCallback(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)

	log := auditRecorder.logs[0]
	require.Equal(t, uint(9), log.UserID)
	require.Equal(t, "finance@example.com", log.Username)
	require.Equal(t, paymentAuditActionProbe, log.Action)
	require.Equal(t, paymentAuditResourceGatewayCallback, log.Resource)
	require.Equal(t, paymentAuditStatusSuccess, log.Status)

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, "paypal", changes["provider"])
	require.Equal(t, "https://payments.example.com/api/v1/payment/webhook/paypal", changes["callback_url"])
	require.Equal(t, true, changes["reachable"])
	require.Equal(t, true, changes["route_reachable"])
	require.Equal(t, true, changes["expected_signature_failure"])
	require.Equal(t, float64(http.StatusUnauthorized), changes["status_code"])
}
