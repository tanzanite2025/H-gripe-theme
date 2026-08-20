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

func TestOpsNetworkSummaryIncludesExplicitRulesAndInferredDomainReferences(t *testing.T) {
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

	hostinger := &ops.Connector{
		Name:        "Hostinger Production",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentProduction,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	cloudflare := &ops.Connector{
		Name:        "Cloudflare Production",
		Provider:    ops.ConnectorProviderCloudflare,
		Environment: ops.ConnectorEnvironmentProduction,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	require.NoError(t, db.Create(hostinger).Error)
	require.NoError(t, db.Create(cloudflare).Error)

	vps := &ops.VPSBinding{
		Name:        "Hostinger Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		ConnectorID: &hostinger.ID,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	require.NoError(t, db.Create(vps).Error)
	project := &ops.ProjectBinding{
		Name:         "commerce-platform",
		VPSBindingID: vps.ID,
		Environment:  ops.ProjectEnvironmentProduction,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthHealthy,
		Enabled:      true,
	}
	require.NoError(t, db.Create(project).Error)
	domain := &ops.DomainBinding{
		Domain:           "learn.example.com",
		ConnectorID:      &cloudflare.ID,
		ProjectBindingID: &project.ID,
		Role:             ops.DomainRoleCanonical,
		Environment:      ops.DomainEnvironmentProduction,
		Provider:         ops.DomainProviderCloudflare,
		Target:           "theme-web:8080",
		ProxyMode:        ops.DomainProxyProxied,
		Status:           ops.DomainStatusActive,
		ObservedStatus:   ops.DomainObservedMatched,
		Enabled:          true,
	}
	require.NoError(t, db.Create(domain).Error)
	require.NoError(t, db.Create(&ops.NetworkRule{
		Name:           "Hostinger ingress",
		Environment:    ops.DomainEnvironmentProduction,
		VPSBindingID:   &vps.ID,
		ConnectorID:    &hostinger.ID,
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
	require.NoError(t, db.Create(&ops.NetworkRule{
		Name:            "Cloudflare zone policy",
		Environment:     ops.DomainEnvironmentProduction,
		DomainBindingID: &domain.ID,
		OwnerKind:       ops.NetworkOwnerDomain,
		OwnerID:         domain.ID,
		ManagedBy:       ops.NetworkManagedByCloudflare,
		SourceKind:      ops.NetworkSourceProvider,
		Scope:           ops.NetworkScopeEdge,
		Direction:       ops.NetworkDirectionIngress,
		Protocol:        ops.NetworkProtocolTCP,
		DesiredState:    ops.NetworkStateOpen,
		ObservedState:   ops.NetworkStateUnknown,
		EffectiveState:  ops.NetworkStateUnknown,
		Status:          ops.NetworkStatusPending,
		Enabled:         true,
	}).Error)

	summaryService := NewOpsNetworkSummaryService(
		repository.NewOpsNetworkRuleRepository(db),
		repository.NewOpsVPSBindingRepository(db),
		repository.NewOpsProjectBindingRepository(db),
		repository.NewOpsDomainBindingRepository(db),
		repository.NewOpsConnectorRepository(db),
	)
	summary, err := summaryService.Get(ops.DomainEnvironmentProduction)
	require.NoError(t, err)
	require.Equal(t, 3, summary.Summary.Total)
	require.Equal(t, 2, summary.Summary.ExplicitRuleCount)
	require.Equal(t, 1, summary.Summary.InferredItemCount)
	require.Equal(t, 1, summary.Summary.VPSCount)
	require.Equal(t, 2, summary.Summary.Attention)

	itemsByName := make(map[string]ops.NetworkSummaryItem, len(summary.Items))
	for _, item := range summary.Items {
		itemsByName[item.Name] = item
	}
	require.Equal(t, "Hostinger Production VPS", itemsByName["Hostinger ingress"].OwnerName)
	require.Equal(t, "Hostinger Production VPS", itemsByName["Hostinger ingress"].VPSName)
	require.Equal(t, "learn.example.com", itemsByName["Cloudflare zone policy"].OwnerName)
	require.Equal(t, "commerce-platform", itemsByName["Cloudflare zone policy"].ProjectName)
	require.Equal(t, "Hostinger Production VPS", itemsByName["Cloudflare zone policy"].VPSName)
	require.Equal(t, "Cloudflare Production", itemsByName["Cloudflare zone policy"].ConnectorName)
	require.Equal(t, ops.NetworkScopeEdge, itemsByName["learn.example.com"].Scope)
	require.Equal(t, "commerce-platform", itemsByName["learn.example.com"].ProjectName)
}
