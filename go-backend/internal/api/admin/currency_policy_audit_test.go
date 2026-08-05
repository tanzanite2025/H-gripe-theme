package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tanzanite/internal/domain/currency"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCurrencyPolicyUpdateAuditRecordsOldAndNewPolicy(t *testing.T) {
	auditRecorder := &fakePaymentAuditRecorder{}
	handler, policyService := newCurrencyPolicyAuditTestHandler(t, auditRecorder)
	require.NoError(t, seedCurrencyPolicy(policyService, currency.Policy{
		AccountingCurrency:   "USD",
		DefaultOrderCurrency: "USD",
		AcceptedCurrencies:   []string{"USD", "EUR"},
	}))

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(7))
	context.Set("username", "ops-admin")
	context.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/admin/settings/currency-policy",
		strings.NewReader(`{"accounting_currency":"usd","default_order_currency":"gbp","accepted_currencies":["gbp","usd","gbp"]}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")
	context.Request.Header.Set("User-Agent", "admin-test-agent")

	handler.UpdatePolicy(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, uint(7), log.UserID)
	require.Equal(t, "ops-admin", log.Username)
	require.Equal(t, adminAuditActionUpdate, log.Action)
	require.Equal(t, adminAuditResourceCurrencyPolicy, log.Resource)
	require.Equal(t, adminAuditStatusSuccess, log.Status)
	require.Equal(t, http.MethodPut, log.Method)
	require.Equal(t, "/api/admin/settings/currency-policy", log.Path)
	require.Equal(t, "admin-test-agent", log.UserAgent)

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, "USD", changes["accounting_currency"])
	require.Equal(t, "GBP", changes["default_order_currency"])
	require.Equal(t, []interface{}{"GBP", "USD"}, changes["accepted_currencies"])
	require.Equal(t, float64(2), changes["accepted_currency_count"])

	var oldValue map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.OldValue), &oldValue))
	require.Equal(t, "USD", oldValue["accounting_currency"])
	require.Equal(t, "USD", oldValue["default_order_currency"])
	require.Equal(t, []interface{}{"EUR", "USD"}, oldValue["accepted_currencies"])

	var newValue map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.NewValue), &newValue))
	require.Equal(t, "GBP", newValue["default_order_currency"])
}

func TestCurrencyPolicyUpdateValidationFailureIsAuditLogged(t *testing.T) {
	auditRecorder := &fakePaymentAuditRecorder{}
	handler, _ := newCurrencyPolicyAuditTestHandler(t, auditRecorder)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Set("user_id", uint(8))
	context.Request = httptest.NewRequest(
		http.MethodPut,
		"/api/admin/settings/currency-policy",
		strings.NewReader(`{"accounting_currency":"USD","default_order_currency":"BTC","accepted_currencies":["USD"]}`),
	)
	context.Request.Header.Set("Content-Type", "application/json")

	handler.UpdatePolicy(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, adminAuditActionUpdate, log.Action)
	require.Equal(t, adminAuditResourceCurrencyPolicy, log.Resource)
	require.Equal(t, adminAuditStatusFailed, log.Status)
	require.Contains(t, log.ErrorMessage, "unsupported default order currency")

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, "USD", changes["accounting_currency"])
	require.Equal(t, "BTC", changes["default_order_currency"])
	require.Equal(t, []interface{}{"USD"}, changes["accepted_currencies"])
}

func newCurrencyPolicyAuditTestHandler(t *testing.T, auditRecorder *fakePaymentAuditRecorder) (*CurrencyPolicyHandler, *service.CurrencyPolicyService) {
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

	require.NoError(t, db.AutoMigrate(&setting.Setting{}))
	policyService := service.NewCurrencyPolicyService(repository.NewSettingRepository(db))
	handler := NewCurrencyPolicyHandler(policyService)
	handler.ConfigureAuditService(auditRecorder)
	return handler, policyService
}

func seedCurrencyPolicy(policyService *service.CurrencyPolicyService, policy currency.Policy) error {
	_, err := policyService.UpdatePolicy(policy)
	return err
}
