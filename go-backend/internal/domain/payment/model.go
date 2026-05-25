package payment

import (
	"time"

	"gorm.io/gorm"
)

// PaymentMethod 支付方式
type PaymentMethod struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Name          string         `gorm:"not null" json:"name"`
	Code          string         `gorm:"uniqueIndex;not null" json:"code"`
	Icon          string         `json:"icon"`
	Description   string         `gorm:"type:text" json:"description"`
	FeeType       string         `gorm:"default:'fixed'" json:"fee_type"` // fixed, percentage
	FeeValue      float64        `gorm:"default:0" json:"fee_value"`
	MinAmount     float64        `gorm:"default:0" json:"min_amount"`
	MaxAmount     float64        `gorm:"default:0" json:"max_amount"`
	SupportedCurrencies string   `json:"supported_currencies"` // 逗号分隔
	Enabled       bool           `gorm:"default:true" json:"enabled"`
	SortOrder     int            `gorm:"default:0" json:"sort_order"`
	Settings      string         `gorm:"type:text" json:"settings"` // JSON格式的额外设置
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
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

// TaxRate 税率
type TaxRate struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Country     string         `gorm:"index" json:"country"`
	State       string         `gorm:"index" json:"state"`
	City        string         `json:"city"`
	PostalCode  string         `json:"postal_code"`
	Rate        float64        `gorm:"not null" json:"rate"` // 百分比，如 7.5 表示 7.5%
	Priority    int            `gorm:"default:0" json:"priority"`
	Compound    bool           `gorm:"default:false" json:"compound"` // 是否复合税率
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (TaxRate) TableName() string {
	return "tax_rates"
}

// CalculateTax 计算税额
func (tr *TaxRate) CalculateTax(amount float64) float64 {
	return amount * tr.Rate / 100
}

// Transaction 支付交易记录
type Transaction struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	OrderID         uint           `gorm:"not null;index" json:"order_id"`
	TransactionID   string         `gorm:"uniqueIndex" json:"transaction_id"` // 第三方交易ID
	PaymentMethod   string         `gorm:"not null" json:"payment_method"`
	Amount          float64        `gorm:"not null" json:"amount"`
	Currency        string         `gorm:"default:'USD'" json:"currency"`
	Status          string         `gorm:"index" json:"status"` // pending, completed, failed, refunded
	GatewayResponse string         `gorm:"type:text" json:"gateway_response"` // JSON格式
	ErrorMessage    string         `gorm:"type:text" json:"error_message"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	CompletedAt     *time.Time     `json:"completed_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Transaction) TableName() string {
	return "transactions"
}

// Refund 退款记录
type Refund struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	OrderID       uint           `gorm:"not null;index" json:"order_id"`
	TransactionID uint           `gorm:"index" json:"transaction_id"`
	Amount        float64        `gorm:"not null" json:"amount"`
	Reason        string         `gorm:"type:text" json:"reason"`
	Status        string         `gorm:"index" json:"status"` // pending, completed, failed
	RefundedBy    uint           `json:"refunded_by"` // 操作人ID
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	CompletedAt   *time.Time     `json:"completed_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Refund) TableName() string {
	return "refunds"
}
