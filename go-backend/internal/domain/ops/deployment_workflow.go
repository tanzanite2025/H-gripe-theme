package ops

import "time"

const (
	DeploymentWorkflowKindDeployment = "deployment"
	DeploymentWorkflowModeDryRun     = "dry_run"
	DeploymentWorkflowModeProduction = "production"

	DeploymentWorkflowStatusDraft            = "draft"
	DeploymentWorkflowStatusAwaitingApproval = "awaiting_approval"
	DeploymentWorkflowStatusValidated        = "validated"
	DeploymentWorkflowStatusRunning          = "running"
	DeploymentWorkflowStatusSucceeded        = "succeeded"
	DeploymentWorkflowStatusFailed           = "failed"
	DeploymentWorkflowStatusCancelled        = "cancelled"
	DeploymentWorkflowStatusPaused           = "paused"
	DeploymentWorkflowStatusRollbackRequired = "rollback_required"
	DeploymentWorkflowStatusRolledBack       = "rolled_back"

	DeploymentWorkflowStepPending   = "pending"
	DeploymentWorkflowStepRunning   = "running"
	DeploymentWorkflowStepSucceeded = "succeeded"
	DeploymentWorkflowStepFailed    = "failed"
	DeploymentWorkflowStepSkipped   = "skipped"

	DeploymentWorkflowStepCheckConnector  = "check_connector"
	DeploymentWorkflowStepDiscoverVPS     = "discover_vps"
	DeploymentWorkflowStepDiscoverProject = "discover_project"
	DeploymentWorkflowStepCheckImage      = "check_image"
	DeploymentWorkflowStepRenderConfig    = "render_config"
	DeploymentWorkflowStepDiffConfig      = "diff_config"
	DeploymentWorkflowStepRecordRollback  = "record_rollback_point"
	DeploymentWorkflowStepUpdateProject   = "update_project"
	DeploymentWorkflowStepHealthCheck     = "health_check"
	DeploymentWorkflowStepPostHealthCheck = "post_deploy_health_check"
	DeploymentWorkflowStepPurgeCache      = "purge_cache"
	DeploymentWorkflowStepRecordRelease   = "record_release"
)

type DeploymentWorkflowRun struct {
	ID                uint                     `gorm:"primarykey" json:"id"`
	Kind              string                   `gorm:"size:32;not null;default:'deployment';index" json:"kind"`
	Mode              string                   `gorm:"size:32;not null;default:'dry_run';index" json:"mode"`
	ProjectID         uint                     `gorm:"not null;index" json:"project_id"`
	ProjectName       string                   `gorm:"size:120;not null;default:''" json:"project"`
	Environment       string                   `gorm:"size:32;not null;default:'';index" json:"environment"`
	RequestedRef      string                   `gorm:"size:160;not null;default:''" json:"requested_ref"`
	Status            string                   `gorm:"size:32;not null;default:'draft';index" json:"status"`
	PreflightStatus   string                   `gorm:"size:32;not null;default:'';index" json:"preflight_status"`
	PreflightSnapshot string                   `gorm:"type:text;not null;default:''" json:"-"`
	CreatedByID       uint                     `gorm:"not null;default:0" json:"created_by_id"`
	CreatedBy         string                   `gorm:"size:160;not null;default:''" json:"created_by"`
	ApprovedByID      *uint                    `json:"approved_by_id,omitempty"`
	ApprovedBy        string                   `gorm:"size:160;not null;default:''" json:"approved_by,omitempty"`
	ApprovedAt        *time.Time               `json:"approved_at,omitempty"`
	StartedAt         *time.Time               `json:"started_at,omitempty"`
	CompletedAt       *time.Time               `json:"completed_at,omitempty"`
	PreviousRef       string                   `gorm:"size:160;not null;default:''" json:"previous_ref,omitempty"`
	RollbackRef       string                   `gorm:"size:160;not null;default:''" json:"rollback_ref,omitempty"`
	RemoteOperationID string                   `gorm:"size:160;not null;default:''" json:"remote_operation_id,omitempty"`
	IdempotencyKey    string                   `gorm:"size:160;not null;default:''" json:"idempotency_key,omitempty"`
	HealthStatus      string                   `gorm:"size:32;not null;default:''" json:"health_status,omitempty"`
	HealthSnapshot    string                   `gorm:"type:text;not null;default:''" json:"-"`
	LastError         string                   `gorm:"type:text;not null;default:''" json:"last_error"`
	Preflight         *DeploymentPreflight     `gorm:"-" json:"preflight,omitempty"`
	Steps             []DeploymentWorkflowStep `gorm:"-" json:"steps"`
	CreatedAt         time.Time                `json:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at"`
}

func (DeploymentWorkflowRun) TableName() string {
	return "ops_deployment_workflow_runs"
}

type DeploymentWorkflowStep struct {
	ID                  uint       `gorm:"primarykey" json:"id"`
	WorkflowRunID       uint       `gorm:"not null;index" json:"workflow_run_id"`
	Sequence            int        `gorm:"not null" json:"sequence"`
	Key                 string     `gorm:"size:64;not null" json:"key"`
	Label               string     `gorm:"size:120;not null;default:''" json:"label"`
	Status              string     `gorm:"size:32;not null;default:'pending';index" json:"status"`
	Retryable           bool       `gorm:"not null;default:false" json:"retryable"`
	ExternalEffect      bool       `gorm:"not null;default:false" json:"external_effect"`
	InputSnapshot       string     `gorm:"type:text;not null;default:''" json:"input_snapshot,omitempty"`
	OutputSummary       string     `gorm:"type:text;not null;default:''" json:"output_summary,omitempty"`
	ErrorMessage        string     `gorm:"type:text;not null;default:''" json:"error_message,omitempty"`
	ExternalOperationID string     `gorm:"size:160;not null;default:''" json:"external_operation_id,omitempty"`
	StartedAt           *time.Time `json:"started_at,omitempty"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (DeploymentWorkflowStep) TableName() string {
	return "ops_deployment_workflow_steps"
}

type DeploymentWorkflowLock struct {
	ResourceKey   string    `gorm:"primaryKey;size:160" json:"resource_key"`
	WorkflowRunID uint      `gorm:"not null;index" json:"workflow_run_id"`
	AcquiredAt    time.Time `json:"acquired_at"`
	ExpiresAt     time.Time `gorm:"index" json:"expires_at"`
}

func (DeploymentWorkflowLock) TableName() string {
	return "ops_deployment_workflow_locks"
}
