package product

import "time"

// ProductSpecOptionItem is a template-level candidate value for a select
// specification. Product-level option values may reference it when a
// template is materialized for a product.
type ProductSpecOptionItem struct {
	ID                     uint      `gorm:"primarykey" json:"id"`
	SpecDefinitionID       uint      `gorm:"not null;index;uniqueIndex:idx_product_spec_option_item_key" json:"spec_definition_id"`
	ValueKey               string    `gorm:"type:varchar(160);not null;uniqueIndex:idx_product_spec_option_item_key" json:"value_key"`
	DefaultLabel           string    `gorm:"type:varchar(160);not null" json:"default_label"`
	ColorHex               string    `gorm:"type:varchar(20)" json:"color_hex,omitempty"`
	SwatchMediaAssetID     *uint     `gorm:"index" json:"swatch_media_asset_id,omitempty"`
	SwatchURL              string    `gorm:"type:varchar(800)" json:"swatch_url,omitempty"`
	IsEnabledByDefault     bool      `gorm:"not null;default:true" json:"is_enabled_by_default"`
	IsDefault              bool      `gorm:"not null;default:false" json:"is_default"`
	DefaultPriceDeltaMinor *int64    `json:"default_price_delta_minor,omitempty"`
	DefaultPriceCurrency   string    `gorm:"type:char(3)" json:"default_price_currency,omitempty"`
	SortOrder              int       `gorm:"not null;default:0" json:"sort_order"`
	Revision               int       `gorm:"not null;default:1" json:"revision"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func (ProductSpecOptionItem) TableName() string {
	return "product_spec_option_items"
}
