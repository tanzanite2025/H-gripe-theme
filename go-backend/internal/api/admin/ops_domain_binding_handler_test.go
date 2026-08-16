package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsDomainBindingHandlerFiltersByEnvironment(t *testing.T) {
	handler := newOpsDomainBindingTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/domains?environment=staging", nil)

	handler.List(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			Domains []struct {
				Domain      string `json:"domain"`
				Environment string `json:"environment"`
			} `json:"domains"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Len(t, body.Data.Domains, 1)
	require.Equal(t, "staging.learn.example.com", body.Data.Domains[0].Domain)
	require.Equal(t, ops.DomainEnvironmentStaging, body.Data.Domains[0].Environment)
}

func TestOpsDomainBindingHandlerRejectsInvalidEnvironment(t *testing.T) {
	handler := newOpsDomainBindingTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/domains?environment=qa", nil)

	handler.List(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var body struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "bad_request", body.Code)
}

func newOpsDomainBindingTestHandler(t *testing.T) *OpsDomainBindingHandler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&ops.Connector{},
		&ops.VPSBinding{},
		&ops.ProjectBinding{},
		&ops.DomainBinding{},
	))
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:      "learn.example.com",
		Role:        ops.DomainRoleInternal,
		Environment: ops.DomainEnvironmentProduction,
		Provider:    ops.DomainProviderCloudflare,
		Status:      ops.DomainStatusActive,
		Enabled:     true,
	}).Error)
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:      "staging.learn.example.com",
		Role:        ops.DomainRoleInternal,
		Environment: ops.DomainEnvironmentStaging,
		Provider:    ops.DomainProviderCloudflare,
		Status:      ops.DomainStatusActive,
		Enabled:     true,
	}).Error)

	return NewOpsDomainBindingHandler(
		service.NewOpsDomainBindingService(
			repository.NewOpsDomainBindingRepository(db),
			repository.NewOpsProjectBindingRepository(db),
			repository.NewOpsConnectorRepository(db),
		),
		nil,
		nil,
		nil,
	)
}
