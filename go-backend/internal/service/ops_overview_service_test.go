package service

import (
	"testing"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsOverviewFiltersByEnvironment(t *testing.T) {
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

	productionConnector := &ops.Connector{
		Name:        "Production Hostinger",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentProduction,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	stagingConnector := &ops.Connector{
		Name:        "Staging Hostinger",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentStaging,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	require.NoError(t, db.Create(productionConnector).Error)
	require.NoError(t, db.Create(stagingConnector).Error)

	productionVPS := &ops.VPSBinding{
		Name:               "Production VPS",
		Provider:           ops.VPSProviderHostinger,
		Environment:        ops.VPSEnvironmentProduction,
		ConnectorID:        &productionConnector.ID,
		ProviderResourceID: "prod-vps",
		Status:             ops.VPSStatusActive,
		ObservedStatus:     ops.VPSObservedHealthy,
		Enabled:            true,
	}
	stagingVPS := &ops.VPSBinding{
		Name:               "Staging VPS",
		Provider:           ops.VPSProviderHostinger,
		Environment:        ops.VPSEnvironmentStaging,
		ConnectorID:        &stagingConnector.ID,
		ProviderResourceID: "staging-vps",
		Status:             ops.VPSStatusActive,
		ObservedStatus:     ops.VPSObservedHealthy,
		Enabled:            true,
	}
	require.NoError(t, db.Create(productionVPS).Error)
	require.NoError(t, db.Create(stagingVPS).Error)

	require.NoError(t, db.Create(&ops.ProjectBinding{
		Name:         "production-project",
		VPSBindingID: productionVPS.ID,
		Environment:  ops.ProjectEnvironmentProduction,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthHealthy,
		Enabled:      true,
	}).Error)
	require.NoError(t, db.Create(&ops.ProjectBinding{
		Name:         "staging-project",
		VPSBindingID: stagingVPS.ID,
		Environment:  ops.ProjectEnvironmentStaging,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthHealthy,
		Enabled:      true,
	}).Error)

	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:         "www.production.example",
		Role:           ops.DomainRoleCanonical,
		Environment:    ops.DomainEnvironmentProduction,
		Provider:       ops.DomainProviderCloudflare,
		Status:         ops.DomainStatusActive,
		ObservedStatus: ops.DomainObservedMatched,
		Enabled:        true,
	}).Error)
	require.NoError(t, db.Create(&ops.DomainBinding{
		Domain:         "www.staging.example",
		Role:           ops.DomainRoleCanonical,
		Environment:    ops.DomainEnvironmentStaging,
		Provider:       ops.DomainProviderCloudflare,
		Status:         ops.DomainStatusActive,
		ObservedStatus: ops.DomainObservedMatched,
		Enabled:        true,
	}).Error)

	service := NewOpsOverviewService(
		repository.NewOpsDomainBindingRepository(db),
		repository.NewOpsConnectorRepository(db),
		repository.NewOpsVPSBindingRepository(db),
		repository.NewOpsProjectBindingRepository(db),
		nil,
	)

	overview, err := service.GetForEnvironment(ops.DomainEnvironmentProduction)
	require.NoError(t, err)
	require.Equal(t, ops.DomainEnvironmentProduction, overview.Environment)
	require.Len(t, overview.Topology.Domains, 1)
	require.Len(t, overview.Topology.VPS, 1)
	require.Len(t, overview.Topology.Projects, 1)
	require.Equal(t, "www.production.example", overview.Topology.Domains[0].Domain)
	require.Equal(t, "Production VPS", overview.Topology.VPS[0].Name)
	require.Equal(t, "production-project", overview.Topology.Projects[0].Name)
	require.Equal(t, 1, overview.Summary["connectors"].Total)

	defaultOverview, err := service.Get()
	require.NoError(t, err)
	require.Equal(t, ops.DomainEnvironmentProduction, defaultOverview.Environment)

	_, err = service.GetForEnvironment("qa")
	require.ErrorIs(t, err, ErrInvalidOpsOverviewEnvironment)
}

func TestSummarizeDomainsKeepsUnknownSeparateFromHealthy(t *testing.T) {
	summary := summarizeDomains([]ops.DomainBinding{
		{Enabled: true, Status: ops.DomainStatusActive, ObservedStatus: ops.DomainObservedUnknown},
		{Enabled: true, Status: ops.DomainStatusActive, ObservedStatus: ops.DomainObservedMatched},
		{Enabled: false, Status: ops.DomainStatusDrifted, ObservedStatus: ops.DomainObservedDrifted},
	})

	if summary.Total != 3 {
		t.Fatalf("Total = %d, want 3", summary.Total)
	}
	if summary.Enabled != 2 {
		t.Fatalf("Enabled = %d, want 2", summary.Enabled)
	}
	if summary.Unknown != 1 {
		t.Fatalf("Unknown = %d, want 1", summary.Unknown)
	}
	if summary.Healthy != 1 {
		t.Fatalf("Healthy = %d, want 1", summary.Healthy)
	}
	if summary.Attention != 2 {
		t.Fatalf("Attention = %d, want 2", summary.Attention)
	}
}

func TestIsDomainAttentionTreatsUnknownAsAttention(t *testing.T) {
	if !isDomainAttention(ops.DomainBinding{
		Status:         ops.DomainStatusActive,
		ObservedStatus: ops.DomainObservedUnknown,
	}) {
		t.Fatal("unknown observed status should require attention")
	}
	if isDomainAttention(ops.DomainBinding{
		Status:         ops.DomainStatusActive,
		ObservedStatus: ops.DomainObservedMatched,
	}) {
		t.Fatal("matched active domain should not require attention")
	}
}
