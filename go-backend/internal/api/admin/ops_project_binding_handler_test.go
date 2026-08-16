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

func TestOpsProjectBindingHandlerFiltersByEnvironment(t *testing.T) {
	handler := newOpsProjectBindingTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/projects?environment=staging", nil)

	handler.List(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			Projects []struct {
				Name        string `json:"name"`
				Environment string `json:"environment"`
			} `json:"projects"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Len(t, body.Data.Projects, 1)
	require.Equal(t, "Staging Project", body.Data.Projects[0].Name)
	require.Equal(t, ops.ProjectEnvironmentStaging, body.Data.Projects[0].Environment)
}

func TestOpsProjectBindingHandlerRejectsInvalidEnvironment(t *testing.T) {
	handler := newOpsProjectBindingTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/projects?environment=qa", nil)

	handler.List(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var body struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "bad_request", body.Code)
}

func newOpsProjectBindingTestHandler(t *testing.T) *OpsProjectBindingHandler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&ops.Connector{}, &ops.VPSBinding{}, &ops.ProjectBinding{}))

	productionVPS := &ops.VPSBinding{
		Name:        "Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	stagingVPS := &ops.VPSBinding{
		Name:        "Staging VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentStaging,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	require.NoError(t, db.Create(productionVPS).Error)
	require.NoError(t, db.Create(stagingVPS).Error)
	require.NoError(t, db.Create(&ops.ProjectBinding{
		Name:         "Production Project",
		VPSBindingID: productionVPS.ID,
		Environment:  ops.ProjectEnvironmentProduction,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthUnknown,
		Enabled:      true,
	}).Error)
	require.NoError(t, db.Create(&ops.ProjectBinding{
		Name:         "Staging Project",
		VPSBindingID: stagingVPS.ID,
		Environment:  ops.ProjectEnvironmentStaging,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthUnknown,
		Enabled:      true,
	}).Error)

	return NewOpsProjectBindingHandler(service.NewOpsProjectBindingService(
		repository.NewOpsProjectBindingRepository(db),
		repository.NewOpsVPSBindingRepository(db),
		repository.NewOpsConnectorRepository(db),
	))
}
