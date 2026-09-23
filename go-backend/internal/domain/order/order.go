package order

import (
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/money"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	FulfillmentModeStock       = "stock"
	FulfillmentModeMadeToOrder = "made_to_order"
	FulfillmentModeMixed       = "mixed"

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
func ResolveSignatureRequired(totalAmount money.Money, fxSnapshot currency.OrderFXSnapshot) bool {
	evaluation, err := EvaluateHighValueOrder(totalAmount, fxSnapshot)
	return err == nil && evaluation.IsHighValue
}

// Order 订单模型
type Order struct {
	ID            uint   `gorm:"primarykey" json:"id"`
	OrderNumber   string `gorm:"uniqueIndex;not null" json:"order_number"`
	UserID        uint   `gorm:"index" json:"user_id"`
	Status        string `gorm:"index;default:'pending'" json:"status"` // pending, paid, processing, needs_review, shipped, completed, cancelled, payment_expired, refunded, disputed
	PaymentMethod string `json:"payment_method"`
	PaymentStatus string `gorm:"index;default:'unpaid'" json:"payment_status"` // unpaid, paid, expired, refunded
	// Provider-settlement payable facts may differ from the storefront order
	// currency for domestic channels such as Alipay and WeChat Pay.
	PaymentCurrency    string `gorm:"size:3;index" json:"payment_currency"`
	PaymentAmountMinor int64  `gorm:"column:payment_amount_minor;not null;default:0" json:"payment_amount_minor"`
	ShippingMethod        string  `json:"shipping_method"`
	ShippingStatus        string  `gorm:"index;default:'pending'" json:"shipping_status"` // pending, processing, shipped, delivered
	FulfillmentHold       bool    `gorm:"not null;default:false;index" json:"fulfillment_hold"`
	DisputePreviousStatus string  `gorm:"size:32" json:"-"`
	DisputePreviousHold   bool    `gorm:"not null;default:false" json:"-"`
	FulfillmentMode       string  `gorm:"size:20;not null;default:'stock';index" json:"fulfillment_mode"`
	ProductionStatus      string  `gorm:"size:20;not null;default:'not_applicable';index" json:"production_status"`
	SignatureRequired     bool    `gorm:"not null;default:false;index" json:"signature_required"`
	ShippingQuoteID       string  `gorm:"type:varchar(36);index" json:"shipping_quote_id,omitempty"`
	ShippingQuotePlanID   string  `gorm:"type:varchar(36)" json:"shipping_quote_plan_id,omitempty"`
	// CheckoutCartID records the cart consumed to create this order so an
	// unpaid cancellation or payment expiration can restore its contents.
	CheckoutCartID *uint `gorm:"index" json:"-"`

	// 金额相关
	SubtotalAmountMinor int64 `gorm:"column:subtotal_amount_minor;not null;default:0" json:"subtotal_amount_minor"`
	ShippingFeeMinor    int64 `gorm:"column:shipping_fee_minor;not null;default:0" json:"shipping_fee_minor"`
	TaxAmountMinor      int64 `gorm:"column:tax_amount_minor;not null;default:0" json:"tax_amount_minor"`
	DiscountAmountMinor int64 `gorm:"column:discount_amount_minor;not null;default:0" json:"discount_amount_minor"`
	TotalAmountMinor    int64 `gorm:"column:total_amount_minor;not null;default:0" json:"total_amount_minor"`
	PointsValueMinor    int64 `gorm:"column:points_value_minor;not null;default:0" json:"points_value_minor"`
	Currency       string  `gorm:"not null;index" json:"currency"`

	// 优惠信息
	CouponCode               string         `json:"coupon_code"`
	PointsUsed               int            `gorm:"default:0" json:"points_used"`
	FXSnapshotData           datatypes.JSON `gorm:"column:fx_snapshot;type:jsonb;not null;default:'{}'" json:"-"`
	ShippingPlanSnapshotData datatypes.JSON `gorm:"column:shipping_plan_snapshot;type:jsonb;not null;default:'{}'" json:"shipping_plan_snapshot,omitempty"`
	PricingSnapshotData      datatypes.JSON `gorm:"column:pricing_snapshot;type:jsonb;not null;default:'{}'" json:"-"`

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
	DeliveredAt           *time.Time     `json:"delivered_at"`
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
	if err := o.validateMinorAmounts(); err != nil {
		return err
	}
	return nil
}

// BeforeSave keeps update paths subject to the same exact-money invariants as
// inserts. Major-unit projection fields are ignored by persistence and cannot
// be used to repair an invalid minor snapshot.
func (o *Order) BeforeSave(tx *gorm.DB) error {
	// GORM partial Updates hydrate only the changed columns; leave financial
	// validation to BeforeCreate/full Save when the currency is absent.
	if strings.TrimSpace(o.Currency) == "" {
		return nil
	}
	o.Currency = currency.NormalizeCode(o.Currency)
	if !currency.IsValidCode(o.Currency) || !currency.IsCatalogCode(o.Currency) {
		return errors.New("order currency must be a supported ISO 4217 code")
	}
	if err := o.validateMinorAmounts(); err != nil {
		return err
	}
	return nil
}

func (o *Order) validateMinorAmounts() error {
	for _, amount := range []int64{o.SubtotalAmountMinor, o.ShippingFeeMinor, o.TaxAmountMinor, o.DiscountAmountMinor, o.TotalAmountMinor, o.PointsValueMinor} {
		if amount < 0 {
			return errors.New("order monetary amounts cannot be negative")
		}
	}
	paymentCurrency := currency.NormalizeCode(o.PaymentCurrency)
	if paymentCurrency == "" {
		paymentCurrency = o.Currency
	}
	if !currency.IsCatalogCode(paymentCurrency) {
		return errors.New("order payment currency must be a supported ISO 4217 code")
	}
	if o.PaymentAmountMinor < 0 {
		return errors.New("order payment amount cannot be negative")
	}
	return nil
}

func (o Order) SubtotalMoney() (money.Money, error) {
	return money.New(o.SubtotalAmountMinor, o.Currency)
}

func (o Order) ShippingFeeMoney() (money.Money, error) {
	return money.New(o.ShippingFeeMinor, o.Currency)
}

func (o Order) TaxMoney() (money.Money, error) {
	return money.New(o.TaxAmountMinor, o.Currency)
}

func (o Order) DiscountMoney() (money.Money, error) {
	return money.New(o.DiscountAmountMinor, o.Currency)
}

func (o Order) TotalMoney() (money.Money, error) {
	return money.New(o.TotalAmountMinor, o.Currency)
}

func (o Order) PaymentMoney() (money.Money, error) {
	code := o.PaymentCurrency
	if code == "" {
		code = o.Currency
	}
	return money.New(o.PaymentAmountMinor, code)
}

// OrderStatusTransition 订单状态流转规则
var OrderStatusTransition = map[string][]string{
	"pending":         {"cancelled", "payment_expired"},
	"paid":            {"processing", "cancelled"},
	"processing":      {"shipped", "cancelled"},
	"needs_review":    {},
	"shipped":         {"completed", "cancelled"},
	"completed":       {},
	"cancelled":       {},
	"payment_expired": {},
	"refunded":        {},
	"disputed":        {},
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
