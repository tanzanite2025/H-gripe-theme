package order

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/currency"

	"gorm.io/gorm"
)

// Order 订单模型
type Order struct {
	ID                       uint   `gorm:"primarykey" json:"id"`
	OrderNumber              string `gorm:"uniqueIndex;not null" json:"order_number"`
	UserID                   uint   `gorm:"index" json:"user_id"`
	Status                   string `gorm:"index;default:'pending'" json:"status"` // pending, paid, processing, shipped, completed, cancelled, payment_expired, refunded
	PaymentMethod            string `json:"payment_method"`
	PaymentStatus            string `gorm:"index;default:'unpaid'" json:"payment_status"` // unpaid, paid, expired, refunded
	ShippingMethod           string `json:"shipping_method"`
	ShippingStatus           string `gorm:"index;default:'pending'" json:"shipping_status"` // pending, processing, shipped, delivered
	TrackingNumber           string `json:"tracking_number"`
	TrackingProviderID       *uint  `gorm:"index" json:"tracking_provider_id"`
	CarrierID                *uint  `gorm:"index" json:"carrier_id"`
	CarrierServiceID         *uint  `gorm:"index" json:"carrier_service_id"`
	TrackingCarrierMappingID *uint  `gorm:"index" json:"tracking_carrier_mapping_id"`
	ProviderCarrierCode      string `json:"provider_carrier_code"`
	ProviderCarrierName      string `json:"provider_carrier_name"`

	// 金额相关
	SubtotalAmount float64 `gorm:"not null" json:"subtotal_amount"`
	ShippingFee    float64 `gorm:"default:0" json:"shipping_fee"`
	TaxAmount      float64 `gorm:"default:0" json:"tax_amount"`
	DiscountAmount float64 `gorm:"default:0" json:"discount_amount"`
	TotalAmount    float64 `gorm:"not null" json:"total_amount"`
	Currency       string  `gorm:"not null;index" json:"currency"`

	// 优惠信息
	CouponCode  string  `json:"coupon_code"`
	PointsUsed  int     `gorm:"default:0" json:"points_used"`
	PointsValue float64 `gorm:"default:0" json:"points_value"`

	// 地址信息
	ShippingAddress Address `gorm:"embedded;embeddedPrefix:shipping_" json:"shipping_address"`
	BillingAddress  Address `gorm:"embedded;embeddedPrefix:billing_" json:"billing_address"`

	// 备注
	CustomerNote string `gorm:"type:text" json:"customer_note"`
	AdminNote    string `gorm:"type:text" json:"admin_note"`

	// 关联
	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items"`

	// 时间戳
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	PaidAt      *time.Time     `json:"paid_at"`
	ShippedAt   *time.Time     `json:"shipped_at"`
	CompletedAt *time.Time     `json:"completed_at"`
	CancelledAt *time.Time     `json:"cancelled_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Address 地址结构
type Address struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Company    string `json:"company"`
	Address1   string `json:"address_1"`
	Address2   string `json:"address_2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
}

// TrackingInfoUpdate is the normalized logistics payload stored on an order.
// Local carrier/service IDs are the editable source; provider carrier code is a resolved snapshot.
type TrackingInfoUpdate struct {
	TrackingNumber           string
	TrackingProviderID       *uint
	CarrierID                *uint
	CarrierServiceID         *uint
	TrackingCarrierMappingID *uint
	ProviderCarrierCode      string
	ProviderCarrierName      string
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// BeforeCreate GORM钩子：创建前
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.OrderNumber == "" {
		return errors.New("order number is required")
	}
	if o.Status == "" {
		o.Status = "pending"
	}
	if o.PaymentStatus == "" {
		o.PaymentStatus = "unpaid"
	}
	if o.ShippingStatus == "" {
		o.ShippingStatus = "pending"
	}
	o.Currency = currency.NormalizeCode(o.Currency)
	if !currency.IsValidCode(o.Currency) || !currency.IsCatalogCode(o.Currency) {
		return errors.New("order currency must be a supported ISO 4217 code")
	}
	return nil
}

// OrderStatusTransition 订单状态流转规则
var OrderStatusTransition = map[string][]string{
	"pending":         {"cancelled", "payment_expired"},
	"paid":            {"processing", "cancelled"},
	"processing":      {"shipped", "cancelled"},
	"shipped":         {"completed", "cancelled"},
	"completed":       {},
	"cancelled":       {},
	"payment_expired": {},
	"refunded":        {},
}

// CanTransitionTo 检查是否可以转换到目标状态
func (o *Order) CanTransitionTo(targetStatus string) bool {
	allowedStatuses, exists := OrderStatusTransition[o.Status]
	if !exists {
		return false
	}
	for _, status := range allowedStatuses {
		if status == targetStatus {
			return true
		}
	}
	return false
}
