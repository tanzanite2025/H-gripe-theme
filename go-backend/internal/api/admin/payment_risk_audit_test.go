package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	paymentdomain "tanzanite/internal/domain/payment"
	"tanzanite/internal/pkg/config"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPaymentRefundExecutionConfirmationFailureIsAuditLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewPaymentRefundExecutionHandler(service.NewPaymentService(nil, nil), nil)
	handler.ConfigureAuditService(auditRecorder)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: "55"}}
	context.Set("user_id", uint(7))
	context.Set("username", "ops-admin")
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/payment/refunds/55/execute", strings.NewReader(`{}`))
	context.Request.Header.Set("Content-Type", "application/json")

	handler.ExecutePendingRefund(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionExecute, log.Action)
	require.Equal(t, paymentAuditResourceRefundExecution, log.Resource)
	require.Equal(t, paymentAuditStatusFailed, log.Status)
	require.Equal(t, uint(55), log.ResourceID)

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, float64(55), changes["refund_id"])
	require.Equal(t, false, changes["confirmation_matched"])
}

func TestPaymentRefundRecommendationPendingRefundAuditMasksOperatorText(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewPaymentRefundRecommendationHandler(service.NewPaymentRefundRecommendationService(nil))
	handler.ConfigureAuditService(auditRecorder)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: "8"}}
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/payment/risk/refund-recommendations/8/pending-refund",
		strings.NewReader(`{"amount":25.5,"reason":"customer private note","decision_notes":"internal decision note"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreatePendingRefund(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionCreate, log.Action)
	require.Equal(t, paymentAuditResourceRefundRecommendation, log.Resource)
	require.Equal(t, paymentAuditStatusFailed, log.Status)
	require.NotContains(t, log.Changes, "customer private note")
	require.NotContains(t, log.Changes, "internal decision note")

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, true, changes["reason_present"])
	require.Equal(t, true, changes["decision_notes_present"])
}

func TestPaymentRiskRecomputeUnsupportedProviderIsAuditLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewPaymentRiskMonitoringHandler(service.NewPaymentRiskMonitoringService(nil, config.PaymentRiskMonitoringConfig{}))
	handler.ConfigureAuditService(auditRecorder)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/payment/risk/recompute",
		strings.NewReader(`{"provider":"stripe,unsupported"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.RecomputeSummary(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionRecompute, log.Action)
	require.Equal(t, paymentAuditResourceRiskMonitoring, log.Resource)
	require.Equal(t, paymentAuditStatusFailed, log.Status)

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, "stripe,unsupported", changes["raw_provider"])
}

func TestPaymentProtectionCreateControlAuditMasksReason(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewPaymentProtectionHandler(service.NewPaymentProtectionService(nil, config.PaymentProtectionConfig{Enabled: true}))
	handler.ConfigureAuditService(auditRecorder)

	expiresAt := time.Now().UTC().Add(time.Hour).Format(time.RFC3339)
	body := `{"action":"pause_payment","scope_type":"global","reason":"private incident details","expires_at":"` + expiresAt + `"}`

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/payment/risk/controls", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreateControl(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionCreate, log.Action)
	require.Equal(t, paymentAuditResourceProtectionControl, log.Resource)
	require.Equal(t, paymentAuditStatusFailed, log.Status)
	require.NotContains(t, log.Changes, "private incident details")

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, true, changes["reason_present"])
	require.Equal(t, false, changes["confirmation_matched"])
}

func TestPaymentCreateRefundAuditMasksReason(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewPaymentHandler(nil, nil)
	handler.ConfigureAuditService(auditRecorder)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/payment/refunds",
		strings.NewReader(`{"order_id":12,"transaction_id":34,"amount":19.99,"reason":"customer private refund reason"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreateRefund(context)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionCreate, log.Action)
	require.Equal(t, paymentAuditResourceRefundDraft, log.Resource)
	require.Equal(t, paymentAuditStatusFailed, log.Status)
	require.NotContains(t, log.Changes, "customer private refund reason")

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, true, changes["reason_present"])
	require.Equal(t, false, changes["gateway_refund_executed"])
}

func TestStripeDisputeEvidenceAuditMasksAdditionalStatement(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("PAYMENT_CONFIG_MASTER_KEY", "")

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewPaymentHandler(nil, nil)
	handler.ConfigureAuditService(auditRecorder)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: "44"}}
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/payment/disputes/44/evidence/submit",
		strings.NewReader(`{"confirm":false,"submit":true,"include_customer_communication":true,"additional_statement":"private dispute narrative","shipping_documentation_file_id":"file_ship"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.SubmitStripeDisputeEvidence(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionSubmit, log.Action)
	require.Equal(t, paymentAuditResourceDisputeEvidence, log.Resource)
	require.Equal(t, paymentAuditStatusFailed, log.Status)
	require.NotContains(t, log.Changes, "private dispute narrative")
	require.NotContains(t, log.Changes, "file_ship")

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, true, changes["additional_statement_present"])
	require.Equal(t, true, changes["shipping_documentation_file_present"])
	require.Equal(t, false, changes["confirmation_matched"])
}

func TestPaymentReviewAuditMasksReasonAndNotes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewPaymentHandler(nil, nil)
	handler.ConfigureAuditService(auditRecorder)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/payment/reviews",
		strings.NewReader(`{"order_id":12,"reason":"private review reason","notes":"private review notes"}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreatePaymentReview(context)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionCreate, log.Action)
	require.Equal(t, paymentAuditResourcePaymentReview, log.Resource)
	require.Equal(t, paymentAuditStatusFailed, log.Status)
	require.NotContains(t, log.Changes, "private review reason")
	require.NotContains(t, log.Changes, "private review notes")

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, true, changes["reason_present"])
	require.Equal(t, true, changes["notes_present"])
}

func TestPaymentMethodCreateAuditMasksSettingsWhenServiceUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewPaymentHandler(nil, nil)
	handler.ConfigureAuditService(auditRecorder)

	settings := `{"api_key":"pm_secret_create","webhook_secret":"whsec_create"}`
	body := paymentMethodAuditRequestBody(t, map[string]interface{}{
		"name":        "Credit Card",
		"code":        "card",
		"fee_type":    "fixed",
		"fee_value":   1.25,
		"enabled":     true,
		"sort_order":  10,
		"description": "primary card provider",
		"settings":    settings,
	})

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/settings/payment-methods", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")

	handler.CreatePaymentMethod(context)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionCreate, log.Action)
	require.Equal(t, paymentAuditResourcePaymentMethod, log.Resource)
	require.Equal(t, paymentAuditStatusFailed, log.Status)
	require.NotContains(t, log.Changes, "pm_secret_create")
	require.NotContains(t, log.Changes, "whsec_create")

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, "card", changes["code"])
	require.Equal(t, true, changes["settings_present"])
	require.Equal(t, float64(len(settings)), changes["settings_length"])
}

func TestPaymentMethodUpdateAuditMasksSettingsInChangesAndOldValue(t *testing.T) {
	auditRecorder := &fakePaymentAuditRecorder{}
	handler, db := newPaymentMethodAuditTestHandler(t, auditRecorder)

	oldSettings := `{"api_key":"pm_secret_old"}`
	existing := paymentdomain.PaymentMethod{
		Name:     "Credit Card",
		Code:     "card",
		FeeType:  "fixed",
		Enabled:  true,
		Settings: oldSettings,
	}
	require.NoError(t, db.Create(&existing).Error)

	newSettings := `{"api_key":"pm_secret_new","merchant_id":"merchant_private"}`
	body := paymentMethodAuditRequestBody(t, map[string]interface{}{
		"name":       "Credit Card Updated",
		"code":       "card",
		"fee_type":   "percentage",
		"fee_value":  2.5,
		"min_amount": 5,
		"max_amount": 500,
		"enabled":    false,
		"settings":   newSettings,
	})

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: "1"}}
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(http.MethodPut, "/api/admin/settings/payment-methods/1", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")

	handler.UpdatePaymentMethod(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionUpdate, log.Action)
	require.Equal(t, paymentAuditResourcePaymentMethod, log.Resource)
	require.Equal(t, paymentAuditStatusSuccess, log.Status)
	require.Equal(t, uint(1), log.ResourceID)

	payload := log.Changes + log.OldValue + log.NewValue
	require.NotContains(t, payload, "pm_secret_old")
	require.NotContains(t, payload, "pm_secret_new")
	require.NotContains(t, payload, "merchant_private")

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, true, changes["settings_present"])
	require.Equal(t, float64(len(newSettings)), changes["settings_length"])

	var oldValue map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.OldValue), &oldValue))
	require.Equal(t, true, oldValue["settings_present"])
	require.Equal(t, float64(len(oldSettings)), oldValue["settings_length"])

	var newValue map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.NewValue), &newValue))
	require.Equal(t, true, newValue["settings_present"])
	require.Equal(t, float64(len(newSettings)), newValue["settings_length"])
}

func TestPaymentMethodDeleteAuditMasksOldSettings(t *testing.T) {
	auditRecorder := &fakePaymentAuditRecorder{}
	handler, db := newPaymentMethodAuditTestHandler(t, auditRecorder)

	settings := `{"private_key":"pm_secret_delete"}`
	existing := paymentdomain.PaymentMethod{
		Name:     "Wallet",
		Code:     "wallet",
		FeeType:  "fixed",
		Enabled:  true,
		Settings: settings,
	}
	require.NoError(t, db.Create(&existing).Error)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: "1"}}
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(http.MethodDelete, "/api/admin/settings/payment-methods/1", nil)

	handler.DeletePaymentMethod(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionDelete, log.Action)
	require.Equal(t, paymentAuditResourcePaymentMethod, log.Resource)
	require.Equal(t, paymentAuditStatusSuccess, log.Status)
	require.NotContains(t, log.Changes+log.OldValue, "pm_secret_delete")

	var oldValue map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.OldValue), &oldValue))
	require.Equal(t, true, oldValue["settings_present"])
	require.Equal(t, float64(len(settings)), oldValue["settings_length"])
}

func TestPaymentMethodUpdateValidationFailureIsAuditLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewPaymentHandler(nil, nil)
	handler.ConfigureAuditService(auditRecorder)

	body := paymentMethodAuditRequestBody(t, map[string]interface{}{
		"name":     "Credit Card",
		"code":     "card",
		"fee_type": "unsupported",
		"settings": `{"api_key":"pm_secret_invalid"}`,
	})

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: "12"}}
	context.Set("user_id", uint(7))
	context.Request = httptest.NewRequest(http.MethodPut, "/api/admin/settings/payment-methods/12", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")

	handler.UpdatePaymentMethod(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, paymentAuditActionUpdate, log.Action)
	require.Equal(t, paymentAuditResourcePaymentMethod, log.Resource)
	require.Equal(t, paymentAuditStatusFailed, log.Status)
	require.Equal(t, uint(12), log.ResourceID)
	require.NotContains(t, log.Changes, "pm_secret_invalid")
	require.Contains(t, log.ErrorMessage, "fee type")
}

func paymentMethodAuditRequestBody(t *testing.T, value map[string]interface{}) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	return string(encoded)
}

func newPaymentMethodAuditTestHandler(t *testing.T, auditRecorder *fakePaymentAuditRecorder) (*PaymentHandler, *gorm.DB) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(&paymentdomain.PaymentMethod{}))

	handler := NewPaymentHandler(service.NewPaymentService(nil, repository.NewPaymentRepository(db)), nil)
	handler.ConfigureAuditService(auditRecorder)
	return handler, db
}
