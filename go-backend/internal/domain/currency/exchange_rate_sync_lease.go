package currency

import "time"

// ExchangeRateSyncLease coordinates the daily sync across application replicas.
type ExchangeRateSyncLease struct {
	LeaseKey       string    `gorm:"primaryKey;size:80"`
	OwnerID        string    `gorm:"size:160;not null"`
	LeaseExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (ExchangeRateSyncLease) TableName() string {
	return "currency_exchange_sync_leases"
}
