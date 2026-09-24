package coupon

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Coupon 优惠券
type Coupon struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Code string `gorm:"uniqueIndex;not null" json:"code"`
	Type string `gorm:"not null" json:"type"` // fixed, percentage
	// ValueRateDecimal is the canonical percentage value (for example
	// "12.5"). It is persisted as NUMERIC and is used for all calculations.
	ValueRateDecimal string `gorm:"column:value_rate_decimal;type:numeric(30,15);not null;default:0" json:"value_rate_decimal"`
	// ValueMinor is the canonical fixed-coupon amount. Percentage coupons use
	// ValueRateDecimal instead, so this field is zero for percentage coupons.
	ValueMinor int64 `gorm:"column:value_minor;not null;default:0" json:"value_minor"`
	// Currency is the currency in which the canonical minor-unit amounts and
	// decimal percentage rate are configured. Empty values are normalized to the
	// configured primary currency at the domain boundary.
	Currency    string `gorm:"size:3;not null;default:'USD'" json:"currency"`
	Description string `gorm:"type:text" json:"description"`
	MinAmountMinor       int64     `gorm:"column:min_amount_minor;not null;default:0" json:"min_amount_minor"`
	MaxDiscountMinor     int64     `gorm:"column:max_discount_minor;not null;default:0" json:"max_discount_minor"`
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
	if strings.EqualFold(strings.TrimSpace(c.Type), "percentage") {
		if strings.TrimSpace(c.ValueRateDecimal) == "" {
			return errors.New("coupon percentage value_rate_decimal is required")
		}
		rate, ok := new(big.Rat).SetString(strings.TrimSpace(c.ValueRateDecimal))
		if !ok || rate.Sign() <= 0 || rate.Cmp(big.NewRat(100, 1)) > 0 {
			return errors.New("coupon percentage must be between 0 and 100")
		}
	} else if strings.EqualFold(strings.TrimSpace(c.Type), "fixed") && c.ValueMinor <= 0 {
		return errors.New("coupon fixed value_minor must be greater than zero")
	}
	if c.ValueMinor < 0 || c.MinAmountMinor < 0 || c.MaxDiscountMinor < 0 {
		return errors.New("coupon monetary values cannot be negative")
	}
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
	minimum, err := domainmoney.New(c.MinAmountMinor, code)
	if err != nil {
		return domainmoney.Money{}, fmt.Errorf("invalid coupon minimum amount: %w", err)
	}
	if amount.AmountMinor() < minimum.AmountMinor() {
		return zero, nil
	}
	var discount domainmoney.Money
	if c.Type == "percentage" {
		rateText := strings.TrimSpace(c.ValueRateDecimal)
		if rateText == "" {
			return domainmoney.Money{}, errors.New("coupon value_rate_decimal is required")
		}
		rate, ok := new(big.Rat).SetString(rateText)
		if !ok || rate.Sign() < 0 {
			return domainmoney.Money{}, errors.New("invalid coupon percentage")
		}
		rate.Quo(rate, big.NewRat(100, 1))
		discount, err = amount.MultiplyRat(rate)
	} else {
		if c.ValueMinor <= 0 {
			return domainmoney.Money{}, errors.New("coupon fixed value_minor must be greater than zero")
		}
		discount, err = domainmoney.New(c.ValueMinor, code)
	}
	if err != nil {
		return domainmoney.Money{}, fmt.Errorf("calculate coupon discount: %w", err)
	}
	if c.MaxDiscountMinor > 0 {
		maxDiscount, maxErr := domainmoney.New(c.MaxDiscountMinor, code)
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
	ID       uint   `gorm:"primarykey" json:"id"`
	CouponID uint   `gorm:"not null;index;index:idx_coupon_usage_coupon_user" json:"coupon_id"`
	UserID   uint   `gorm:"not null;index;index:idx_coupon_usage_coupon_user" json:"user_id"`
	Email    string `gorm:"size:255;not null;default:'';index:idx_coupon_usage_coupon_email" json:"email"`
	OrderID  uint   `gorm:"not null;index" json:"order_id"`
	DiscountMinor  int64      `gorm:"column:discount_minor;not null;default:0" json:"discount_minor"`
	Currency       string     `gorm:"size:3;not null;default:'USD'" json:"currency"`
	Status         string     `gorm:"not null;default:applied;index" json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	ReversedAt     *time.Time `json:"reversed_at"`
	ReversalReason string     `gorm:"type:text" json:"reversal_reason"`
}

func (u *CouponUsage) BeforeSave(tx *gorm.DB) error {
	code := currency.NormalizeCode(u.Currency)
	if code == "" {
		code = currency.DefaultPrimaryCurrency
	}
	if !currency.IsCatalogCode(code) {
		return errors.New("coupon usage currency must be a supported catalog currency")
	}
	u.Currency = code
	if u.DiscountMinor < 0 {
		return errors.New("coupon usage discount cannot be negative")
	}
	return nil
}

func (u CouponUsage) DiscountMoney(code string) (domainmoney.Money, error) {
	code = currency.NormalizeCode(code)
	if code == "" {
		code = currency.NormalizeCode(u.Currency)
	}
	return domainmoney.New(u.DiscountMinor, code)
}

// TableName 指定表名
func (CouponUsage) TableName() string {
	return "coupon_usage"
}

// NormalizeEmail returns the canonical email identity used for guest coupon limits.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
