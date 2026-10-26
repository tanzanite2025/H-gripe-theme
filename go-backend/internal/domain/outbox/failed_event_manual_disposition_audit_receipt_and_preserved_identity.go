package outbox

import "time"

const FailureDispositionAuditContract = "outbox.failure_manual_disposition.v1"

// FailureDispositionAuditSnapshot keeps routing and failure evidence without
// copying a customer payload or a possibly sensitive error into audit_logs.
// PostgreSQL fingerprints cover the complete persisted row, including fields
// outside the domain model. Other dialects explicitly identify their narrower
// domain-JSON fingerprint and cannot supply a PostgreSQL recovery receipt.
type FailureDispositionAuditSnapshot struct {
	EventID              uint      `json:"event_id"`
	EventKey             string    `json:"event_key"`
	EventType            string    `json:"event_type"`
	AggregateType        string    `json:"aggregate_type"`
	AggregateID          string    `json:"aggregate_id"`
	PayloadSHA256        string    `json:"payload_sha256"`
	Status               string    `json:"status"`
	Attempts             int       `json:"attempts"`
	MaxAttempts          int       `json:"max_attempts"`
	LastErrorSHA256      string    `json:"last_error_sha256"`
	LastErrorBytes       int       `json:"last_error_bytes"`
	UpdatedAt            time.Time `json:"updated_at"`
	RowFingerprintFormat string    `json:"row_fingerprint_format"`
	RowSHA256            string    `json:"row_sha256"`
}

type FailureDispositionAuditReceipt struct {
	Contract  string                          `json:"contract"`
	Operation string                          `json:"operation"`
	Note      string                          `json:"note"`
	Event     FailureDispositionAuditSnapshot `json:"event"`
}

type FailureDispositionResult struct {
	EventID   uint      `json:"id"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
	AuditID   uint      `json:"audit_id"`
}
