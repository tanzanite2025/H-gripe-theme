package payment

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/currency"

	"gorm.io/gorm"
)

// Transaction 支付交易记录
type Transaction struct {
	ID              uint           `gorm:"primarykey" json:"id"`
	OrderID         uint           `gorm:"not null;index" json:"order_id"`
	TransactionID   string         `gorm:"uniqueIndex" json:"transaction_id"` // 第三方交易ID
	PaymentMethod   string         `gorm:"not null" json:"payment_method"`
	Amount          float64        `gorm:"not null" json:"amount"`
	Currency        string         `gorm:"not null" json:"currency"`
	Status          string         `gorm:"index" json:"status"`               // pending, processing, requires_action, completed, failed, expired, refunded
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

func (t *Transaction) BeforeSave(tx *gorm.DB) error {
	t.Currency = currency.NormalizeCode(t.Currency)
	if !currency.IsValidCode(t.Currency) || !currency.IsCatalogCode(t.Currency) {
		return errors.New("payment transaction currency must be a supported ISO 4217 code")
	}
	return nil
}
