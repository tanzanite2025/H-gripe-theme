package product

import "time"

// ProductVariantOptionValue stores product-specific display metadata for a
// variant option. The stable value_key is what product_variants.option_values
// stores; labels and swatches are presentation data.
type ProductVariantOptionValue struct {
	ID                 uint      `gorm:"primarykey" json:"id"`
	ProductID          uint      `gorm:"not null;index;uniqueIndex:idx_product_variant_option_value_key" json:"product_id"`
	SpecDefinitionID   uint      `gorm:"not null;index;uniqueIndex:idx_product_variant_option_value_key" json:"spec_definition_id"`
	ValueKey           string    `gorm:"type:varchar(160);not null;uniqueIndex:idx_product_variant_option_value_key" json:"value_key"`
	Label              string    `gorm:"type:varchar(160);not null" json:"label"`
	ColorHex           string    `gorm:"type:varchar(20)" json:"color_hex,omitempty"`
	SwatchMediaAssetID *uint     `gorm:"index" json:"swatch_media_asset_id,omitempty"`
	SwatchURL          string    `gorm:"type:varchar(800)" json:"swatch_url,omitempty"`
	SortOrder          int       `gorm:"default:0;not null" json:"sort_order"`
	IsEnabled          bool      `gorm:"default:true;not null" json:"is_enabled"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (ProductVariantOptionValue) TableName() string {
	return "product_variant_option_values"
}
