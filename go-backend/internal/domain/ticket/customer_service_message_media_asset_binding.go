package ticket

import "time"

const (
	TicketMessageLegacyURLAttachments = "legacy_urls"
	TicketMessageMediaAssetBindings   = "media_asset_bindings"
)

// TicketMessageMediaAttachment is the authoritative ordered business relation
// for messages using media_asset_bindings. URLs are generated only on reads.
type TicketMessageMediaAttachment struct {
	ID              uint      `gorm:"primarykey" json:"-"`
	TicketMessageID uint      `gorm:"not null;uniqueIndex:idx_ticket_message_media_attachment_position,priority:1;index" json:"-"`
	MediaAssetID    uint      `gorm:"not null;index" json:"-"`
	AttachmentIndex int       `gorm:"not null;uniqueIndex:idx_ticket_message_media_attachment_position,priority:2" json:"-"`
	UploadSessionID *string   `gorm:"type:varchar(36);uniqueIndex" json:"-"`
	CreatedAt       time.Time `json:"-"`
}

func (TicketMessageMediaAttachment) TableName() string { return "ticket_message_media_attachments" }

// The transport may select a public asset or an owned upload receipt, never
// both. These requests are not stored as a second attachment representation.
type CustomerServiceMessageMediaAttachmentBindingRequest struct {
	MediaAssetID    uint   `json:"media_asset_id,omitempty"`
	UploadSessionID string `json:"upload_session_id,omitempty"`
}
