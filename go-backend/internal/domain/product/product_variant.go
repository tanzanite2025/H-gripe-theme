package product

import (
	"time"

	"gorm.io/gorm"
)

type ProductVariant struct {
	ID                 uint           `gorm:"primarykey" json:"id"`
	ProductID          uint           `gorm:"not null;index;uniqueIndex:idx_product_variant_options" json:"product_id"`
	ShippingTemplateID *uint          `gorm:"index" json:"shipping_template_id"`
	SKU                string         `gorm:"type:varchar(120);uniqueIndex;not null" json:"sku"`
	Title              string         `gorm:"type:varchar(160)" json:"title"`
	OptionValues       string         `gorm:"type:text;not null;uniqueIndex:idx_product_variant_options" json:"option_values"`
	Price              float64        `gorm:"not null" json:"price"`
	SalePrice          *float64       `json:"sale_price"`
	Stock              int            `gorm:"default:0;not null" json:"stock"`
	Weight             int            `gorm:"column:weight_grams" json:"weight_grams"`
	IsDefault          bool           `gorm:"default:false;not null" json:"is_default"`
	IsActive           bool           `gorm:"default:true;not null" json:"is_active"`
	SortOrder          int            `gorm:"default:0;not null" json:"sort_order"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductVariant) TableName() string {
	return "product_variants"
}

func (v *ProductVariant) EffectivePrice() float64 {
	if v.SalePrice != nil {
		return *v.SalePrice
	}
	return v.Price
}

func (v *ProductVariant) BeforeCreate(tx *gorm.DB) error {
	if v.OptionValues == "" {
		v.OptionValues = "{}"
	}
	return nil
}
