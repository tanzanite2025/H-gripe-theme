package repository

import (
	"errors"

	"commerce-platform/internal/domain/orderevidence"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderEvidenceExportSnapshotRepository struct {
	db *gorm.DB
}

func NewOrderEvidenceExportSnapshotRepository(db *gorm.DB) *OrderEvidenceExportSnapshotRepository {
	return &OrderEvidenceExportSnapshotRepository{db: db}
}

func (r *OrderEvidenceExportSnapshotRepository) WithTx(tx *gorm.DB) *OrderEvidenceExportSnapshotRepository {
	return &OrderEvidenceExportSnapshotRepository{db: tx}
}

// CreateOrGetForPackage creates the first immutable export manifest for a
// package version and returns the canonical row when concurrent requests race.
func (r *OrderEvidenceExportSnapshotRepository) CreateOrGetForPackage(
	snapshot *orderevidence.OrderEvidenceExportSnapshot,
) (*orderevidence.OrderEvidenceExportSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if snapshot == nil {
		return nil, errors.New("order evidence export snapshot is required")
	}
	if snapshot.Version == 0 {
		snapshot.Version = 1
	}
	if err := snapshot.Validate(); err != nil {
		return nil, err
	}

	var result *orderevidence.OrderEvidenceExportSnapshot
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existing orderevidence.OrderEvidenceExportSnapshot
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("evidence_package_id = ?", snapshot.EvidencePackageID).
			Order("version DESC, id DESC").
			First(&existing).Error
		if err == nil {
			if validateErr := existing.Validate(); validateErr != nil {
				return validateErr
			}
			result = &existing
			return nil
		}
		if !IsRecordNotFound(err) {
			return err
		}

		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(snapshot).Error; err != nil {
			return err
		}

		var canonical orderevidence.OrderEvidenceExportSnapshot
		if err := tx.Where(
			"evidence_package_id = ?",
			snapshot.EvidencePackageID,
		).Order("version DESC, id DESC").First(&canonical).Error; err != nil {
			return err
		}
		if err := canonical.Validate(); err != nil {
			return err
		}
		result = &canonical
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *OrderEvidenceExportSnapshotRepository) FindByPackageID(
	packageID uint,
) (*orderevidence.OrderEvidenceExportSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if packageID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var snapshot orderevidence.OrderEvidenceExportSnapshot
	if err := r.db.
		Where("evidence_package_id = ?", packageID).
		Order("version DESC, id DESC").
		First(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *OrderEvidenceExportSnapshotRepository) FindLatestByOrderID(
	orderID uint,
) (*orderevidence.OrderEvidenceExportSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if orderID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var snapshot orderevidence.OrderEvidenceExportSnapshot
	if err := r.db.
		Where("order_id = ?", orderID).
		Order("evidence_package_version DESC, version DESC, id DESC").
		First(&snapshot).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}
