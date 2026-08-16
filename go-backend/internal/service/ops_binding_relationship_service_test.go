package service

import (
	"context"
	"testing"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsProjectBindingRejectsMismatchedVPSEnvironment(t *testing.T) {
	repos := newOpsBindingRelationshipRepositories(t)
	vps := &ops.VPSBinding{
		Name:        "Staging VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentStaging,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(vps).Error)

	projectService := NewOpsProjectBindingService(repos.projects, repos.vps, repos.connectors)
	_, err := projectService.Create(OpsProjectBindingInput{
		Name:         "production-project",
		VPSBindingID: vps.ID,
		Environment:  ops.ProjectEnvironmentProduction,
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "does not match VPS environment")
}

func TestOpsProjectBindingPreservesOmittedConnectorAndClearsExplicitNull(t *testing.T) {
	repos := newOpsBindingRelationshipRepositories(t)
	connector := &ops.Connector{
		Name:        "Hostinger Production",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentProduction,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(connector).Error)
	vps := &ops.VPSBinding{
		Name:        "Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		ConnectorID: &connector.ID,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(vps).Error)
	project := &ops.ProjectBinding{
		Name:         "production-project",
		VPSBindingID: vps.ID,
		ConnectorID:  &connector.ID,
		Environment:  ops.ProjectEnvironmentProduction,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthUnknown,
		Enabled:      true,
	}
	require.NoError(t, repos.db.Create(project).Error)

	projectService := NewOpsProjectBindingService(repos.projects, repos.vps, repos.connectors)
	updated, err := projectService.Update(project.ID, OpsProjectBindingInput{
		Name:         project.Name,
		VPSBindingID: project.VPSBindingID,
		Environment:  project.Environment,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.ConnectorID)
	require.Equal(t, connector.ID, *updated.ConnectorID)

	updated, err = projectService.Update(project.ID, OpsProjectBindingInput{
		Name:           project.Name,
		VPSBindingID:   project.VPSBindingID,
		Environment:    project.Environment,
		ConnectorIDSet: true,
		ConnectorID:    nil,
	})
	require.NoError(t, err)
	require.Nil(t, updated.ConnectorID)
}

func TestOpsVPSBindingRejectsMismatchedConnectorAndPreservesOmittedConnector(t *testing.T) {
	repos := newOpsBindingRelationshipRepositories(t)
	stagingConnector := &ops.Connector{
		Name:        "Hostinger Staging",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentStaging,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(stagingConnector).Error)
	vps := &ops.VPSBinding{
		Name:        "Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(vps).Error)

	vpsService := NewOpsVPSBindingService(repos.vps, repos.connectors, repos.projects)
	_, err := vpsService.Update(vps.ID, OpsVPSBindingInput{
		Name:           vps.Name,
		Provider:       vps.Provider,
		Environment:    vps.Environment,
		ConnectorIDSet: true,
		ConnectorID:    &stagingConnector.ID,
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "does not match connector environment")

	productionConnector := &ops.Connector{
		Name:        "Hostinger Production",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentProduction,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(productionConnector).Error)
	require.NoError(t, repos.vps.Update(&ops.VPSBinding{
		ID:          vps.ID,
		Name:        vps.Name,
		Provider:    vps.Provider,
		Environment: vps.Environment,
		ConnectorID: &productionConnector.ID,
		Status:      vps.Status,
		Enabled:     true,
	}))

	updated, err := vpsService.Update(vps.ID, OpsVPSBindingInput{
		Name:        vps.Name,
		Provider:    vps.Provider,
		Environment: vps.Environment,
	})
	require.NoError(t, err)
	require.NotNil(t, updated.ConnectorID)
	require.Equal(t, productionConnector.ID, *updated.ConnectorID)
}

func TestOpsVPSBindingRejectsProviderOrEnvironmentChangeWhenProjectsAreBound(t *testing.T) {
	repos := newOpsBindingRelationshipRepositories(t)
	vps := &ops.VPSBinding{
		Name:        "Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(vps).Error)
	project := &ops.ProjectBinding{
		Name:         "production-project",
		VPSBindingID: vps.ID,
		Environment:  ops.ProjectEnvironmentProduction,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthUnknown,
		Enabled:      true,
	}
	require.NoError(t, repos.db.Create(project).Error)

	vpsService := NewOpsVPSBindingService(repos.vps, repos.connectors, repos.projects)
	_, err := vpsService.Update(vps.ID, OpsVPSBindingInput{
		Name:        vps.Name,
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentStaging,
	})
	require.ErrorIs(t, err, ErrInvalidOpsVPSBinding)
	require.ErrorContains(t, err, "projects are bound")

	updated, err := vpsService.Update(vps.ID, OpsVPSBindingInput{
		Name:        vps.Name,
		Provider:    vps.Provider,
		Environment: vps.Environment,
		Hostname:    "prod-1.example.com",
	})
	require.NoError(t, err)
	require.Equal(t, "prod-1.example.com", updated.Hostname)
}

func TestOpsVPSBindingListsOnlyRequestedEnvironment(t *testing.T) {
	repos := newOpsBindingRelationshipRepositories(t)
	require.NoError(t, repos.db.Create(&ops.VPSBinding{
		Name:        "Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}).Error)
	require.NoError(t, repos.db.Create(&ops.VPSBinding{
		Name:        "Staging VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentStaging,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}).Error)

	vpsService := NewOpsVPSBindingService(repos.vps, repos.connectors, repos.projects)
	records, err := vpsService.ListForEnvironment(" PRODUCTION ")
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, "Production VPS", records[0].Name)

	records, err = vpsService.List()
	require.NoError(t, err)
	require.Len(t, records, 2)

	_, err = vpsService.ListForEnvironment("qa")
	require.ErrorIs(t, err, ErrInvalidOpsVPSEnvironment)
}

func TestOpsProjectBindingListsOnlyRequestedEnvironment(t *testing.T) {
	repos := newOpsBindingRelationshipRepositories(t)
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
	require.NoError(t, repos.db.Create(productionVPS).Error)
	require.NoError(t, repos.db.Create(stagingVPS).Error)
	require.NoError(t, repos.db.Create(&ops.ProjectBinding{
		Name:         "Production Project",
		VPSBindingID: productionVPS.ID,
		Environment:  ops.ProjectEnvironmentProduction,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthUnknown,
		Enabled:      true,
	}).Error)
	require.NoError(t, repos.db.Create(&ops.ProjectBinding{
		Name:         "Staging Project",
		VPSBindingID: stagingVPS.ID,
		Environment:  ops.ProjectEnvironmentStaging,
		Status:       ops.ProjectStatusActive,
		HealthStatus: ops.ProjectHealthUnknown,
		Enabled:      true,
	}).Error)

	projectService := NewOpsProjectBindingService(repos.projects, repos.vps, repos.connectors)
	records, err := projectService.ListForEnvironment(" STAGING ")
	require.NoError(t, err)
	require.Len(t, records, 1)
	require.Equal(t, "Staging Project", records[0].Name)

	records, err = projectService.List()
	require.NoError(t, err)
	require.Len(t, records, 2)

	_, err = projectService.ListForEnvironment("qa")
	require.ErrorIs(t, err, ErrInvalidOpsProjectEnvironment)
}

func TestOpsHostingerSyncRejectsMismatchedProjectConnectorEnvironment(t *testing.T) {
	repos := newOpsBindingRelationshipRepositories(t)
	connector := &ops.Connector{
		Name:        "Hostinger Staging",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentStaging,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(connector).Error)
	vps := &ops.VPSBinding{
		Name:               "Production VPS",
		Provider:           ops.VPSProviderHostinger,
		Environment:        ops.VPSEnvironmentProduction,
		ConnectorID:        &connector.ID,
		ProviderResourceID: "production-vps",
		Status:             ops.VPSStatusActive,
		ObservedStatus:     ops.VPSObservedUnknown,
		Enabled:            true,
	}
	require.NoError(t, repos.db.Create(vps).Error)
	project := &ops.ProjectBinding{
		Name:               "Production Project",
		VPSBindingID:       vps.ID,
		Environment:        ops.ProjectEnvironmentProduction,
		ComposeProjectName: "production-project",
		Status:             ops.ProjectStatusActive,
		HealthStatus:       ops.ProjectHealthUnknown,
		Enabled:            true,
	}
	require.NoError(t, repos.db.Create(project).Error)

	syncService := NewOpsHostingerSyncService(
		repos.vps,
		repos.projects,
		NewOpsConnectorService(repos.connectors),
	)
	result, err := syncService.SyncProject(context.Background(), project.ID)
	require.ErrorIs(t, err, ErrOpsHostingerSync)
	require.NotNil(t, result)
	require.Contains(t, result.ObservedError, "环境与项目环境不一致")

	refreshed, err := repos.projects.FindByID(project.ID)
	require.NoError(t, err)
	require.Equal(t, ops.ProjectHealthUnknown, refreshed.HealthStatus)
	require.Equal(t, result.ObservedError, refreshed.LastError)
}

func TestOpsProjectBindingRejectsMismatchedInheritedConnector(t *testing.T) {
	repos := newOpsBindingRelationshipRepositories(t)
	connector := &ops.Connector{
		Name:        "Hostinger Staging",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentStaging,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(connector).Error)
	vps := &ops.VPSBinding{
		Name:        "Production VPS",
		Provider:    ops.VPSProviderHostinger,
		Environment: ops.VPSEnvironmentProduction,
		ConnectorID: &connector.ID,
		Status:      ops.VPSStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(vps).Error)

	projectService := NewOpsProjectBindingService(repos.projects, repos.vps, repos.connectors)
	_, err := projectService.Create(OpsProjectBindingInput{
		Name:               "Production Project",
		VPSBindingID:       vps.ID,
		Environment:        ops.ProjectEnvironmentProduction,
		ComposeProjectName: "production-project",
	})
	require.ErrorIs(t, err, ErrInvalidOpsProjectBinding)
	require.ErrorContains(t, err, "does not match connector environment")
}

func TestOpsDomainBindingRejectsMismatchedConnectorProvider(t *testing.T) {
	repos := newOpsBindingRelationshipRepositories(t)
	connector := &ops.Connector{
		Name:        "Hostinger Production",
		Provider:    ops.ConnectorProviderHostinger,
		Environment: ops.ConnectorEnvironmentProduction,
		Status:      ops.ConnectorStatusActive,
		Enabled:     true,
	}
	require.NoError(t, repos.db.Create(connector).Error)
	domainService := NewOpsDomainBindingService(repos.domains, repos.projects, repos.connectors)
	_, err := domainService.Create(OpsDomainBindingInput{
		Domain:         "learn.example.com",
		ConnectorID:    &connector.ID,
		ConnectorIDSet: true,
		Role:           ops.DomainRoleInternal,
		Environment:    ops.DomainEnvironmentProduction,
		Provider:       ops.DomainProviderCloudflare,
	})
	require.Error(t, err)
	require.ErrorContains(t, err, "does not match connector provider")
}

type opsBindingRelationshipRepositories struct {
	db         *gorm.DB
	projects   *repository.OpsProjectBindingRepository
	vps        *repository.OpsVPSBindingRepository
	connectors *repository.OpsConnectorRepository
	domains    *repository.OpsDomainBindingRepository
}

func newOpsBindingRelationshipRepositories(t *testing.T) opsBindingRelationshipRepositories {
	t.Helper()
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
	return opsBindingRelationshipRepositories{
		db:         db,
		projects:   repository.NewOpsProjectBindingRepository(db),
		vps:        repository.NewOpsVPSBindingRepository(db),
		connectors: repository.NewOpsConnectorRepository(db),
		domains:    repository.NewOpsDomainBindingRepository(db),
	}
}
