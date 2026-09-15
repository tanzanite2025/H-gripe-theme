package product

import (
	"strings"
	"time"
)

type SpecDefinition struct {
	ID                             uint   `gorm:"primarykey" json:"id"`
	ProductSpecificationTemplateID uint   `gorm:"not null;index;uniqueIndex:idx_product_specification_template_spec_slug" json:"product_specification_template_id"`
	Group                          string `gorm:"type:varchar(80);default:'specs';not null" json:"group"`
	Name                           string `gorm:"type:varchar(120);not null" json:"name"`
	Slug                           string `gorm:"type:varchar(120);not null;uniqueIndex:idx_product_specification_template_spec_slug" json:"slug"`
	FieldType                      string `gorm:"type:varchar(32);default:'text';not null" json:"field_type"`
	Presentation                   string `gorm:"type:varchar(32);default:'text';not null" json:"presentation"`
	Unit                           string `gorm:"type:varchar(32)" json:"unit"`
	IsRequired                     bool   `gorm:"default:false;not null" json:"is_required"`
	IsFilterable                   bool   `gorm:"default:false;not null" json:"is_filterable"`
	IsVisible                      bool   `gorm:"default:true;not null" json:"is_visible"`
	// Role is the explicit runtime role of this definition.
	Role          string                  `gorm:"type:varchar(24);default:'attribute';not null;index" json:"role"`
	SelectionMode string                  `gorm:"type:varchar(16);default:'single';not null" json:"selection_mode"`
	MinSelections int                     `gorm:"default:0;not null" json:"min_selections"`
	MaxSelections *int                    `gorm:"" json:"max_selections,omitempty"`
	SortOrder     int                     `gorm:"default:0;not null" json:"sort_order"`
	Validation    string                  `gorm:"type:text" json:"validation"`
	OptionItems   []ProductSpecOptionItem `gorm:"foreignKey:SpecDefinitionID" json:"option_items,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

func (SpecDefinition) TableName() string {
	return "product_spec_definitions"
}

// RuntimeRole returns the explicit template role used by product, catalog and
// checkout logic.
func (s SpecDefinition) RuntimeRole() string {
	role := strings.TrimSpace(s.Role)
	if role != "" {
		return role
	}
	return "attribute"
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
