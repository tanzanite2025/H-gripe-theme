package media

import (
	"time"
)

const (
	MediaDerivativeRebuildJobStatusPending   = "pending"
	MediaDerivativeRebuildJobStatusRunning   = "running"
	MediaDerivativeRebuildJobStatusSucceeded = "succeeded"
)

// MediaDerivativeRebuildJob is a durable, resumable synchronization request.
// It is created whenever the active conversion contract changes and processes
// image assets in bounded batches, so a configuration save never performs a
// synchronous catalog-wide resize.
type MediaDerivativeRebuildJob struct {
	ID                      uint       `gorm:"primaryKey" json:"id"`
	Reason                  string     `gorm:"size:80;not null;default:''" json:"reason"`
	Status                  string     `gorm:"size:16;not null;index" json:"status"`
	CursorAssetID           uint       `gorm:"not null;default:0;index" json:"cursor_asset_id"`
	ScannedAssets           int        `gorm:"not null;default:0" json:"scanned_assets"`
	GeneratedAssets         int        `gorm:"not null;default:0" json:"generated_assets"`
	GeneratedDerivatives    int        `gorm:"not null;default:0" json:"generated_derivatives"`
	FailedAssets            int        `gorm:"not null;default:0" json:"failed_assets"`
	UpdatedProductMediaRows int64      `gorm:"not null;default:0" json:"updated_product_media_rows"`
	LastError               string     `gorm:"type:text;not null;default:''" json:"last_error,omitempty"`
	LockedAt                *time.Time `gorm:"index" json:"locked_at,omitempty"`
	LockedBy                string     `gorm:"size:128;not null;default:'';index" json:"locked_by"`
	LeaseGeneration         int64      `gorm:"not null;default:0;index" json:"lease_generation"`
	LeaseExpiresAt          *time.Time `gorm:"index" json:"lease_expires_at,omitempty"`
	StartedAt               *time.Time `gorm:"index" json:"started_at,omitempty"`
	FinishedAt              *time.Time `gorm:"index" json:"finished_at,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at"`
}

func (MediaDerivativeRebuildJob) TableName() string {
	return "media_derivative_rebuild_jobs"
}
