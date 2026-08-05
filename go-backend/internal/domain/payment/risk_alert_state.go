package payment

import "time"

// PaymentRiskAlertState keeps the latest evaluated level per provider. It is
// maintained even when external alert delivery is disabled so future changes
// are evaluated against the actual current state.
type PaymentRiskAlertState struct {
	Provider          string           `gorm:"primaryKey;size:32" json:"provider"`
	CurrentLevel      PaymentRiskLevel `gorm:"not null;default:'normal'" json:"current_level"`
	CurrentSnapshotID uint             `gorm:"not null;default:0" json:"current_snapshot_id"`
	UpdatedAt         time.Time        `json:"updated_at"`
}

func (PaymentRiskAlertState) TableName() string {
	return "payment_risk_alert_states"
}
