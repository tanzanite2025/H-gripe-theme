package sitequality

import "time"

const (
	SiteQualityFindingStateOpen         = "open"
	SiteQualityFindingStateAcknowledged = "acknowledged"
	SiteQualityFindingStateResolved     = "resolved"
	SiteQualityFindingStateVerified     = "verified"

	SiteQualityFindingEventDetected           = "detected"
	SiteQualityFindingEventReopened           = "reopened"
	SiteQualityFindingEventAcknowledged       = "acknowledged"
	SiteQualityFindingEventResolutionRecorded = "resolution_recorded"
	SiteQualityFindingEventVerificationPassed = "verification_passed"
)

// SiteQualityFinding is the durable work-item projection of an actionable
// Lighthouse audit. SiteQualityRun remains the immutable record of every run.
type SiteQualityFinding struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	TargetID           *uint      `gorm:"uniqueIndex:uq_site_quality_findings_target_strategy_audit;index" json:"target_id,omitempty"`
	TargetURL          string     `gorm:"type:text;not null;index" json:"target_url"`
	Strategy           string     `gorm:"size:16;not null;uniqueIndex:uq_site_quality_findings_target_strategy_audit;index" json:"strategy"`
	AuditID            string     `gorm:"size:128;not null;uniqueIndex:uq_site_quality_findings_target_strategy_audit;index" json:"audit_id"`
	RuleID             string     `gorm:"size:128;not null;default:'';index" json:"rule_id"`
	ProviderAuditID    string     `gorm:"size:128;not null;default:''" json:"provider_audit_id,omitempty"`
	FindingKind        string     `gorm:"size:32;not null;default:'opportunity';index" json:"finding_kind"`
	RuleVersion        string     `gorm:"size:64;not null;default:''" json:"rule_version"`
	Confidence         float64    `gorm:"not null;default:0" json:"confidence"`
	SampleCount        int        `gorm:"not null;default:1" json:"sample_count"`
	Confirmations      int        `gorm:"not null;default:1" json:"confirmations"`
	ConsecutiveClean   int        `gorm:"not null;default:0" json:"consecutive_clean"`
	Title              string     `gorm:"type:text;not null;default:''" json:"title"`
	Description        string     `gorm:"type:text;not null;default:''" json:"description"`
	Severity           string     `gorm:"size:16;not null;index" json:"severity"`
	State              string     `gorm:"size:32;not null;default:'open';index" json:"state"`
	FirstDetectedAt    time.Time  `gorm:"not null;index" json:"first_detected_at"`
	LastDetectedAt     time.Time  `gorm:"not null;index" json:"last_detected_at"`
	LatestRunID        uint       `gorm:"not null;index" json:"latest_run_id"`
	LatestSavingsMS    *float64   `json:"latest_savings_ms,omitempty"`
	LatestSavingsBytes *int64     `json:"latest_savings_bytes,omitempty"`
	ResourceCount      int        `gorm:"not null;default:0" json:"resource_count"`
	LatestEvidence     string     `gorm:"type:jsonb;not null;default:'{}'" json:"latest_evidence"`
	ResolutionNote     string     `gorm:"type:text;not null;default:''" json:"resolution_note"`
	ResolvedAt         *time.Time `gorm:"index" json:"resolved_at,omitempty"`
	VerifiedAt         *time.Time `gorm:"index" json:"verified_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (SiteQualityFinding) TableName() string {
	return "site_quality_findings"
}

// SiteQualityFindingEvent records a lifecycle decision. The run itself stores
// the complete provider evidence, so routine observation updates do not need
// duplicate event rows.
type SiteQualityFindingEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	FindingID   uint      `gorm:"not null;index" json:"finding_id"`
	RunID       *uint     `gorm:"index" json:"run_id,omitempty"`
	EventType   string    `gorm:"size:64;not null;index" json:"event_type"`
	ActorUserID uint      `gorm:"not null;default:0;index" json:"actor_user_id"`
	Note        string    `gorm:"type:text;not null;default:''" json:"note"`
	Metadata    string    `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt   time.Time `gorm:"not null;index" json:"created_at"`
}

func (SiteQualityFindingEvent) TableName() string {
	return "site_quality_finding_events"
}

// SiteQualityFindingEvidence is the normalized provider evidence exposed with
// a finding. It deliberately contains only details needed to diagnose the
// work item; the full SiteQuality response stays on the run.
type SiteQualityFindingEvidence struct {
	AuditID         string                              `json:"audit_id"`
	RuleID          string                              `json:"rule_id"`
	ProviderAuditID string                              `json:"provider_audit_id,omitempty"`
	Title           string                              `json:"title"`
	Description     string                              `json:"description,omitempty"`
	Score           *float64                            `json:"score,omitempty"`
	DisplayValue    string                              `json:"display_value,omitempty"`
	NumericValue    *float64                            `json:"numeric_value,omitempty"`
	SavingsMS       *float64                            `json:"savings_ms,omitempty"`
	SavingsBytes    *int64                              `json:"savings_bytes,omitempty"`
	Resources       []SiteQualityFindingResource        `json:"resources,omitempty"`
	Links           []SiteQualityLinkEvidence           `json:"links,omitempty"`
	Headings        []SiteQualityHeadingEvidence        `json:"headings,omitempty"`
	StructuredData  []SiteQualityStructuredDataEvidence `json:"structured_data,omitempty"`
}

type SiteQualityFindingResource struct {
	URL        string   `json:"url"`
	TotalBytes *int64   `json:"total_bytes,omitempty"`
	WastedMS   *float64 `json:"wasted_ms,omitempty"`
}

// SiteQualityLinkEvidence mirrors Lighthouse's link-text table details.
// Lighthouse intentionally exposes only the destination and rendered text.
type SiteQualityLinkEvidence struct {
	Href     string `json:"href"`
	Text     string `json:"text"`
	TextLang string `json:"text_lang,omitempty"`
}

// SiteQualityHeadingEvidence identifies a heading node reported by the
// accessibility audit so an operator can locate the semantic violation.
type SiteQualityHeadingEvidence struct {
	Level       int    `json:"level,omitempty"`
	Text        string `json:"text,omitempty"`
	Snippet     string `json:"snippet,omitempty"`
	Selector    string `json:"selector,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

// SiteQualityStructuredDataEvidence identifies structured data entities or
// script nodes that explain why a schema finding was raised.
type SiteQualityStructuredDataEvidence struct {
	Format      string `json:"format,omitempty"`
	Type        string `json:"type,omitempty"`
	ID          string `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	URL         string `json:"url,omitempty"`
	Selector    string `json:"selector,omitempty"`
	Snippet     string `json:"snippet,omitempty"`
	Property    string `json:"property,omitempty"`
	Explanation string `json:"explanation,omitempty"`
}

type SiteQualityFindingDetection struct {
	TargetID           *uint
	TargetURL          string
	Strategy           string
	AuditID            string
	RuleID             string
	ProviderAuditID    string
	FindingKind        string
	RuleVersion        string
	Confidence         float64
	SampleCount        int
	Confirmations      int
	Title              string
	Description        string
	Severity           string
	LatestRunID        uint
	LatestSavingsMS    *float64
	LatestSavingsBytes *int64
	ResourceCount      int
	LatestEvidence     string
	DetectedAt         time.Time
}

// SiteQualityFindingEvaluationInput is the confirmed result of a multi-sample
// job. The repository applies it in the same transaction as the job outcome.
type SiteQualityFindingEvaluationInput struct {
	TargetID                 uint
	FindingID                *uint
	TargetURL                string
	Strategy                 string
	Detections               []SiteQualityFindingDetection
	CleanAuditIDs            []string
	ObservedAuditIDs         []string
	RequiredCleanEvaluations int
	LatestRunID              uint
	DecidedAt                time.Time
}

type SiteQualityFindingActionInput struct {
	Note string `json:"note"`
}

type SiteQualityFindingResolutionInput struct {
	ResolutionNote string `json:"resolution_note"`
}

type SiteQualityFindingStats struct {
	Active       int64 `json:"active"`
	Open         int64 `json:"open"`
	Acknowledged int64 `json:"acknowledged"`
	Resolved     int64 `json:"resolved"`
	Verified     int64 `json:"verified"`
	Critical     int64 `json:"critical"`
	High         int64 `json:"high"`
}
