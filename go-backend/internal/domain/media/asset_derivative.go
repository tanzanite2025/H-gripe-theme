package media

import (
	"time"

	"gorm.io/gorm"
)

// MediaAssetDerivative is a generated, immutable display asset derived from a
// source MediaAsset. Storefront rendering consumes these URLs instead of asking
// SSR workers to resize uploaded originals at request time.
type MediaAssetDerivative struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	MediaAssetID  uint           `gorm:"not null;uniqueIndex:idx_media_asset_derivatives_asset_preset,where:deleted_at IS NULL;index" json:"media_asset_id"`
	Preset        string         `gorm:"type:varchar(40);not null;uniqueIndex:idx_media_asset_derivatives_asset_preset,where:deleted_at IS NULL" json:"preset"`
	PresetVersion int            `gorm:"default:1;not null;index" json:"preset_version"`
	URL           string         `gorm:"not null;uniqueIndex" json:"url"`
	StorageKey    string         `gorm:"not null;index" json:"storage_key"`
	MimeType      string         `gorm:"type:varchar(120);not null" json:"mime_type"`
	Width         int            `gorm:"default:0;not null" json:"width"`
	Height        int            `gorm:"default:0;not null" json:"height"`
	Size          int64          `gorm:"default:0;not null" json:"size"`
	Asset         *MediaAsset    `gorm:"foreignKey:MediaAssetID" json:"-"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (MediaAssetDerivative) TableName() string {
	return "media_asset_derivatives"
}
