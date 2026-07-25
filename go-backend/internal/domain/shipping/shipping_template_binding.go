package shipping

import (
	"time"

	"gorm.io/gorm"
)

// ShippingTemplateBinding maps a shipping template to a default, product type, product, or variant scope.
type ShippingTemplateBinding struct {
	ID            uint              `gorm:"primarykey" json:"id"`
	TemplateID    uint              `gorm:"not null;index" json:"template_id"`
	Scope         string            `gorm:"type:varchar(30);not null;index" json:"scope"` // default, product_type, product, variant
	ProductTypeID *uint             `gorm:"index" json:"product_type_id"`
	ProductID     *uint             `gorm:"index" json:"product_id"`
	VariantID     *uint             `gorm:"index" json:"variant_id"`
	Priority      int               `gorm:"default:0;not null;index" json:"priority"`
	Enabled       bool              `gorm:"default:true;not null;index" json:"enabled"`
	Template      *ShippingTemplate `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
	DeletedAt     gorm.DeletedAt    `gorm:"index" json:"-"`
}

func (ShippingTemplateBinding) TableName() string {
	return "shipping_template_bindings"
}
