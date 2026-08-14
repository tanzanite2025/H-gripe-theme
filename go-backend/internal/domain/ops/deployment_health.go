package ops

import "time"

const (
	DeploymentHealthHealthy  = "healthy"
	DeploymentHealthDegraded = "degraded"
	DeploymentHealthFailed   = "failed"
)

type DeploymentHealthCheckReport struct {
	ProjectID   uint                   `json:"project_id"`
	Project     string                 `json:"project"`
	Environment string                 `json:"environment"`
	Status      string                 `json:"status"`
	CheckedAt   time.Time              `json:"checked_at"`
	Summary     string                 `json:"summary"`
	Checks      []DeploymentHealthItem `json:"checks"`
}

type DeploymentHealthItem struct {
	Domain     string `json:"domain"`
	Check      string `json:"check"`
	Target     string `json:"target"`
	Status     string `json:"status"`
	StatusCode int    `json:"status_code,omitempty"`
	Message    string `json:"message,omitempty"`
	DurationMS int64  `json:"duration_ms,omitempty"`
}
