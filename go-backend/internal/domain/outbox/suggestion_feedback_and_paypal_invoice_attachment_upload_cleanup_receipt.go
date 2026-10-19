package outbox

import "time"

const SuggestionAndPayPalAttachmentReceiptEventKeyPrefix = "suggestion-paypal-attachment-write:"
const SuggestionAndPayPalAttachmentCleanupResourceType = "suggestion_paypal_attachment_upload"

// This receipt is committed before provider IO and retained until the exact
// uploaded object is atomically exchanged for a feedback row or invoice snapshot.
type SuggestionAndPayPalAttachmentUploadReceiptPayload struct {
	ObjectStorageCleanupPayload
	OwnerID          uint       `json:"owner_id"`
	Name             string     `json:"name"`
	Size             int64      `json:"size"`
	MimeType         string     `json:"mime_type"`
	ContentSHA256    string     `json:"content_sha256"`
	UploadState      string     `json:"upload_state"`
	WriteDeadline    time.Time  `json:"write_deadline"`
	BindingExpiresAt *time.Time `json:"binding_expires_at,omitempty"`
}
