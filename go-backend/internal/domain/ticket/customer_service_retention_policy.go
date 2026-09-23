package ticket

import "time"

// CustomerServiceRetentionPolicy is the single runtime policy owned by the
// customer-service conversation domain. It intentionally does not use the
// generic settings aggregate.
type CustomerServiceRetentionPolicy struct {
	ID                   uint      `gorm:"primarykey" json:"id"`
	Enabled              bool      `gorm:"not null;default:false" json:"enabled"`
	IntervalSeconds      int       `gorm:"not null;default:86400" json:"interval_seconds"`
	MinimumRetentionDays int       `gorm:"not null;default:730" json:"minimum_retention_days"`
	RecoveryWindowDays   int       `gorm:"not null;default:30" json:"recovery_window_days"`
	BatchLimit           int       `gorm:"not null;default:100" json:"batch_limit"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (CustomerServiceRetentionPolicy) TableName() string {
	return "customer_service_retention_policies"
}
