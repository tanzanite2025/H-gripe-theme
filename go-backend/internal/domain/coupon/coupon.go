package coupon

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Coupon 优惠券
type Coupon struct {
	ID    uint    `gorm:"primarykey" json:"id"`
	Code  string  `gorm:"uniqueIndex;not null" json:"code"`
	Type  string  `gorm:"not null" json:"type"` // fixed, percentage
	Value float64 `gorm:"not null" json:"value"`
	// Currency is the currency in which Value, MinAmount and MaxDiscount are
	// configured. Historically coupons were always entered in the primary
	// currency, so an empty value is treated as the primary currency by
	// checkout for backwards compatibility.
	Currency             string    `gorm:"size:3;not null;default:'USD'" json:"currency"`
	Description          string    `gorm:"type:text" json:"description"`
	MinAmount            float64   `gorm:"default:0" json:"min_amount"`
	MaxDiscount          float64   `gorm:"default:0" json:"max_discount"`
	UsageLimit           int       `gorm:"default:0;check:coupon_usage_limit_non_negative,usage_limit >= 0" json:"usage_limit"` // 0表示无限制
	UsageLimitPerUser    int       `gorm:"default:0" json:"usage_limit_per_user"`
	UsedCount            int       `gorm:"default:0;check:coupon_used_count_valid,used_count >= 0 AND (usage_limit = 0 OR used_count <= usage_limit)" json:"used_count"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	ApplicableProducts   string    `gorm:"type:text" json:"applicable_products"`   // JSON数组
	ExcludedProducts     string    `gorm:"type:text" json:"excluded_products"`     // JSON数组
	ApplicableCategories string    `gorm:"type:text" json:"applicable_categories"` // JSON数组
	// ReferralRecipientUserID makes a generated referral coupon private to the
	// referee. It is omitted from public responses and enforced at validation.
	ReferralRecipientUserID *uint          `gorm:"index" json:"-"`
	Enabled                 bool           `gorm:"default:true" json:"enabled"`
	CreatedAt               time.Time      `json:"created_at"`
	UpdatedAt               time.Time      `json:"updated_at"`
	DeletedAt               gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeSave normalizes and validates coupon monetary currency so malformed
// admin input cannot be persisted and later interpreted as another currency.
func (c *Coupon) BeforeSave(tx *gorm.DB) error {
	code := currency.NormalizeCode(c.Currency)
	if code == "" {
		code = currency.DefaultPrimaryCurrency
	}
	if !currency.IsCatalogCode(code) {
		return errors.New("coupon currency must be a supported catalog currency")
	}
	c.Currency = code
	return nil
}

// TableName 指定表名
func (Coupon) TableName() string {
	return "coupons"
}

// IsValid 检查优惠券是否有效
func (c *Coupon) IsValid() bool {
	now := time.Now()
	if !c.Enabled {
		return false
	}
	if now.Before(c.StartDate) || now.After(c.EndDate) {
		return false
	}
	if c.UsageLimit > 0 && c.UsedCount >= c.UsageLimit {
		return false
	}
	return true
}

// CalculateDiscountMoney computes a coupon discount in the amount's currency
// and minor units. An omitted coupon currency is resolved to the amount
// currency at this domain boundary; an explicit mismatch is rejected so FX
// conversion remains the responsibility of the caller.
func (c *Coupon) CalculateDiscountMoney(amount domainmoney.Money) (domainmoney.Money, error) {
	if c == nil {
		return domainmoney.Money{}, errors.New("coupon is required")
	}
	code := currency.NormalizeCode(c.Currency)
	if code == "" {
		code = amount.Currency().String()
	}
	if code != amount.Currency().String() {
		return domainmoney.Money{}, fmt.Errorf("coupon currency %s does not match amount currency %s", code, amount.Currency())
	}
	zero, err := domainmoney.New(0, code)
	if err != nil {
		return domainmoney.Money{}, err
	}
	minimum, err := domainmoney.FromMajorFloat(c.MinAmount, code)
	if err != nil {
		return domainmoney.Money{}, fmt.Errorf("invalid coupon minimum amount: %w", err)
	}
	if amount.AmountMinor() < minimum.AmountMinor() {
		return zero, nil
	}
	var discount domainmoney.Money
	if c.Type == "percentage" {
		rate, ok := new(big.Rat).SetString(strconv.FormatFloat(c.Value, 'f', -1, 64))
		if !ok || rate.Sign() < 0 {
			return domainmoney.Money{}, errors.New("invalid coupon percentage")
		}
		rate.Quo(rate, big.NewRat(100, 1))
		discount, err = amount.MultiplyRat(rate)
	} else {
		discount, err = domainmoney.FromMajorFloat(c.Value, code)
	}
	if err != nil {
		return domainmoney.Money{}, fmt.Errorf("calculate coupon discount: %w", err)
	}
	if c.MaxDiscount > 0 {
		maxDiscount, maxErr := domainmoney.FromMajorFloat(c.MaxDiscount, code)
		if maxErr != nil {
			return domainmoney.Money{}, fmt.Errorf("invalid coupon max discount: %w", maxErr)
		}
		if discount.AmountMinor() > maxDiscount.AmountMinor() {
			discount = maxDiscount
		}
	}
	if discount.AmountMinor() > amount.AmountMinor() {
		discount = amount
	}
	return discount, nil
}

// CouponUsage 优惠券使用记录
const (
	CouponUsageStatusApplied  = "applied"
	CouponUsageStatusReversed = "reversed"
)

type CouponUsage struct {
	ID             uint       `gorm:"primarykey" json:"id"`
	CouponID       uint       `gorm:"not null;index;index:idx_coupon_usage_coupon_user" json:"coupon_id"`
	UserID         uint       `gorm:"not null;index;index:idx_coupon_usage_coupon_user" json:"user_id"`
	Email          string     `gorm:"size:255;not null;default:'';index:idx_coupon_usage_coupon_email" json:"email"`
	OrderID        uint       `gorm:"not null;index" json:"order_id"`
	Discount       float64    `gorm:"not null" json:"discount"`
	Status         string     `gorm:"not null;default:applied;index" json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	ReversedAt     *time.Time `json:"reversed_at"`
	ReversalReason string     `gorm:"type:text" json:"reversal_reason"`
}

// TableName 指定表名
func (CouponUsage) TableName() string {
	return "coupon_usage"
}

// NormalizeEmail returns the canonical email identity used for guest coupon limits.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
