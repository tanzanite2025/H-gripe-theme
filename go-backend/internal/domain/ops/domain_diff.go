package ops

import "time"

const (
	DomainDiffStatusMatched = "matched"
	DomainDiffStatusDrifted = "drifted"
	DomainDiffStatusUnknown = "unknown"
	DomainDiffStatusError   = "error"
)

type DomainDiff struct {
	DomainID       uint             `json:"domain_id"`
	Domain         string           `json:"domain"`
	Environment    string           `json:"environment"`
	GeneratedAt    time.Time        `json:"generated_at"`
	Status         string           `json:"status"`
	Summary        string           `json:"summary"`
	ObservedSource string           `json:"observed_source,omitempty"`
	LastObservedAt *time.Time       `json:"last_observed_at,omitempty"`
	ObservedError  string           `json:"observed_error,omitempty"`
	Items          []DomainDiffItem `json:"items"`
}

type DomainDiffItem struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Desired  string `json:"desired"`
	Observed string `json:"observed"`
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
}
