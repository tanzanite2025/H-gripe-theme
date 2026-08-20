package product

import "time"

type ProductSpecificationTemplate struct {
	ID              uint             `gorm:"primarykey" json:"id"`
	Name            string           `gorm:"type:varchar(120);not null" json:"name"`
	Slug            string           `gorm:"type:varchar(120);uniqueIndex;not null" json:"slug"`
	Description     string           `gorm:"type:text" json:"description"`
	SortOrder       int              `gorm:"default:0;not null" json:"sort_order"`
	IsEnabled       bool             `gorm:"default:true;not null" json:"is_enabled"`
	IsSystemManaged bool             `gorm:"default:false;not null;index" json:"is_system_managed"`
	SpecDefinitions []SpecDefinition `gorm:"foreignKey:ProductSpecificationTemplateID" json:"spec_definitions,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

func (ProductSpecificationTemplate) TableName() string {
	return "product_specification_templates"
}
