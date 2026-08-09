package currency

import (
	"time"

	"gorm.io/gorm"
)

type ExchangeRate struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	BaseCurrency  string         `gorm:"size:3;not null;uniqueIndex:idx_currency_exchange_rate_pair" json:"base_currency"`
	QuoteCurrency string         `gorm:"size:3;not null;uniqueIndex:idx_currency_exchange_rate_pair" json:"quote_currency"`
	Rate          float64        `gorm:"not null" json:"rate"`
	Source        string         `gorm:"size:80;not null;default:''" json:"source"`
	FetchedAt     time.Time      `gorm:"not null" json:"fetched_at"`
	ExpiresAt     *time.Time     `json:"expires_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ExchangeRate) TableName() string {
	return "currency_exchange_rates"
}
