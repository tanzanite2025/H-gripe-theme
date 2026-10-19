package media

import (
	"time"

	"gorm.io/gorm"
)

// MediaAsset stores uploaded file metadata. It is a generic asset record,
// not a gallery image and not a product-media relation by itself.
type MediaAsset struct {
	ID                                      uint                       `gorm:"primarykey" json:"id"`
	Filename                                string                     `gorm:"not null" json:"filename"`
	OriginalFilename                        string                     `json:"original_filename"`
	URL                                     string                     `gorm:"not null;uniqueIndex" json:"url"`
	StorageKey                              string                     `json:"storage_key"`
	MimeType                                string                     `json:"mime_type"`
	MediaType                               string                     `gorm:"default:'image';not null;index" json:"media_type"` // image, video, document
	Size                                    int64                      `gorm:"default:0" json:"size"`
	ContentSHA256                           string                     `gorm:"type:char(64);index" json:"content_sha256"`
	CopyrightClaimJSON                      string                     `gorm:"type:text" json:"-"`
	Width                                   int                        `gorm:"default:0" json:"width"`
	Height                                  int                        `gorm:"default:0" json:"height"`
	DurationSeconds                         int                        `gorm:"default:0" json:"duration_seconds"`
	Alt                                     string                     `json:"alt"`
	Caption                                 string                     `gorm:"type:text" json:"caption"`
	UploaderID                              uint                       `gorm:"index;index:idx_media_assets_uploader_actor_account,priority:2" json:"uploader_id"`
	UploaderActorType                       MediaUploadActorType       `gorm:"type:varchar(32);not null;default:'legacy_account';index:idx_media_assets_uploader_actor_account,priority:1" json:"uploader_actor_type"`
	UploaderVisitorSessionHash              string                     `gorm:"type:varchar(64);not null;default:'';index:idx_media_assets_uploader_visitor_session" json:"-"`
	Status                                  string                     `gorm:"default:'active';not null" json:"status"`
	Visibility                              string                     `gorm:"default:'public';not null" json:"visibility"`
	LibraryPartition                        MediaAssetLibraryPartition `gorm:"type:varchar(32);not null;default:'unclassified';index:idx_media_assets_library_partition_created,priority:1" json:"library_partition"`
	UploadOrigin                            MediaAssetUploadOrigin     `gorm:"type:varchar(64);not null;default:'legacy_unknown';index" json:"upload_origin"`
	ClassificationMethod                    string                     `gorm:"type:varchar(32);not null;default:'historical_unknown'" json:"classification_method"`
	ClassificationRevision                  int64                      `gorm:"not null;default:0" json:"classification_revision"`
	CustomerServiceMediaPurpose             string                     `gorm:"type:varchar(32);not null;default:''" json:"customer_service_media_purpose"`
	UploadPurpose                           string                     `gorm:"type:varchar(64);not null;default:''" json:"upload_purpose"`
	AccessURL                               string                     `gorm:"-" json:"access_url,omitempty"`
	BusinessDraftUploadExpiresAt            *time.Time                 `gorm:"-" json:"business_draft_upload_expires_at,omitempty"`
	RequiresConversationAccess              bool                       `gorm:"-" json:"requires_conversation_access,omitempty"`
	PendingCustomerServiceUploadSessionID   string                     `gorm:"-" json:"-"`
	CustomerServiceDerivativeWriteToken     string                     `gorm:"type:varchar(36);not null;default:''" json:"-"`
	CustomerServiceDerivativeWriteExpiresAt *time.Time                 `json:"-"`
	Derivatives                             []MediaAssetDerivative     `gorm:"foreignKey:MediaAssetID" json:"derivatives,omitempty"`
	CreatedAt                               time.Time                  `gorm:"index:idx_media_assets_library_partition_created,priority:2" json:"created_at"`
	UpdatedAt                               time.Time                  `json:"updated_at"`
	DeletedAt                               gorm.DeletedAt             `gorm:"index" json:"-"`
}

func (MediaAsset) TableName() string {
	return "media_assets"
}
