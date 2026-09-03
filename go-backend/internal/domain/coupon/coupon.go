package coupon

import (
	"time"

	"gorm.io/gorm"
)

// Coupon 优惠券
type Coupon struct {
	ID                   uint           `gorm:"primarykey" json:"id"`
	Code                 string         `gorm:"uniqueIndex;not null" json:"code"`
	Type                 string         `gorm:"not null" json:"type"` // fixed, percentage
	Value                float64        `gorm:"not null" json:"value"`
	Description          string         `gorm:"type:text" json:"description"`
	MinAmount            float64        `gorm:"default:0" json:"min_amount"`
	MaxDiscount          float64        `gorm:"default:0" json:"max_discount"`
	UsageLimit           int            `gorm:"default:0;check:coupon_usage_limit_non_negative,usage_limit >= 0" json:"usage_limit"` // 0表示无限制
	UsageLimitPerUser    int            `gorm:"default:0" json:"usage_limit_per_user"`
	UsedCount            int            `gorm:"default:0;check:coupon_used_count_valid,used_count >= 0 AND (usage_limit = 0 OR used_count <= usage_limit)" json:"used_count"`
	StartDate            time.Time      `json:"start_date"`
	EndDate              time.Time      `json:"end_date"`
	ApplicableProducts   string         `gorm:"type:text" json:"applicable_products"`   // JSON数组
	ExcludedProducts     string         `gorm:"type:text" json:"excluded_products"`     // JSON数组
	ApplicableCategories string         `gorm:"type:text" json:"applicable_categories"` // JSON数组
	Enabled              bool           `gorm:"default:true" json:"enabled"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
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

// CalculateDiscount 计算折扣金额
func (c *Coupon) CalculateDiscount(amount float64) float64 {
	if amount < c.MinAmount {
		return 0
	}

	var discount float64
	if c.Type == "percentage" {
		discount = amount * c.Value / 100
	} else {
		discount = c.Value
	}

	if c.MaxDiscount > 0 && discount > c.MaxDiscount {
		discount = c.MaxDiscount
	}
	if discount > amount {
		discount = amount
	}

	return discount
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
