package payment

import (
	"time"

	"gorm.io/gorm"
)

// Refund 退款记录
type Refund struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	OrderID         uint           `gorm:"not null;index" json:"order_id"`
	TransactionID   uint           `gorm:"index" json:"transaction_id"`
	RefundID        *string        `gorm:"uniqueIndex" json:"refund_id,omitempty"`
	Amount          float64        `gorm:"not null" json:"amount"`
	Reason          string         `gorm:"type:text" json:"reason"`
	Status          string         `gorm:"index" json:"status"` // pending, completed, failed
	RefundedBy      uint           `json:"refunded_by"`         // 操作人ID
	GatewayResponse string         `gorm:"type:text" json:"gateway_response"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CompletedAt     *time.Time     `json:"completed_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Refund) TableName() string {
	return "refunds"
}
