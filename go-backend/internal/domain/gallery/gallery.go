package gallery

import (
	"time"

	"gorm.io/gorm"
)

// Gallery 图片库
type Gallery struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"not null" json:"name"`
	Slug        string         `gorm:"uniqueIndex" json:"slug"`
	Description string         `gorm:"type:text" json:"description"`
	CoverImage  string         `json:"cover_image"`
	Images      []GalleryImage `gorm:"foreignKey:GalleryID" json:"images"`
	ViewCount   int            `gorm:"default:0" json:"view_count"`
	Status      string         `gorm:"default:'published'" json:"status"` // draft, published
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Gallery) TableName() string {
	return "galleries"
}

// GalleryImage 图片库图片
type GalleryImage struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	GalleryID   uint           `gorm:"not null;index" json:"gallery_id"`
	URL         string         `gorm:"not null" json:"url"`
	Thumbnail   string         `json:"thumbnail"`
	Title       string         `json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	Alt         string         `json:"alt"`
	Width       int            `json:"width"`
	Height      int            `json:"height"`
	Size        int64          `json:"size"` // bytes
	Tags        string         `json:"tags"` // 逗号分隔
	Order       int            `gorm:"default:0" json:"order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (GalleryImage) TableName() string {
	return "gallery_images"
}
