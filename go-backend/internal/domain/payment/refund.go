package payment

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Refund 退款记录
type Refund struct {
	ID                     uint             `gorm:"primarykey" json:"id"`
	OrderID                uint             `gorm:"not null;index" json:"order_id"`
	TransactionID          uint             `gorm:"index" json:"transaction_id"`
	RefundID               *string          `gorm:"uniqueIndex" json:"refund_id,omitempty"`
	Amount                 float64          `gorm:"not null" json:"amount"` // net amount actually sent to the payment gateway
	RequestedAmount        float64          `gorm:"not null;default:0" json:"requested_amount"`
	DiscountClawbackAmount float64          `gorm:"not null;default:0" json:"discount_clawback_amount"`
	CalculationSnapshot    string           `gorm:"type:text" json:"calculation_snapshot"`
	FXSnapshotData         datatypes.JSON   `gorm:"column:fx_snapshot;type:jsonb;not null;default:'{}'" json:"-"`
	LineItems              []RefundLineItem `gorm:"foreignKey:RefundID" json:"line_items,omitempty"`
	Reason                 string           `gorm:"type:text" json:"reason"`
	Status                 string           `gorm:"index" json:"status"` // pending, completed, failed
	RefundedBy             uint             `json:"refunded_by"`         // 操作人ID
	GatewayResponse        string           `gorm:"type:text" json:"gateway_response"`
	CreatedAt              time.Time        `json:"created_at"`
	UpdatedAt              time.Time        `json:"updated_at"`
	CompletedAt            *time.Time       `json:"completed_at"`
	DeletedAt              gorm.DeletedAt   `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Refund) TableName() string {
	return "refunds"
}

// RefundLineItem records the exact order item facts that produced a refund.
// Product snapshots are copied from order_items so later catalog edits cannot
// alter historical refund accounting.
type RefundLineItem struct {
	ID                 uint       `gorm:"primarykey" json:"id"`
	RefundID           uint       `gorm:"not null;index" json:"refund_id"`
	OrderID            uint       `gorm:"not null;index" json:"order_id"`
	OrderItemID        uint       `gorm:"not null;index" json:"order_item_id"`
	ProductID          uint       `gorm:"not null;index" json:"product_id"`
	VariantID          *uint      `gorm:"index" json:"variant_id,omitempty"`
	ProductName        string     `gorm:"not null" json:"product_name"`
	SKU                string     `json:"sku"`
	Quantity           int        `gorm:"not null" json:"quantity"`
	UnitPrice          float64    `gorm:"not null;default:0" json:"unit_price"`
	LineSubtotalAmount float64    `gorm:"not null;default:0" json:"line_subtotal_amount"`
	LineTaxAmount      float64    `gorm:"not null;default:0" json:"line_tax_amount"`
	LineDiscountAmount float64    `gorm:"not null;default:0" json:"line_discount_amount"`
	LineTotalAmount    float64    `gorm:"not null;default:0" json:"line_total_amount"`
	Restock            bool       `gorm:"not null;default:false" json:"restock"`
	RestockedAt        *time.Time `json:"restocked_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (RefundLineItem) TableName() string {
	return "refund_line_items"
}
