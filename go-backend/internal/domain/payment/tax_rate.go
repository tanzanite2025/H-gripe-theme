package payment

import (
	domainmoney "commerce-platform/internal/domain/money"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// TaxRate 税率
type TaxRate struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	Name       string         `gorm:"not null" json:"name"`
	Country    string         `gorm:"index" json:"country"`
	State      string         `gorm:"index" json:"state"`
	City       string         `json:"city"`
	PostalCode string         `json:"postal_code"`
	Rate       float64        `gorm:"not null" json:"rate"` // 百分比，如 7.5 表示 7.5%
	Priority   int            `gorm:"default:0" json:"priority"`
	Compound   bool           `gorm:"default:false" json:"compound"` // 是否复合税率
	Enabled    bool           `gorm:"default:true" json:"enabled"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
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
	rate, ok := new(big.Rat).SetString(strconv.FormatFloat(tr.Rate, 'f', -1, 64))
	if !ok || rate.Sign() < 0 {
		return domainmoney.Money{}, fmt.Errorf("invalid tax rate percentage")
	}
	rate.Quo(rate, big.NewRat(100, 1))
	return amount.MultiplyRat(rate)
}
