package payment

import "time"

type PaymentRiskLevel string

const (
	PaymentRiskLevelNormal   PaymentRiskLevel = "normal"
	PaymentRiskLevelWarning  PaymentRiskLevel = "warning"
	PaymentRiskLevelCritical PaymentRiskLevel = "critical"
)

// PaymentRiskSnapshot is an internal rolling-window view. The rate fields are
// operational indicators, not claims about a provider's official program.
type PaymentRiskSnapshot struct {
	ID                      uint             `gorm:"primarykey" json:"id"`
	Provider                string           `gorm:"index;not null" json:"provider"`
	WindowDays              int              `gorm:"not null" json:"window_days"`
	WindowStart             time.Time        `gorm:"index;not null" json:"window_start"`
	WindowEnd               time.Time        `gorm:"index;not null" json:"window_end"`
	SuccessfulPaymentCount  int64            `gorm:"not null;default:0" json:"successful_payment_count"`
	SuccessfulPaymentAmount float64          `gorm:"not null;default:0" json:"successful_payment_amount"`
	DisputeCount            int64            `gorm:"not null;default:0" json:"dispute_count"`
	DisputeAmount           float64          `gorm:"not null;default:0" json:"dispute_amount"`
	EarlyFraudWarningCount  int64            `gorm:"not null;default:0" json:"early_fraud_warning_count"`
	RefundCount             int64            `gorm:"not null;default:0" json:"refund_count"`
	RefundAmount            float64          `gorm:"not null;default:0" json:"refund_amount"`
	DisputeActivityRate     float64          `gorm:"not null;default:0" json:"dispute_activity_rate"`
	EarlyFraudWarningRate   float64          `gorm:"not null;default:0" json:"early_fraud_warning_rate"`
	RefundRate              float64          `gorm:"not null;default:0" json:"refund_rate"`
	Level                   PaymentRiskLevel `gorm:"index;not null" json:"level"`
	RecommendedAction       string           `gorm:"not null;default:''" json:"recommended_action"`
	ReasonsJSON             string           `gorm:"type:text" json:"-"`
	ComputedAt              time.Time        `gorm:"index;not null" json:"computed_at"`
	CreatedAt               time.Time        `json:"created_at"`
}

func (PaymentRiskSnapshot) TableName() string {
	return "payment_risk_snapshots"
}
