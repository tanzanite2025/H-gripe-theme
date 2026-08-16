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

func TestOpsOverviewHandlerFiltersByEnvironment(t *testing.T) {
	handler := newOpsOverviewTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/overview?environment=staging", nil)

	handler.Get(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			Environment string `json:"environment"`
			Topology    struct {
				VPS []struct {
					Name string `json:"name"`
				} `json:"vps"`
			} `json:"topology"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, ops.DomainEnvironmentStaging, body.Data.Environment)
	require.Len(t, body.Data.Topology.VPS, 1)
	require.Equal(t, "Staging VPS", body.Data.Topology.VPS[0].Name)
}

func TestOpsOverviewHandlerRejectsInvalidEnvironment(t *testing.T) {
	handler := newOpsOverviewTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/overview?environment=qa", nil)

	handler.Get(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var body struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "bad_request", body.Code)
}

func newOpsOverviewTestHandler(t *testing.T) *OpsOverviewHandler {
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
	require.NoError(t, db.Create(&ops.VPSBinding{
		Name:        "Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}).Error)
	require.NoError(t, db.Create(&ops.VPSBinding{
		Name:        "Staging VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentStaging,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}).Error)

	return NewOpsOverviewHandler(service.NewOpsOverviewService(
		repository.NewOpsDomainBindingRepository(db),
		repository.NewOpsConnectorRepository(db),
		repository.NewOpsVPSBindingRepository(db),
		repository.NewOpsProjectBindingRepository(db),
		nil,
	))
}
