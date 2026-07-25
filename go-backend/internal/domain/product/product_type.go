package product

import "time"

type ProductType struct {
	ID              uint             `gorm:"primarykey" json:"id"`
	Name            string           `gorm:"type:varchar(120);not null" json:"name"`
	Slug            string           `gorm:"type:varchar(120);uniqueIndex;not null" json:"slug"`
	Description     string           `gorm:"type:text" json:"description"`
	SortOrder       int              `gorm:"default:0;not null" json:"sort_order"`
	IsEnabled       bool             `gorm:"default:true;not null" json:"is_enabled"`
	SpecDefinitions []SpecDefinition `gorm:"foreignKey:ProductTypeID" json:"spec_definitions,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

func (ProductType) TableName() string {
	return "product_types"
}
