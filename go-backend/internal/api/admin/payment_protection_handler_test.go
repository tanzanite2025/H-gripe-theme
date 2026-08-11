package admin

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestPaymentProtectionCreateControlRequiresConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPaymentProtectionHandler(service.NewPaymentProtectionService(nil, config.PaymentProtectionConfig{Enabled: true}))
	expiresAt := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	body := `{"action":"pause_payment","scope_type":"global","reason":"temporary incident review","expires_at":"` + expiresAt + `"}`

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/payment/risk/controls", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreateControl(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "confirmation")
}

func TestPaymentProtectionRevokeControlRequiresConfirmation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewPaymentProtectionHandler(service.NewPaymentProtectionService(nil, config.PaymentProtectionConfig{Enabled: true}))

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: "1"}}
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/payment/risk/controls/1/revoke", strings.NewReader(`{}`))
	context.Request.Header.Set("Content-Type", "application/json")

	handler.RevokeControl(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Contains(t, recorder.Body.String(), "confirmation")
}
