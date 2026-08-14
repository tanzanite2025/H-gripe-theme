package product

import (
	"time"

	"commerce-platform/internal/domain/currency"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const DefaultPriceCurrency = "USD"

type Product struct {
	ID                   uint                        `gorm:"primarykey" json:"id"`
	ProductTypeID        *uint                       `gorm:"index" json:"product_type_id"`
	BrandID              *uint                       `gorm:"index" json:"brand_id,omitempty"`
	ShippingTemplateID   *uint                       `gorm:"index" json:"shipping_template_id"`
	AfterSalesTemplateID *uint                       `gorm:"index" json:"after_sales_template_id"`
	PackagingTemplateID  *uint                       `gorm:"index" json:"packaging_template_id"`
	SKU                  string                      `gorm:"uniqueIndex;not null" json:"sku"`
	Name                 string                      `gorm:"not null;index" json:"name"`
	Slug                 string                      `gorm:"uniqueIndex:idx_product_slug_locale;not null" json:"slug"`
	Description          string                      `gorm:"type:text" json:"description"`
	ShortDesc            string                      `gorm:"type:text" json:"short_description"`
	Currency             string                      `gorm:"size:3;not null;default:'USD';index" json:"currency"`
	Price                float64                     `gorm:"not null" json:"price"`
	SalePrice            *float64                    `json:"sale_price"`
	DisplayPriceData     datatypes.JSON              `gorm:"column:display_prices;type:json;not null;default:'[]'" json:"display_prices,omitempty"`
	Stock                int                         `gorm:"default:0" json:"stock"`
	Status               string                      `gorm:"default:'active';index" json:"status"` // active, inactive, out_of_stock
	Locale               string                      `gorm:"uniqueIndex:idx_product_slug_locale;default:'en';index" json:"locale"`
	ParentID             *uint                       `gorm:"index" json:"parent_id"` // 翻译关联
	Featured             bool                        `gorm:"default:false" json:"featured"`
	ViewCount            int                         `gorm:"default:0" json:"view_count"`
	MetaTitle            string                      `json:"meta_title"`
	MetaDesc             string                      `gorm:"type:text" json:"meta_description"`
	Media                []ProductMedia              `gorm:"foreignKey:ProductID" json:"media,omitempty"`
	ProductType          *ProductType                `gorm:"foreignKey:ProductTypeID" json:"product_type,omitempty"`
	Brand                *ProductBrand               `gorm:"foreignKey:BrandID;constraint:OnDelete:RESTRICT" json:"brand,omitempty"`
	AfterSalesTemplate   *ProductInformationTemplate `gorm:"foreignKey:AfterSalesTemplateID;constraint:OnDelete:SET NULL" json:"after_sales_template,omitempty"`
	PackagingTemplate    *ProductInformationTemplate `gorm:"foreignKey:PackagingTemplateID;constraint:OnDelete:SET NULL" json:"packaging_template,omitempty"`
	SpecValues           []ProductSpecValue          `gorm:"foreignKey:ProductID" json:"spec_values,omitempty"`
	Variants             []ProductVariant            `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
	VariantOptionValues  []ProductVariantOptionValue `gorm:"foreignKey:ProductID" json:"variant_option_values,omitempty"`
	TranslationGroup     *ProductTranslationGroup    `gorm:"-" json:"translation_group,omitempty"`
	CreatedAt            time.Time                   `json:"created_at"`
	UpdatedAt            time.Time                   `json:"updated_at"`
	DeletedAt            gorm.DeletedAt              `gorm:"index" json:"-"`
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
	return p.normalizeCurrency()
}

func (p *Product) BeforeSave(tx *gorm.DB) error {
	return p.normalizeCurrency()
}

func (p *Product) normalizeCurrency() error {
	p.Currency = currency.NormalizeCode(p.Currency)
	if p.Currency == "" {
		p.Currency = DefaultPriceCurrency
	}
	if !currency.IsValidCode(p.Currency) || !currency.IsCatalogCode(p.Currency) {
		return gorm.ErrInvalidData
	}
	if len(p.DisplayPriceData) == 0 {
		p.DisplayPriceData = datatypes.JSON([]byte("[]"))
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

func (p *Product) DisplaySKU() string {
	if variant := p.DefaultVariant(); variant != nil {
		return variant.SKU
	}
	return p.SKU
}

func (p *Product) DisplayPrices() (float64, *float64) {
	if variant := p.DefaultVariant(); variant != nil {
		return variant.Price, variant.SalePrice
	}
	return p.Price, p.SalePrice
}

func (p *Product) DisplayPriceCurrency() string {
	if variant := p.DefaultVariant(); variant != nil && variant.Currency != "" {
		return variant.Currency
	}
	return p.Currency
}

func (p *Product) TotalVariantStock() int {
	activeVariants := p.ActiveVariants()
	if len(activeVariants) == 0 {
		return p.Stock
	}

	total := 0
	for _, variant := range activeVariants {
		total += variant.Stock
	}
	return total
}
