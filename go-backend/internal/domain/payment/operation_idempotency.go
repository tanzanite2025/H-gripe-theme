package payment

import "time"

const (
	PaymentOperationIdempotencyPending     = "pending"
	PaymentOperationIdempotencyReconciling = "reconciling"
	PaymentOperationIdempotencyCompleted   = "completed"
)

// PaymentOperationIdempotency is the durable idempotency record for
// provider-facing capture and confirmation endpoints. Redis may accelerate
// retries, but this row remains the correctness boundary when Redis loses a
// key or becomes unavailable.
type PaymentOperationIdempotency struct {
	ID                      uint       `gorm:"primarykey" json:"id"`
	UserID                  uint       `gorm:"not null;uniqueIndex:idx_payment_operation_idempotency_scope_key" json:"user_id"`
	Scope                   string     `gorm:"type:varchar(64);not null;uniqueIndex:idx_payment_operation_idempotency_scope_key" json:"scope"`
	IdempotencyKey          string     `gorm:"type:varchar(255);not null;uniqueIndex:idx_payment_operation_idempotency_scope_key" json:"idempotency_key"`
	RequestHash             string     `gorm:"type:varchar(64);not null" json:"request_hash"`
	Status                  string     `gorm:"type:varchar(16);not null;index" json:"status"`
	StatusCode              int        `gorm:"not null;default:0" json:"status_code"`
	ContentType             string     `gorm:"type:varchar(255);not null;default:''" json:"content_type"`
	ResponseBody            string     `gorm:"type:text;not null;default:''" json:"response_body"`
	ClaimToken              string     `gorm:"type:varchar(64);not null;default:'';index" json:"-"`
	LeaseExpiresAt          *time.Time `gorm:"index" json:"lease_expires_at,omitempty"`
	ReconciliationStartedAt *time.Time `gorm:"index" json:"reconciliation_started_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

func (PaymentOperationIdempotency) TableName() string {
	return "payment_operation_idempotencies"
}
