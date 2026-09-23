package payment

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"

	"gorm.io/gorm"
)

// PaymentRiskCheckoutDecision records the decision made before a provider
// payment is created. It makes adaptive 3DS behaviour observable even when
// the payment later fails or remains in requires_action.
type PaymentRiskCheckoutDecision struct {
	ID                 uint   `gorm:"primarykey" json:"id"`
	Provider           string `gorm:"index;not null" json:"provider"`
	OrderID            *uint  `gorm:"index" json:"order_id,omitempty"`
	ProviderPaymentID  string `gorm:"index;not null" json:"provider_payment_id"`
	Mode               string `gorm:"not null" json:"mode"`
	Strategy           string `gorm:"index;not null" json:"strategy"`
	ExemptionCandidate bool   `gorm:"not null;default:false" json:"exemption_candidate"`
	RiskLevel          string `gorm:"index;not null" json:"risk_level"`
	RiskScore          int    `gorm:"not null;default:0" json:"risk_score"`
	PortfolioRiskLevel string `gorm:"index;not null" json:"portfolio_risk_level"`
	ReasonsJSON        string `gorm:"type:text;not null;default:'[]'" json:"-"`
	AmountMinor        int64  `gorm:"column:amount_minor;not null;default:0" json:"amount_minor"`
	Currency   string    `gorm:"not null;default:''" json:"currency"`
	OccurredAt time.Time `gorm:"index;not null" json:"occurred_at"`
	CreatedAt  time.Time `json:"created_at"`
}

func (d *PaymentRiskCheckoutDecision) BeforeSave(tx *gorm.DB) error {
	d.Currency = currency.NormalizeCode(d.Currency)
	if d.Currency != "" && !currency.IsCatalogCode(d.Currency) {
		return errors.New("payment risk checkout decision currency must be a supported ISO 4217 code")
	}
	if d.AmountMinor < 0 {
		return errors.New("payment risk checkout decision amount cannot be negative")
	}
	return nil
}

func (d PaymentRiskCheckoutDecision) AmountMoney() (domainmoney.Money, error) {
	return domainmoney.New(d.AmountMinor, d.Currency)
}

func (PaymentRiskCheckoutDecision) TableName() string {
	return "payment_risk_checkout_decisions"
}
