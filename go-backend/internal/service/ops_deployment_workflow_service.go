package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/ops"
	"commerce-platform/internal/repository"
)

var (
	ErrOpsDeploymentWorkflowInvalidTransition = errors.New("invalid operations deployment workflow transition")
	ErrOpsDeploymentWorkflowPreflightBlocked  = errors.New("operations deployment workflow is blocked by preflight")
	ErrOpsDeploymentWorkflowUnsupportedMode   = errors.New("operations deployment workflow mode is not supported")
	ErrOpsDeploymentWorkflowUnsupportedRef    = errors.New("operations production deployment ref is not supported")
)

type OpsDeploymentWorkflowService struct {
	workflowRepo     *repository.OpsDeploymentWorkflowRepository
	projectRepo      *repository.OpsProjectBindingRepository
	vpsRepo          *repository.OpsVPSBindingRepository
	preflightService *OpsDeploymentPreflightService
	connectorService *OpsConnectorService
	hostingerSync    *OpsHostingerSyncService
	healthCheck      *OpsDeploymentHealthCheckService
	cachePurge       *OpsCloudflareCachePurgeService
}

type OpsDeploymentWorkflowCreateInput struct {
	ProjectID    uint
	RequestedRef string
	CreatedByID  uint
	CreatedBy    string
}

type OpsDeploymentWorkflowActor struct {
	UserID   uint
	Username string
}

func NewOpsDeploymentWorkflowService(
	workflowRepo *repository.OpsDeploymentWorkflowRepository,
	projectRepo *repository.OpsProjectBindingRepository,
	preflightService *OpsDeploymentPreflightService,
) *OpsDeploymentWorkflowService {
	return &OpsDeploymentWorkflowService{
		workflowRepo:     workflowRepo,
		projectRepo:      projectRepo,
		preflightService: preflightService,
	}
}

func (s *OpsDeploymentWorkflowService) ConfigureProductionDependencies(
	vpsRepo *repository.OpsVPSBindingRepository,
	connectorService *OpsConnectorService,
	hostingerSync *OpsHostingerSyncService,
	healthCheck *OpsDeploymentHealthCheckService,
) {
	if s == nil {
		return
	}
	s.vpsRepo = vpsRepo
	s.connectorService = connectorService
	s.hostingerSync = hostingerSync
	s.healthCheck = healthCheck
}

func (s *OpsDeploymentWorkflowService) ConfigureCachePurgeService(cachePurge *OpsCloudflareCachePurgeService) {
	if s == nil {
		return
	}
	s.cachePurge = cachePurge
}

func (s *OpsDeploymentWorkflowService) List(projectID uint) ([]ops.DeploymentWorkflowRun, error) {
	if s == nil || s.workflowRepo == nil {
		return nil, errors.New("operations deployment workflow service is not configured")
	}
	return s.workflowRepo.List(projectID, 25)
}

func (s *OpsDeploymentWorkflowService) Get(id uint) (*ops.DeploymentWorkflowRun, error) {
	if s == nil || s.workflowRepo == nil {
		return nil, errors.New("operations deployment workflow service is not configured")
	}
	return s.workflowRepo.FindByID(id)
}

func (s *OpsDeploymentWorkflowService) CreateDryRun(input OpsDeploymentWorkflowCreateInput) (*ops.DeploymentWorkflowRun, error) {
	inputMode := ops.DeploymentWorkflowModeDryRun
	return s.createWorkflow(input, inputMode)
}

func (s *OpsDeploymentWorkflowService) CreateProduction(input OpsDeploymentWorkflowCreateInput) (*ops.DeploymentWorkflowRun, error) {
	return s.createWorkflow(input, ops.DeploymentWorkflowModeProduction)
}

func (s *OpsDeploymentWorkflowService) createWorkflow(input OpsDeploymentWorkflowCreateInput, mode string) (*ops.DeploymentWorkflowRun, error) {
	if s == nil || s.workflowRepo == nil || s.projectRepo == nil || s.preflightService == nil {
		return nil, errors.New("operations deployment workflow service is not configured")
	}
	project, err := s.projectRepo.FindByID(input.ProjectID)
	if err != nil {
		return nil, err
	}
	report, err := s.preflightService.EvaluateProject(input.ProjectID)
	if err != nil {
		return nil, err
	}

	requestedRef := strings.TrimSpace(input.RequestedRef)
	if requestedRef == "" {
		if mode == ops.DeploymentWorkflowModeProduction {
			requestedRef = "master"
		} else {
			requestedRef = firstNonEmptyWorkflowValue(project.CurrentCommitSHA, project.CurrentImageTag, "current")
		}
	}
	preflightSnapshot, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("encode deployment preflight snapshot: %w", err)
	}
	run := &ops.DeploymentWorkflowRun{
		Kind:              ops.DeploymentWorkflowKindDeployment,
		Mode:              mode,
		ProjectID:         project.ID,
		ProjectName:       project.Name,
		Environment:       project.Environment,
		RequestedRef:      requestedRef,
		Status:            ops.DeploymentWorkflowStatusDraft,
		PreflightStatus:   report.StatusLevel,
		PreflightSnapshot: string(preflightSnapshot),
		CreatedByID:       input.CreatedByID,
		CreatedBy:         normalizeWorkflowActor(input.CreatedBy),
		Preflight:         report,
	}
	steps := buildWorkflowSteps(project, requestedRef, report, mode)
	if err := s.workflowRepo.Create(run, steps); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindByID(run.ID)
}

func (s *OpsDeploymentWorkflowService) Validate(id uint) (*ops.DeploymentWorkflowRun, error) {
	if s == nil || s.workflowRepo == nil || s.preflightService == nil {
		return nil, errors.New("operations deployment workflow service is not configured")
	}
	run, err := s.workflowRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if run.Mode != ops.DeploymentWorkflowModeDryRun && run.Mode != ops.DeploymentWorkflowModeProduction {
		return nil, ErrOpsDeploymentWorkflowUnsupportedMode
	}
	if !workflowStatusAllowed(run.Status,
		ops.DeploymentWorkflowStatusDraft,
		ops.DeploymentWorkflowStatusAwaitingApproval,
		ops.DeploymentWorkflowStatusValidated,
	) {
		return nil, workflowTransitionError(run.Status, "validate")
	}

	report, err := s.preflightService.EvaluateProject(run.ProjectID)
	if err != nil {
		return nil, err
	}
	snapshot, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("encode deployment preflight snapshot: %w", err)
	}
	updates := map[string]interface{}{
		"preflight_snapshot": string(snapshot),
		"preflight_status":   report.StatusLevel,
		"last_error":         "",
	}
	if report.BlockingCount > 0 {
		updates["status"] = ops.DeploymentWorkflowStatusDraft
		updates["last_error"] = "Preflight contains blocking checks."
	} else {
		updates["status"] = ops.DeploymentWorkflowStatusAwaitingApproval
	}
	if err := s.workflowRepo.UpdateRun(id, updates); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindByID(id)
}

func (s *OpsDeploymentWorkflowService) Approve(id uint, actor OpsDeploymentWorkflowActor) (*ops.DeploymentWorkflowRun, error) {
	if s == nil || s.workflowRepo == nil {
		return nil, errors.New("operations deployment workflow service is not configured")
	}
	run, err := s.workflowRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if run.Mode != ops.DeploymentWorkflowModeDryRun && run.Mode != ops.DeploymentWorkflowModeProduction {
		return nil, ErrOpsDeploymentWorkflowUnsupportedMode
	}
	if run.Status != ops.DeploymentWorkflowStatusAwaitingApproval {
		return nil, workflowTransitionError(run.Status, "approve")
	}
	if run.Preflight == nil || run.Preflight.BlockingCount > 0 || run.PreflightStatus == ops.DeploymentStatusBlocked {
		return nil, ErrOpsDeploymentWorkflowPreflightBlocked
	}
	now := time.Now().UTC()
	approvedBy := normalizeWorkflowActor(actor.Username)
	updates := map[string]interface{}{
		"status":         ops.DeploymentWorkflowStatusValidated,
		"approved_by_id": actor.UserID,
		"approved_by":    approvedBy,
		"approved_at":    now,
		"last_error":     "",
	}
	if err := s.workflowRepo.UpdateRun(id, updates); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindByID(id)
}

func (s *OpsDeploymentWorkflowService) ExecuteDryRun(id uint) (*ops.DeploymentWorkflowRun, error) {
	if s == nil || s.workflowRepo == nil {
		return nil, errors.New("operations deployment workflow service is not configured")
	}
	run, err := s.workflowRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if run.Mode != ops.DeploymentWorkflowModeDryRun {
		return nil, ErrOpsDeploymentWorkflowUnsupportedMode
	}
	return s.execute(context.Background(), id)
}

func (s *OpsDeploymentWorkflowService) Execute(ctx context.Context, id uint) (*ops.DeploymentWorkflowRun, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	return s.execute(ctx, id)
}

func (s *OpsDeploymentWorkflowService) execute(ctx context.Context, id uint) (*ops.DeploymentWorkflowRun, error) {
	if s == nil || s.workflowRepo == nil || s.preflightService == nil {
		return nil, errors.New("operations deployment workflow service is not configured")
	}
	run, err := s.workflowRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if run.Mode != ops.DeploymentWorkflowModeDryRun && run.Mode != ops.DeploymentWorkflowModeProduction {
		return nil, ErrOpsDeploymentWorkflowUnsupportedMode
	}
	if run.Status != ops.DeploymentWorkflowStatusValidated {
		return nil, workflowTransitionError(run.Status, "execute")
	}

	report, err := s.preflightService.EvaluateProject(run.ProjectID)
	if err != nil {
		return nil, err
	}
	snapshot, err := json.Marshal(report)
	if err != nil {
		return nil, fmt.Errorf("encode deployment preflight snapshot: %w", err)
	}
	if report.BlockingCount > 0 {
		if updateErr := s.workflowRepo.UpdateRun(id, map[string]interface{}{
			"status":             ops.DeploymentWorkflowStatusDraft,
			"preflight_status":   report.StatusLevel,
			"preflight_snapshot": string(snapshot),
			"last_error":         "Preflight changed and now contains blocking checks.",
		}); updateErr != nil {
			return nil, updateErr
		}
		return nil, ErrOpsDeploymentWorkflowPreflightBlocked
	}
	if run.Mode == ops.DeploymentWorkflowModeProduction {
		return s.executeProduction(ctx, run, report, snapshot)
	}

	startedAt := time.Now().UTC()
	if err := s.workflowRepo.UpdateRun(id, map[string]interface{}{
		"status":             ops.DeploymentWorkflowStatusRunning,
		"preflight_status":   report.StatusLevel,
		"preflight_snapshot": string(snapshot),
		"started_at":         startedAt,
		"completed_at":       nil,
		"last_error":         "",
	}); err != nil {
		return nil, err
	}

	for _, step := range run.Steps {
		if err := contextError(ctx); err != nil {
			return s.failWorkflow(id, step.ID, err)
		}
		stepStartedAt := time.Now().UTC()
		if err := s.workflowRepo.UpdateStep(step.ID, map[string]interface{}{
			"status":        ops.DeploymentWorkflowStepRunning,
			"started_at":    stepStartedAt,
			"error_message": "",
		}); err != nil {
			return s.failWorkflow(id, step.ID, err)
		}
		summary := dryRunStepSummary(step.Key, report)
		if err := s.workflowRepo.UpdateStep(step.ID, map[string]interface{}{
			"status":         ops.DeploymentWorkflowStepSucceeded,
			"output_summary": summary,
			"completed_at":   time.Now().UTC(),
		}); err != nil {
			return s.failWorkflow(id, step.ID, err)
		}
	}

	if err := s.workflowRepo.UpdateRun(id, map[string]interface{}{
		"status":       ops.DeploymentWorkflowStatusSucceeded,
		"completed_at": time.Now().UTC(),
		"last_error":   "",
	}); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindByID(id)
}

func (s *OpsDeploymentWorkflowService) executeProduction(
	ctx context.Context,
	run *ops.DeploymentWorkflowRun,
	report *ops.DeploymentPreflight,
	snapshot []byte,
) (*ops.DeploymentWorkflowRun, error) {
	if s.vpsRepo == nil || s.connectorService == nil || s.hostingerSync == nil || s.healthCheck == nil {
		return nil, errors.New("operations production deployment dependencies are not configured")
	}
	if strings.TrimSpace(run.RequestedRef) != "" && strings.TrimSpace(run.RequestedRef) != "master" {
		return nil, fmt.Errorf("%w: Hostinger project update currently follows the project IMAGE_TAG=master policy", ErrOpsDeploymentWorkflowUnsupportedRef)
	}
	project, err := s.projectRepo.FindByID(run.ProjectID)
	if err != nil {
		return nil, err
	}
	vps, err := s.vpsRepo.FindByID(project.VPSBindingID)
	if err != nil {
		return nil, err
	}
	if vps.ConnectorID == nil || *vps.ConnectorID == 0 {
		return nil, errors.New("production deployment VPS is not bound to a Hostinger connector")
	}
	if strings.TrimSpace(vps.ProviderResourceID) == "" || strings.TrimSpace(project.ComposeProjectName) == "" {
		return nil, errors.New("production deployment target is missing Hostinger resource or project name")
	}
	if err := s.workflowRepo.AcquireProjectLock(project.ID, run.ID, 30*time.Minute); err != nil {
		if errors.Is(err, repository.ErrOpsDeploymentWorkflowLockHeld) {
			return nil, fmt.Errorf("%w: project %s", repository.ErrOpsDeploymentWorkflowLockHeld, project.Name)
		}
		return nil, err
	}
	defer s.workflowRepo.ReleaseWorkflowLocks(run.ID)

	previousRef := firstNonEmptyWorkflowValue(project.CurrentImageTag, project.CurrentCommitSHA)
	idempotencyKey := fmt.Sprintf("ops-deploy-%d-%d", run.ID, project.ID)
	if err := s.workflowRepo.UpdateRun(run.ID, map[string]interface{}{
		"status":             ops.DeploymentWorkflowStatusRunning,
		"preflight_status":   report.StatusLevel,
		"preflight_snapshot": string(snapshot),
		"previous_ref":       previousRef,
		"rollback_ref":       previousRef,
		"idempotency_key":    idempotencyKey,
		"started_at":         time.Now().UTC(),
		"completed_at":       nil,
		"last_error":         "",
	}); err != nil {
		return nil, err
	}

	for _, step := range run.Steps {
		if err := contextError(ctx); err != nil {
			return s.failWorkflowWithStatus(run.ID, step.ID, ops.DeploymentWorkflowStatusPaused, err)
		}
		if err := s.workflowRepo.UpdateStep(step.ID, map[string]interface{}{
			"status":        ops.DeploymentWorkflowStepRunning,
			"started_at":    time.Now().UTC(),
			"error_message": "",
		}); err != nil {
			return s.failWorkflow(run.ID, step.ID, err)
		}

		summary, externalOperationID, stepErr := s.executeProductionStep(ctx, run, project, vps, report, step)
		if stepErr != nil {
			if step.Key == ops.DeploymentWorkflowStepPostHealthCheck {
				return s.failWorkflowWithStatus(run.ID, step.ID, ops.DeploymentWorkflowStatusRollbackRequired, stepErr)
			}
			return s.failWorkflow(run.ID, step.ID, stepErr)
		}
		if err := s.workflowRepo.UpdateStep(step.ID, map[string]interface{}{
			"status":                ops.DeploymentWorkflowStepSucceeded,
			"output_summary":        summary,
			"external_operation_id": externalOperationID,
			"completed_at":          time.Now().UTC(),
		}); err != nil {
			return s.failWorkflow(run.ID, step.ID, err)
		}
		if externalOperationID != "" && step.Key == ops.DeploymentWorkflowStepUpdateProject {
			if err := s.workflowRepo.UpdateRun(run.ID, map[string]interface{}{
				"remote_operation_id": externalOperationID,
			}); err != nil {
				return s.failWorkflow(run.ID, step.ID, err)
			}
		}
	}
	if err := s.workflowRepo.UpdateRun(run.ID, map[string]interface{}{
		"status":        ops.DeploymentWorkflowStatusSucceeded,
		"health_status": ops.DeploymentHealthHealthy,
		"completed_at":  time.Now().UTC(),
		"last_error":    "",
	}); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindByID(run.ID)
}

func (s *OpsDeploymentWorkflowService) Cancel(id uint) (*ops.DeploymentWorkflowRun, error) {
	if s == nil || s.workflowRepo == nil {
		return nil, errors.New("operations deployment workflow service is not configured")
	}
	run, err := s.workflowRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if !workflowStatusAllowed(run.Status,
		ops.DeploymentWorkflowStatusDraft,
		ops.DeploymentWorkflowStatusAwaitingApproval,
		ops.DeploymentWorkflowStatusValidated,
		ops.DeploymentWorkflowStatusPaused,
	) {
		return nil, workflowTransitionError(run.Status, "cancel")
	}
	if err := s.workflowRepo.UpdateRun(id, map[string]interface{}{
		"status":       ops.DeploymentWorkflowStatusCancelled,
		"completed_at": time.Now().UTC(),
		"last_error":   "",
	}); err != nil {
		return nil, err
	}
	return s.workflowRepo.FindByID(id)
}

func (s *OpsDeploymentWorkflowService) failWorkflow(runID, stepID uint, cause error) (*ops.DeploymentWorkflowRun, error) {
	return s.failWorkflowWithStatus(runID, stepID, ops.DeploymentWorkflowStatusFailed, cause)
}

func (s *OpsDeploymentWorkflowService) failWorkflowWithStatus(
	runID,
	stepID uint,
	status string,
	cause error,
) (*ops.DeploymentWorkflowRun, error) {
	message := cause.Error()
	_ = s.workflowRepo.UpdateStep(stepID, map[string]interface{}{
		"status":        ops.DeploymentWorkflowStepFailed,
		"error_message": message,
		"completed_at":  time.Now().UTC(),
	})
	_ = s.workflowRepo.UpdateRun(runID, map[string]interface{}{
		"status":       status,
		"last_error":   message,
		"completed_at": time.Now().UTC(),
	})
	return nil, cause
}

func (s *OpsDeploymentWorkflowService) executeProductionStep(
	ctx context.Context,
	run *ops.DeploymentWorkflowRun,
	project *ops.ProjectBindingView,
	vps *ops.VPSBinding,
	report *ops.DeploymentPreflight,
	step ops.DeploymentWorkflowStep,
) (string, string, error) {
	switch step.Key {
	case ops.DeploymentWorkflowStepCheckConnector:
		connectorID := vps.ConnectorID
		if project.ConnectorID != nil && *project.ConnectorID != 0 {
			connectorID = project.ConnectorID
		}
		if connectorID == nil || *connectorID == 0 {
			return "", "", errors.New("production deployment has no Hostinger connector")
		}
		connector, err := s.connectorService.Get(*connectorID)
		if err != nil {
			return "", "", err
		}
		if !connector.Enabled || connector.Status != ops.ConnectorStatusActive {
			return "", "", fmt.Errorf("Hostinger connector is not ready: %s", connector.Name)
		}
		return fmt.Sprintf("Hostinger connector ready: %s.", connector.Name), "", nil
	case ops.DeploymentWorkflowStepDiscoverVPS:
		return fmt.Sprintf("VPS target confirmed: %s / %s.", vps.Name, vps.ProviderResourceID), "", nil
	case ops.DeploymentWorkflowStepDiscoverProject:
		result, err := s.hostingerSync.SyncProject(ctx, project.ID)
		if err != nil {
			return "", "", err
		}
		return fmt.Sprintf(
			"Remote project observed: %s, containers %d, running %d, healthy %d.",
			result.RemoteState,
			result.ContainerCount,
			result.RunningContainerCount,
			result.HealthyContainerCount,
		), "", nil
	case ops.DeploymentWorkflowStepCheckImage:
		if strings.TrimSpace(run.RequestedRef) != "" && strings.TrimSpace(run.RequestedRef) != "master" {
			return "", "", fmt.Errorf("%w: %s", ErrOpsDeploymentWorkflowUnsupportedRef, run.RequestedRef)
		}
		return "Hostinger project update will pull the current validated master image set.", "", nil
	case ops.DeploymentWorkflowStepHealthCheck:
		if report == nil {
			return "", "", errors.New("deployment preflight report is missing")
		}
		return fmt.Sprintf("Preflight remains clear: blocking %d, warnings %d.", report.BlockingCount, report.WarningCount), "", nil
	case ops.DeploymentWorkflowStepRecordRollback:
		return fmt.Sprintf("Rollback point recorded: %s.", firstNonEmptyWorkflowValue(run.PreviousRef, project.CurrentImageTag, "unknown")), "", nil
	case ops.DeploymentWorkflowStepUpdateProject:
		connectorID := vps.ConnectorID
		if project.ConnectorID != nil && *project.ConnectorID != 0 {
			connectorID = project.ConnectorID
		}
		result, err := s.connectorService.HostingerUpdateProject(
			ctx,
			*connectorID,
			vps.ProviderResourceID,
			project.ComposeProjectName,
			run.IdempotencyKey,
		)
		if err != nil {
			return "", "", err
		}
		return result.Message, result.OperationID, nil
	case ops.DeploymentWorkflowStepPostHealthCheck:
		if _, err := s.hostingerSync.SyncProject(ctx, project.ID); err != nil {
			return "", "", err
		}
		latestProject, err := s.projectRepo.FindByID(project.ID)
		if err != nil {
			return "", "", err
		}
		health, err := s.healthCheck.CheckProject(ctx, latestProject)
		if err != nil {
			return "", "", err
		}
		encoded, encodeErr := json.Marshal(health)
		if encodeErr == nil {
			_ = s.workflowRepo.UpdateRun(run.ID, map[string]interface{}{
				"health_status":   health.Status,
				"health_snapshot": string(encoded),
			})
		}
		if health.Status != ops.DeploymentHealthHealthy {
			return health.Summary, "", errors.New(health.Summary)
		}
		return health.Summary, "", nil
	case ops.DeploymentWorkflowStepPurgeCache:
		if s.cachePurge == nil {
			return "", "", errors.New("Cloudflare cache purge service is not configured")
		}
		result, err := s.cachePurge.PurgeProject(ctx, project.ID)
		if err != nil {
			return "", "", err
		}
		return result.Summary, cachePurgeOperationID(result), nil
	case ops.DeploymentWorkflowStepRecordRelease:
		now := time.Now().UTC()
		if err := s.projectRepo.RecordDeployment(project.ID, now); err != nil {
			return "", "", err
		}
		return fmt.Sprintf("Production deployment evidence recorded at %s.", now.Format(time.RFC3339)), "", nil
	default:
		return "Production deployment step completed.", "", nil
	}
}

func buildWorkflowSteps(project *ops.ProjectBindingView, requestedRef string, report *ops.DeploymentPreflight, mode string) []ops.DeploymentWorkflowStep {
	input := map[string]interface{}{
		"project_id":    project.ID,
		"project":       project.Name,
		"requested_ref": requestedRef,
		"mode":          mode,
	}
	encodedInput, _ := json.Marshal(input)
	specs := []struct {
		key            string
		label          string
		externalEffect bool
		retryable      bool
	}{
		{ops.DeploymentWorkflowStepCheckConnector, "检查连接器", false, true},
		{ops.DeploymentWorkflowStepDiscoverVPS, "读取 VPS 观察证据", false, true},
		{ops.DeploymentWorkflowStepDiscoverProject, "读取项目观察证据", false, true},
		{ops.DeploymentWorkflowStepCheckImage, "检查镜像和版本证据", false, true},
		{ops.DeploymentWorkflowStepRenderConfig, "生成配置预览", false, false},
		{ops.DeploymentWorkflowStepDiffConfig, "计算配置差异", false, false},
		{ops.DeploymentWorkflowStepHealthCheck, "检查发布前健康证据", false, true},
		{ops.DeploymentWorkflowStepRecordRelease, "固化 dry-run 证据", false, false},
	}
	if mode == ops.DeploymentWorkflowModeProduction {
		specs = []struct {
			key            string
			label          string
			externalEffect bool
			retryable      bool
		}{
			{ops.DeploymentWorkflowStepCheckConnector, "检查 Hostinger 执行连接器", false, true},
			{ops.DeploymentWorkflowStepDiscoverVPS, "确认 VPS 发布目标", false, true},
			{ops.DeploymentWorkflowStepDiscoverProject, "读取发布前项目状态", false, true},
			{ops.DeploymentWorkflowStepCheckImage, "确认 master 镜像发布策略", false, true},
			{ops.DeploymentWorkflowStepRecordRollback, "记录发布前回滚点", false, false},
			{ops.DeploymentWorkflowStepHealthCheck, "确认发布前 Preflight", false, true},
			{ops.DeploymentWorkflowStepUpdateProject, "更新 Hostinger Docker 项目", true, false},
			{ops.DeploymentWorkflowStepPostHealthCheck, "执行发布后健康检查", false, true},
			{ops.DeploymentWorkflowStepPurgeCache, "按域名清理 Cloudflare 缓存", true, true},
			{ops.DeploymentWorkflowStepRecordRelease, "固化生产发布证据", false, false},
		}
	}
	steps := make([]ops.DeploymentWorkflowStep, 0, len(specs))
	for index, spec := range specs {
		steps = append(steps, ops.DeploymentWorkflowStep{
			Sequence:       index + 1,
			Key:            spec.key,
			Label:          spec.label,
			Status:         ops.DeploymentWorkflowStepPending,
			Retryable:      spec.retryable,
			ExternalEffect: spec.externalEffect,
			InputSnapshot:  string(encodedInput),
			OutputSummary:  "",
		})
	}
	return steps
}

func cachePurgeOperationID(result *OpsCloudflareCachePurgeResult) string {
	if result == nil {
		return ""
	}
	for _, group := range result.Groups {
		for _, operationID := range group.OperationIDs {
			if value := strings.TrimSpace(operationID); value != "" {
				return value
			}
		}
	}
	return ""
}

func contextError(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func dryRunStepSummary(key string, report *ops.DeploymentPreflight) string {
	if report == nil {
		return "未生成 Preflight 证据。"
	}
	switch key {
	case ops.DeploymentWorkflowStepCheckConnector:
		return fmt.Sprintf("只读连接器检查完成：阻断 %d，警告 %d。", report.BlockingCount, report.WarningCount)
	case ops.DeploymentWorkflowStepDiscoverVPS, ops.DeploymentWorkflowStepDiscoverProject:
		return "读取台账和已持久化 Observed State；未写入外部平台。"
	case ops.DeploymentWorkflowStepCheckImage:
		return "检查镜像标签、Commit SHA 和版本一致性；未拉取或更新镜像。"
	case ops.DeploymentWorkflowStepRenderConfig:
		return "生成配置计划边界；未写入 Compose、Caddy、Nginx 或 DNS。"
	case ops.DeploymentWorkflowStepDiffConfig:
		return "记录 Desired/Observed 差异；未应用任何变更。"
	case ops.DeploymentWorkflowStepHealthCheck:
		return "使用当前只读健康证据；未执行发布后的外部健康探测。"
	case ops.DeploymentWorkflowStepRecordRelease:
		return "dry-run 证据已固化；该步骤不代表生产部署已执行。"
	default:
		return "只读 dry-run 步骤完成。"
	}
}

func workflowStatusAllowed(status string, allowed ...string) bool {
	for _, candidate := range allowed {
		if status == candidate {
			return true
		}
	}
	return false
}

func workflowTransitionError(status, action string) error {
	return fmt.Errorf("%w: cannot %s from status %s", ErrOpsDeploymentWorkflowInvalidTransition, action, status)
}

func normalizeWorkflowActor(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "system"
	}
	return value
}

func firstNonEmptyWorkflowValue(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
