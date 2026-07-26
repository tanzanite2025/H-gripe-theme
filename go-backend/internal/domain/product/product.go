package product

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID                 uint               `gorm:"primarykey" json:"id"`
	ProductTypeID      *uint              `gorm:"index" json:"product_type_id"`
	ShippingTemplateID *uint              `gorm:"index" json:"shipping_template_id"`
	SKU                string             `gorm:"uniqueIndex;not null" json:"sku"`
	Name               string             `gorm:"not null;index" json:"name"`
	Slug               string             `gorm:"uniqueIndex:idx_product_slug_locale;not null" json:"slug"`
	Description        string             `gorm:"type:text" json:"description"`
	ShortDesc          string             `gorm:"type:text" json:"short_description"`
	Price              float64            `gorm:"not null" json:"price"`
	SalePrice          *float64           `json:"sale_price"`
	Stock              int                `gorm:"default:0" json:"stock"`
	Status             string             `gorm:"default:'active';index" json:"status"` // active, inactive, out_of_stock
	Locale             string             `gorm:"uniqueIndex:idx_product_slug_locale;default:'en';index" json:"locale"`
	ParentID           *uint              `gorm:"index" json:"parent_id"` // 翻译关联
	Featured           bool               `gorm:"default:false" json:"featured"`
	ViewCount          int                `gorm:"default:0" json:"view_count"`
	MetaTitle          string             `json:"meta_title"`
	MetaDesc           string             `gorm:"type:text" json:"meta_description"`
	Media              []ProductMedia     `gorm:"foreignKey:ProductID" json:"media,omitempty"`
	ProductType        *ProductType       `gorm:"foreignKey:ProductTypeID" json:"product_type,omitempty"`
	SpecValues         []ProductSpecValue `gorm:"foreignKey:ProductID" json:"spec_values,omitempty"`
	Variants           []ProductVariant   `gorm:"foreignKey:ProductID" json:"variants,omitempty"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
	DeletedAt          gorm.DeletedAt     `gorm:"index" json:"-"`
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
