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
