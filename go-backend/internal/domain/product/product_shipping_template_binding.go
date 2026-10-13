package product

import (
	"time"

	"gorm.io/gorm"
)

// ProductShippingTemplateBinding connects a product or one of its SKUs to a
// customer-facing shipping billing template. Multiple enabled bindings are
// intentional: Checkout uses them to produce alternative shipping plans.
type ProductShippingTemplateBinding struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	ProductID          uint           `gorm:"not null;index;uniqueIndex:idx_product_shipping_template_bindings_product_scope,where:variant_id IS NULL AND deleted_at IS NULL;uniqueIndex:idx_product_shipping_template_bindings_variant_scope,where:variant_id IS NOT NULL AND deleted_at IS NULL" json:"product_id"`
	VariantID          *uint          `gorm:"index;uniqueIndex:idx_product_shipping_template_bindings_variant_scope,where:variant_id IS NOT NULL AND deleted_at IS NULL" json:"variant_id,omitempty"`
	ShippingTemplateID uint           `gorm:"not null;index;uniqueIndex:idx_product_shipping_template_bindings_product_scope,where:variant_id IS NULL AND deleted_at IS NULL;uniqueIndex:idx_product_shipping_template_bindings_variant_scope,where:variant_id IS NOT NULL AND deleted_at IS NULL" json:"shipping_template_id"`
	Priority           int            `gorm:"not null;default:0" json:"priority"`
	Enabled            bool           `gorm:"not null;default:true;index" json:"enabled"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductShippingTemplateBinding) TableName() string {
	return "product_shipping_template_bindings"
}
