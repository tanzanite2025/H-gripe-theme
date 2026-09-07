package orderevidence

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	OrderEvidencePackageSchemaVersion = 1

	PackageStatusIncomplete = "incomplete"
	PackageStatusReady      = "ready"
	PackageStatusLocked     = "locked"
	PackageStatusSuperseded = "superseded"
)

// OrderEvidencePackage is a versioned, order-scoped container for structured
// fulfillment evidence. It is separate from the immutable order snapshot.
type OrderEvidencePackage struct {
	ID                    uint                `gorm:"primaryKey" json:"id"`
	OrderID               uint                `gorm:"not null;index" json:"order_id"`
	SnapshotID            uint                `gorm:"not null;index" json:"snapshot_id"`
	PackageVersion        int                 `gorm:"not null" json:"package_version"`
	Status                string              `gorm:"size:20;not null;index" json:"status"`
	OrderTotalUSDSnapshot float64             `gorm:"column:order_total_usd_snapshot;not null" json:"order_total_usd_snapshot"`
	IsHighValue           bool                `gorm:"not null;index" json:"is_high_value"`
	HasSpokeTensionQC     bool                `gorm:"not null;index" json:"has_spoke_tension_qc"`
	SchemaVersion         int                 `gorm:"not null" json:"schema_version"`
	CreatedBy             uint                `gorm:"not null;default:0" json:"created_by"`
	LockedAt              *time.Time          `json:"locked_at,omitempty"`
	Items                 []OrderEvidenceItem `gorm:"foreignKey:PackageID" json:"items,omitempty"`
	CreatedAt             time.Time           `json:"created_at"`
	UpdatedAt             time.Time           `json:"updated_at"`
}

func (OrderEvidencePackage) TableName() string {
	return "order_evidence_packages"
}

func (p OrderEvidencePackage) Validate() error {
	if p.OrderID == 0 {
		return errors.New("order evidence package order_id is required")
	}
	if p.SnapshotID == 0 {
		return errors.New("order evidence package snapshot_id is required")
	}
	if p.PackageVersion <= 0 {
		return errors.New("order evidence package package_version must be greater than zero")
	}
	status := strings.ToLower(strings.TrimSpace(p.Status))
	switch status {
	case PackageStatusIncomplete, PackageStatusReady, PackageStatusLocked, PackageStatusSuperseded:
	default:
		return fmt.Errorf("unsupported order evidence package status %q", p.Status)
	}
	if p.SchemaVersion != OrderEvidencePackageSchemaVersion {
		return fmt.Errorf("unsupported order evidence package schema version %d", p.SchemaVersion)
	}
	if math.IsNaN(p.OrderTotalUSDSnapshot) || math.IsInf(p.OrderTotalUSDSnapshot, 0) || p.OrderTotalUSDSnapshot < 0 {
		return errors.New("order evidence package order_total_usd_snapshot must be finite and non-negative")
	}
	if status == PackageStatusLocked && p.LockedAt == nil {
		return errors.New("locked order evidence package requires locked_at")
	}
	if status != PackageStatusLocked && p.LockedAt != nil {
		return errors.New("unlocked order evidence package cannot have locked_at")
	}
	return nil
}

func (p *OrderEvidencePackage) BeforeCreate(tx *gorm.DB) error {
	if p == nil {
		return errors.New("order evidence package is required")
	}
	p.Status = strings.ToLower(strings.TrimSpace(p.Status))
	return p.Validate()
}

func (p *OrderEvidencePackage) BeforeSave(tx *gorm.DB) error {
	if p == nil {
		return errors.New("order evidence package is required")
	}
	p.Status = strings.ToLower(strings.TrimSpace(p.Status))
	return p.Validate()
}
