package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsDeploymentPreflightHandlerReturnsProjectReportContract(t *testing.T) {
	handler, projectID := newOpsDeploymentPreflightTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: strconv.FormatUint(uint64(projectID), 10)}}
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/projects/"+strconv.FormatUint(uint64(projectID), 10)+"/preflight", nil)

	handler.GetProjectReport(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			ProjectID   uint   `json:"project_id"`
			StatusLevel string `json:"status_level"`
			Ready       bool   `json:"ready"`
			Checks      []struct {
				Key string `json:"key"`
			} `json:"checks"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, projectID, body.Data.ProjectID)
	require.Equal(t, "review", body.Data.StatusLevel)
	require.False(t, body.Data.Ready)
	require.NotEmpty(t, body.Data.Checks)
}

func TestOpsDeploymentPreflightHandlerKeepsQueryCompatibilityEndpoint(t *testing.T) {
	handler, projectID := newOpsDeploymentPreflightTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(
		http.MethodGet,
		"/api/admin/ops/deployments/preflight?project_id="+strconv.FormatUint(uint64(projectID), 10),
		nil,
	)

	handler.GetProjectReportByQuery(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			ProjectID uint `json:"project_id"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, projectID, body.Data.ProjectID)
}

func TestOpsDeploymentPreflightHandlerRejectsInvalidProjectIDs(t *testing.T) {
	handler, _ := newOpsDeploymentPreflightTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: "not-a-number"}}
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/projects/not-a-number/preflight", nil)

	handler.GetProjectReport(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var body struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "bad_request", body.Code)
}

func TestOpsDeploymentPreflightHandlerReturnsNotFoundForMissingProject(t *testing.T) {
	handler, _ := newOpsDeploymentPreflightTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: "9999"}}
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/projects/9999/preflight", nil)

	handler.GetProjectReport(context)

	require.Equal(t, http.StatusNotFound, recorder.Code)
	var body struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "not_found", body.Code)
}

func TestOpsDeploymentPreflightHandlerOverviewPreservesReviewCount(t *testing.T) {
	handler, projectID := newOpsDeploymentPreflightTestHandler(t)
	require.NotZero(t, projectID)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/deployments/preflight-overview", nil)

	handler.GetOverview(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			ProjectCount int `json:"project_count"`
			ReviewCount  int `json:"review_count"`
			Projects     []struct {
				StatusLevel string `json:"status_level"`
			} `json:"projects"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, 1, body.Data.ProjectCount)
	require.Equal(t, 1, body.Data.ReviewCount)
	require.Len(t, body.Data.Projects, 1)
	require.Equal(t, "review", body.Data.Projects[0].StatusLevel)
}

func TestOpsDeploymentPreflightHandlerOverviewFiltersByEnvironment(t *testing.T) {
	handler, projectID := newOpsDeploymentPreflightTestHandler(t)
	require.NotZero(t, projectID)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/deployments/preflight-overview?environment=staging", nil)

	handler.GetOverview(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			Environment  string `json:"environment"`
			ProjectCount int    `json:"project_count"`
			Projects     []struct {
				ProjectID   uint   `json:"project_id"`
				Environment string `json:"environment"`
			} `json:"projects"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, ops.ProjectEnvironmentStaging, body.Data.Environment)
	require.Equal(t, 1, body.Data.ProjectCount)
	require.Len(t, body.Data.Projects, 1)
	require.Equal(t, projectID, body.Data.Projects[0].ProjectID)
	require.Equal(t, ops.ProjectEnvironmentStaging, body.Data.Projects[0].Environment)
}

func TestOpsDeploymentPreflightHandlerOverviewRejectsInvalidEnvironment(t *testing.T) {
	handler, _ := newOpsDeploymentPreflightTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/ops/deployments/preflight-overview?environment=qa", nil)

	handler.GetOverview(context)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	var body struct {
		Code string `json:"code"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, "bad_request", body.Code)
}

func newOpsDeploymentPreflightTestHandler(t *testing.T) (*OpsDeploymentPreflightHandler, uint) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")

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
	require.NoError(t, db.AutoMigrate(
		&ops.Connector{},
		&ops.VPSBinding{},
		&ops.ProjectBinding{},
		&ops.DomainBinding{},
	))

	connector := &ops.Connector{
		Name:          "Hostinger Staging",
		Provider:      ops.ConnectorProviderHostinger,
		Environment:   ops.ConnectorEnvironmentStaging,
		AuthType:      ops.ConnectorAuthEnvironment,
		CredentialRef: "HOSTINGER_API_TOKEN",
		Scopes:        "vps:read,project:read",
		Status:        ops.ConnectorStatusActive,
		Enabled:       true,
	}
	require.NoError(t, db.Create(connector).Error)

	now := time.Date(2026, 8, 13, 14, 0, 0, 0, time.UTC)
	vps := &ops.VPSBinding{
		Name:               "Staging VPS",
		Provider:           ops.VPSProviderHostinger,
		Environment:        ops.VPSEnvironmentStaging,
		ConnectorID:        &connector.ID,
		ProviderResourceID: "staging-vm-1",
		Hostname:           "staging.example.com",
		ObservedHostname:   "staging.example.com",
		IPv4:               "203.0.113.20",
		ObservedIPv4:       "203.0.113.20",
		Status:             ops.VPSStatusActive,
		ObservedStatus:     ops.VPSObservedHealthy,
		ObservedState:      "running",
		LastObservedAt:     &now,
		Enabled:            true,
	}
	require.NoError(t, db.Create(vps).Error)

	project := &ops.ProjectBinding{
		Name:                   "staging-project",
		VPSBindingID:           vps.ID,
		Environment:            ops.ProjectEnvironmentStaging,
		ComposeSource:          "compose.staging.yml",
		ComposeProjectName:     "staging-project",
		Services:               "api",
		Networks:               "app",
		Volumes:                "uploads, site_logo_uploads",
		CurrentImageTag:        "sha-cccccccccccccccccccccccccccccccccccccccc",
		CurrentCommitSHA:       "cccccccccccccccccccccccccccccccccccccccc",
		GatewayNetwork:         "staging-edge",
		GatewayAlias:           "staging-web",
		Status:                 ops.ProjectStatusActive,
		HealthStatus:           ops.ProjectHealthHealthy,
		ObservedContainerCount: 1,
		ObservedRunningCount:   1,
		ObservedHealthyCount:   1,
		LastCheckedAt:          &now,
		LastDeploymentAt:       &now,
		BackupPolicy:           "Daily staging backup.",
		RestoreNotes:           "2026-08-13 staging restore exercise completed.",
		Enabled:                true,
	}
	require.NoError(t, db.Create(project).Error)

	return NewOpsDeploymentPreflightHandler(service.NewOpsDeploymentPreflightService(
		repository.NewOpsProjectBindingRepository(db),
		repository.NewOpsVPSBindingRepository(db),
		repository.NewOpsConnectorRepository(db),
		repository.NewOpsDomainBindingRepository(db),
	)), project.ID
}
