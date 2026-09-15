package payment

import (
	domainmoney "commerce-platform/internal/domain/money"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"gorm.io/gorm"
)

// PaymentMethod 支付方式
type PaymentMethod struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Code        string         `gorm:"uniqueIndex;not null" json:"code"`
	Icon        string         `json:"icon"`
	Description string         `gorm:"type:text" json:"description"`
	FeeType     string         `gorm:"default:'fixed'" json:"fee_type"` // fixed, percentage
	FeeValue    float64        `gorm:"default:0" json:"fee_value"`
	MinAmount   float64        `gorm:"default:0" json:"min_amount"`
	MaxAmount   float64        `gorm:"default:0" json:"max_amount"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	Settings    string         `gorm:"type:text" json:"settings"` // JSON格式的额外设置
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
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
		rate, ok := new(big.Rat).SetString(strconv.FormatFloat(pm.FeeValue, 'f', -1, 64))
		if !ok || rate.Sign() < 0 {
			return domainmoney.Money{}, fmt.Errorf("invalid payment method fee percentage")
		}
		rate.Quo(rate, big.NewRat(100, 1))
		return amount.MultiplyRat(rate)
	}
	return domainmoney.FromMajorFloat(pm.FeeValue, amount.Currency().String())
}
