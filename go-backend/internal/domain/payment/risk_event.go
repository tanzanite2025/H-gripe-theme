package payment

import "time"

import (
	"errors"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"

	"gorm.io/gorm"
)

type PaymentRiskProvider string

const (
	PaymentRiskProviderStripe PaymentRiskProvider = "stripe"
	PaymentRiskProviderPayPal PaymentRiskProvider = "paypal"
)

type PaymentRiskEventKind string

const (
	PaymentRiskEventEarlyFraudWarning PaymentRiskEventKind = "early_fraud_warning"
	PaymentRiskEventDispute           PaymentRiskEventKind = "dispute"
)

// PaymentRiskEvent is the provider-neutral event ledger used for idempotent
// monitoring. Provider payloads stay here for audit and later investigation;
// checkout decisions use normalized fields only.
type PaymentRiskEvent struct {
	ID                uint                 `gorm:"primarykey" json:"id"`
	Provider          string               `gorm:"index;not null" json:"provider"`
	Kind              PaymentRiskEventKind `gorm:"index;not null" json:"kind"`
	ExternalReference string               `gorm:"not null" json:"external_reference"`
	WebhookEventID    string               `gorm:"index" json:"webhook_event_id"`
	ProviderPaymentID string               `gorm:"index" json:"provider_payment_id"`
	PaymentIntentID   string               `gorm:"index" json:"payment_intent_id"`
	ChargeID          string               `gorm:"index" json:"charge_id"`
	OrderID           *uint                `gorm:"index" json:"order_id,omitempty"`
	TransactionID     *uint                `gorm:"index" json:"transaction_id,omitempty"`
	AmountMinor       int64                `gorm:"column:amount_minor;not null;default:0" json:"amount_minor"`
	Currency     string    `gorm:"not null;default:''" json:"currency"`
	OccurredAt   time.Time `gorm:"index;not null" json:"occurred_at"`
	Payload      string    `gorm:"type:text" json:"-"`
	MetadataJSON string    `gorm:"type:text" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (e *PaymentRiskEvent) BeforeSave(tx *gorm.DB) error {
	e.Currency = currency.NormalizeCode(e.Currency)
	if e.Currency != "" && !currency.IsCatalogCode(e.Currency) {
		return errors.New("payment risk event currency must be a supported ISO 4217 code")
	}
	if e.AmountMinor < 0 {
		return errors.New("payment risk event amount cannot be negative")
	}
	return nil
}

func (e PaymentRiskEvent) AmountMoney() (domainmoney.Money, error) {
	return domainmoney.New(e.AmountMinor, e.Currency)
}

func (PaymentRiskEvent) TableName() string {
	return "payment_risk_events"
}
