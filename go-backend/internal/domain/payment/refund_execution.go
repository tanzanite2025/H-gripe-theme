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
	ID                             uint       `gorm:"primarykey" json:"id"`
	RefundID                       uint       `gorm:"uniqueIndex;not null" json:"refund_id"`
	OrderID                        uint       `gorm:"index;not null" json:"order_id"`
	TransactionID                  uint       `gorm:"index;not null" json:"transaction_id"`
	Provider                       string     `gorm:"index;not null" json:"provider"`
	ProviderPaymentID              string     `gorm:"index;not null" json:"provider_payment_id"`
	MerchantOrderNumber            string     `gorm:"index;not null" json:"merchant_order_number"`
	ProviderTransactionID          string     `gorm:"index;not null" json:"provider_transaction_id"`
	AmountMinor                    int64      `gorm:"column:amount_minor;not null;default:0" json:"amount_minor"`
	Currency                       string     `gorm:"not null;default:''" json:"currency"`
	Status                         string     `gorm:"index;not null;default:'processing'" json:"status"`
	IdempotencyKey                 string     `gorm:"uniqueIndex;not null" json:"idempotency_key"`
	AttemptCount                   int        `gorm:"not null;default:1" json:"attempt_count"`
	RequestedByID                  uint       `gorm:"index;not null" json:"requested_by_id"`
	RequestedAt                    time.Time  `gorm:"not null" json:"requested_at"`
	CompletedAt                    *time.Time `json:"completed_at,omitempty"`
	ProviderRefundID               string     `gorm:"index;not null;default:''" json:"provider_refund_id"`
	ProviderStatus                 string     `gorm:"not null;default:''" json:"provider_status"`
	SettlementAmountMinor          int64      `gorm:"column:settlement_amount_minor;not null;default:0" json:"settlement_amount_minor"`
	SettlementCurrency             string     `gorm:"column:settlement_currency;size:3;not null;default:''" json:"settlement_currency"`
	SettlementBalanceTransactionID string     `gorm:"column:settlement_balance_transaction_id;size:255;not null;default:''" json:"settlement_balance_transaction_id"`
	FXGainLossMinor                int64      `gorm:"column:fx_gain_loss_minor;not null;default:0" json:"fx_gain_loss_minor"`
	FXGainLossCurrency             string     `gorm:"column:fx_gain_loss_currency;size:3;not null;default:''" json:"fx_gain_loss_currency"`
	GatewayResponseJSON            string     `gorm:"type:text" json:"-"`
	ErrorMessage                   string     `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt                      time.Time  `json:"created_at"`
	UpdatedAt                      time.Time  `json:"updated_at"`
}

func (e *PaymentRefundExecution) BeforeSave(tx *gorm.DB) error {
	e.Currency = currency.NormalizeCode(e.Currency)
	e.SettlementCurrency = currency.NormalizeCode(e.SettlementCurrency)
	e.FXGainLossCurrency = currency.NormalizeCode(e.FXGainLossCurrency)
	if !currency.IsCatalogCode(e.Currency) {
		return errors.New("payment refund execution currency must be a supported ISO 4217 code")
	}
	if e.AmountMinor < 0 {
		return errors.New("payment refund execution amount cannot be negative")
	}
	if e.SettlementAmountMinor < 0 {
		return errors.New("payment refund execution settlement amount cannot be negative")
	}
	if e.SettlementAmountMinor > 0 && !currency.IsCatalogCode(e.SettlementCurrency) {
		return errors.New("payment refund execution settlement currency must be a supported ISO 4217 code")
	}
	if e.FXGainLossMinor != 0 && !currency.IsCatalogCode(e.FXGainLossCurrency) {
		return errors.New("payment refund execution FX gain/loss currency must be a supported ISO 4217 code")
	}
	return nil
}

func (e PaymentRefundExecution) AmountMoney() (domainmoney.Money, error) {
	return domainmoney.New(e.AmountMinor, e.Currency)
}

func (PaymentRefundExecution) TableName() string {
	return "payment_refund_executions"
}
