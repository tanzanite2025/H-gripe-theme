package order

import (
	"errors"
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
	ProductCategoryID   *uint  `gorm:"index" json:"-"`
	ProductCategorySlug string `gorm:"size:120" json:"-"`
	VariantID           *uint  `gorm:"not null;index" json:"variant_id"`
	ProductName         string `gorm:"not null" json:"product_name"`
	SKU                 string `json:"sku"`
	Quantity            int    `gorm:"not null" json:"quantity"`
	Currency            string `gorm:"size:3;not null;default:'USD'" json:"currency"`
	PriceMinor          int64  `gorm:"column:price_minor;not null;default:0" json:"price_minor"`
	SubtotalMinor       int64  `gorm:"column:subtotal_minor;not null;default:0" json:"subtotal_minor"`
	TaxAmountMinor      int64  `gorm:"column:tax_amount_minor;not null;default:0" json:"tax_amount_minor"`
	DiscountMinor       int64  `gorm:"column:discount_minor;not null;default:0" json:"discount_minor"`
	TotalMinor          int64  `gorm:"column:total_minor;not null;default:0" json:"total_minor"`
	// Legacy major-unit columns are retained only for historical rows while
	// callers migrate to the exact Money accessors below.
	Price                  float64  `gorm:"not null" json:"price"`
	Subtotal               float64  `gorm:"not null" json:"subtotal"`
	TaxAmount              float64  `gorm:"default:0" json:"tax_amount"`
	Discount               float64  `gorm:"default:0" json:"discount"`
	Total                  float64  `gorm:"not null" json:"total"`
	Attributes             string   `gorm:"type:text" json:"attributes"` // JSON格式的商品属性
	WeightGrams            int      `gorm:"column:weight_grams;not null;default:0" json:"weight_grams"`
	FulfillmentMode        string   `gorm:"size:20;not null;default:'stock';index" json:"fulfillment_mode"`
	HSCode                 string   `gorm:"column:hs_code;size:12" json:"hs_code"`
	CNCode                 string   `gorm:"column:cn_code;size:12" json:"cn_code"`
	CountryOfOrigin        string   `gorm:"column:country_of_origin;size:2" json:"country_of_origin"`
	CustomsDescription     string   `gorm:"column:customs_description;size:255" json:"customs_description"`
	DeclaredValue          *float64 `gorm:"column:declared_value;type:numeric(12,2)" json:"declared_value"`
	DeclaredValueConfirmed bool     `gorm:"column:declared_value_confirmed;not null;default:false" json:"declared_value_confirmed"`
	// PricingSnapshotData is the immutable order-time line pricing contract.
	// Legacy float columns remain for compatibility during the migration.
	PricingSnapshotData       datatypes.JSON `gorm:"column:pricing_snapshot;type:jsonb;not null;default:'{}'" json:"-"`
	ConfigurationSnapshotData datatypes.JSON `gorm:"column:configuration_snapshot;type:jsonb;not null;default:'{}'" json:"-"`
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
	fields := []struct {
		major float64
		minor *int64
	}{
		{major: i.Price, minor: &i.PriceMinor},
		{major: i.Subtotal, minor: &i.SubtotalMinor},
		{major: i.TaxAmount, minor: &i.TaxAmountMinor},
		{major: i.Discount, minor: &i.DiscountMinor},
		{major: i.Total, minor: &i.TotalMinor},
	}
	for _, field := range fields {
		if *field.minor == 0 && field.major != 0 {
			value, err := domainmoney.FromMajorFloat(field.major, i.Currency)
			if err != nil {
				return err
			}
			*field.minor = value.AmountMinor()
		}
		if *field.minor < 0 {
			return errors.New("order item monetary amounts cannot be negative")
		}
	}
	return nil
}

func (i OrderItem) PriceMoney() (domainmoney.Money, error) {
	if i.PriceMinor == 0 && i.Price != 0 {
		return domainmoney.FromMajorFloat(i.Price, i.Currency)
	}
	return domainmoney.New(i.PriceMinor, i.Currency)
}

func (i OrderItem) SubtotalMoney() (domainmoney.Money, error) {
	if i.SubtotalMinor == 0 && i.Subtotal != 0 {
		return domainmoney.FromMajorFloat(i.Subtotal, i.Currency)
	}
	return domainmoney.New(i.SubtotalMinor, i.Currency)
}

func (i OrderItem) TaxAmountMoney() (domainmoney.Money, error) {
	if i.TaxAmountMinor == 0 && i.TaxAmount != 0 {
		return domainmoney.FromMajorFloat(i.TaxAmount, i.Currency)
	}
	return domainmoney.New(i.TaxAmountMinor, i.Currency)
}

func (i OrderItem) DiscountMoney() (domainmoney.Money, error) {
	if i.DiscountMinor == 0 && i.Discount != 0 {
		return domainmoney.FromMajorFloat(i.Discount, i.Currency)
	}
	return domainmoney.New(i.DiscountMinor, i.Currency)
}

func (i OrderItem) TotalMoney() (domainmoney.Money, error) {
	if i.TotalMinor == 0 && i.Total != 0 {
		return domainmoney.FromMajorFloat(i.Total, i.Currency)
	}
	return domainmoney.New(i.TotalMinor, i.Currency)
}
