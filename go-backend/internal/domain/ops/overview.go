package ops

import "time"

type OverviewSummary struct {
	Total      int `json:"total"`
	Enabled    int `json:"enabled"`
	Attention  int `json:"attention"`
	Unknown    int `json:"unknown"`
	Healthy    int `json:"healthy"`
	Configured int `json:"configured"`
}

type OverviewResource struct {
	Kind           string    `json:"kind"`
	ID             uint      `json:"id"`
	Name           string    `json:"name"`
	Environment    string    `json:"environment"`
	Status         string    `json:"status"`
	ObservedStatus string    `json:"observed_status,omitempty"`
	HealthStatus   string    `json:"health_status,omitempty"`
	Message        string    `json:"message"`
	Target         string    `json:"target,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type OverviewTopology struct {
	VPS      []VPSBinding         `json:"vps"`
	Projects []ProjectBindingView `json:"projects"`
	Domains  []DomainBinding      `json:"domains"`
}

type Overview struct {
	Environment string                     `json:"environment"`
	GeneratedAt time.Time                  `json:"generated_at"`
	Summary     map[string]OverviewSummary `json:"summary"`
	Topology    OverviewTopology           `json:"topology"`
	Attention   []OverviewResource         `json:"attention"`
	RecentAudit interface{}                `json:"recent_audit"`
}
