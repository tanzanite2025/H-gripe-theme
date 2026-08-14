package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestOpsDeploymentWorkflowDryRunLifecycle(t *testing.T) {
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")
	db, projectID := newOpsDeploymentWorkflowTestDB(t)
	workflowService := NewOpsDeploymentWorkflowService(
		repository.NewOpsDeploymentWorkflowRepository(db),
		repository.NewOpsProjectBindingRepository(db),
		NewOpsDeploymentPreflightService(
			repository.NewOpsProjectBindingRepository(db),
			repository.NewOpsVPSBindingRepository(db),
			repository.NewOpsConnectorRepository(db),
			repository.NewOpsDomainBindingRepository(db),
		),
	)

	run, err := workflowService.CreateDryRun(OpsDeploymentWorkflowCreateInput{
		ProjectID:    projectID,
		RequestedRef: "sha-cccccccccccccccccccccccccccccccccccccccc",
		CreatedByID:  7,
		CreatedBy:    "ops-admin",
	})
	require.NoError(t, err)
	require.NotZero(t, run.ID)
	require.Equal(t, ops.DeploymentWorkflowStatusDraft, run.Status)
	require.Equal(t, ops.DeploymentWorkflowModeDryRun, run.Mode)
	require.Len(t, run.Steps, 8)
	require.NotZero(t, run.Steps[0].ID)

	run, err = workflowService.Validate(run.ID)
	require.NoError(t, err)
	require.Equal(t, ops.DeploymentWorkflowStatusAwaitingApproval, run.Status)
	require.NotEmpty(t, run.Preflight)

	run, err = workflowService.Approve(run.ID, OpsDeploymentWorkflowActor{
		UserID:   9,
		Username: "release-manager",
	})
	require.NoError(t, err)
	require.Equal(t, ops.DeploymentWorkflowStatusValidated, run.Status)
	require.Equal(t, "release-manager", run.ApprovedBy)
	require.NotNil(t, run.ApprovedAt)

	run, err = workflowService.ExecuteDryRun(run.ID)
	require.NoError(t, err)
	require.Equal(t, ops.DeploymentWorkflowStatusSucceeded, run.Status)
	require.NotNil(t, run.StartedAt)
	require.NotNil(t, run.CompletedAt)
	for _, step := range run.Steps {
		require.Equal(t, ops.DeploymentWorkflowStepSucceeded, step.Status)
		require.NotEmpty(t, step.OutputSummary)
		require.False(t, step.ExternalEffect)
	}
}

func TestOpsDeploymentWorkflowCannotApproveBlockedPreflight(t *testing.T) {
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")
	db, projectID := newOpsDeploymentWorkflowTestDB(t)
	require.NoError(t, db.Model(&ops.ProjectBinding{}).Where("id = ?", projectID).Updates(map[string]interface{}{
		"compose_source": "",
	}).Error)
	workflowService := NewOpsDeploymentWorkflowService(
		repository.NewOpsDeploymentWorkflowRepository(db),
		repository.NewOpsProjectBindingRepository(db),
		NewOpsDeploymentPreflightService(
			repository.NewOpsProjectBindingRepository(db),
			repository.NewOpsVPSBindingRepository(db),
			repository.NewOpsConnectorRepository(db),
			repository.NewOpsDomainBindingRepository(db),
		),
	)

	run, err := workflowService.CreateDryRun(OpsDeploymentWorkflowCreateInput{ProjectID: projectID})
	require.NoError(t, err)
	run, err = workflowService.Validate(run.ID)
	require.NoError(t, err)
	require.Equal(t, ops.DeploymentWorkflowStatusDraft, run.Status)
	require.Equal(t, ops.DeploymentStatusBlocked, run.PreflightStatus)

	_, err = workflowService.Approve(run.ID, OpsDeploymentWorkflowActor{UserID: 1, Username: "admin"})
	require.ErrorIs(t, err, ErrOpsDeploymentWorkflowInvalidTransition)
}

func TestOpsDeploymentWorkflowProductionCreatesControlledSteps(t *testing.T) {
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")
	db, projectID := newOpsDeploymentWorkflowTestDB(t)
	workflowService := NewOpsDeploymentWorkflowService(
		repository.NewOpsDeploymentWorkflowRepository(db),
		repository.NewOpsProjectBindingRepository(db),
		NewOpsDeploymentPreflightService(
			repository.NewOpsProjectBindingRepository(db),
			repository.NewOpsVPSBindingRepository(db),
			repository.NewOpsConnectorRepository(db),
			repository.NewOpsDomainBindingRepository(db),
		),
	)

	run, err := workflowService.CreateProduction(OpsDeploymentWorkflowCreateInput{
		ProjectID: projectID,
		CreatedBy: "release-manager",
	})
	require.NoError(t, err)
	require.Equal(t, ops.DeploymentWorkflowModeProduction, run.Mode)
	require.Equal(t, "master", run.RequestedRef)
	require.Len(t, run.Steps, 10)

	var externalSteps int
	for _, step := range run.Steps {
		if step.ExternalEffect {
			externalSteps++
			require.Contains(t, []string{
				ops.DeploymentWorkflowStepUpdateProject,
				ops.DeploymentWorkflowStepPurgeCache,
			}, step.Key)
		}
	}
	require.Equal(t, 2, externalSteps)
}

func TestOpsDeploymentWorkflowProjectLock(t *testing.T) {
	db, projectID := newOpsDeploymentWorkflowTestDB(t)
	repo := repository.NewOpsDeploymentWorkflowRepository(db)
	first := &ops.DeploymentWorkflowRun{ProjectID: projectID}
	second := &ops.DeploymentWorkflowRun{ProjectID: projectID}
	require.NoError(t, db.Create(first).Error)
	require.NoError(t, db.Create(second).Error)

	require.NoError(t, repo.AcquireProjectLock(projectID, first.ID, time.Minute))
	require.ErrorIs(t, repo.AcquireProjectLock(projectID, second.ID, time.Minute), repository.ErrOpsDeploymentWorkflowLockHeld)
	require.NoError(t, repo.ReleaseWorkflowLocks(first.ID))
	require.NoError(t, repo.AcquireProjectLock(projectID, second.ID, time.Minute))
}

func newOpsDeploymentWorkflowTestDB(t *testing.T) (*gorm.DB, uint) {
	t.Helper()
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
		&ops.DeploymentWorkflowRun{},
		&ops.DeploymentWorkflowStep{},
		&ops.DeploymentWorkflowLock{},
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

	now := time.Date(2026, 8, 14, 5, 0, 0, 0, time.UTC)
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
		Volumes:                "uploads",
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
		RestoreNotes:           "2026-08-14 staging restore exercise completed.",
		Enabled:                true,
	}
	require.NoError(t, db.Create(project).Error)
	return db, project.ID
}
