package sitelogo

import "time"

const CurrentAssetID uint = 1

type Asset struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Filename         string    `gorm:"not null" json:"filename"`
	OriginalFilename string    `gorm:"not null;default:''" json:"original_filename"`
	URL              string    `gorm:"not null" json:"url"`
	StorageKey       string    `gorm:"not null" json:"storage_key"`
	MimeType         string    `gorm:"type:varchar(120);not null;default:'image/svg+xml'" json:"mime_type"`
	Size             int64     `gorm:"not null;default:0" json:"size"`
	ContentSHA256    string    `gorm:"type:char(64);not null;default:''" json:"content_sha256"`
	Width            int       `gorm:"not null;default:48" json:"width"`
	Height           int       `gorm:"not null;default:48" json:"height"`
	UploaderID       uint      `gorm:"not null;default:0" json:"uploader_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (Asset) TableName() string {
	return "site_logo_assets"
}
