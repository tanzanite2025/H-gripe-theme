package order

import (
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	FulfillmentModeStock       = "stock"
	FulfillmentModeMadeToOrder = "made_to_order"
	FulfillmentModeMixed       = "mixed"

	HighValueSignatureThresholdUSD = 750

	ProductionStatusNotApplicable = "not_applicable"
	ProductionStatusNotStarted    = "not_started"
	ProductionStatusStarted       = "started"
	ProductionStatusCompleted     = "completed"
	ProductionStatusCancelled     = "cancelled"
)

func NormalizeFulfillmentMode(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return FulfillmentModeStock
	}
	return value
}

func IsValidFulfillmentMode(value string) bool {
	switch NormalizeFulfillmentMode(value) {
	case FulfillmentModeStock, FulfillmentModeMadeToOrder, FulfillmentModeMixed:
		return true
	default:
		return false
	}
}

func NormalizeProductionStatus(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ProductionStatusNotApplicable
	}
	return value
}

func IsValidProductionStatus(value string) bool {
	switch NormalizeProductionStatus(value) {
	case ProductionStatusNotApplicable,
		ProductionStatusNotStarted,
		ProductionStatusStarted,
		ProductionStatusCompleted,
		ProductionStatusCancelled:
		return true
	default:
		return false
	}
}

func ResolveFulfillmentMode(items []OrderItem) string {
	hasStock := false
	hasMadeToOrder := false
	for _, item := range items {
		switch NormalizeFulfillmentMode(item.FulfillmentMode) {
		case FulfillmentModeMadeToOrder:
			hasMadeToOrder = true
		default:
			hasStock = true
		}
	}

	switch {
	case hasStock && hasMadeToOrder:
		return FulfillmentModeMixed
	case hasMadeToOrder:
		return FulfillmentModeMadeToOrder
	default:
		return FulfillmentModeStock
	}
}

func DefaultProductionStatus(fulfillmentMode string) string {
	mode := NormalizeFulfillmentMode(fulfillmentMode)
	if mode == FulfillmentModeMadeToOrder || mode == FulfillmentModeMixed {
		return ProductionStatusNotStarted
	}
	return ProductionStatusNotApplicable
}

// ResolveSignatureRequired evaluates the high-value shipping policy against
// the immutable FX snapshot captured during checkout. The policy is defined
// in USD, so an order whose snapshot uses another base currency is left
// unmarked until a USD-based order policy is available.
func ResolveSignatureRequired(totalAmount float64, fxSnapshot currency.OrderFXSnapshot) bool {
	evaluation, err := EvaluateHighValueOrder(totalAmount, fxSnapshot)
	return err == nil && evaluation.IsHighValue
}

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
	FulfillmentMode          string `gorm:"size:20;not null;default:'stock';index" json:"fulfillment_mode"`
	ProductionStatus         string `gorm:"size:20;not null;default:'not_applicable';index" json:"production_status"`
	SignatureRequired        bool   `gorm:"not null;default:false;index" json:"signature_required"`
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
	CouponCode     string         `json:"coupon_code"`
	PointsUsed     int            `gorm:"default:0" json:"points_used"`
	PointsValue    float64        `gorm:"default:0" json:"points_value"`
	FXSnapshotData datatypes.JSON `gorm:"column:fx_snapshot;type:jsonb;not null;default:'{}'" json:"-"`

	// 地址信息
	ShippingAddress Address `gorm:"embedded;embeddedPrefix:shipping_" json:"shipping_address"`
	BillingAddress  Address `gorm:"embedded;embeddedPrefix:billing_" json:"billing_address"`

	// 备注
	CustomerNote string `gorm:"type:text" json:"customer_note"`
	AdminNote    string `gorm:"type:text" json:"admin_note"`

	// 关联
	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items"`

	// 时间戳
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
	PaidAt                *time.Time     `json:"paid_at"`
	ShippedAt             *time.Time     `json:"shipped_at"`
	CompletedAt           *time.Time     `json:"completed_at"`
	CancelledAt           *time.Time     `json:"cancelled_at"`
	ProductionStartedAt   *time.Time     `json:"production_started_at"`
	ProductionCompletedAt *time.Time     `json:"production_completed_at"`
	DeletedAt             gorm.DeletedAt `gorm:"index" json:"-"`
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
	o.FulfillmentMode = NormalizeFulfillmentMode(o.FulfillmentMode)
	if !IsValidFulfillmentMode(o.FulfillmentMode) {
		return errors.New("order fulfillment mode is invalid")
	}
	if o.ProductionStatus == "" {
		o.ProductionStatus = DefaultProductionStatus(o.FulfillmentMode)
	} else {
		o.ProductionStatus = NormalizeProductionStatus(o.ProductionStatus)
	}
	if !IsValidProductionStatus(o.ProductionStatus) {
		return errors.New("order production status is invalid")
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
