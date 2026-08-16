package payment

import "time"

// PaymentRiskCheckoutDecision records the decision made before a provider
// payment is created. It makes adaptive 3DS behaviour observable even when
// the payment later fails or remains in requires_action.
type PaymentRiskCheckoutDecision struct {
	ID                 uint      `gorm:"primarykey" json:"id"`
	Provider           string    `gorm:"index;not null" json:"provider"`
	OrderID            *uint     `gorm:"index" json:"order_id,omitempty"`
	ProviderPaymentID  string    `gorm:"index;not null" json:"provider_payment_id"`
	Mode               string    `gorm:"not null" json:"mode"`
	Strategy           string    `gorm:"index;not null" json:"strategy"`
	ExemptionCandidate bool      `gorm:"not null;default:false" json:"exemption_candidate"`
	RiskLevel          string    `gorm:"index;not null" json:"risk_level"`
	RiskScore          int       `gorm:"not null;default:0" json:"risk_score"`
	PortfolioRiskLevel string    `gorm:"index;not null" json:"portfolio_risk_level"`
	ReasonsJSON        string    `gorm:"type:text;not null;default:'[]'" json:"-"`
	Amount             float64   `gorm:"not null;default:0" json:"amount"`
	Currency           string    `gorm:"not null;default:''" json:"currency"`
	OccurredAt         time.Time `gorm:"index;not null" json:"occurred_at"`
	CreatedAt          time.Time `json:"created_at"`
}

func (PaymentRiskCheckoutDecision) TableName() string {
	return "payment_risk_checkout_decisions"
}
