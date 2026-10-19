package homevisualtile

import (
	"time"

	"commerce-platform/internal/domain/outbox"
)

const VisualShowcaseDraftUploadReceiptEventKeyPrefix = "visual-showcase-draft-upload:"

// This is a temporary workflow receipt in the existing cleanup Outbox. The
// saved visual_showcase_items rows remain the authoritative business owners.
type VisualShowcaseDraftUploadCleanupReceiptPayload struct {
	outbox.ObjectStorageCleanupPayload
	UploadState    string     `json:"upload_state"`
	WriteDeadline  time.Time  `json:"write_deadline"`
	DraftExpiresAt *time.Time `json:"draft_expires_at,omitempty"`
	Size           int64      `json:"size"`
	ContentSHA256  string     `json:"content_sha256"`
}
