package product

import (
	"time"

	"gorm.io/datatypes"
)

// ProductDisplayPriceSnapshot is a storefront read model. Source pricing
// remains owned by products/product_variants; this table is rebuilt from that
// source and is never part of the transactional product update.
type ProductDisplayPriceSnapshot struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	ScopeKey             string         `gorm:"size:64;uniqueIndex;not null" json:"-"`
	ProductID            uint           `gorm:"not null;index" json:"product_id"`
	VariantID            *uint          `gorm:"index" json:"variant_id,omitempty"`
	SourceCurrency       string         `gorm:"size:3;not null" json:"source_currency"`
	SourcePriceMinor     int64          `gorm:"not null" json:"source_price_minor"`
	SourceSalePriceMinor *int64         `gorm:"column:source_sale_price_minor" json:"source_sale_price_minor,omitempty"`
	DisplayPriceData     datatypes.JSON `gorm:"column:display_prices;type:json;not null;default:'[]'" json:"display_prices"`
	CreatedAt            time.Time      `json:"-"`
	UpdatedAt            time.Time      `json:"-"`
}

func (ProductDisplayPriceSnapshot) TableName() string { return "product_display_price_snapshots" }
