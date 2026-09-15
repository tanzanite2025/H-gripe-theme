package payment

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"gorm.io/gorm"
)

const (
	PaymentRefundExecutionStatusProcessing = "processing"
	PaymentRefundExecutionStatusSucceeded  = "succeeded"
	PaymentRefundExecutionStatusFailed     = "failed"
)

type PaymentRefundExecution struct {
	ID                    uint       `gorm:"primarykey" json:"id"`
	RefundID              uint       `gorm:"uniqueIndex;not null" json:"refund_id"`
	OrderID               uint       `gorm:"index;not null" json:"order_id"`
	TransactionID         uint       `gorm:"index;not null" json:"transaction_id"`
	Provider              string     `gorm:"index;not null" json:"provider"`
	ProviderPaymentID     string     `gorm:"index;not null" json:"provider_payment_id"`
	MerchantOrderNumber   string     `gorm:"index;not null" json:"merchant_order_number"`
	ProviderTransactionID string     `gorm:"index;not null" json:"provider_transaction_id"`
	AmountMinor           int64      `gorm:"column:amount_minor;not null;default:0" json:"amount_minor"`
	Amount                float64    `gorm:"not null" json:"-"`
	Currency              string     `gorm:"not null;default:''" json:"currency"`
	Status                string     `gorm:"index;not null;default:'processing'" json:"status"`
	IdempotencyKey        string     `gorm:"uniqueIndex;not null" json:"idempotency_key"`
	AttemptCount          int        `gorm:"not null;default:1" json:"attempt_count"`
	RequestedByID         uint       `gorm:"index;not null" json:"requested_by_id"`
	RequestedAt           time.Time  `gorm:"not null" json:"requested_at"`
	CompletedAt           *time.Time `json:"completed_at,omitempty"`
	ProviderRefundID      string     `gorm:"index;not null;default:''" json:"provider_refund_id"`
	ProviderStatus        string     `gorm:"not null;default:''" json:"provider_status"`
	GatewayResponseJSON   string     `gorm:"type:text" json:"-"`
	ErrorMessage          string     `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

func (e *PaymentRefundExecution) BeforeSave(tx *gorm.DB) error {
	e.Currency = currency.NormalizeCode(e.Currency)
	if !currency.IsCatalogCode(e.Currency) {
		return errors.New("payment refund execution currency must be a supported ISO 4217 code")
	}
	if e.AmountMinor == 0 && e.Amount != 0 {
		value, err := domainmoney.FromMajorFloat(e.Amount, e.Currency)
		if err != nil {
			return err
		}
		e.AmountMinor = value.AmountMinor()
	}
	if e.AmountMinor < 0 {
		return errors.New("payment refund execution amount cannot be negative")
	}
	return nil
}

func (e PaymentRefundExecution) AmountMoney() (domainmoney.Money, error) {
	if e.AmountMinor == 0 && e.Amount != 0 {
		return domainmoney.FromMajorFloat(e.Amount, e.Currency)
	}
	return domainmoney.New(e.AmountMinor, e.Currency)
}

func (PaymentRefundExecution) TableName() string {
	return "payment_refund_executions"
}
