package order

import (
	"errors"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OrderItem 订单商品项
type OrderItem struct {
	ID        uint `gorm:"primarykey" json:"id"`
	OrderID   uint `gorm:"not null;index" json:"order_id"`
	ProductID uint `gorm:"not null;index" json:"product_id"`
	// ProductCategoryID is captured at checkout so coupon category restrictions
	// can be evaluated against the immutable order line. It is nullable for
	// legacy orders created before category-aware coupon validation.
	ProductCategoryID      *uint  `gorm:"index" json:"-"`
	ProductCategorySlug    string `gorm:"size:120" json:"-"`
	VariantID              *uint  `gorm:"not null;index" json:"variant_id"`
	ProductName            string `gorm:"not null" json:"product_name"`
	SKU                    string `json:"sku"`
	Quantity               int    `gorm:"not null" json:"quantity"`
	Currency               string `gorm:"size:3;not null;default:'USD'" json:"currency"`
	PriceMinor             int64  `gorm:"column:price_minor;not null;default:0" json:"price_minor"`
	SubtotalMinor          int64  `gorm:"column:subtotal_minor;not null;default:0" json:"subtotal_minor"`
	TaxAmountMinor         int64  `gorm:"column:tax_amount_minor;not null;default:0" json:"tax_amount_minor"`
	DiscountMinor          int64  `gorm:"column:discount_minor;not null;default:0" json:"discount_minor"`
	TotalMinor             int64  `gorm:"column:total_minor;not null;default:0" json:"total_minor"`
	WeightGrams            int    `gorm:"column:weight_grams;not null;default:0" json:"weight_grams"`
	FulfillmentMode        string `gorm:"size:20;not null;default:'stock';index" json:"fulfillment_mode"`
	HSCode                 string `gorm:"column:hs_code;size:12" json:"hs_code"`
	CNCode                 string `gorm:"column:cn_code;size:12" json:"cn_code"`
	CountryOfOrigin        string `gorm:"column:country_of_origin;size:2" json:"country_of_origin"`
	CustomsDescription     string `gorm:"column:customs_description;size:255" json:"customs_description"`
	DeclaredValueMinor     *int64 `gorm:"column:declared_value_minor" json:"declared_value_minor"`
	DeclaredValueConfirmed bool   `gorm:"column:declared_value_confirmed;not null;default:false" json:"declared_value_confirmed"`
	// PricingSnapshotData is the immutable order-time line pricing contract.
	// It is the source for order-time discount, tax, and net amount facts.
	PricingSnapshotData       datatypes.JSON `gorm:"column:pricing_snapshot;type:jsonb;not null;default:'{}'" json:"-"`
	ConfigurationSnapshotData datatypes.JSON `gorm:"column:configuration_snapshot;type:jsonb;not null;default:'{}'" json:"configuration_snapshot,omitempty"`
	// ConfigurationData is carried from a cart/checkout request and is never
	// persisted directly; the service writes ConfigurationSnapshotData after
	// revalidating and normalizing it against the current catalog.
	ConfigurationData datatypes.JSON `gorm:"-" json:"-"`
	ConfigurationHash string         `gorm:"-" json:"-"`
	CreatedAt         time.Time      `json:"created_at"`
}

// TableName 指定表名
func (OrderItem) TableName() string {
	return "order_items"
}

func (i *OrderItem) BeforeCreate(tx *gorm.DB) error {
	i.Currency = currency.NormalizeCode(i.Currency)
	if i.Currency == "" {
		i.Currency = currency.DefaultPrimaryCurrency
	}
	if !currency.IsCatalogCode(i.Currency) {
		return errors.New("order item currency must be a supported ISO 4217 code")
	}
	for _, amount := range []int64{i.PriceMinor, i.SubtotalMinor, i.TaxAmountMinor, i.DiscountMinor, i.TotalMinor} {
		if amount < 0 {
			return errors.New("order item monetary amounts cannot be negative")
		}
	}
	if i.DeclaredValueMinor != nil && *i.DeclaredValueMinor < 0 {
		return errors.New("order item declared value cannot be negative")
	}
	return nil
}

// BeforeSave applies the exact minor-unit invariant to updates as well as
// inserts.
func (i *OrderItem) BeforeSave(tx *gorm.DB) error {
	i.Currency = currency.NormalizeCode(i.Currency)
	if i.Currency == "" {
		i.Currency = currency.DefaultPrimaryCurrency
	}
	if !currency.IsCatalogCode(i.Currency) {
		return errors.New("order item currency must be a supported ISO 4217 code")
	}
	for _, amount := range []int64{i.PriceMinor, i.SubtotalMinor, i.TaxAmountMinor, i.DiscountMinor, i.TotalMinor} {
		if amount < 0 {
			return errors.New("order item monetary amounts cannot be negative")
		}
	}
	if i.DeclaredValueMinor != nil && *i.DeclaredValueMinor < 0 {
		return errors.New("order item declared value cannot be negative")
	}
	return nil
}

func (i OrderItem) PriceMoney() (domainmoney.Money, error) {
	return domainmoney.New(i.PriceMinor, i.Currency)
}

func (i OrderItem) SubtotalMoney() (domainmoney.Money, error) {
	return domainmoney.New(i.SubtotalMinor, i.Currency)
}

func (i OrderItem) TaxAmountMoney() (domainmoney.Money, error) {
	return domainmoney.New(i.TaxAmountMinor, i.Currency)
}

func (i OrderItem) DiscountMoney() (domainmoney.Money, error) {
	return domainmoney.New(i.DiscountMinor, i.Currency)
}

func (i OrderItem) TotalMoney() (domainmoney.Money, error) {
	return domainmoney.New(i.TotalMinor, i.Currency)
}

// ConfigurationEvidenceJSON returns the immutable order-time configuration
// snapshot. Configuration snapshots are the sole source of order-line
// configuration evidence.
func (i OrderItem) ConfigurationEvidenceJSON() string {
	raw := strings.TrimSpace(string(i.ConfigurationSnapshotData))
	if raw == "" {
		return "{}"
	}
	return raw
}
