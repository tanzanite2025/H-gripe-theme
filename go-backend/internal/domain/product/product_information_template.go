package product

import (
	"time"

	"gorm.io/gorm"
)

const (
	ProductInformationTemplateKindAfterSales = "after_sales"
	ProductInformationTemplateKindPackaging  = "packaging"
)

type ProductInformationTemplate struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	Kind      string         `gorm:"size:32;not null;index" json:"kind"`
	Name      string         `gorm:"size:160;not null" json:"name"`
	Slug      string         `gorm:"size:160;not null" json:"slug"`
	Content   string         `gorm:"type:text;not null" json:"content"`
	Locale    string         `gorm:"size:32;not null;default:'en';index" json:"locale"`
	IsEnabled bool           `gorm:"not null;index" json:"is_enabled"`
	SortOrder int            `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductInformationTemplate) TableName() string {
	return "product_information_templates"
}

func IsProductInformationTemplateKind(value string) bool {
	return value == ProductInformationTemplateKindAfterSales || value == ProductInformationTemplateKindPackaging
}
