package product

import "time"

// ProductVariantOptionValue stores product-specific display metadata for a
// variant option. The stable value_key is what product_variants.option_values
// stores; labels and swatches are presentation data.
type ProductVariantOptionValue struct {
	ID                     uint                       `gorm:"primarykey" json:"id"`
	ProductID              uint                       `gorm:"not null;index;uniqueIndex:idx_product_variant_option_value_key" json:"product_id"`
	SpecDefinitionID       uint                       `gorm:"not null;index;uniqueIndex:idx_product_variant_option_value_key" json:"spec_definition_id"`
	TemplateOptionItemID   *uint                      `gorm:"index" json:"template_option_item_id,omitempty"`
	SourceTemplateRevision int                        `gorm:"default:0;not null" json:"source_template_revision"`
	ValueKey               string                     `gorm:"type:varchar(160);not null;uniqueIndex:idx_product_variant_option_value_key" json:"value_key"`
	Label                  string                     `gorm:"type:varchar(160);not null" json:"label"`
	ColorHex               string                     `gorm:"type:varchar(20)" json:"color_hex,omitempty"`
	SwatchMediaAssetID     *uint                      `gorm:"index" json:"swatch_media_asset_id,omitempty"`
	SwatchURL              string                     `gorm:"type:varchar(800)" json:"swatch_url,omitempty"`
	SortOrder              int                        `gorm:"default:0;not null" json:"sort_order"`
	IsEnabled              bool                       `gorm:"default:true;not null" json:"is_enabled"`
	CreatedAt              time.Time                  `json:"created_at"`
	UpdatedAt              time.Time                  `json:"updated_at"`
	CustomOptionPolicy     *ProductCustomOptionPolicy `gorm:"foreignKey:ProductVariantOptionValueID" json:"custom_option_policy,omitempty"`
}

func (ProductVariantOptionValue) TableName() string {
	return "product_variant_option_values"
}

// ProductCustomOptionPolicy stores transaction semantics for a materialized
// custom option. The initial vertical slice only permits inventory_policy=none.
type ProductCustomOptionPolicy struct {
	ID                          uint      `gorm:"primarykey" json:"id"`
	ProductVariantOptionValueID uint      `gorm:"not null;uniqueIndex" json:"product_variant_option_value_id"`
	PriceDeltaMinor             int64     `gorm:"not null;default:0" json:"price_delta_minor"`
	IsDefault                   bool      `gorm:"not null;default:false" json:"is_default"`
	InventoryPolicy             string    `gorm:"type:varchar(24);not null;default:'none'" json:"inventory_policy"`
	ComponentVariantID          *uint     `gorm:"index" json:"component_variant_id,omitempty"`
	ComponentQuantity           int       `gorm:"not null;default:0" json:"component_quantity"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

func (ProductCustomOptionPolicy) TableName() string {
	return "product_custom_option_policies"
}
