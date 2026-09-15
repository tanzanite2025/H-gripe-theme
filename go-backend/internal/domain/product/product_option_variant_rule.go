package product

import "time"

type ProductOptionGroupVariantRule struct {
	ID                    uint      `gorm:"primarykey" json:"id"`
	VariantID             uint      `gorm:"not null;index;uniqueIndex:idx_product_option_group_variant_rule" json:"variant_id"`
	SpecDefinitionID      uint      `gorm:"not null;index;uniqueIndex:idx_product_option_group_variant_rule" json:"spec_definition_id"`
	IsApplicable          bool      `gorm:"not null;default:true" json:"is_applicable"`
	MinSelectionsOverride *int      `json:"min_selections_override,omitempty"`
	MaxSelectionsOverride *int      `json:"max_selections_override,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (ProductOptionGroupVariantRule) TableName() string { return "product_option_group_variant_rules" }

type ProductOptionValueVariantRule struct {
	ID                          uint      `gorm:"primarykey" json:"id"`
	VariantID                   uint      `gorm:"not null;index;uniqueIndex:idx_product_option_value_variant_rule" json:"variant_id"`
	ProductVariantOptionValueID uint      `gorm:"not null;index;uniqueIndex:idx_product_option_value_variant_rule" json:"product_variant_option_value_id"`
	IsEnabled                   bool      `gorm:"not null;default:true" json:"is_enabled"`
	PriceDeltaMinorOverride     *int64    `json:"price_delta_minor_override,omitempty"`
	UnavailableReason           string    `gorm:"type:varchar(240)" json:"unavailable_reason,omitempty"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

func (ProductOptionValueVariantRule) TableName() string { return "product_option_value_variant_rules" }
