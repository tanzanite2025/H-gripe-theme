package orderevidence

import (
	"encoding/hex"
	"errors"
	"fmt"
	"mime"
	"strings"
	"time"

	"gorm.io/gorm"
)

// OrderEvidenceAttachment stores only a validated object-storage reference.
// Uploading, hashing, and object deletion belong to the storage-aware service.
type OrderEvidenceAttachment struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	EvidenceItemID   uint      `gorm:"not null;index" json:"evidence_item_id"`
	StorageKey       string    `gorm:"not null" json:"storage_key"`
	OriginalFilename string    `gorm:"not null" json:"original_filename"`
	MimeType         string    `gorm:"size:160;not null" json:"mime_type"`
	SizeBytes        int64     `gorm:"not null" json:"size_bytes"`
	SHA256           string    `gorm:"column:sha256;type:char(64);not null" json:"sha256"`
	UploadedBy       uint      `gorm:"not null;default:0" json:"uploaded_by"`
	CreatedAt        time.Time `json:"created_at"`
}

func (OrderEvidenceAttachment) TableName() string {
	return "order_evidence_attachments"
}

func (a OrderEvidenceAttachment) Validate() error {
	if a.EvidenceItemID == 0 {
		return errors.New("order evidence attachment evidence_item_id is required")
	}
	if _, err := NormalizeEvidenceStorageKey(a.StorageKey); err != nil {
		return err
	}
	if strings.TrimSpace(a.OriginalFilename) == "" {
		return errors.New("order evidence attachment original_filename is required")
	}
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(a.MimeType))
	if err != nil || strings.TrimSpace(mediaType) == "" || !strings.Contains(mediaType, "/") {
		return errors.New("order evidence attachment mime_type is invalid")
	}
	if a.SizeBytes < 0 {
		return errors.New("order evidence attachment size_bytes must be non-negative")
	}
	sha256Value := strings.TrimSpace(a.SHA256)
	if len(sha256Value) != 64 {
		return errors.New("order evidence attachment sha256 is invalid")
	}
	if _, err := hex.DecodeString(sha256Value); err != nil {
		return errors.New("order evidence attachment sha256 is invalid")
	}
	return nil
}

func (a *OrderEvidenceAttachment) BeforeCreate(tx *gorm.DB) error {
	if a == nil {
		return errors.New("order evidence attachment is required")
	}
	a.StorageKey, _ = NormalizeEvidenceStorageKey(a.StorageKey)
	a.MimeType = strings.ToLower(strings.TrimSpace(a.MimeType))
	a.OriginalFilename = strings.TrimSpace(a.OriginalFilename)
	a.SHA256 = strings.ToLower(strings.TrimSpace(a.SHA256))
	return a.Validate()
}

func (a *OrderEvidenceAttachment) BeforeUpdate(tx *gorm.DB) error {
	return ErrOrderEvidenceAttachmentImmutable
}

func (a *OrderEvidenceAttachment) BeforeDelete(tx *gorm.DB) error {
	return ErrOrderEvidenceAttachmentImmutable
}

var ErrOrderEvidenceAttachmentImmutable = errors.New("order evidence attachment is immutable")

func NormalizeEvidenceStorageKey(value string) (string, error) {
	normalized := strings.Trim(strings.ReplaceAll(strings.TrimSpace(value), "\\", "/"), "/")
	if normalized == "" {
		return "", errors.New("order evidence attachment storage_key is required")
	}
	if strings.Contains(normalized, "://") || strings.ContainsAny(normalized, "?#") {
		return "", errors.New("order evidence attachment storage_key must be an object key, not a URL")
	}
	for _, segment := range strings.Split(normalized, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", fmt.Errorf("order evidence attachment storage_key contains an invalid path segment")
		}
	}
	return normalized, nil
}
