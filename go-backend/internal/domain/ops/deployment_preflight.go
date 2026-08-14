package ops

import "time"

const (
	DeploymentCheckPass    = "pass"
	DeploymentCheckWarning = "warning"
	DeploymentCheckBlock   = "block"
	DeploymentCheckInfo    = "info"

	DeploymentStatusReady   = "ready"
	DeploymentStatusReview  = "review"
	DeploymentStatusBlocked = "blocked"
)

type DeploymentPreflight struct {
	ProjectID     uint                       `json:"project_id"`
	Project       string                     `json:"project"`
	Environment   string                     `json:"environment"`
	GeneratedAt   time.Time                  `json:"generated_at"`
	Ready         bool                       `json:"ready"`
	StatusLevel   string                     `json:"status_level"`
	BlockingCount int                        `json:"blocking_count"`
	WarningCount  int                        `json:"warning_count"`
	PassCount     int                        `json:"pass_count"`
	InfoCount     int                        `json:"info_count"`
	Summary       string                     `json:"summary"`
	NextActions   []string                   `json:"next_actions"`
	Categories    []DeploymentPreflightGroup `json:"categories"`
	Checks        []DeploymentPreflightCheck `json:"checks"`
}

type DeploymentPreflightOverview struct {
	Environment  string                       `json:"environment"`
	GeneratedAt  time.Time                    `json:"generated_at"`
	ProjectCount int                          `json:"project_count"`
	ReadyCount   int                          `json:"ready_count"`
	ReviewCount  int                          `json:"review_count"`
	BlockedCount int                          `json:"blocked_count"`
	WarningCount int                          `json:"warning_count"`
	Categories   []DeploymentPreflightGroup   `json:"categories"`
	Projects     []DeploymentPreflightSummary `json:"projects"`
}

type DeploymentPreflightSummary struct {
	ProjectID     uint                       `json:"project_id"`
	Project       string                     `json:"project"`
	Environment   string                     `json:"environment"`
	Ready         bool                       `json:"ready"`
	StatusLevel   string                     `json:"status_level"`
	BlockingCount int                        `json:"blocking_count"`
	WarningCount  int                        `json:"warning_count"`
	PassCount     int                        `json:"pass_count"`
	InfoCount     int                        `json:"info_count"`
	Summary       string                     `json:"summary"`
	BlockReasons  []string                   `json:"block_reasons"`
	WarnReasons   []string                   `json:"warn_reasons"`
	NextActions   []string                   `json:"next_actions"`
	Categories    []DeploymentPreflightGroup `json:"categories"`
	GeneratedAt   time.Time                  `json:"generated_at"`
}

type DeploymentPreflightGroup struct {
	Category      string `json:"category"`
	Label         string `json:"label"`
	TotalCount    int    `json:"total_count"`
	BlockingCount int    `json:"blocking_count"`
	WarningCount  int    `json:"warning_count"`
	PassCount     int    `json:"pass_count"`
	InfoCount     int    `json:"info_count"`
}

type DeploymentPreflightCheck struct {
	Key      string `json:"key"`
	Category string `json:"category,omitempty"`
	Label    string `json:"label"`
	Status   string `json:"status"`
	Message  string `json:"message"`
	Detail   string `json:"detail,omitempty"`
}
