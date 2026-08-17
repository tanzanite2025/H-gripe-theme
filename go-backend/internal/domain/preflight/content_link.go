package preflight

import "time"

const (
	ContentLinkRunStatusSuccess = "success"
	ContentLinkRunStatusFailed  = "failed"

	ContentLinkIssueStateOpen     = "open"
	ContentLinkIssueStateResolved = "resolved"
	ContentLinkIssueStateVerified = "verified"
	ContentLinkIssueStateIgnored  = "ignored"

	ContentLinkFixStatusNotFixable = "not_fixable"
	ContentLinkFixStatusPending    = "pending"
	ContentLinkFixStatusApplied    = "applied"
	ContentLinkFixStatusFailed     = "failed"

	ContentLinkEventDetected           = "detected"
	ContentLinkEventReopened           = "reopened"
	ContentLinkEventResolutionRecorded = "resolution_recorded"
	ContentLinkEventFixApplied         = "fix_applied"
	ContentLinkEventFixFailed          = "fix_failed"
	ContentLinkEventVerificationPassed = "verification_passed"
	ContentLinkEventManualRecheck      = "manual_recheck"
)

type ContentLinkRun struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TargetURL    string    `gorm:"type:text;not null;index" json:"target_url"`
	RouteEntryID *uint     `gorm:"index" json:"route_entry_id,omitempty"`
	Status       string    `gorm:"size:24;not null;index" json:"status"`
	CheckedAt    time.Time `gorm:"not null;index" json:"checked_at"`
	IssueCount   int       `gorm:"not null;default:0" json:"issue_count"`
	FixableCount int       `gorm:"not null;default:0" json:"fixable_count"`
	ErrorMessage string    `gorm:"type:text;not null;default:''" json:"error_message"`
	CreatedAt    time.Time `json:"created_at"`
}

func (ContentLinkRun) TableName() string {
	return "preflight_content_link_runs"
}

type ContentLinkIssue struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	RouteEntryID    *uint      `gorm:"index" json:"route_entry_id,omitempty"`
	RunID           uint       `gorm:"not null;index" json:"run_id"`
	TargetURL       string     `gorm:"type:text;not null;index" json:"target_url"`
	FinalURL        string     `gorm:"type:text;not null;default:''" json:"final_url"`
	LinkURL         string     `gorm:"type:text;not null;default:''" json:"link_url"`
	LinkText        string     `gorm:"type:text;not null;default:''" json:"link_text"`
	Selector        string     `gorm:"type:text;not null;default:''" json:"selector"`
	Snippet         string     `gorm:"type:text;not null;default:''" json:"snippet"`
	SourceType      string     `gorm:"size:48;not null;default:'';index" json:"source_type"`
	SourceID        *uint      `gorm:"index" json:"source_id,omitempty"`
	SourceKey       string     `gorm:"type:text;not null;default:''" json:"source_key"`
	SourceField     string     `gorm:"size:80;not null;default:''" json:"source_field"`
	IssueKey        string     `gorm:"size:128;not null;uniqueIndex" json:"issue_key"`
	Severity        string     `gorm:"size:16;not null;default:'medium';index" json:"severity"`
	State           string     `gorm:"size:24;not null;default:'open';index" json:"state"`
	SuggestedText   string     `gorm:"type:text;not null;default:''" json:"suggested_text"`
	FixStatus       string     `gorm:"size:24;not null;default:'not_fixable';index" json:"fix_status"`
	FixError        string     `gorm:"type:text;not null;default:''" json:"fix_error"`
	LatestEvidence  string     `gorm:"type:jsonb;not null;default:'{}'" json:"latest_evidence"`
	FirstDetectedAt time.Time  `gorm:"not null;index" json:"first_detected_at"`
	LastDetectedAt  time.Time  `gorm:"not null;index" json:"last_detected_at"`
	ResolvedAt      *time.Time `gorm:"index" json:"resolved_at,omitempty"`
	VerifiedAt      *time.Time `gorm:"index" json:"verified_at,omitempty"`
	FixedAt         *time.Time `gorm:"index" json:"fixed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (ContentLinkIssue) TableName() string {
	return "preflight_content_link_issues"
}

type ContentLinkIssueEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	IssueID     uint      `gorm:"not null;index" json:"issue_id"`
	EventType   string    `gorm:"size:64;not null;index" json:"event_type"`
	ActorUserID uint      `gorm:"not null;default:0;index" json:"actor_user_id"`
	Note        string    `gorm:"type:text;not null;default:''" json:"note"`
	Metadata    string    `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt   time.Time `gorm:"not null;index" json:"created_at"`
}

func (ContentLinkIssueEvent) TableName() string {
	return "preflight_content_link_issue_events"
}

type ContentLinkIssueStats struct {
	Active      int64 `json:"active"`
	Open        int64 `json:"open"`
	Resolved    int64 `json:"resolved"`
	Verified    int64 `json:"verified"`
	Ignored     int64 `json:"ignored"`
	Fixable     int64 `json:"fixable"`
	NeedsSource int64 `json:"needs_source"`
	Applied     int64 `json:"applied"`
}

type ContentLinkDetection struct {
	RouteEntryID    *uint
	RunID           uint
	TargetURL       string
	FinalURL        string
	LinkURL         string
	LinkText        string
	Selector        string
	Snippet         string
	SourceType      string
	SourceID        *uint
	SourceKey       string
	SourceField     string
	IssueKey        string
	Severity        string
	SuggestedText   string
	FixStatus       string
	LatestEvidence  string
	FirstDetectedAt time.Time
	LastDetectedAt  time.Time
}
