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

func TestServiceCenterHandlerCloudflareFiltersResources(t *testing.T) {
	handler := newServiceCenterTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/services/cloudflare?environment=production", nil)

	handler.Cloudflare(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			ConnectionCount int `json:"connection_count"`
			DomainCount     int `json:"domain_count"`
			ZoneCount       int `json:"zone_count"`
			Connections     []struct {
				Name string `json:"name"`
			} `json:"connections"`
			Domains []struct {
				Domain string `json:"domain"`
			} `json:"domains"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, 1, body.Data.ConnectionCount)
	require.Equal(t, 1, body.Data.DomainCount)
	require.Equal(t, 1, body.Data.ZoneCount)
	require.Equal(t, "Cloudflare Production", body.Data.Connections[0].Name)
	require.Equal(t, "www.example.com", body.Data.Domains[0].Domain)
}

func TestServiceCenterHandlerOverviewIncludesProviders(t *testing.T) {
	handler := newServiceCenterTestHandler(t)

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/api/admin/services/overview", nil)

	handler.Overview(context)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Code int `json:"code"`
		Data struct {
			Providers []struct {
				ID            string `json:"id"`
				ResourceCount int    `json:"resource_count"`
			} `json:"providers"`
			Network struct {
				Summary struct {
					ExplicitRuleCount int `json:"explicit_rule_count"`
					InferredItemCount int `json:"inferred_item_count"`
					VPSCount          int `json:"vps_count"`
				} `json:"summary"`
			} `json:"network"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Len(t, body.Data.Providers, 4)
	require.Equal(t, serviceProviderCloudflare, body.Data.Providers[0].ID)
	require.Equal(t, 2, body.Data.Providers[0].ResourceCount)
	require.Equal(t, "github", body.Data.Providers[2].ID)
	require.Equal(t, 1, body.Data.Providers[2].ResourceCount)
	require.Equal(t, 1, body.Data.Network.Summary.ExplicitRuleCount)
	require.Equal(t, 2, body.Data.Network.Summary.InferredItemCount)
	require.Equal(t, 1, body.Data.Network.Summary.VPSCount)
}

func newServiceCenterTestHandler(t *testing.T) *ServiceCenterHandler {
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
		&ops.NetworkRule{},
	))

	connector := ops.Connector{
		Name:        "Cloudflare Production",
		Provider:    ops.ConnectorProviderCloudflare,
		Environment: ops.ConnectorEnvironmentProduction,
		AuthType:    ops.ConnectorAuthManual,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	require.NoError(t, db.Create(&connector).Error)
	require.NoError(t, db.Create(&ops.Connector{
		Name:        "Hostinger Production",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentProduction,
		AuthType:    ops.ConnectorAuthManual,
		Status:      ops.ConnectorStatusPending,
		Enabled:     true,
	}).Error)
	githubConnector := &ops.Connector{
		Name:        "GitHub Production",
		Provider:    ops.ConnectorProviderGitHub,
		Environment: ops.ConnectorEnvironmentProduction,
		AuthType:    ops.ConnectorAuthBearer,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	require.NoError(t, db.Create(githubConnector).Error)
	require.NoError(t, db.Create(&ops.Connector{
		Name:        "GHCR Production",
		Provider:    ops.ConnectorProviderGHCR,
		Environment: ops.ConnectorEnvironmentProduction,
		AuthType:    ops.ConnectorAuthBearer,
		Status:      ops.ConnectorStatusPending,
		Enabled:     true,
	}).Error)
	vps := &ops.VPSBinding{
		Name:        "Hostinger Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	require.NoError(t, db.Create(vps).Error)
	project := &ops.ProjectBinding{
		Name:         "commerce-platform",
		VPSBindingID: vps.ID,
		ConnectorID:  &githubConnector.ID,
		Environment:  ops.ProjectEnvironmentProduction,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthUnknown,
		Enabled:      true,
	}
	require.NoError(t, db.Create(project).Error)
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:           "www.example.com",
		ConnectorID:      &connector.ID,
		ProjectBindingID: &project.ID,
		Role:             ops.DomainRoleCanonical,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Zone:             "example.com",
		Status:           ops.DomainStatusActive,
		ObservedStatus:   ops.DomainObservedMatched,
		Enabled:          true,
	}).Error)
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:         "staging.example.com",
		Role:           ops.DomainRoleCanonical,
		Environment:    ops.DomainEnvironmentStaging,
		Provider:       ops.DomainProviderCloudflare,
		Zone:           "example.com",
		Status:         ops.DomainStatusActive,
		ObservedStatus: ops.DomainObservedMatched,
		Enabled:        true,
	}).Error)
	require.NoError(t, db.Create(&ops.NetworkRule{
		Name:           "Hostinger ingress",
		Environment:    ops.DomainEnvironmentProduction,
		VPSBindingID:   &vps.ID,
		OwnerKind:      ops.NetworkOwnerVPS,
		OwnerID:        vps.ID,
		ManagedBy:      ops.NetworkManagedByHostinger,
		SourceKind:     ops.NetworkSourceFirewallRule,
		Scope:          ops.NetworkScopeOSFirewall,
		Direction:      ops.NetworkDirectionIngress,
		Protocol:       ops.NetworkProtocolTCP,
		Ports:          "80,443",
		DesiredState:   ops.NetworkStateOpen,
		ObservedState:  ops.NetworkStateUnknown,
		EffectiveState: ops.NetworkStateUnknown,
		Status:         ops.NetworkStatusPending,
		Enabled:        true,
	}).Error)

	connectorService := service.NewOpsConnectorService(repository.NewOpsConnectorRepository(db))
	domainService := service.NewOpsDomainBindingService(
		repository.NewOpsDomainBindingRepository(db),
		repository.NewOpsProjectBindingRepository(db),
		repository.NewOpsConnectorRepository(db),
	)
	handler := NewServiceCenterHandler(connectorService, domainService, NewOpsConnectorHandler(connectorService))
	handler.ConfigureOpsNetworkSummaryService(service.NewOpsNetworkSummaryService(
		repository.NewOpsNetworkRuleRepository(db),
		repository.NewOpsVPSBindingRepository(db),
		repository.NewOpsProjectBindingRepository(db),
		repository.NewOpsDomainBindingRepository(db),
		repository.NewOpsConnectorRepository(db),
	))
	handler.ConfigureOpsOverviewService(service.NewOpsOverviewService(
		repository.NewOpsDomainBindingRepository(db),
		repository.NewOpsConnectorRepository(db),
		repository.NewOpsVPSBindingRepository(db),
		repository.NewOpsProjectBindingRepository(db),
		nil,
	))
	return handler
}
