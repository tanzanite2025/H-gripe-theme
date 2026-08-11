package gallery

import (
	"time"

	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

// Gallery 图片库
type Gallery struct {
	ID           uint                 `gorm:"primarykey" json:"id"`
	Name         string               `gorm:"not null" json:"name"`
	Slug         string               `gorm:"uniqueIndex" json:"slug"`
	Description  string               `gorm:"type:text" json:"description"`
	CoverImage   string               `json:"cover_image"`
	Images       []GalleryImage       `gorm:"foreignKey:GalleryID" json:"images"`
	ProductLinks []GalleryProductLink `gorm:"foreignKey:GalleryID" json:"product_links,omitempty"`
	ImageCount   int64                `gorm:"-" json:"image_count"`
	ViewCount    int                  `gorm:"default:0" json:"view_count"`
	Status       string               `gorm:"default:'published'" json:"status"` // draft, published
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
	DeletedAt    gorm.DeletedAt       `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Gallery) TableName() string {
	return "galleries"
}

// GalleryImage 图片库图片
type GalleryImage struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	GalleryID    uint           `gorm:"not null;index" json:"gallery_id"`
	MediaAssetID *uint          `gorm:"index" json:"media_asset_id,omitempty"`
	URL          string         `gorm:"not null" json:"url"`
	Thumbnail    string         `json:"thumbnail"`
	Title        string         `json:"title"`
	Description  string         `gorm:"type:text" json:"description"`
	Alt          string         `json:"alt"`
	Width        int            `json:"width"`
	Height       int            `json:"height"`
	Size         int64          `json:"size"` // bytes
	Tags         string         `json:"tags"` // 逗号分隔
	Order        int            `gorm:"default:0" json:"order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (GalleryImage) TableName() string {
	return "gallery_images"
}

// GalleryProductLink connects a gallery to an existing catalog product.
// Product details are preloaded only for API response mapping.
type GalleryProductLink struct {
	ID        uint             `gorm:"primarykey" json:"id"`
	GalleryID uint             `gorm:"not null;index" json:"gallery_id"`
	ProductID uint             `gorm:"not null;index" json:"product_id"`
	SortOrder int              `gorm:"not null;default:0" json:"sort_order"`
	Product   *product.Product `gorm:"foreignKey:ProductID;references:ID" json:"-"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

func (GalleryProductLink) TableName() string {
	return "gallery_product_links"
}
