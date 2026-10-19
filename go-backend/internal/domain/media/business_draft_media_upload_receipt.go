package media

import (
	"gorm.io/datatypes"
	"time"
)

// A short-lived workflow receipt, never an authority for business references.
// Bound receipts stop automatic expiry; normal reference scanning governs deletion.
type BusinessDraftMediaUploadReceipt struct {
	ID                  string                 `gorm:"type:uuid;primaryKey"`
	AssetID             *uint                  `gorm:"uniqueIndex"`
	UploaderID          uint                   `gorm:"not null"`
	UploadOrigin        MediaAssetUploadOrigin `gorm:"not null"`
	OriginalObjectKey   string                 `gorm:"not null;uniqueIndex"`
	ObjectKeysJSON      datatypes.JSON         `gorm:"column:object_keys_json;not null"`
	ContentSHA256       string                 `gorm:"not null"`
	Size                int64                  `gorm:"not null"`
	MimeType            string                 `gorm:"not null"`
	Status              string                 `gorm:"not null;index:idx_business_draft_media_receipt_expiry,priority:1"`
	ExpiresAt           time.Time              `gorm:"not null;index:idx_business_draft_media_receipt_expiry,priority:2"`
	NextExpiryAttemptAt time.Time              `gorm:"not null;default:CURRENT_TIMESTAMP"`
	CreatedAt           time.Time              `gorm:"not null"`
}
