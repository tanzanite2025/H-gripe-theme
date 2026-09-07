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
	OrderEvidenceExportSnapshotSchemaVersion = 1
	ExportSnapshotStatusLocked               = "locked"
)

var ErrOrderEvidenceExportSnapshotImmutable = errors.New("order evidence export snapshot is immutable")

// OrderEvidenceExportSnapshot freezes the admin-export manifest for one locked
// evidence package version. The manifest is separate from payment-provider
// submission snapshots because an order export has no dispute or provider.
type OrderEvidenceExportSnapshot struct {
	ID                     uint           `gorm:"primaryKey" json:"id"`
	OrderID                uint           `gorm:"not null;index" json:"order_id"`
	EvidencePackageID      uint           `gorm:"not null;index;uniqueIndex:uq_order_evidence_export_snapshot_version" json:"evidence_package_id"`
	EvidencePackageVersion int            `gorm:"not null" json:"evidence_package_version"`
	Version                int            `gorm:"not null;uniqueIndex:uq_order_evidence_export_snapshot_version" json:"version"`
	Status                 string         `gorm:"size:16;not null;index" json:"status"`
	SchemaVersion          int            `gorm:"not null" json:"schema_version"`
	LockedAt               time.Time      `gorm:"not null" json:"locked_at"`
	SnapshotData           datatypes.JSON `gorm:"column:snapshot_data;type:jsonb;not null" json:"-"`
	SnapshotSHA256         string         `gorm:"column:snapshot_sha256;type:char(64);not null" json:"snapshot_sha256"`
	CreatedBy              uint           `gorm:"not null;default:0" json:"created_by"`
	CreatedAt              time.Time      `json:"created_at"`
}

func (OrderEvidenceExportSnapshot) TableName() string {
	return "order_evidence_export_snapshots"
}

func (s OrderEvidenceExportSnapshot) Validate() error {
	if s.OrderID == 0 {
		return errors.New("order evidence export snapshot order_id is required")
	}
	if s.EvidencePackageID == 0 {
		return errors.New("order evidence export snapshot evidence_package_id is required")
	}
	if s.EvidencePackageVersion <= 0 {
		return errors.New("order evidence export snapshot evidence_package_version must be greater than zero")
	}
	if s.Version <= 0 {
		return errors.New("order evidence export snapshot version must be greater than zero")
	}
	if s.Status != ExportSnapshotStatusLocked {
		return fmt.Errorf("unsupported order evidence export snapshot status %q", s.Status)
	}
	if s.SchemaVersion != OrderEvidenceExportSnapshotSchemaVersion {
		return fmt.Errorf("unsupported order evidence export snapshot schema version %d", s.SchemaVersion)
	}
	if s.LockedAt.IsZero() {
		return errors.New("order evidence export snapshot locked_at is required")
	}
	if len(s.SnapshotData) == 0 || !json.Valid(s.SnapshotData) {
		return errors.New("order evidence export snapshot data must be valid JSON")
	}
	if len(strings.TrimSpace(s.SnapshotSHA256)) != sha256.Size*2 {
		return errors.New("order evidence export snapshot sha256 is invalid")
	}
	return s.VerifyIntegrity()
}

func (s OrderEvidenceExportSnapshot) VerifyIntegrity() error {
	hash := sha256.Sum256(s.SnapshotData)
	expected := hex.EncodeToString(hash[:])
	if !strings.EqualFold(strings.TrimSpace(s.SnapshotSHA256), expected) {
		return errors.New("order evidence export snapshot sha256 does not match snapshot data")
	}
	return nil
}

func (s *OrderEvidenceExportSnapshot) BeforeCreate(tx *gorm.DB) error {
	if s == nil {
		return errors.New("order evidence export snapshot is required")
	}
	s.Status = strings.ToLower(strings.TrimSpace(s.Status))
	return s.Validate()
}

func (s *OrderEvidenceExportSnapshot) BeforeUpdate(tx *gorm.DB) error {
	return ErrOrderEvidenceExportSnapshotImmutable
}

func (s *OrderEvidenceExportSnapshot) BeforeDelete(tx *gorm.DB) error {
	return ErrOrderEvidenceExportSnapshotImmutable
}
