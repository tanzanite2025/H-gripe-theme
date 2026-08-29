package admin

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/security"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGlobalIPBlockHandlerAuditsDuplicateAsUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db, blockService := newGlobalIPBlockHandler(t)
	auditRecorder := &fakePaymentAuditRecorder{}
	handler.ConfigureAuditService(auditRecorder)

	firstRecorder := httptest.NewRecorder()
	firstContext, _ := gin.CreateTestContext(firstRecorder)
	firstContext.Set("user_id", uint(7))
	firstContext.Set("username", "first-admin")
	firstContext.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/security/ip-blocks",
		strings.NewReader(`{"cidr":"198.51.100.52","reason":"initial review"}`),
	)
	firstContext.Request.Header.Set("Content-Type", "application/json")

	handler.Create(firstContext)

	require.Equal(t, http.StatusOK, firstRecorder.Code, firstRecorder.Body.String())
	var firstPayload struct {
		Data struct {
			Rule service.IPBlockRuleSnapshot `json:"rule"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(firstRecorder.Body.Bytes(), &firstPayload))
	require.NotZero(t, firstPayload.Data.Rule.ID)
	require.Equal(t, "198.51.100.52/32", firstPayload.Data.Rule.CIDR)
	require.Len(t, auditRecorder.logs, 1)
	require.Equal(t, adminAuditActionCreate, auditRecorder.logs[0].Action)

	secondRecorder := httptest.NewRecorder()
	secondContext, _ := gin.CreateTestContext(secondRecorder)
	secondContext.Set("user_id", uint(8))
	secondContext.Set("username", "second-admin")
	secondContext.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/security/ip-blocks",
		strings.NewReader(`{"cidr":"198.51.100.52/32","reason":"updated review"}`),
	)
	secondContext.Request.Header.Set("Content-Type", "application/json")

	handler.Create(secondContext)

	require.Equal(t, http.StatusOK, secondRecorder.Code, secondRecorder.Body.String())
	var secondPayload struct {
		Data struct {
			Rule service.IPBlockRuleSnapshot `json:"rule"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(secondRecorder.Body.Bytes(), &secondPayload))
	require.Equal(t, firstPayload.Data.Rule.ID, secondPayload.Data.Rule.ID)
	require.Len(t, auditRecorder.logs, 2)
	require.Equal(t, adminAuditActionUpdate, auditRecorder.logs[1].Action)
	require.Equal(t, adminAuditStatusSuccess, auditRecorder.logs[1].Status)

	var oldValue map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(auditRecorder.logs[1].OldValue), &oldValue))
	require.Equal(t, float64(firstPayload.Data.Rule.ID), oldValue["rule_id"])
	require.Equal(t, "198.51.100.52/32", oldValue["cidr"])
	require.Equal(t, security.IPBlockRuleStatusActive, oldValue["status"])

	var newValue map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(auditRecorder.logs[1].NewValue), &newValue))
	require.Equal(t, float64(secondPayload.Data.Rule.ID), newValue["rule_id"])
	require.Equal(t, "198.51.100.52/32", newValue["cidr"])
	require.Equal(t, security.IPBlockRuleStatusActive, newValue["status"])
	require.NotContains(t, auditRecorder.logs[1].Changes, "updated review")

	blocked, match, err := blockService.IsBlocked(context.Background(), "198.51.100.52", time.Now())
	require.NoError(t, err)
	require.True(t, blocked)
	require.NotNil(t, match)
	require.Equal(t, secondPayload.Data.Rule.ID, match.ID)

	var count int64
	require.NoError(t, db.Model(&security.IPBlockRule{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestGlobalIPBlockHandlerReturns503AndRollsBackWhenAuditWriteFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db, _ := newGlobalIPBlockHandler(t)
	auditErr := errors.New("audit database is down")
	recorder := &failingGlobalIPBlockAuditRecorder{err: auditErr}
	handler.ConfigureAuditService(recorder)

	responseRecorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(responseRecorder)
	c.Set("user_id", uint(7))
	c.Set("username", "ops-admin")
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/security/ip-blocks",
		strings.NewReader(`{"cidr":"198.51.100.64","reason":"audit failure response test"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.Create(c)

	require.Equal(t, http.StatusServiceUnavailable, responseRecorder.Code, responseRecorder.Body.String())
	require.Contains(t, responseRecorder.Body.String(), "ip_block_audit_unavailable")
	require.Equal(t, 2, recorder.attempts)

	var count int64
	require.NoError(t, db.Model(&security.IPBlockRule{}).Count(&count).Error)
	require.Zero(t, count)
}

func TestGlobalIPBlockHandlerDisableReturns503AndRollsBackWhenAuditWriteFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db, blockService := newGlobalIPBlockHandler(t)
	rule, err := blockService.Block(context.Background(), service.IPBlockRuleInput{
		CIDR:   "198.51.100.65",
		Reason: "disable audit failure response test",
	})
	require.NoError(t, err)

	auditErr := errors.New("audit database is down")
	recorder := &failingGlobalIPBlockAuditRecorder{err: auditErr}
	handler.ConfigureAuditService(recorder)

	responseRecorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(responseRecorder)
	c.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(rule.ID), 10)}}
	c.Set("user_id", uint(7))
	c.Set("username", "ops-admin")
	c.Request = httptest.NewRequest(
		http.MethodDelete,
		"/api/admin/security/ip-blocks/"+strconv.FormatUint(uint64(rule.ID), 10),
		nil,
	)

	handler.Disable(c)

	require.Equal(t, http.StatusServiceUnavailable, responseRecorder.Code, responseRecorder.Body.String())
	require.Contains(t, responseRecorder.Body.String(), "ip_block_audit_unavailable")
	require.Equal(t, 2, recorder.attempts)

	var stored security.IPBlockRule
	require.NoError(t, db.First(&stored, rule.ID).Error)
	require.True(t, stored.Enabled)
	require.Nil(t, stored.DisabledAt)
}

func newGlobalIPBlockHandler(t *testing.T) (*GlobalIPBlockHandler, *gorm.DB, *service.GlobalIPBlockService) {
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

	require.NoError(t, db.AutoMigrate(&security.IPBlockRule{}))
	blockService := service.NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))
	return NewGlobalIPBlockHandler(blockService), db, blockService
}

type failingGlobalIPBlockAuditRecorder struct {
	err      error
	attempts int
}

func (r *failingGlobalIPBlockAuditRecorder) CreateAuditLog(_ *audit.AuditLog) error {
	r.attempts++
	return r.err
}
