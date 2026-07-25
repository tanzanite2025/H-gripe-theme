package product

import (
	"time"

	"gorm.io/gorm"
)

// ProductMedia stores product-owned media placement.
// Galleries are intentionally separate showcase collections and should not be
// used as the source of truth for product images or videos.
type ProductMedia struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	ProductID    uint           `gorm:"not null;index" json:"product_id"`
	VariantID    *uint          `gorm:"index" json:"variant_id,omitempty"`
	MediaAssetID *uint          `gorm:"index" json:"media_asset_id,omitempty"`
	MediaType    string         `gorm:"default:'image';not null;index" json:"media_type"`
	Role         string         `gorm:"default:'gallery';not null;index" json:"role"`
	URL          string         `gorm:"not null" json:"url"`
	ThumbnailURL string         `json:"thumbnail_url"`
	PosterURL    string         `json:"poster_url"`
	Alt          string         `json:"alt"`
	Title        string         `json:"title"`
	Locale       string         `gorm:"index" json:"locale"`
	SortOrder    int            `gorm:"default:0;not null" json:"sort_order"`
	IsPrimary    bool           `gorm:"default:false;not null" json:"is_primary"`
	IsVisible    bool           `gorm:"default:true;not null" json:"is_visible"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (ProductMedia) TableName() string {
	return "product_media"
}
