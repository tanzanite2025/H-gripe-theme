package payment

import (
	domainmoney "commerce-platform/internal/domain/money"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"
)

// TaxRate 税率
type TaxRate struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	Name       string `gorm:"not null" json:"name"`
	Country    string `gorm:"index" json:"country"`
	State      string `gorm:"index" json:"state"`
	City       string `json:"city"`
	PostalCode string `json:"postal_code"`
	// RateDecimal is the exact percentage used for tax arithmetic and is the
	// sole persisted source of truth (for example "7.5" means 7.5%).
	RateDecimal string         `gorm:"column:rate_decimal;type:numeric(30,15);not null;default:0" json:"rate_decimal"`
	Priority    int            `gorm:"default:0" json:"priority"`
	Compound    bool           `gorm:"default:false" json:"compound"` // 是否复合税率
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

func (tr *TaxRate) BeforeSave(tx *gorm.DB) error {
	value := strings.TrimSpace(tr.RateDecimal)
	if value == "" {
		// GORM partial updates (for example toggling Enabled) execute hooks
		// with a zero-value model. Preserve the existing database decimal in
		// that case; only validate when the canonical field is being written.
		if tx != nil && tx.Statement != nil && !tx.Statement.Changed("RateDecimal") {
			return nil
		}
		return fmt.Errorf("tax rate decimal is required")
	}
	rate, ok := new(big.Rat).SetString(value)
	if !ok || rate.Sign() < 0 || rate.Cmp(big.NewRat(100, 1)) > 0 {
		return fmt.Errorf("tax rate decimal must be between 0 and 100")
	}
	// Keep the in-memory value aligned with NUMERIC(30,15); PostgreSQL must not
	// silently round a more precise rate during persistence.
	scaled := new(big.Rat).Mul(rate, new(big.Rat).SetInt64(1_000_000_000_000_000))
	if scaled.Denom().Cmp(big.NewInt(1)) != 0 {
		return fmt.Errorf("tax rate decimal supports at most 15 fractional digits")
	}
	tr.RateDecimal = value
	return nil
}

// TableName 指定表名
func (TaxRate) TableName() string {
	return "tax_rates"
}

// CalculateTaxMoney calculates tax in minor units using the configured rate.
func (tr *TaxRate) CalculateTaxMoney(amount domainmoney.Money) (domainmoney.Money, error) {
	if tr == nil {
		return domainmoney.Money{}, fmt.Errorf("tax rate is required")
	}
	rateText := strings.TrimSpace(tr.RateDecimal)
	rate, ok := new(big.Rat).SetString(rateText)
	if !ok || rate.Sign() < 0 || rate.Cmp(big.NewRat(100, 1)) > 0 {
		return domainmoney.Money{}, fmt.Errorf("invalid tax rate percentage")
	}
	rate.Quo(rate, big.NewRat(100, 1))
	return amount.MultiplyRat(rate)
}
