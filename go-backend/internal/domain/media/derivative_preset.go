package media

import (
	"time"

	"gorm.io/gorm"
)

// MediaDerivativePreset defines one generated image conversion contract.
// Changing rendering-affecting fields must increment GenerationVersion so
// preflight can distinguish stale generated files from current ones.
type MediaDerivativePreset struct {
	ID                   uint           `gorm:"primarykey" json:"id"`
	Code                 string         `gorm:"type:varchar(40);not null;uniqueIndex:idx_media_derivative_presets_code,where:deleted_at IS NULL" json:"code"`
	Label                string         `gorm:"type:varchar(120);not null;default:''" json:"label"`
	MaxWidth             int            `gorm:"not null" json:"max_width"`
	SortOrder            int            `gorm:"default:0;not null;index" json:"sort_order"`
	Enabled              bool           `gorm:"default:true;not null;index" json:"enabled"`
	GenerationVersion    int            `gorm:"default:1;not null" json:"generation_version"`
	IsSystem             bool           `gorm:"default:false;not null;index" json:"is_system"`
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`
	GeneratedDerivatives int64          `gorm:"-" json:"generated_derivatives"`
}

func (MediaDerivativePreset) TableName() string {
	return "media_derivative_presets"
}
