package product

import "time"

type SpecDefinition struct {
	ID                             uint      `gorm:"primarykey" json:"id"`
	ProductSpecificationTemplateID uint      `gorm:"not null;index;uniqueIndex:idx_product_specification_template_spec_slug" json:"product_specification_template_id"`
	Group                          string    `gorm:"type:varchar(80);default:'specs';not null" json:"group"`
	Name                           string    `gorm:"type:varchar(120);not null" json:"name"`
	Slug                           string    `gorm:"type:varchar(120);not null;uniqueIndex:idx_product_specification_template_spec_slug" json:"slug"`
	FieldType                      string    `gorm:"type:varchar(32);default:'text';not null" json:"field_type"`
	Presentation                   string    `gorm:"type:varchar(32);default:'text';not null" json:"presentation"`
	Unit                           string    `gorm:"type:varchar(32)" json:"unit"`
	IsRequired                     bool      `gorm:"default:false;not null" json:"is_required"`
	IsFilterable                   bool      `gorm:"default:false;not null" json:"is_filterable"`
	IsVisible                      bool      `gorm:"default:true;not null" json:"is_visible"`
	IsVariantOption                bool      `gorm:"default:false;not null" json:"is_variant_option"`
	SortOrder                      int       `gorm:"default:0;not null" json:"sort_order"`
	Options                        string    `gorm:"type:text" json:"options"`
	Validation                     string    `gorm:"type:text" json:"validation"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`
}

func (SpecDefinition) TableName() string {
	return "product_spec_definitions"
}

type ProductSpecValue struct {
	ID               uint            `gorm:"primarykey" json:"id"`
	ProductID        uint            `gorm:"not null;index;uniqueIndex:idx_product_spec_value" json:"product_id"`
	SpecDefinitionID uint            `gorm:"not null;index;uniqueIndex:idx_product_spec_value" json:"spec_definition_id"`
	Value            string          `gorm:"type:text;not null" json:"value"`
	SpecDefinition   *SpecDefinition `gorm:"foreignKey:SpecDefinitionID" json:"definition,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

func (ProductSpecValue) TableName() string {
	return "product_spec_values"
}
