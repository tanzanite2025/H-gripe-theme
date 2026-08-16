package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAdminAccountHandlerEnsureRecordsAuthenticatedOperator(t *testing.T) {
	handler, db := newAdminAccountHandlerTestFixture(t)
	body := bytes.NewBufferString(`{
		"email":"new-admin@example.com",
		"password":"N3w-Admin-Secret!",
		"role":"admin"
	}`)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/ops/admin-accounts/ensure", body)
	context.Request.RemoteAddr = "198.51.100.10:1234"
	context.Request.Header.Set("Content-Type", "application/json")
	context.Request.Header.Set("User-Agent", "admin-account-handler-test")
	context.Set("user_id", uint(77))
	context.Set("user_role", "admin")
	context.Set("username", "platform-owner")

	handler.Ensure(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var auditLog audit.AuditLog
	require.NoError(t, db.Last(&auditLog).Error)
	require.Equal(t, uint(77), auditLog.UserID)
	require.Equal(t, "platform-owner", auditLog.Username)
	require.Equal(t, "user", auditLog.Resource)
	require.NotZero(t, auditLog.ResourceID)
	require.NotEqual(t, auditLog.UserID, auditLog.ResourceID)
	require.Equal(t, http.MethodPost, auditLog.Method)
	require.Equal(t, "/api/admin/ops/admin-accounts/ensure", auditLog.Path)
	require.Equal(t, "198.51.100.10", auditLog.IPAddress)
	require.Equal(t, "admin-account-handler-test", auditLog.UserAgent)
	require.NotContains(t, auditLog.NewValue, "N3w-Admin-Secret!")
}

func TestAdminAccountHandlerListRejectsNonAdminActor(t *testing.T) {
	handler, _ := newAdminAccountHandlerTestFixture(t)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/admin-accounts", nil)
	context.Set("user_id", uint(11))
	context.Set("user_role", "manager")

	handler.List(context)

	require.Equal(t, http.StatusForbidden, recorder.Code)
}

func TestAdminAccountHandlerListReturnsAccountsForAdmin(t *testing.T) {
	handler, db := newAdminAccountHandlerTestFixture(t)
	require.NoError(t, db.Create(&user.User{
		Email:    "admin@example.com",
		Username: "primary-admin",
		Role:     "admin",
		Status:   "active",
		Locale:   "en",
	}).Error)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/admin-accounts", nil)
	context.Set("user_id", uint(11))
	context.Set("user_role", "admin")

	handler.List(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var responseBody struct {
		Data struct {
			Accounts []service.AdminAccountMaintenanceAccount `json:"accounts"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &responseBody))
	require.Len(t, responseBody.Data.Accounts, 1)
	require.Equal(t, "admin@example.com", responseBody.Data.Accounts[0].Email)
}

func TestAdminAccountHandlerEnsureRejectsUnsupportedLocale(t *testing.T) {
	handler, _ := newAdminAccountHandlerTestFixture(t)
	body := bytes.NewBufferString(`{
		"email":"new-admin@example.com",
		"password":"N3w-Admin-Secret!",
		"role":"admin",
		"locale":"xx_unsupported"
	}`)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/ops/admin-accounts/ensure", body)
	context.Request.Header.Set("Content-Type", "application/json")
	context.Set("user_id", uint(77))
	context.Set("user_role", "admin")

	handler.Ensure(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
}

func TestAdminAccountHandlerEnsureRejectsSelfRoleChange(t *testing.T) {
	handler, db := newAdminAccountHandlerTestFixture(t)
	operator := user.User{
		Email:    "admin@example.com",
		Username: "primary-admin",
		Role:     "admin",
		Status:   "active",
		Locale:   "en",
	}
	require.NoError(t, operator.HashPassword("0ld-Admin-Secret!"))
	require.NoError(t, db.Create(&operator).Error)
	body := bytes.NewBufferString(`{
		"email":"admin@example.com",
		"password":"N3w-Admin-Secret!",
		"role":"manager"
	}`)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/admin/ops/admin-accounts/ensure", body)
	context.Request.Header.Set("Content-Type", "application/json")
	context.Set("user_id", operator.ID)
	context.Set("user_role", "admin")

	handler.Ensure(context)

	require.Equal(t, http.StatusForbidden, recorder.Code)
	var stored user.User
	require.NoError(t, db.First(&stored, operator.ID).Error)
	require.Equal(t, "admin", stored.Role)
	require.True(t, stored.CheckPassword("0ld-Admin-Secret!"))
}

func newAdminAccountHandlerTestFixture(t *testing.T) (*AdminAccountHandler, *gorm.DB) {
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
	require.NoError(t, db.AutoMigrate(&user.User{}, &audit.AuditLog{}))
	return NewAdminAccountHandler(service.NewAdminAccountMaintenanceService(db)), db
}
