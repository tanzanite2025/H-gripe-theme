package currency

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ExchangeRate struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	BaseCurrency  string         `gorm:"size:3;not null;uniqueIndex:idx_currency_exchange_rate_pair,where:deleted_at IS NULL" json:"base_currency"`
	QuoteCurrency string         `gorm:"size:3;not null;uniqueIndex:idx_currency_exchange_rate_pair,where:deleted_at IS NULL" json:"quote_currency"`
	RateDecimal   string         `gorm:"column:rate_decimal;type:numeric(30,15);not null;default:0" json:"rate_decimal"`
	Source        string         `gorm:"size:80;not null;default:''" json:"source"`
	FetchedAt     time.Time      `gorm:"not null" json:"fetched_at"`
	ExpiresAt     *time.Time     `json:"expires_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (r *ExchangeRate) NormalizeRateDecimal() error {
	if r == nil {
		return fmt.Errorf("exchange rate is required")
	}
	value := strings.TrimSpace(r.RateDecimal)
	if value == "" {
		return fmt.Errorf("exchange rate must be positive")
	}
	if strings.Contains(value, "/") {
		return fmt.Errorf("invalid exchange rate decimal %q", value)
	}
	rat, ok := new(big.Rat).SetString(value)
	if !ok || rat.Sign() <= 0 {
		return fmt.Errorf("invalid exchange rate decimal %q", value)
	}
	r.RateDecimal = value
	return nil
}

func (ExchangeRate) TableName() string {
	return "currency_exchange_rates"
}
