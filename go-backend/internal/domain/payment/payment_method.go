package payment

import (
	domainmoney "commerce-platform/internal/domain/money"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"
)

// PaymentMethod 支付方式
type PaymentMethod struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	Name           string         `gorm:"not null" json:"name"`
	Code           string         `gorm:"uniqueIndex;not null" json:"code"`
	Icon           string         `json:"icon"`
	Description    string         `gorm:"type:text" json:"description"`
	FeeType        string         `gorm:"default:'fixed'" json:"fee_type"` // fixed, percentage
	FeeRateDecimal string         `gorm:"column:fee_rate_decimal;type:numeric(30,15);not null;default:0" json:"fee_rate_decimal"`
	FeeValueMinor  int64          `gorm:"column:fee_value_minor;not null;default:0" json:"fee_value_minor"`
	MinAmountMinor int64          `gorm:"column:min_amount_minor;not null;default:0" json:"min_amount_minor"`
	MaxAmountMinor int64          `gorm:"column:max_amount_minor;not null;default:0" json:"max_amount_minor"`
	Enabled        bool           `gorm:"default:true" json:"enabled"`
	SortOrder      int            `gorm:"default:0" json:"sort_order"`
	Settings       string         `gorm:"type:text" json:"settings"` // JSON格式的额外设置
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (pm *PaymentMethod) BeforeSave(tx *gorm.DB) error {
	if pm == nil {
		return fmt.Errorf("payment method is required")
	}
	if strings.TrimSpace(pm.FeeType) == "" {
		pm.FeeType = "fixed"
	}
	if pm.FeeType != "fixed" && pm.FeeType != "percentage" {
		return fmt.Errorf("payment method fee type must be fixed or percentage")
	}
	if pm.FeeType == "percentage" && strings.TrimSpace(pm.FeeRateDecimal) == "" {
		return fmt.Errorf("payment method fee rate is required")
	}
	if pm.FeeValueMinor < 0 || pm.MinAmountMinor < 0 || pm.MaxAmountMinor < 0 {
		return fmt.Errorf("payment method monetary fields cannot be negative")
	}
	if pm.MaxAmountMinor > 0 && pm.MinAmountMinor > pm.MaxAmountMinor {
		return fmt.Errorf("minimum amount cannot exceed maximum amount")
	}
	rate, err := pm.FeeRate()
	if err != nil {
		return err
	}
	pm.FeeRateDecimal = rate
	return nil
}

// FeeRate validates and returns the canonical decimal percentage used by the
// payment fee calculator. NUMERIC(30,15) must not silently round API input.
func (pm PaymentMethod) FeeRate() (string, error) {
	value := strings.TrimSpace(pm.FeeRateDecimal)
	if value == "" {
		value = "0"
	}
	if strings.Contains(value, "/") {
		return "", fmt.Errorf("payment method fee rate must be a decimal")
	}
	rate, ok := new(big.Rat).SetString(value)
	if !ok || rate.Sign() < 0 || rate.Cmp(big.NewRat(100, 1)) > 0 {
		return "", fmt.Errorf("payment method fee rate must be between 0 and 100")
	}
	scaled := new(big.Rat).Mul(rate, new(big.Rat).SetInt64(1_000_000_000_000_000))
	if scaled.Denom().Cmp(big.NewInt(1)) != 0 {
		return "", fmt.Errorf("payment method fee rate supports at most 15 fractional digits")
	}
	return value, nil
}

// TableName 指定表名
func (PaymentMethod) TableName() string {
	return "payment_methods"
}

// CalculateFeeMoney calculates the gateway fee in the amount's currency.
func (pm *PaymentMethod) CalculateFeeMoney(amount domainmoney.Money) (domainmoney.Money, error) {
	if pm == nil {
		return domainmoney.Money{}, fmt.Errorf("payment method is required")
	}
	if pm.FeeType == "percentage" {
		rateText, err := pm.FeeRate()
		if err != nil {
			return domainmoney.Money{}, err
		}
		rate, ok := new(big.Rat).SetString(rateText)
		if !ok {
			return domainmoney.Money{}, fmt.Errorf("invalid payment method fee percentage")
		}
		rate.Quo(rate, big.NewRat(100, 1))
		return amount.MultiplyRat(rate)
	}
	return domainmoney.New(pm.FeeValueMinor, amount.Currency().String())
}
