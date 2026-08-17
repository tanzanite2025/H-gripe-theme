package sitequality

import "time"

const (
	SiteQualityTargetSourceRouteCatalog = "route_catalog"
	SiteQualityTargetSourceOperator     = "operator"

	SiteQualityTargetTierCritical   = "critical"
	SiteQualityTargetTierStandard   = "standard"
	SiteQualityTargetTierBackground = "background"

	SiteQualityJobKindScheduled = "scheduled"
	SiteQualityJobKindManual    = "manual"
	SiteQualityJobKindRecheck   = "recheck"

	SiteQualityJobStatusQueued     = "queued"
	SiteQualityJobStatusProcessing = "processing"
	SiteQualityJobStatusSucceeded  = "succeeded"
	SiteQualityJobStatusFailed     = "failed"
	SiteQualityJobStatusDeadLetter = "dead_letter"

	SiteQualityEvaluationStatusCompleted           = "completed"
	SiteQualityEvaluationStatusInsufficientSamples = "insufficient_samples"
	SiteQualityEvaluationStatusFailed              = "failed"
)

// SiteQualityTarget is the Site Quality-owned identity of a storefront
// page. URL Management can supply candidates, but it never owns quality state.
type SiteQualityTarget struct {
	ID                      uint       `gorm:"primaryKey" json:"id"`
	RouteEntryID            *uint      `gorm:"uniqueIndex;index" json:"route_entry_id,omitempty"`
	CanonicalURL            string     `gorm:"type:text;not null;uniqueIndex" json:"canonical_url"`
	Locale                  string     `gorm:"size:20;not null;default:'';index" json:"locale"`
	Source                  string     `gorm:"size:32;not null;default:'operator';index" json:"source"`
	SourceType              string     `gorm:"size:32;not null;default:'';index" json:"source_type"`
	Title                   string     `gorm:"type:text;not null;default:''" json:"title"`
	SamplingTier            string     `gorm:"size:16;not null;default:'standard';index" json:"sampling_tier"`
	SamplingIntervalSeconds int        `gorm:"not null;default:604800" json:"sampling_interval_seconds"`
	Enabled                 bool       `gorm:"not null;default:true;index" json:"enabled"`
	LedgerSynced            bool       `gorm:"not null;default:false;index" json:"ledger_synced"`
	LedgerSyncMarker        string     `gorm:"size:128;not null;default:'';index" json:"ledger_sync_marker,omitempty"`
	LedgerSyncedAt          *time.Time `gorm:"index" json:"ledger_synced_at,omitempty"`
	DisabledAt              *time.Time `gorm:"index" json:"disabled_at,omitempty"`
	DisableReason           string     `gorm:"type:text;not null;default:''" json:"disable_reason,omitempty"`
	NextScheduledAt         *time.Time `gorm:"index" json:"next_scheduled_at,omitempty"`
	LastScheduledAt         *time.Time `gorm:"index" json:"last_scheduled_at,omitempty"`
	LastCompletedAt         *time.Time `gorm:"index" json:"last_completed_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

func (SiteQualityTarget) TableName() string {
	return "site_quality_targets"
}

// SiteQualityJob is a leaseable provider-work request. It is the source
// of execution truth across replicas; a scheduler only claims due jobs.
type SiteQualityJob struct {
	ID                    uint       `gorm:"primaryKey" json:"id"`
	TargetID              uint       `gorm:"not null;index" json:"target_id"`
	FindingID             *uint      `gorm:"index" json:"finding_id,omitempty"`
	Strategy              string     `gorm:"size:16;not null;index" json:"strategy"`
	Kind                  string     `gorm:"size:16;not null;index" json:"kind"`
	Status                string     `gorm:"size:16;not null;index" json:"status"`
	IdempotencyKey        string     `gorm:"size:255;not null;uniqueIndex" json:"idempotency_key"`
	SampleCount           int        `gorm:"not null;default:3" json:"sample_count"`
	RequiredConfirmations int        `gorm:"not null;default:2" json:"required_confirmations"`
	Attempts              int        `gorm:"not null;default:0" json:"attempts"`
	MaxAttempts           int        `gorm:"not null;default:4" json:"max_attempts"`
	AvailableAt           time.Time  `gorm:"not null;index" json:"available_at"`
	LockedAt              *time.Time `gorm:"index" json:"locked_at,omitempty"`
	LockedBy              string     `gorm:"size:128;not null;default:'';index" json:"locked_by"`
	LeaseGeneration       int64      `gorm:"not null;default:0;index" json:"lease_generation"`
	LeaseExpiresAt        *time.Time `gorm:"index" json:"lease_expires_at,omitempty"`
	HeartbeatAt           *time.Time `gorm:"index" json:"heartbeat_at,omitempty"`
	StartedAt             *time.Time `gorm:"index" json:"started_at,omitempty"`
	FinishedAt            *time.Time `gorm:"index" json:"finished_at,omitempty"`
	InitiatedByUserID     uint       `gorm:"not null;default:0;index" json:"initiated_by_user_id"`
	ReleaseID             string     `gorm:"size:128;not null;default:''" json:"release_id"`
	LastError             string     `gorm:"type:text;not null;default:''" json:"last_error,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

func (SiteQualityJob) TableName() string {
	return "site_quality_jobs"
}

// SiteQualityEvaluation is the durable statistical decision for one
// completed job. Individual SiteQualityRun records remain immutable samples.
type SiteQualityEvaluation struct {
	ID                    uint      `gorm:"primaryKey" json:"id"`
	JobID                 uint      `gorm:"not null;uniqueIndex" json:"job_id"`
	TargetID              uint      `gorm:"not null;index" json:"target_id"`
	Strategy              string    `gorm:"size:16;not null;index" json:"strategy"`
	Status                string    `gorm:"size:32;not null;index" json:"status"`
	SampleCount           int       `gorm:"not null" json:"sample_count"`
	SuccessfulSamples     int       `gorm:"not null" json:"successful_samples"`
	RequiredConfirmations int       `gorm:"not null" json:"required_confirmations"`
	ConfirmedAuditIDs     string    `gorm:"type:jsonb;not null;default:'[]'" json:"confirmed_audit_ids"`
	CleanAuditIDs         string    `gorm:"type:jsonb;not null;default:'[]'" json:"clean_audit_ids"`
	DecisionJSON          string    `gorm:"type:jsonb;not null;default:'{}'" json:"decision"`
	DecidedAt             time.Time `gorm:"not null;index" json:"decided_at"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (SiteQualityEvaluation) TableName() string {
	return "site_quality_evaluations"
}

// SiteQualityProviderSlot is a globally leased provider concurrency slot. It
// protects shared runner capacity across multiple API replicas.
type SiteQualityProviderSlot struct {
	Provider    string     `gorm:"size:64;primaryKey" json:"provider"`
	Slot        int        `gorm:"primaryKey" json:"slot"`
	AvailableAt time.Time  `gorm:"not null;index" json:"available_at"`
	LockedAt    *time.Time `gorm:"index" json:"locked_at,omitempty"`
	LockedBy    string     `gorm:"size:128;not null;default:'';index" json:"locked_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (SiteQualityProviderSlot) TableName() string {
	return "site_quality_provider_slots"
}

type SiteQualityTargetInput struct {
	RouteEntryID            *uint
	CanonicalURL            string
	Locale                  string
	Source                  string
	SourceType              string
	Title                   string
	SamplingTier            string
	SamplingIntervalSeconds int
	Enabled                 bool
	LedgerSynced            bool
	LedgerSyncMarker        string
	LedgerSyncedAt          *time.Time
	DisableReason           string
}

type SiteQualityJobLease struct {
	WorkerID       string
	Generation     int64
	LeaseExpiresAt time.Time
}
