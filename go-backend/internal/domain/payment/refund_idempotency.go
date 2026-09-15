package payment

import "time"

// RefundIdempotency is the durable idempotency fact for an admin refund
// request. The row and the refund are committed in the same transaction.
type RefundIdempotency struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	AdminUserID    uint      `gorm:"not null;index" json:"admin_user_id"`
	Scope          string    `gorm:"type:varchar(64);not null;uniqueIndex:idx_refund_idempotency_scope_key" json:"scope"`
	IdempotencyKey string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_refund_idempotency_scope_key" json:"idempotency_key"`
	RequestHash    string    `gorm:"type:varchar(64);not null" json:"request_hash"`
	RefundID       *uint     `gorm:"index" json:"refund_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (RefundIdempotency) TableName() string {
	return "refund_idempotencies"
}
