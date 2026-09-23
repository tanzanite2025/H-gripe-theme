package loyalty

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"
)

// MemberLevel 会员等级
type MemberLevel struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	Name      string `gorm:"not null" json:"name"`
	MinPoints int    `gorm:"not null" json:"min_points"`
	MaxPoints int    `gorm:"not null" json:"max_points"`
	// DiscountRateDecimal is the exact percentage discount (for example "5.5"
	// means a 5.5% discount). It is the sole persisted pricing value.
	DiscountRateDecimal string         `gorm:"column:discount_rate_decimal;type:numeric(30,15);not null;default:0" json:"discount_rate_decimal"`
	Benefits            string         `gorm:"type:text" json:"benefits"` // JSON格式的权益说明
	Icon                string         `json:"icon"`
	Color               string         `json:"color"`
	SortOrder           int            `gorm:"default:0" json:"sort_order"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (MemberLevel) TableName() string {
	return "member_levels"
}

func (l *MemberLevel) BeforeSave(tx *gorm.DB) error {
	if l == nil {
		return fmt.Errorf("member level is required")
	}
	value := strings.TrimSpace(l.DiscountRateDecimal)
	if value == "" {
		value = "0"
	}
	if strings.Contains(value, "/") {
		return fmt.Errorf("member level discount rate must be a decimal")
	}
	rate, ok := new(big.Rat).SetString(value)
	if !ok || rate.Sign() < 0 || rate.Cmp(big.NewRat(100, 1)) > 0 {
		return fmt.Errorf("member level discount rate must be between 0 and 100")
	}
	// Keep the in-memory invariant aligned with NUMERIC(30,15); values with a
	// finer scale would be rounded by PostgreSQL and must be rejected up front.
	scaled := new(big.Rat).Mul(rate, new(big.Rat).SetInt64(1_000_000_000_000_000))
	if scaled.Denom().Cmp(big.NewInt(1)) != 0 {
		return fmt.Errorf("member level discount rate supports at most 15 fractional digits")
	}
	l.DiscountRateDecimal = value
	return nil
}
