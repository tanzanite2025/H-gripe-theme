package urlmanagement

import (
	"time"

	seodomain "commerce-platform/internal/domain/seo"
)

const (
	URLIssueTypeRedirectChain          = "redirect_chain"
	URLIssueTypeRedirectTargetMismatch = "redirect_target_mismatch"
	URLIssueTypeRedirectStatusMismatch = "redirect_status_mismatch"
	URLIssueTypeNotFound               = "not_found"
	URLIssueTypeServerError            = "server_error"
	URLIssueTypeCanonicalMismatch      = "canonical_mismatch"
	URLIssueTypePathCollision          = "path_collision"
	URLIssueTypeStaleRoute             = "stale_route"
	URLIssueTypeCheckError             = "check_error"

	URLIssueSeverityLow      = "low"
	URLIssueSeverityMedium   = "medium"
	URLIssueSeverityHigh     = "high"
	URLIssueSeverityCritical = "critical"

	URLIssueStateOpen         = "open"
	URLIssueStateAcknowledged = "acknowledged"
	URLIssueStateResolved     = "resolved"
	URLIssueStateVerified     = "verified"
	URLIssueStateSuppressed   = "suppressed"

	URLIssueEventDetected           = "detected"
	URLIssueEventReopened           = "reopened"
	URLIssueEventAcknowledged       = "acknowledged"
	URLIssueEventClaimed            = "claimed"
	URLIssueEventCommented          = "commented"
	URLIssueEventRedirectLinked     = "redirect_linked"
	URLIssueEventResolutionRecorded = "resolution_recorded"
	URLIssueEventSuppressed         = "suppressed"
	URLIssueEventVerificationPassed = "verification_passed"
	URLIssueEventVerificationFailed = "verification_failed"

	URLIssueResolutionRedirectPublished = "redirect_published"
	URLIssueResolutionSourceRestored    = "source_restored"
	URLIssueResolutionSourcePathChanged = "source_path_changed"
	URLIssueResolutionCanonicalFixed    = "canonical_fixed"
	URLIssueResolutionRuntimeFixed      = "runtime_fixed"
	URLIssueResolutionRetired           = "retired"
	URLIssueResolutionNotApplicable     = "not_applicable"
)

// StorefrontURLIssue is the human-owned lifecycle record for a detected
// storefront URL problem. Route observations and check evidence remain in
// their own append-only/catalog tables.
type StorefrontURLIssue struct {
	ID                   uint                                   `gorm:"primaryKey" json:"id"`
	RouteEntryID         uint                                   `gorm:"not null;uniqueIndex:uq_storefront_url_issues_route_type;index" json:"route_entry_id"`
	IssueType            string                                 `gorm:"size:64;not null;uniqueIndex:uq_storefront_url_issues_route_type;index" json:"issue_type"`
	Severity             string                                 `gorm:"size:16;not null;index" json:"severity"`
	State                string                                 `gorm:"size:32;not null;default:'open';index" json:"state"`
	AssigneeID           *uint                                  `gorm:"index" json:"assignee_id,omitempty"`
	ResolutionType       string                                 `gorm:"size:64;not null;default:''" json:"resolution_type"`
	ResolutionNote       string                                 `gorm:"type:text;not null;default:''" json:"resolution_note"`
	LinkedRedirectRuleID *uint                                  `gorm:"index" json:"linked_redirect_rule_id,omitempty"`
	LatestCheckResultID  *uint                                  `gorm:"index" json:"latest_check_result_id,omitempty"`
	FirstDetectedAt      time.Time                              `gorm:"not null;index" json:"first_detected_at"`
	LastDetectedAt       time.Time                              `gorm:"not null;index" json:"last_detected_at"`
	ResolvedAt           *time.Time                             `gorm:"index" json:"resolved_at,omitempty"`
	VerifiedAt           *time.Time                             `gorm:"index" json:"verified_at,omitempty"`
	SuppressedUntil      *time.Time                             `gorm:"index" json:"suppressed_until,omitempty"`
	SuppressionReason    string                                 `gorm:"type:text;not null;default:''" json:"suppression_reason"`
	CreatedAt            time.Time                              `json:"created_at"`
	UpdatedAt            time.Time                              `json:"updated_at"`
	RouteEntry           *seodomain.StorefrontRouteCatalogEntry `gorm:"foreignKey:RouteEntryID;references:ID" json:"route_entry,omitempty"`
}

func (StorefrontURLIssue) TableName() string {
	return "storefront_url_issues"
}

// StorefrontURLIssueEvent is an immutable timeline entry for a URL issue.
type StorefrontURLIssueEvent struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	IssueID     uint      `gorm:"not null;index" json:"issue_id"`
	EventType   string    `gorm:"size:64;not null;index" json:"event_type"`
	ActorUserID uint      `gorm:"not null;default:0;index" json:"actor_user_id"`
	Note        string    `gorm:"type:text;not null;default:''" json:"note"`
	Metadata    string    `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	CreatedAt   time.Time `gorm:"not null;index" json:"created_at"`
}

func (StorefrontURLIssueEvent) TableName() string {
	return "storefront_url_issue_events"
}

type StorefrontURLIssueStats struct {
	Active       int64 `json:"active"`
	Open         int64 `json:"open"`
	Acknowledged int64 `json:"acknowledged"`
	Resolved     int64 `json:"resolved"`
	Verified     int64 `json:"verified"`
	Suppressed   int64 `json:"suppressed"`
	Critical     int64 `json:"critical"`
	High         int64 `json:"high"`
}

type StorefrontURLIssueActionInput struct {
	Note string `json:"note"`
}

type StorefrontURLIssueResolutionInput struct {
	ResolutionType       string `json:"resolution_type"`
	ResolutionNote       string `json:"resolution_note"`
	LinkedRedirectRuleID *uint  `json:"linked_redirect_rule_id"`
}

type StorefrontURLIssueSuppressionInput struct {
	Reason          string `json:"reason"`
	SuppressedUntil string `json:"suppressed_until"`
}

type StorefrontURLIssueLinkRedirectInput struct {
	RedirectRuleID uint `json:"redirect_rule_id"`
}
