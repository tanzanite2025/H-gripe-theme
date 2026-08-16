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

func TestOpsVPSBindingHandlerFiltersByEnvironment(t *testing.T) {
	handler := newOpsVPSBindingTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/vps?environment=staging", nil)

	handler.List(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			VPS []struct {
				Name        string `json:"name"`
				Environment string `json:"environment"`
			} `json:"vps"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Len(t, body.Data.VPS, 1)
	require.Equal(t, "Staging VPS", body.Data.VPS[0].Name)
	require.Equal(t, ops.VPSEnvironmentStaging, body.Data.VPS[0].Environment)
}

func TestOpsVPSBindingHandlerRejectsInvalidEnvironment(t *testing.T) {
	handler := newOpsVPSBindingTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/vps?environment=qa", nil)

	handler.List(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var body struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "bad_request", body.Code)
}

func newOpsVPSBindingTestHandler(t *testing.T) *OpsVPSBindingHandler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&ops.Connector{}, &ops.VPSBinding{}, &ops.ProjectBinding{}))
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

	return NewOpsVPSBindingHandler(service.NewOpsVPSBindingService(
		repository.NewOpsVPSBindingRepository(db),
		repository.NewOpsConnectorRepository(db),
		repository.NewOpsProjectBindingRepository(db),
	))
}
