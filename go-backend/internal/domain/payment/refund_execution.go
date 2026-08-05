package payment

import "time"

const (
	PaymentRefundExecutionStatusProcessing = "processing"
	PaymentRefundExecutionStatusSucceeded  = "succeeded"
	PaymentRefundExecutionStatusFailed     = "failed"
)

type PaymentRefundExecution struct {
	ID                  uint       `gorm:"primarykey" json:"id"`
	RefundID            uint       `gorm:"uniqueIndex;not null" json:"refund_id"`
	OrderID             uint       `gorm:"index;not null" json:"order_id"`
	TransactionID       uint       `gorm:"index;not null" json:"transaction_id"`
	Provider            string     `gorm:"index;not null" json:"provider"`
	ProviderPaymentID   string     `gorm:"index;not null" json:"provider_payment_id"`
	Amount              float64    `gorm:"not null" json:"amount"`
	Currency            string     `gorm:"not null;default:''" json:"currency"`
	Status              string     `gorm:"index;not null;default:'processing'" json:"status"`
	IdempotencyKey      string     `gorm:"uniqueIndex;not null" json:"idempotency_key"`
	AttemptCount        int        `gorm:"not null;default:1" json:"attempt_count"`
	RequestedByID       uint       `gorm:"index;not null" json:"requested_by_id"`
	RequestedAt         time.Time  `gorm:"not null" json:"requested_at"`
	CompletedAt         *time.Time `json:"completed_at,omitempty"`
	ProviderRefundID    string     `gorm:"index;not null;default:''" json:"provider_refund_id"`
	ProviderStatus      string     `gorm:"not null;default:''" json:"provider_status"`
	GatewayResponseJSON string     `gorm:"type:text" json:"-"`
	ErrorMessage        string     `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (PaymentRefundExecution) TableName() string {
	return "payment_refund_executions"
}
