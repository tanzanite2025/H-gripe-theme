package order

import (
	"time"

	"gorm.io/gorm"
)

// Order 订单模型
type Order struct {
	ID             uint   `gorm:"primarykey" json:"id"`
	OrderNumber    string `gorm:"uniqueIndex;not null" json:"order_number"`
	UserID         uint   `gorm:"index" json:"user_id"`
	Status         string `gorm:"index;default:'pending'" json:"status"` // pending, paid, processing, shipped, completed, cancelled, refunded
	PaymentMethod  string `json:"payment_method"`
	PaymentStatus  string `gorm:"index;default:'unpaid'" json:"payment_status"` // unpaid, paid, refunded
	ShippingMethod string `json:"shipping_method"`
	ShippingStatus string `gorm:"index;default:'pending'" json:"shipping_status"` // pending, processing, shipped, delivered
	TrackingNumber string `json:"tracking_number"`
	CarrierCode    string `json:"carrier_code"`

	// 金额相关
	SubtotalAmount float64 `gorm:"not null" json:"subtotal_amount"`
	ShippingFee    float64 `gorm:"default:0" json:"shipping_fee"`
	TaxAmount      float64 `gorm:"default:0" json:"tax_amount"`
	DiscountAmount float64 `gorm:"default:0" json:"discount_amount"`
	TotalAmount    float64 `gorm:"not null" json:"total_amount"`

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

// OrderItem 订单商品项
type OrderItem struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	OrderID     uint      `gorm:"not null;index" json:"order_id"`
	ProductID   uint      `gorm:"not null;index" json:"product_id"`
	ProductName string    `gorm:"not null" json:"product_name"`
	SKU         string    `json:"sku"`
	Quantity    int       `gorm:"not null" json:"quantity"`
	Price       float64   `gorm:"not null" json:"price"`
	Subtotal    float64   `gorm:"not null" json:"subtotal"`
	TaxAmount   float64   `gorm:"default:0" json:"tax_amount"`
	Discount    float64   `gorm:"default:0" json:"discount"`
	Total       float64   `gorm:"not null" json:"total"`
	Attributes  string    `gorm:"type:text" json:"attributes"` // JSON格式的商品属性
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}

// BeforeCreate GORM钩子：创建前
func (o *Order) BeforeCreate(tx *gorm.DB) error {
	if o.OrderNumber == "" {
		o.OrderNumber = generateOrderNumber()
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
	return nil
}

// generateOrderNumber 生成订单号
func generateOrderNumber() string {
	// 格式: TZ + YYYYMMDD + 6位随机数
	now := time.Now()
	return now.Format("TZ20060102") + randomString(6)
}

func randomString(n int) string {
	const letters = "0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// OrderStatusTransition 订单状态流转规则
var OrderStatusTransition = map[string][]string{
	"pending":    {"cancelled"},
	"paid":       {"processing", "cancelled"},
	"processing": {"shipped", "cancelled"},
	"shipped":    {"completed", "cancelled"},
	"completed":  {},
	"cancelled":  {},
	"refunded":   {},
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

// OrderSummary 订单摘要（用于列表）
type OrderSummary struct {
	ID             uint       `json:"id"`
	OrderNumber    string     `json:"order_number"`
	UserID         uint       `json:"user_id"`
	Status         string     `json:"status"`
	PaymentStatus  string     `json:"payment_status"`
	ShippingStatus string     `json:"shipping_status"`
	TotalAmount    float64    `json:"total_amount"`
	ItemCount      int        `json:"item_count"`
	CustomerName   string     `json:"customer_name"`
	CreatedAt      time.Time  `json:"created_at"`
	PaidAt         *time.Time `json:"paid_at"`
}

// ToSummary 转换为摘要
func (o *Order) ToSummary() *OrderSummary {
	customerName := o.ShippingAddress.FirstName + " " + o.ShippingAddress.LastName
	return &OrderSummary{
		ID:             o.ID,
		OrderNumber:    o.OrderNumber,
		UserID:         o.UserID,
		Status:         o.Status,
		PaymentStatus:  o.PaymentStatus,
		ShippingStatus: o.ShippingStatus,
		TotalAmount:    o.TotalAmount,
		ItemCount:      len(o.Items),
		CustomerName:   customerName,
		CreatedAt:      o.CreatedAt,
		PaidAt:         o.PaidAt,
	}
}
