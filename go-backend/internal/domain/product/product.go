package product

import (
	"encoding/json"
	"strings"
	"time"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const DefaultPriceCurrency = "USD"

const (
	FulfillmentModeStock       = "stock"
	FulfillmentModeMadeToOrder = "made_to_order"
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
	case FulfillmentModeStock, FulfillmentModeMadeToOrder:
		return true
	default:
		return false
	}
}

type Product struct {
	ID                             uint   `gorm:"primarykey" json:"id"`
	ProductSpecificationTemplateID *uint  `gorm:"index" json:"product_specification_template_id"`
	ProductCategoryID              *uint  `gorm:"index" json:"product_category_id"`
	BrandID                        *uint  `gorm:"index" json:"brand_id,omitempty"`
	ShippingTemplateID             *uint  `gorm:"index" json:"shipping_template_id"`
	AfterSalesTemplateID           *uint  `gorm:"index" json:"after_sales_template_id"`
	PackagingTemplateID            *uint  `gorm:"index" json:"packaging_template_id"`
	CustomsClassificationProfileID *uint  `gorm:"index" json:"customs_classification_profile_id,omitempty"`
	HSCode                         string `gorm:"column:hs_code;size:12" json:"hs_code"`
	CNCode                         string `gorm:"column:cn_code;size:12" json:"cn_code"`
	CountryOfOrigin                string `gorm:"column:country_of_origin;size:2" json:"country_of_origin"`
	CustomsDescription             string `gorm:"column:customs_description;size:255" json:"customs_description"`
	// Product identity belongs to product_variants.SKU. This field is a
	// presentation-only projection kept for in-memory catalog responses.
	SKU            string `gorm:"-" json:"-"`
	Name           string `gorm:"not null;index" json:"name"`
	Slug           string `gorm:"uniqueIndex:idx_product_slug_locale_active,where:deleted_at IS NULL;not null" json:"slug"`
	Description    string `gorm:"type:text" json:"description"`
	ShortDesc      string `gorm:"type:text" json:"short_description"`
	Currency       string `gorm:"size:3;not null;default:'USD';index" json:"currency"`
	PriceMinor     int64  `gorm:"column:price_minor;not null;default:0" json:"price_minor"`
	SalePriceMinor *int64 `gorm:"column:sale_price_minor" json:"sale_price_minor,omitempty"`
	// DisplayPriceData is hydrated from ProductDisplayPriceSnapshot by the
	// repository read model loader. It is intentionally not persisted on the
	// transactional products table.
	DisplayPriceData     datatypes.JSON               `gorm:"-" json:"-"`
	DisplayPriceSnapshot *ProductDisplayPriceSnapshot `gorm:"-" json:"-"`
	FulfillmentMode      string                       `gorm:"size:20;not null;default:'stock';index" json:"fulfillment_mode"`
	// Inventory is owned by product_variants. This field is a presentation-only
	// projection and is never persisted on the transactional products table.
	Stock                        int                           `gorm:"-" json:"-"`
	Status                       string                        `gorm:"default:'active';index" json:"status"` // active, inactive, out_of_stock
	Locale                       string                        `gorm:"uniqueIndex:idx_product_slug_locale_active,where:deleted_at IS NULL;default:'en';index" json:"locale"`
	ParentID                     *uint                         `gorm:"index" json:"parent_id"` // 翻译关联
	Featured                     bool                          `gorm:"default:false" json:"featured"`
	ViewCount                    int                           `gorm:"default:0" json:"view_count"`
	MetaTitle                    string                        `json:"meta_title"`
	MetaDesc                     string                        `gorm:"type:text" json:"meta_description"`
	Media                        []ProductMedia                `gorm:"foreignKey:ProductID" json:"media,omitempty"`
	ProductSpecificationTemplate *ProductSpecificationTemplate `gorm:"foreignKey:ProductSpecificationTemplateID" json:"product_specification_template,omitempty"`
	ProductCategory              *ProductCategory              `gorm:"foreignKey:ProductCategoryID;constraint:OnDelete:SET NULL" json:"product_category,omitempty"`
	Brand                        *ProductBrand                 `gorm:"foreignKey:BrandID;constraint:OnDelete:RESTRICT" json:"brand,omitempty"`
	AfterSalesTemplate           *ProductInformationTemplate   `gorm:"foreignKey:AfterSalesTemplateID;constraint:OnDelete:SET NULL" json:"after_sales_template,omitempty"`
	PackagingTemplate            *ProductInformationTemplate   `gorm:"foreignKey:PackagingTemplateID;constraint:OnDelete:SET NULL" json:"packaging_template,omitempty"`
	CustomsClassificationProfile *CustomsClassificationProfile `gorm:"foreignKey:CustomsClassificationProfileID;constraint:OnDelete:SET NULL" json:"customs_classification_profile,omitempty"`
	SpecValues                   []ProductSpecValue            `gorm:"foreignKey:ProductID" json:"spec_values,omitempty"`
	Variants                     []ProductVariant              `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
	VariantOptionValues          []ProductVariantOptionValue   `gorm:"foreignKey:ProductID" json:"variant_option_values,omitempty"`
	OptionValueRelations         []ProductOptionValueRelation  `gorm:"foreignKey:ProductID" json:"option_value_relations,omitempty"`
	OptionValueRelationsDirty    bool                          `gorm:"-" json:"-"`
	TranslationGroup             *ProductTranslationGroup      `gorm:"-" json:"translation_group,omitempty"`
	CreatedAt                    time.Time                     `json:"created_at"`
	UpdatedAt                    time.Time                     `json:"updated_at"`
	DeletedAt                    gorm.DeletedAt                `gorm:"index" json:"-"`
}

// MarshalJSON keeps the admin/catalog response shape while deriving legacy
// summary fields from active variants. These fields are presentation values,
// never persistence or transaction inputs.
func (p Product) MarshalJSON() ([]byte, error) {
	type productAlias Product
	price, salePrice := "0", (*string)(nil)
	if variant := p.StartingPriceVariant(); variant != nil {
		if money, err := variant.PriceMoney(); err == nil {
			price, _ = money.FormatMajor()
		}
		if money, err := variant.SalePriceMoney(); err == nil && money != nil {
			value, _ := money.FormatMajor()
			salePrice = &value
		}
	} else if money, err := p.PriceMoney(); err == nil {
		price, _ = money.FormatMajor()
	}
	return json.Marshal(struct {
		productAlias
		SKU              string  `json:"sku"`
		PriceDecimal     string  `json:"price_decimal"`
		SalePriceDecimal *string `json:"sale_price_decimal,omitempty"`
		Stock            int     `json:"stock"`
	}{
		productAlias:     productAlias(p),
		SKU:              p.DisplaySKU(),
		PriceDecimal:     price,
		SalePriceDecimal: salePrice,
		Stock:            p.TotalVariantStock(),
	})
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}

// BeforeCreate GORM钩子：创建前
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.Locale == "" {
		p.Locale = "en"
	}
	if p.Status == "" {
		p.Status = "active"
	}
	if err := p.normalizeFulfillmentMode(); err != nil {
		return err
	}
	return p.normalizeCurrency()
}

func (p *Product) BeforeSave(tx *gorm.DB) error {
	if err := p.normalizeFulfillmentMode(); err != nil {
		return err
	}
	return p.normalizeCurrency()
}

func (p *Product) normalizeFulfillmentMode() error {
	p.FulfillmentMode = NormalizeFulfillmentMode(p.FulfillmentMode)
	if !IsValidFulfillmentMode(p.FulfillmentMode) {
		return gorm.ErrInvalidData
	}
	return nil
}

func (p *Product) normalizeCurrency() error {
	p.Currency = currency.NormalizeCode(p.Currency)
	if p.Currency == "" {
		p.Currency = DefaultPriceCurrency
	}
	if !currency.IsValidCode(p.Currency) || !currency.IsCatalogCode(p.Currency) {
		return gorm.ErrInvalidData
	}
	if p.PriceMinor < 0 || (p.SalePriceMinor != nil && *p.SalePriceMinor < 0) {
		return gorm.ErrInvalidData
	}
	return nil
}

func (p *Product) ActiveVariants() []ProductVariant {
	var variants []ProductVariant
	for _, variant := range p.Variants {
		if variant.IsActive {
			variants = append(variants, variant)
		}
	}
	return variants
}

func (p *Product) DefaultVariant() *ProductVariant {
	activeVariants := p.ActiveVariants()
	if len(activeVariants) == 0 {
		return nil
	}
	for i := range activeVariants {
		if activeVariants[i].IsDefault {
			return &activeVariants[i]
		}
	}
	return &activeVariants[0]
}

// StartingPriceVariant returns the active variant with the lowest effective
// price. The default variant wins deterministic ties so existing catalog
// presentation remains stable when variants share the same price.
func (p *Product) StartingPriceVariant() *ProductVariant {
	var selected *ProductVariant
	for i := range p.Variants {
		candidate := &p.Variants[i]
		if !candidate.IsActive {
			continue
		}
		if selected == nil || startingPriceVariantPrecedes(candidate, selected) {
			selected = candidate
		}
	}
	return selected
}

func startingPriceVariantPrecedes(candidate, selected *ProductVariant) bool {
	candidateMoney, candidateErr := candidate.EffectivePriceMoney()
	selectedMoney, selectedErr := selected.EffectivePriceMoney()
	if candidateErr == nil && selectedErr == nil && candidateMoney.Currency() == selectedMoney.Currency() {
		if candidateMoney.AmountMinor() != selectedMoney.AmountMinor() {
			return candidateMoney.AmountMinor() < selectedMoney.AmountMinor()
		}
	} else {
		// Invalid source rows are not purchasable. Preserve deterministic
		// ordering without reintroducing a float fallback.
		return false
	}
	if candidate.IsDefault != selected.IsDefault {
		return candidate.IsDefault
	}
	if candidate.SortOrder != selected.SortOrder {
		return candidate.SortOrder < selected.SortOrder
	}
	if candidate.ID != 0 && selected.ID != 0 {
		return candidate.ID < selected.ID
	}
	return false
}

func (p *Product) DisplaySKU() string {
	if variant := p.DefaultVariant(); variant != nil {
		return variant.SKU
	}
	return p.SKU
}

func (p *Product) DisplayPrices() (string, *string) {
	if variant := p.StartingPriceVariant(); variant != nil {
		priceMoney, err := variant.PriceMoney()
		if err != nil {
			return "", nil
		}
		price, err := priceMoney.FormatMajor()
		if err != nil {
			return "", nil
		}
		saleMoney, err := variant.SalePriceMoney()
		if err != nil || saleMoney == nil {
			return price, nil
		}
		sale, err := saleMoney.FormatMajor()
		if err != nil {
			return price, nil
		}
		return price, &sale
	}
	priceMoney, err := p.PriceMoney()
	if err != nil {
		return "", nil
	}
	price, err := priceMoney.FormatMajor()
	if err != nil {
		return "", nil
	}
	saleMoney, err := p.SalePriceMoney()
	if err != nil || saleMoney == nil {
		return price, nil
	}
	sale, err := saleMoney.FormatMajor()
	if err != nil {
		return price, nil
	}
	return price, &sale
}

// PriceMoney returns the source catalog price in the currency's minor unit.
// It is the only price representation that pricing code should consume.
func (p *Product) PriceMoney() (domainmoney.Money, error) {
	return domainmoney.New(p.PriceMinor, p.Currency)
}

func (p *Product) SalePriceMoney() (*domainmoney.Money, error) {
	if p.SalePriceMinor == nil {
		return nil, nil
	}
	value, err := domainmoney.New(*p.SalePriceMinor, p.Currency)
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func (p *Product) DisplayPriceCurrency() string {
	if variant := p.StartingPriceVariant(); variant != nil && variant.Currency != "" {
		return variant.Currency
	}
	return p.Currency
}

func (p *Product) TotalVariantStock() int {
	activeVariants := p.ActiveVariants()
	if len(activeVariants) == 0 {
		return 0
	}

	total := 0
	for _, variant := range activeVariants {
		total += variant.Stock
	}
	return total
}
