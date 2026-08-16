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

func TestOpsConnectorHandlerFiltersByEnvironment(t *testing.T) {
	handler := newOpsConnectorTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/connectors?environment=staging", nil)

	handler.List(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			Connectors []struct {
				Name        string `json:"name"`
				Environment string `json:"environment"`
			} `json:"connectors"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Len(t, body.Data.Connectors, 1)
	require.Equal(t, "Cloudflare Staging", body.Data.Connectors[0].Name)
	require.Equal(t, ops.ConnectorEnvironmentStaging, body.Data.Connectors[0].Environment)
}

func TestOpsConnectorHandlerRejectsInvalidEnvironment(t *testing.T) {
	handler := newOpsConnectorTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/connectors?environment=qa", nil)

	handler.List(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var body struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "bad_request", body.Code)
}

func newOpsConnectorTestHandler(t *testing.T) *OpsConnectorHandler {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&ops.Connector{}))
	require.NoError(t, db.Create(&ops.Connector{
		Name:        "Cloudflare Production",
		Provider:    ops.ConnectorProviderCloudflare,
		Environment: ops.ConnectorEnvironmentProduction,
		AuthType:    ops.ConnectorAuthManual,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}).Error)
	require.NoError(t, db.Create(&ops.Connector{
		Name:        "Cloudflare Staging",
		Provider:    ops.ConnectorProviderCloudflare,
		Environment: ops.ConnectorEnvironmentStaging,
		AuthType:    ops.ConnectorAuthManual,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}).Error)

	return NewOpsConnectorHandler(service.NewOpsConnectorService(repository.NewOpsConnectorRepository(db)))
}
