package product

import (
	"time"

	"gorm.io/gorm"
)

// ProductShippingTemplateBinding connects a specific SKU to a
// customer-facing shipping billing template. Multiple enabled bindings are
// intentional: Checkout uses them to produce alternative shipping plans.
type ProductShippingTemplateBinding struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	ProductID          uint           `gorm:"not null;index;uniqueIndex:idx_product_shipping_template_bindings_variant_scope,where:variant_id IS NOT NULL AND deleted_at IS NULL" json:"product_id"`
	VariantID          uint           `gorm:"not null;check:shipping_template_binding_requires_sku,variant_id > 0;index;uniqueIndex:idx_product_shipping_template_bindings_variant_scope,where:variant_id IS NOT NULL AND deleted_at IS NULL" json:"variant_id"`
	ShippingTemplateID uint           `gorm:"not null;index;uniqueIndex:idx_product_shipping_template_bindings_variant_scope,where:variant_id IS NOT NULL AND deleted_at IS NULL" json:"shipping_template_id"`
	Priority           int            `gorm:"not null;default:0" json:"priority"`
	Enabled            bool           `gorm:"not null;default:true;index" json:"enabled"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductShippingTemplateBinding) TableName() string {
	return "product_shipping_template_bindings"
}
