package media

import (
	"time"

	"gorm.io/gorm"
)

type Media struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	Filename   string         `gorm:"not null" json:"filename"`
	URL        string         `gorm:"not null" json:"url"`
	MimeType   string         `json:"mime_type"`
	Size       int64          `json:"size"` // bytes
	Width      int            `json:"width"`
	Height     int            `json:"height"`
	Alt        string         `json:"alt"`
	Caption    string         `gorm:"type:text" json:"caption"`
	UploaderID uint           `gorm:"index" json:"uploader_id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName 指定表名
func (Media) TableName() string {
	return "media"
}

// MediaAsset stores uploaded file metadata. It is a generic asset record,
// not a gallery image and not a product-media relation by itself.
type MediaAsset struct {
	ID               uint           `gorm:"primarykey" json:"id"`
	Filename         string         `gorm:"not null" json:"filename"`
	OriginalFilename string         `json:"original_filename"`
	URL              string         `gorm:"not null;uniqueIndex" json:"url"`
	StorageKey       string         `json:"storage_key"`
	MimeType         string         `json:"mime_type"`
	MediaType        string         `gorm:"default:'image';not null;index" json:"media_type"` // image, video, document
	Size             int64          `gorm:"default:0" json:"size"`
	Width            int            `gorm:"default:0" json:"width"`
	Height           int            `gorm:"default:0" json:"height"`
	DurationSeconds  int            `gorm:"default:0" json:"duration_seconds"`
	Alt              string         `json:"alt"`
	Caption          string         `gorm:"type:text" json:"caption"`
	UploaderID       uint           `gorm:"index" json:"uploader_id"`
	Status           string         `gorm:"default:'active';not null" json:"status"`
	Visibility       string         `gorm:"default:'public';not null" json:"visibility"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MediaAsset) TableName() string {
	return "media_assets"
}
