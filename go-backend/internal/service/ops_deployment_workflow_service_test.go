package service

import (
	"context"
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
	promoteOpsDeploymentWorkflowFixtureToProduction(t, db, projectID)
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

func TestOpsDeploymentWorkflowProductionRejectsNonProductionProject(t *testing.T) {
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

	_, err := workflowService.CreateProduction(OpsDeploymentWorkflowCreateInput{ProjectID: projectID})
	require.ErrorIs(t, err, ErrOpsDeploymentWorkflowProductionEnv)
}

func TestOpsDeploymentWorkflowProductionRejectsUnsupportedRefAtCreate(t *testing.T) {
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")
	db, projectID := newOpsDeploymentWorkflowTestDB(t)
	promoteOpsDeploymentWorkflowFixtureToProduction(t, db, projectID)
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

	_, err := workflowService.CreateProduction(OpsDeploymentWorkflowCreateInput{
		ProjectID:    projectID,
		RequestedRef: "sha-cccccccccccccccccccccccccccccccccccccccc",
	})
	require.ErrorIs(t, err, ErrOpsDeploymentWorkflowUnsupportedRef)
}

func TestProductionWorkflowStartUpdatesHydratesRuntimeFields(t *testing.T) {
	run := &ops.DeploymentWorkflowRun{ID: 44}
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			ID:              9,
			CurrentImageTag: "sha-previous",
		},
	}
	report := &ops.DeploymentPreflight{StatusLevel: ops.DeploymentStatusReady}

	updates := productionWorkflowStartUpdates(run, project, report, []byte(`{"status_level":"ready"}`))

	require.Equal(t, ops.DeploymentWorkflowStatusRunning, run.Status)
	require.Equal(t, ops.DeploymentStatusReady, run.PreflightStatus)
	require.Equal(t, "sha-previous", run.PreviousRef)
	require.Equal(t, "sha-previous", run.RollbackRef)
	require.Equal(t, "ops-deploy-44-9", run.IdempotencyKey)
	require.NotNil(t, run.StartedAt)
	require.Nil(t, run.CompletedAt)
	require.Equal(t, "ops-deploy-44-9", updates["idempotency_key"])
	require.Equal(t, run.PreviousRef, updates["rollback_ref"])
}

func TestProductionRollbackRefPrefersDeployableCommit(t *testing.T) {
	const commit = "cccccccccccccccccccccccccccccccccccccccc"
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			CurrentImageTag:  "sha-dddddddddddddddddddddddddddddddddddddddd",
			CurrentCommitSHA: commit,
		},
	}

	require.Equal(t, commit, productionRollbackRef(project))
}

func TestProductionRollbackRefNormalizesSHAImageTag(t *testing.T) {
	const commit = "cccccccccccccccccccccccccccccccccccccccc"
	project := &ops.ProjectBindingView{
		ProjectBinding: ops.ProjectBinding{
			CurrentImageTag: "sha-" + commit,
		},
	}

	require.Equal(t, commit, productionRollbackRef(project))
}

func TestOpsDeploymentWorkflowRetryFailedStepContinuesAfterSucceededSteps(t *testing.T) {
	t.Setenv("HOSTINGER_API_TOKEN", "configured-token")
	db, projectID := newOpsDeploymentWorkflowTestDB(t)
	workflowRepo := repository.NewOpsDeploymentWorkflowRepository(db)
	workflowService := NewOpsDeploymentWorkflowService(
		workflowRepo,
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
	})
	require.NoError(t, err)
	run, err = workflowService.Validate(run.ID)
	require.NoError(t, err)
	run, err = workflowService.Approve(run.ID, OpsDeploymentWorkflowActor{UserID: 1, Username: "admin"})
	require.NoError(t, err)
	require.Len(t, run.Steps, 8)

	require.NoError(t, workflowRepo.UpdateStep(run.Steps[0].ID, map[string]interface{}{
		"status":         ops.DeploymentWorkflowStepSucceeded,
		"output_summary": "already checked",
		"completed_at":   time.Now().UTC(),
	}))
	require.NoError(t, workflowRepo.UpdateStep(run.Steps[1].ID, map[string]interface{}{
		"status":        ops.DeploymentWorkflowStepFailed,
		"error_message": "temporary read failure",
		"completed_at":  time.Now().UTC(),
	}))
	require.NoError(t, workflowRepo.UpdateRun(run.ID, map[string]interface{}{
		"status":       ops.DeploymentWorkflowStatusFailed,
		"last_error":   "temporary read failure",
		"completed_at": time.Now().UTC(),
	}))

	run, err = workflowService.RetryFailedStep(context.Background(), run.ID)
	require.NoError(t, err)
	require.Equal(t, ops.DeploymentWorkflowStatusSucceeded, run.Status)
	require.Equal(t, ops.DeploymentWorkflowStepSucceeded, run.Steps[0].Status)
	require.Equal(t, "already checked", run.Steps[0].OutputSummary)
	for _, step := range run.Steps[1:] {
		require.Equal(t, ops.DeploymentWorkflowStepSucceeded, step.Status)
		require.NotEmpty(t, step.OutputSummary)
	}
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

func TestOpsDeploymentWorkflowRollbackExecutesPinnedReleaseAndPersistsEvidence(t *testing.T) {
	db, projectID := newOpsDeploymentWorkflowTestDB(t)
	const commit = "dddddddddddddddddddddddddddddddddddddddd"
	require.NoError(t, db.Model(&ops.VPSBinding{}).Where("id = ?", 1).Updates(map[string]interface{}{
		"environment": "production",
	}).Error)
	require.NoError(t, db.Model(&ops.ProjectBinding{}).Where("id = ?", projectID).Updates(map[string]interface{}{
		"environment": "production",
	}).Error)

	workflowRepo := repository.NewOpsDeploymentWorkflowRepository(db)
	run := &ops.DeploymentWorkflowRun{
		Kind:         ops.DeploymentWorkflowKindDeployment,
		Mode:         ops.DeploymentWorkflowModeProduction,
		ProjectID:    projectID,
		ProjectName:  "staging-project",
		Environment:  "production",
		RequestedRef: "master",
		Status:       ops.DeploymentWorkflowStatusRollbackRequired,
		RollbackRef:  commit,
		PreviousRef:  commit,
		CreatedBy:    "release-manager",
	}
	require.NoError(t, workflowRepo.Create(run, []ops.DeploymentWorkflowStep{
		{
			Sequence:       1,
			Key:            ops.DeploymentWorkflowStepPostHealthCheck,
			Label:          "执行发布后健康检查",
			Status:         ops.DeploymentWorkflowStepFailed,
			Retryable:      true,
			ErrorMessage:   "new release is unhealthy",
			ExternalEffect: false,
		},
		{
			Sequence:       2,
			Key:            ops.DeploymentWorkflowStepPurgeCache,
			Label:          "按域名清理 Cloudflare 缓存",
			Status:         ops.DeploymentWorkflowStepPending,
			Retryable:      true,
			ExternalEffect: true,
		},
		{
			Sequence:       3,
			Key:            ops.DeploymentWorkflowStepRecordRelease,
			Label:          "固化生产发布证据",
			Status:         ops.DeploymentWorkflowStepPending,
			Retryable:      false,
			ExternalEffect: false,
		},
	}))
	run, err := workflowRepo.FindByID(run.ID)
	require.NoError(t, err)

	executor := &fakeOpsDeploymentRollbackExecutor{
		result: &OpsDeploymentRollbackExecutionResult{
			OperationID:   "ssh-rollback-1-dddddddddddd",
			Target:        "prod.example.com:22",
			OutputSummary: "Deploying sha-dddddddddddddddddddddddddddddddddddddddd.",
		},
	}
	healthChecker := &fakeOpsDeploymentHealthChecker{
		report: &ops.DeploymentHealthCheckReport{
			ProjectID: projectID,
			Status:    ops.DeploymentHealthHealthy,
			Summary:   "Rollback health verification passed.",
		},
	}
	workflowService := NewOpsDeploymentWorkflowService(
		workflowRepo,
		repository.NewOpsProjectBindingRepository(db),
		nil,
	)
	workflowService.ConfigureProductionDependencies(
		repository.NewOpsVPSBindingRepository(db),
		nil,
		&fakeOpsDeploymentProjectSyncer{},
		healthChecker,
	)
	workflowService.ConfigureCachePurgeService(&fakeOpsDeploymentCachePurger{
		result: &OpsCloudflareCachePurgeResult{
			Summary: "Cloudflare cache purge completed.",
		},
	})
	workflowService.ConfigureRollbackExecutor(executor)

	result, err := workflowService.Rollback(nil, run.ID)
	require.NoError(t, err)
	require.Equal(t, ops.DeploymentWorkflowStatusRolledBack, result.Status)
	require.Equal(t, executor.result.OperationID, result.RemoteOperationID)
	require.Equal(t, commit, executor.input.RollbackRef)
	require.Equal(t, "production", executor.input.Environment)
	require.Equal(t, 1, executor.calls)

	rollbackStep := workflowStepByKey(result, ops.DeploymentWorkflowStepExecuteRollback)
	require.NotNil(t, rollbackStep)
	require.Equal(t, ops.DeploymentWorkflowStepSucceeded, rollbackStep.Status)
	require.Equal(t, executor.result.OperationID, rollbackStep.ExternalOperationID)
	require.Contains(t, rollbackStep.OutputSummary, "Deploying")
	require.Equal(t, ops.DeploymentWorkflowStepSucceeded, workflowStepByKey(result, ops.DeploymentWorkflowStepPostHealthCheck).Status)
	require.Equal(t, ops.DeploymentWorkflowStepSucceeded, workflowStepByKey(result, ops.DeploymentWorkflowStepPurgeCache).Status)
	require.Equal(t, ops.DeploymentWorkflowStepSucceeded, workflowStepByKey(result, ops.DeploymentWorkflowStepRecordRelease).Status)
}

func TestOpsDeploymentWorkflowRollbackRejectsNonCommitRollbackPoint(t *testing.T) {
	db, projectID := newOpsDeploymentWorkflowTestDB(t)
	require.NoError(t, db.Model(&ops.VPSBinding{}).Where("id = ?", 1).Updates(map[string]interface{}{
		"environment": "production",
	}).Error)
	require.NoError(t, db.Model(&ops.ProjectBinding{}).Where("id = ?", projectID).Updates(map[string]interface{}{
		"environment": "production",
	}).Error)
	workflowRepo := repository.NewOpsDeploymentWorkflowRepository(db)
	run := &ops.DeploymentWorkflowRun{
		Kind:        ops.DeploymentWorkflowKindDeployment,
		Mode:        ops.DeploymentWorkflowModeProduction,
		ProjectID:   projectID,
		ProjectName: "staging-project",
		Environment: "production",
		Status:      ops.DeploymentWorkflowStatusRollbackRequired,
		RollbackRef: "sha-dddddddddddddddddddddddddddddddddddddddd",
	}
	require.NoError(t, workflowRepo.Create(run, nil))
	executor := &fakeOpsDeploymentRollbackExecutor{}
	workflowService := NewOpsDeploymentWorkflowService(
		workflowRepo,
		repository.NewOpsProjectBindingRepository(db),
		nil,
	)
	workflowService.ConfigureRollbackExecutor(executor)

	_, err := workflowService.Rollback(nil, run.ID)
	require.ErrorIs(t, err, ErrOpsDeploymentRollbackInvalidRef)
	require.Zero(t, executor.calls)
}

type fakeOpsDeploymentRollbackExecutor struct {
	calls  int
	input  OpsDeploymentRollbackExecutionInput
	result *OpsDeploymentRollbackExecutionResult
	err    error
}

func (f *fakeOpsDeploymentRollbackExecutor) ExecuteRollback(_ context.Context, input OpsDeploymentRollbackExecutionInput) (*OpsDeploymentRollbackExecutionResult, error) {
	f.calls++
	f.input = input
	return f.result, f.err
}

type fakeOpsDeploymentProjectSyncer struct {
	err error
}

func (f *fakeOpsDeploymentProjectSyncer) SyncProject(_ context.Context, projectID uint) (*ops.HostingerProjectSyncResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &ops.HostingerProjectSyncResult{ProjectID: projectID, HealthStatus: ops.ProjectHealthHealthy}, nil
}

type fakeOpsDeploymentHealthChecker struct {
	report *ops.DeploymentHealthCheckReport
	err    error
}

func (f *fakeOpsDeploymentHealthChecker) CheckProject(_ context.Context, project *ops.ProjectBindingView) (*ops.DeploymentHealthCheckReport, error) {
	if f.err != nil {
		return nil, f.err
	}
	report := *f.report
	report.ProjectID = project.ID
	return &report, nil
}

type fakeOpsDeploymentCachePurger struct {
	result *OpsCloudflareCachePurgeResult
	err    error
}

func (f *fakeOpsDeploymentCachePurger) PurgeProject(_ context.Context, projectID uint) (*OpsCloudflareCachePurgeResult, error) {
	if f.err != nil {
		return nil, f.err
	}
	result := *f.result
	result.ProjectID = projectID
	return &result, nil
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

func promoteOpsDeploymentWorkflowFixtureToProduction(t *testing.T, db *gorm.DB, projectID uint) {
	t.Helper()
	var project ops.ProjectBinding
	require.NoError(t, db.First(&project, projectID).Error)
	require.NoError(t, db.Model(&ops.ProjectBinding{}).Where("id = ?", projectID).Update("environment", ops.ProjectEnvironmentProduction).Error)

	var vps ops.VPSBinding
	require.NoError(t, db.First(&vps, project.VPSBindingID).Error)
	require.NoError(t, db.Model(&ops.VPSBinding{}).Where("id = ?", vps.ID).Update("environment", ops.VPSEnvironmentProduction).Error)
	if vps.ConnectorID != nil {
		require.NoError(t, db.Model(&ops.Connector{}).Where("id = ?", *vps.ConnectorID).Update("environment", ops.ConnectorEnvironmentProduction).Error)
	}
}
