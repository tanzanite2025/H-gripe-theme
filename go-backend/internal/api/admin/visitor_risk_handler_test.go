package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestVisitorRiskHandlerCreatesAndReadsDecision(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, riskService := newAdminVisitorRiskTestService(t)
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	fact := visitor.RiskDailyFact{
		Day:           now,
		IPHash:        "ip-hash",
		UserAgentHash: "ua-hash",
		FirstSeenAt:   now,
		LastSeenAt:    now,
		SamplePaths:   []byte(`[]`),
	}
	require.NoError(t, db.Create(&fact).Error)

	handler := NewVisitorRiskHandler(riskService)
	expiresAt := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	body := bytes.NewBufferString(`{"action":"temporary_block","reason":"Repeated checkout failures","expires_at":"` + expiresAt + `"}`)
	w := httptest.NewRecorder()
	createContext, _ := gin.CreateTestContext(w)
	createContext.Request = httptest.NewRequest(http.MethodPost, "/api/admin/customer-service/visitor-risk-facts/1/decision", body)
	createContext.Request.Header.Set("Content-Type", "application/json")
	createContext.Params = gin.Params{{Key: "id", Value: "1"}}
	createContext.Set("user_id", uint(7))

	handler.CreateVisitorRiskDecision(createContext)

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var createPayload struct {
		Code int `json:"code"`
		Data struct {
			Decision service.VisitorRiskDecisionSnapshot `json:"decision"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &createPayload))
	require.Equal(t, 0, createPayload.Code)
	require.Equal(t, visitor.RiskDecisionActionTemporaryBlock, createPayload.Data.Decision.Action)
	require.Equal(t, uint(7), *createPayload.Data.Decision.CreatedBy)

	readRecorder := httptest.NewRecorder()
	readContext, _ := gin.CreateTestContext(readRecorder)
	readContext.Request = httptest.NewRequest(http.MethodGet, "/api/admin/customer-service/visitor-risk-facts/1/decision", nil)
	readContext.Params = gin.Params{{Key: "id", Value: "1"}}

	handler.GetVisitorRiskDecision(readContext)

	require.Equal(t, http.StatusOK, readRecorder.Code, readRecorder.Body.String())
	var readPayload struct {
		Code int `json:"code"`
		Data struct {
			Decision *service.VisitorRiskDecisionSnapshot `json:"decision"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(readRecorder.Body.Bytes(), &readPayload))
	require.Equal(t, 0, readPayload.Code)
	require.NotNil(t, readPayload.Data.Decision)
	require.Equal(t, createPayload.Data.Decision.ID, readPayload.Data.Decision.ID)
}

func TestVisitorRiskDecisionAuditMasksReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, riskService := newAdminVisitorRiskTestService(t)
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	fact := visitor.RiskDailyFact{
		Day:           now,
		IPHash:        "ip-hash",
		UserAgentHash: "ua-hash",
		FirstSeenAt:   now,
		LastSeenAt:    now,
		SamplePaths:   []byte(`[]`),
	}
	require.NoError(t, db.Create(&fact).Error)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewVisitorRiskHandler(riskService)
	handler.ConfigureAuditService(auditRecorder)
	expiresAt := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	reason := "private checkout abuse investigation note"
	body := `{"action":"temporary_block","reason":"` + reason + `","expires_at":"` + expiresAt + `"}`

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/customer-service/visitor-risk-facts/1/decision", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Params = gin.Params{{Key: "id", Value: "1"}}
	context.Set("user_id", uint(7))
	context.Set("username", "risk-admin")

	handler.CreateVisitorRiskDecision(context)

	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, adminAuditActionCreate, log.Action)
	require.Equal(t, adminAuditResourceVisitorRiskDecision, log.Resource)
	require.Equal(t, adminAuditStatusSuccess, log.Status)
	require.Equal(t, "risk-admin", log.Username)
	require.NotContains(t, log.Changes, reason)
	require.NotContains(t, log.NewValue, reason)

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, float64(1), changes["fact_id"])
	require.Equal(t, "temporary_block", changes["requested_action"])
	require.Equal(t, "temporary_block", changes["action"])
	require.Equal(t, true, changes["reason_present"])
	require.Equal(t, float64(len(reason)), changes["reason_length"])
	require.Equal(t, true, changes["expires"])
}

func TestVisitorRiskDecisionValidationFailureIsAuditLogged(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, riskService := newAdminVisitorRiskTestService(t)
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	fact := visitor.RiskDailyFact{
		Day:         now,
		IPHash:      "ip-hash",
		FirstSeenAt: now,
		LastSeenAt:  now,
		SamplePaths: []byte(`[]`),
	}
	require.NoError(t, db.Create(&fact).Error)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler := NewVisitorRiskHandler(riskService)
	handler.ConfigureAuditService(auditRecorder)
	body := `{"action":"temporary_block","reason":"private no expiry reason"}`

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/customer-service/visitor-risk-facts/1/decision", strings.NewReader(body))
	context.Request.Header.Set("Content-Type", "application/json")
	context.Params = gin.Params{{Key: "id", Value: "1"}}
	context.Set("user_id", uint(7))

	handler.CreateVisitorRiskDecision(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, adminAuditActionCreate, log.Action)
	require.Equal(t, adminAuditResourceVisitorRiskDecision, log.Resource)
	require.Equal(t, adminAuditStatusFailed, log.Status)
	require.NotContains(t, log.Changes, "private no expiry reason")
	require.Contains(t, log.ErrorMessage, "temporary_block requires an expiry")
}

func newAdminVisitorRiskTestService(t *testing.T) (*gorm.DB, *service.VisitorRiskService) {
	t.Helper()

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

	require.NoError(t, db.AutoMigrate(&visitor.RiskDailyFact{}, &visitor.RiskDecision{}))

	return db, service.NewVisitorRiskService(
		repository.NewVisitorRiskFactRepository(db),
		config.VisitorRiskConfig{
			Enabled:       true,
			HashSalt:      "admin-test-risk-salt",
			RetentionDays: 365,
		},
		"",
	)
}
