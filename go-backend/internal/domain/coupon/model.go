package coupon

import (
	"time"

	"gorm.io/gorm"
)

// Coupon 优惠券
type Coupon struct {
	ID                uint           `gorm:"primarykey" json:"id"`
	Code              string         `gorm:"uniqueIndex;not null" json:"code"`
	Type              string         `gorm:"not null" json:"type"` // fixed, percentage
	Value             float64        `gorm:"not null" json:"value"`
	Description       string         `gorm:"type:text" json:"description"`
	MinAmount         float64        `gorm:"default:0" json:"min_amount"`
	MaxDiscount       float64        `gorm:"default:0" json:"max_discount"`
	UsageLimit        int            `gorm:"default:0" json:"usage_limit"` // 0表示无限制
	UsageLimitPerUser int            `gorm:"default:0" json:"usage_limit_per_user"`
	UsedCount         int            `gorm:"default:0" json:"used_count"`
	StartDate         time.Time      `json:"start_date"`
	EndDate           time.Time      `json:"end_date"`
	ApplicableProducts string        `gorm:"type:text" json:"applicable_products"` // JSON数组
	ExcludedProducts   string        `gorm:"type:text" json:"excluded_products"`   // JSON数组
	ApplicableCategories string      `gorm:"type:text" json:"applicable_categories"` // JSON数组
	Enabled           bool           `gorm:"default:true" json:"enabled"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
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
	
	return discount
}

// CouponUsage 优惠券使用记录
type CouponUsage struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	CouponID  uint      `gorm:"not null;index" json:"coupon_id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	OrderID   uint      `gorm:"not null;index" json:"order_id"`
	Discount  float64   `gorm:"not null" json:"discount"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 指定表名
func (CouponUsage) TableName() string {
	return "coupon_usage"
}

// GiftCard 礼品卡
type GiftCard struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Code          string         `gorm:"uniqueIndex;not null" json:"code"`
	InitialValue  float64        `gorm:"not null" json:"initial_value"`
	Balance       float64        `gorm:"not null" json:"balance"`
	Currency      string         `gorm:"default:'USD'" json:"currency"`
	Status        string         `gorm:"index" json:"status"` // active, used, expired, cancelled
	RecipientEmail string        `json:"recipient_email"`
	RecipientName  string        `json:"recipient_name"`
	SenderName     string        `json:"sender_name"`
	Message        string        `gorm:"type:text" json:"message"`
	CoverImage     string        `json:"cover_image"`
	ExpiresAt      *time.Time    `json:"expires_at"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (GiftCard) TableName() string {
	return "gift_cards"
}

// IsValid 检查礼品卡是否有效
func (gc *GiftCard) IsValid() bool {
	if gc.Status != "active" {
		return false
	}
	if gc.Balance <= 0 {
		return false
	}
	if gc.ExpiresAt != nil && time.Now().After(*gc.ExpiresAt) {
		return false
	}
	return true
}

// GiftCardTransaction 礼品卡交易记录
type GiftCardTransaction struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	GiftCardID uint      `gorm:"not null;index" json:"gift_card_id"`
	OrderID    uint      `gorm:"index" json:"order_id"`
	Type       string    `gorm:"not null" json:"type"` // issue, use, refund
	Amount     float64   `gorm:"not null" json:"amount"`
	Balance    float64   `gorm:"not null" json:"balance"` // 交易后余额
	Note       string    `gorm:"type:text" json:"note"`
	CreatedAt  time.Time `json:"created_at"`
}

// TableName 指定表名
func (GiftCardTransaction) TableName() string {
	return "gift_card_transactions"
}
