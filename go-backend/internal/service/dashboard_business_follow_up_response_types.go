package service

import "time"

// DashboardCommandCenterResponse grows by verified business slices. Unimplemented
// financial/logistics metrics are omitted rather than represented as healthy zeros.
type DashboardCommandCenterResponse struct {
	Timestamp           time.Time                         `json:"timestamp"`
	ActionCenter        DashboardCommandCenterActionItems `json:"action_center"`
	SystemDegradedNodes []string                          `json:"system_degraded_nodes"`
}

type DashboardCommandCenterActionItems struct {
	OutboxFailures DashboardCommandCenterOutboxFailures `json:"outbox_failures"`
}

type DashboardCommandCenterOutboxFailures struct {
	Status               string                          `json:"status"`
	TotalCount           *int64                          `json:"total_count"`
	FailedCount          *int64                          `json:"failed_count"`
	DeadLetterCount      *int64                          `json:"dead_letter_count"`
	ManualAttentionCount *int64                          `json:"manual_attention_count"`
	AutomaticRetryCount  *int64                          `json:"automatic_retry_count"`
	PreviewLimit         int                             `json:"preview_limit"`
	Items                []DashboardBusinessFollowUpItem `json:"items"`
}

type DashboardBusinessFollowUpRelatedRecord struct {
	Label              string `json:"label"`
	Path               string `json:"path"`
	RequiredPermission string `json:"required_permission"`
}

type DashboardBusinessFollowUpItem struct {
	ID                  uint                                    `json:"id"`
	Title               string                                  `json:"title"`
	Category            string                                  `json:"category"`
	BusinessImpact      string                                  `json:"business_impact"`
	ResponsibleTeam     string                                  `json:"responsible_team"`
	SalesNextSteps      []string                                `json:"sales_next_steps"`
	ReasonSummary       string                                  `json:"reason_summary"`
	ProcessingState     string                                  `json:"processing_state"`
	Attempts            int                                     `json:"attempts"`
	MaxAttempts         int                                     `json:"max_attempts"`
	CreatedAt           time.Time                               `json:"created_at"`
	UpdatedAt           time.Time                               `json:"updated_at"`
	LastAttemptAt       *time.Time                              `json:"last_attempt_at"`
	NextRetryAt         *time.Time                              `json:"next_retry_at"`
	RelatedRecord       *DashboardBusinessFollowUpRelatedRecord `json:"related_record"`
	RelatedRecordNotice string                                  `json:"related_record_notice"`
}
