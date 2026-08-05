package payment

import (
	"time"

	"gorm.io/gorm"
)

// PaymentMethod 支付方式
type PaymentMethod struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Code        string         `gorm:"uniqueIndex;not null" json:"code"`
	Icon        string         `json:"icon"`
	Description string         `gorm:"type:text" json:"description"`
	FeeType     string         `gorm:"default:'fixed'" json:"fee_type"` // fixed, percentage
	FeeValue    float64        `gorm:"default:0" json:"fee_value"`
	MinAmount   float64        `gorm:"default:0" json:"min_amount"`
	MaxAmount   float64        `gorm:"default:0" json:"max_amount"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	Settings    string         `gorm:"type:text" json:"settings"` // JSON格式的额外设置
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (PaymentMethod) TableName() string {
	return "payment_methods"
}

// CalculateFee 计算手续费
func (pm *PaymentMethod) CalculateFee(amount float64) float64 {
	if pm.FeeType == "percentage" {
		return amount * pm.FeeValue / 100
	}
	return pm.FeeValue
}
