package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"commerce-platform/internal/domain/security"
	"commerce-platform/internal/domain/visitor"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestVisitorProfileIPBlockWithoutRetainedIPReturnsConflictAndMasksReason(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db, _ := newVisitorProfileIPBlockHandler(t)
	profile := &visitor.Profile{
		CustomerServiceVisitorHash: "visitor-without-ip",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 time.Now().UTC(),
	}
	require.NoError(t, db.Create(profile).Error)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler.ConfigureAuditService(auditRecorder)
	reason := "private incident details that must not be copied into audit metadata"
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Set("user_id", uint(7))
	c.Set("username", "ops-admin")
	c.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/customer-service/visitor-profiles/1/ip-block",
		strings.NewReader(`{"reason":"`+reason+`"}`),
	)
	c.Request.Header.Set("Content-Type", "application/json")

	handler.BlockVisitorProfileIP(c)

	require.Equal(t, http.StatusConflict, recorder.Code, recorder.Body.String())
	require.Len(t, auditRecorder.logs, 1)
	log := auditRecorder.logs[0]
	require.Equal(t, adminAuditActionCreate, log.Action)
	require.Equal(t, adminAuditResourceGlobalIPBlockRule, log.Resource)
	require.Equal(t, adminAuditStatusFailed, log.Status)
	require.NotContains(t, log.Changes, reason)
	require.Contains(t, log.ErrorMessage, "no retained IP address")

	var changes map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(log.Changes), &changes))
	require.Equal(t, float64(1), changes["profile_id"])
	require.Equal(t, true, changes["reason_present"])
	require.Equal(t, float64(len([]rune(reason))), changes["reason_length"])
}

func TestVisitorProfileIPBlockHandlerCreatesAndRemovesGlobalRule(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler, db, blockService := newVisitorProfileIPBlockHandler(t)
	profile := &visitor.Profile{
		CustomerServiceVisitorHash: "visitor-with-ip",
		IPAddress:                  "198.51.100.24",
		IPHash:                     "retained-ip-hash",
		ProfileStatus:              visitor.ProfileStatusActive,
		LastSeenAt:                 time.Now().UTC(),
	}
	require.NoError(t, db.Create(profile).Error)

	auditRecorder := &fakePaymentAuditRecorder{}
	handler.ConfigureAuditService(auditRecorder)
	reason := "manual visitor abuse review"
	createRecorder := httptest.NewRecorder()
	createContext, _ := gin.CreateTestContext(createRecorder)
	createContext.Params = gin.Params{{Key: "id", Value: "1"}}
	createContext.Set("user_id", uint(7))
	createContext.Set("username", "ops-admin")
	createContext.Request = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/customer-service/visitor-profiles/1/ip-block",
		strings.NewReader(`{"reason":"`+reason+`"}`),
	)
	createContext.Request.Header.Set("Content-Type", "application/json")

	handler.BlockVisitorProfileIP(createContext)

	require.Equal(t, http.StatusOK, createRecorder.Code, createRecorder.Body.String())
	var createPayload struct {
		Code int `json:"code"`
		Data struct {
			Rule service.IPBlockRuleSnapshot `json:"rule"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(createRecorder.Body.Bytes(), &createPayload))
	require.Equal(t, 0, createPayload.Code)
	require.Equal(t, "198.51.100.24/32", createPayload.Data.Rule.CIDR)
	require.Equal(t, security.IPBlockRuleSourceVisitorProfile, createPayload.Data.Rule.Source)
	require.Equal(t, uint(7), *createPayload.Data.Rule.CreatedBy)
	require.Len(t, auditRecorder.logs, 1)
	require.NotContains(t, auditRecorder.logs[0].Changes, reason)

	blocked, match, err := blockService.IsBlocked(context.Background(), "198.51.100.24", time.Now())
	require.NoError(t, err)
	require.True(t, blocked)
	require.NotNil(t, match)
	require.Equal(t, createPayload.Data.Rule.ID, match.ID)

	deleteRecorder := httptest.NewRecorder()
	deleteContext, _ := gin.CreateTestContext(deleteRecorder)
	deleteContext.Params = gin.Params{{Key: "id", Value: "1"}}
	deleteContext.Set("user_id", uint(8))
	deleteContext.Set("username", "ops-admin")
	deleteContext.Request = httptest.NewRequest(
		http.MethodDelete,
		"/api/admin/customer-service/visitor-profiles/1/ip-block",
		nil,
	)

	handler.UnblockVisitorProfileIP(deleteContext)

	require.Equal(t, http.StatusOK, deleteRecorder.Code, deleteRecorder.Body.String())
	require.Len(t, auditRecorder.logs, 2)
	require.Equal(t, adminAuditActionDelete, auditRecorder.logs[1].Action)
	require.Equal(t, adminAuditStatusSuccess, auditRecorder.logs[1].Status)
	var oldValue map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(auditRecorder.logs[1].OldValue), &oldValue))
	require.Equal(t, float64(createPayload.Data.Rule.ID), oldValue["rule_id"])
	require.Equal(t, "198.51.100.24/32", oldValue["cidr"])
	require.Equal(t, security.IPBlockRuleStatusActive, oldValue["status"])
	var newValue map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(auditRecorder.logs[1].NewValue), &newValue))
	require.Equal(t, security.IPBlockRuleStatusDisabled, newValue["status"])

	blocked, match, err = blockService.IsBlocked(context.Background(), "198.51.100.24", time.Now())
	require.NoError(t, err)
	require.False(t, blocked)
	require.Nil(t, match)
}

func newVisitorProfileIPBlockHandler(t *testing.T) (*VisitorProfileHandler, *gorm.DB, *service.GlobalIPBlockService) {
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

	require.NoError(t, db.AutoMigrate(&visitor.Profile{}, &security.IPBlockRule{}))
	blockService := service.NewGlobalIPBlockService(repository.NewGlobalIPBlockRuleRepository(db))
	visitorProfileService := service.NewVisitorProfileService(repository.NewVisitorProfileRepository(db))
	visitorProfileService.ConfigureGlobalIPBlockService(blockService)
	return NewVisitorProfileHandler(visitorProfileService), db, blockService
}
