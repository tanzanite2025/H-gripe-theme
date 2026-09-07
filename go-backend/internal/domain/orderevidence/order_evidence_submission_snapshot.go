package orderevidence

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const (
	OrderEvidenceSubmissionSnapshotSchemaVersion = 1
	SubmissionSnapshotStatusLocked               = "locked"
)

var ErrOrderEvidenceSubmissionSnapshotImmutable = errors.New("order evidence submission snapshot is immutable")

// OrderEvidenceSubmissionSnapshot freezes the exact evidence material selected
// for a provider submission or later evidence export. The payload is
// provider-neutral at the storage boundary and is interpreted by the payment
// adapter that created it.
type OrderEvidenceSubmissionSnapshot struct {
	ID                     uint           `gorm:"primaryKey" json:"id"`
	Provider               string         `gorm:"size:32;not null;index;uniqueIndex:uq_order_evidence_submission_snapshot_version" json:"provider"`
	DisputeID              uint           `gorm:"not null;index;uniqueIndex:uq_order_evidence_submission_snapshot_version" json:"dispute_id"`
	OrderID                uint           `gorm:"not null;index" json:"order_id"`
	EvidencePackageID      *uint          `gorm:"index" json:"evidence_package_id,omitempty"`
	EvidencePackageVersion int            `gorm:"not null;default:0" json:"evidence_package_version"`
	Version                int            `gorm:"not null;uniqueIndex:uq_order_evidence_submission_snapshot_version" json:"version"`
	Status                 string         `gorm:"size:16;not null;index" json:"status"`
	SchemaVersion          int            `gorm:"not null" json:"schema_version"`
	LockedAt               time.Time      `gorm:"not null" json:"locked_at"`
	SnapshotData           datatypes.JSON `gorm:"column:snapshot_data;type:jsonb;not null" json:"-"`
	SnapshotSHA256         string         `gorm:"column:snapshot_sha256;type:char(64);not null" json:"snapshot_sha256"`
	CreatedBy              uint           `gorm:"not null;default:0" json:"created_by"`
	CreatedAt              time.Time      `json:"created_at"`
}

func (OrderEvidenceSubmissionSnapshot) TableName() string {
	return "order_evidence_submission_snapshots"
}

func (s OrderEvidenceSubmissionSnapshot) Validate() error {
	if s.Provider != "stripe" && s.Provider != "paypal" {
		return fmt.Errorf("unsupported order evidence submission snapshot provider %q", s.Provider)
	}
	if s.DisputeID == 0 {
		return errors.New("order evidence submission snapshot dispute_id is required")
	}
	if s.OrderID == 0 {
		return errors.New("order evidence submission snapshot order_id is required")
	}
	if s.EvidencePackageID == nil && s.EvidencePackageVersion != 0 {
		return errors.New("order evidence submission snapshot package version requires package id")
	}
	if s.EvidencePackageID != nil && *s.EvidencePackageID == 0 {
		return errors.New("order evidence submission snapshot package id is invalid")
	}
	if s.Version <= 0 {
		return errors.New("order evidence submission snapshot version must be greater than zero")
	}
	if s.Status != SubmissionSnapshotStatusLocked {
		return fmt.Errorf("unsupported order evidence submission snapshot status %q", s.Status)
	}
	if s.SchemaVersion != OrderEvidenceSubmissionSnapshotSchemaVersion {
		return fmt.Errorf("unsupported order evidence submission snapshot schema version %d", s.SchemaVersion)
	}
	if s.LockedAt.IsZero() {
		return errors.New("order evidence submission snapshot locked_at is required")
	}
	if len(s.SnapshotData) == 0 || !json.Valid(s.SnapshotData) {
		return errors.New("order evidence submission snapshot data must be valid JSON")
	}
	if len(strings.TrimSpace(s.SnapshotSHA256)) != sha256.Size*2 {
		return errors.New("order evidence submission snapshot sha256 is invalid")
	}
	return s.VerifyIntegrity()
}

func (s OrderEvidenceSubmissionSnapshot) VerifyIntegrity() error {
	hash := sha256.Sum256(s.SnapshotData)
	expected := hex.EncodeToString(hash[:])
	if !strings.EqualFold(strings.TrimSpace(s.SnapshotSHA256), expected) {
		return errors.New("order evidence submission snapshot sha256 does not match snapshot data")
	}
	return nil
}

func (s *OrderEvidenceSubmissionSnapshot) BeforeCreate(tx *gorm.DB) error {
	if s == nil {
		return errors.New("order evidence submission snapshot is required")
	}
	s.Provider = strings.ToLower(strings.TrimSpace(s.Provider))
	s.Status = strings.ToLower(strings.TrimSpace(s.Status))
	return s.Validate()
}

func (s *OrderEvidenceSubmissionSnapshot) BeforeUpdate(tx *gorm.DB) error {
	return ErrOrderEvidenceSubmissionSnapshotImmutable
}

func (s *OrderEvidenceSubmissionSnapshot) BeforeDelete(tx *gorm.DB) error {
	return ErrOrderEvidenceSubmissionSnapshotImmutable
}
