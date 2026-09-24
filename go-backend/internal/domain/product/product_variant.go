package product

import (
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ProductVariant struct {
	ID                 uint   `gorm:"primarykey" json:"id"`
	ProductID          uint   `gorm:"not null;index;uniqueIndex:idx_product_variant_options" json:"product_id"`
	MasterVariantID    *uint  `gorm:"index" json:"master_variant_id,omitempty"`
	ShippingTemplateID *uint  `gorm:"index" json:"shipping_template_id"`
	SKU                string `gorm:"type:varchar(120);uniqueIndex:idx_product_variants_sku_active,where:deleted_at IS NULL;not null" json:"sku"`
	Title              string `gorm:"type:varchar(160)" json:"title"`
	OptionValues       string `gorm:"type:jsonb;not null;default:'{}';uniqueIndex:idx_product_variant_options" json:"option_values"`
	Currency           string `gorm:"size:3;not null;default:'USD';index" json:"currency"`
	PriceMinor         int64  `gorm:"column:price_minor;not null;default:0" json:"price_minor"`
	SalePriceMinor     *int64 `gorm:"column:sale_price_minor" json:"sale_price_minor,omitempty"`
	// DisplayPriceData is hydrated from ProductDisplayPriceSnapshot by the
	// repository read model loader. It is intentionally not persisted on the
	// transactional product_variants table.
	DisplayPriceData     datatypes.JSON                  `gorm:"-" json:"-"`
	DisplayPriceSnapshot *ProductDisplayPriceSnapshot    `gorm:"-" json:"-"`
	Stock                int                             `gorm:"default:0;not null" json:"stock"`
	Weight               int                             `gorm:"column:weight_grams" json:"weight_grams"`
	IsDefault            bool                            `gorm:"default:false;not null" json:"is_default"`
	IsActive             bool                            `gorm:"default:true;not null" json:"is_active"`
	SortOrder            int                             `gorm:"default:0;not null" json:"sort_order"`
	CreatedAt            time.Time                       `json:"created_at"`
	UpdatedAt            time.Time                       `json:"updated_at"`
	DeletedAt            gorm.DeletedAt                  `gorm:"index" json:"-"`
	OptionGroupRules     []ProductOptionGroupVariantRule `gorm:"foreignKey:VariantID" json:"option_group_rules,omitempty"`
	OptionValueRules     []ProductOptionValueVariantRule `gorm:"foreignKey:VariantID" json:"option_value_rules,omitempty"`
}

func (ProductVariant) TableName() string {
	return "product_variants"
}

func (v *ProductVariant) EffectivePrice() string {
	value, err := v.EffectivePriceMoney()
	if err != nil {
		return ""
	}
	major, err := value.FormatMajor()
	if err != nil {
		return ""
	}
	return major
}

func (v *ProductVariant) PriceMoney() (domainmoney.Money, error) {
	return domainmoney.New(v.PriceMinor, v.Currency)
}

func (v *ProductVariant) SalePriceMoney() (*domainmoney.Money, error) {
	if v.SalePriceMinor == nil {
		return nil, nil
	}
	value, err := domainmoney.New(*v.SalePriceMinor, v.Currency)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (v *ProductVariant) EffectivePriceMoney() (domainmoney.Money, error) {
	if v.SalePriceMinor != nil {
		return domainmoney.New(*v.SalePriceMinor, v.Currency)
	}
	return v.PriceMoney()
}

func (v *ProductVariant) BeforeCreate(tx *gorm.DB) error {
	if v.OptionValues == "" {
		v.OptionValues = "{}"
	}
	return v.normalizeCurrency()
}

func (v *ProductVariant) BeforeSave(tx *gorm.DB) error {
	return v.normalizeCurrency()
}

// AfterFind exposes the inventory owned by the master variant while retaining
// the translated row's identity, SKU, price, and option metadata.
func (v *ProductVariant) AfterFind(tx *gorm.DB) error {
	if v.MasterVariantID == nil || *v.MasterVariantID == 0 {
		return nil
	}
	var master struct {
		Stock int
	}
	if err := tx.Model(&ProductVariant{}).Select("stock").Where("id = ?", *v.MasterVariantID).First(&master).Error; err != nil {
		return err
	}
	v.Stock = master.Stock
	return nil
}

func (v *ProductVariant) normalizeCurrency() error {
	v.Currency = currency.NormalizeCode(v.Currency)
	if v.Currency == "" {
		v.Currency = DefaultPriceCurrency
	}
	if !currency.IsValidCode(v.Currency) || !currency.IsCatalogCode(v.Currency) {
		return gorm.ErrInvalidData
	}
	if v.PriceMinor < 0 || (v.SalePriceMinor != nil && *v.SalePriceMinor < 0) {
		return gorm.ErrInvalidData
	}
	return nil
}
