package payment

import "time"

type PaymentProtectionAction string

const (
	PaymentProtectionActionForce3DS     PaymentProtectionAction = "force_3ds"
	PaymentProtectionActionPausePayment PaymentProtectionAction = "pause_payment"
)

type PaymentProtectionScope string

const (
	PaymentProtectionScopeGlobal        PaymentProtectionScope = "global"
	PaymentProtectionScopeProvider      PaymentProtectionScope = "provider"
	PaymentProtectionScopeCountry       PaymentProtectionScope = "country"
	PaymentProtectionScopePaymentMethod PaymentProtectionScope = "payment_method"
)

// PaymentProtectionControl is a manually managed, time-bounded override.
// Active is derived at read time and is not persisted.
type PaymentProtectionControl struct {
	ID         uint                    `gorm:"primarykey" json:"id"`
	Action     PaymentProtectionAction `gorm:"size:48;not null;index" json:"action"`
	ScopeType  PaymentProtectionScope  `gorm:"size:32;not null;index" json:"scope_type"`
	ScopeValue string                  `gorm:"size:128;not null;default:''" json:"scope_value"`
	Reason     string                  `gorm:"type:text;not null" json:"reason"`
	ExpiresAt  time.Time               `gorm:"not null;index" json:"expires_at"`
	Enabled    bool                    `gorm:"not null;default:true;index" json:"enabled"`
	CreatedBy  uint                    `gorm:"not null;index" json:"created_by"`
	UpdatedBy  uint                    `gorm:"not null;index" json:"updated_by"`
	CreatedAt  time.Time               `json:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at"`
	Active     bool                    `gorm:"-" json:"active"`
	Status     string                  `gorm:"-" json:"status"`
}

func (PaymentProtectionControl) TableName() string {
	return "payment_protection_controls"
}

type PaymentProtectionEvaluationInput struct {
	Provider      string
	Country       string
	PaymentMethod string
}

type PaymentProtectionDecision struct {
	Force3DS     bool
	PausePayment bool
	Reasons      []string
}
