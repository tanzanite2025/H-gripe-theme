package repository

import (
	"errors"
	"strings"

	"commerce-platform/internal/domain/orderevidence"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderEvidenceSubmissionSnapshotRepository struct {
	db *gorm.DB
}

func NewOrderEvidenceSubmissionSnapshotRepository(db *gorm.DB) *OrderEvidenceSubmissionSnapshotRepository {
	return &OrderEvidenceSubmissionSnapshotRepository{db: db}
}

func (r *OrderEvidenceSubmissionSnapshotRepository) WithTx(tx *gorm.DB) *OrderEvidenceSubmissionSnapshotRepository {
	return &OrderEvidenceSubmissionSnapshotRepository{db: tx}
}

func (r *OrderEvidenceSubmissionSnapshotRepository) Create(
	snapshot *orderevidence.OrderEvidenceSubmissionSnapshot,
) error {
	if r == nil || r.db == nil {
		return gorm.ErrInvalidDB
	}
	if snapshot == nil {
		return errors.New("order evidence submission snapshot is required")
	}
	return r.db.Create(snapshot).Error
}

// CreateOrGetLatest creates version one once and returns the existing immutable
// version for concurrent retries. A later revision must be an explicit new
// record; this method never overwrites a locked snapshot.
func (r *OrderEvidenceSubmissionSnapshotRepository) CreateOrGetLatest(
	snapshot *orderevidence.OrderEvidenceSubmissionSnapshot,
) (*orderevidence.OrderEvidenceSubmissionSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if snapshot == nil {
		return nil, errors.New("order evidence submission snapshot is required")
	}
	snapshot.Provider = strings.ToLower(strings.TrimSpace(snapshot.Provider))
	if snapshot.Version == 0 {
		snapshot.Version = 1
	}
	if err := snapshot.Validate(); err != nil {
		return nil, err
	}

	var result *orderevidence.OrderEvidenceSubmissionSnapshot
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var existing orderevidence.OrderEvidenceSubmissionSnapshot
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("provider = ? AND dispute_id = ?", snapshot.Provider, snapshot.DisputeID).
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

		var canonical orderevidence.OrderEvidenceSubmissionSnapshot
		if findErr := tx.Where(
			"provider = ? AND dispute_id = ?",
			snapshot.Provider,
			snapshot.DisputeID,
		).Order("version DESC, id DESC").First(&canonical).Error; findErr != nil {
			return findErr
		}
		if validateErr := canonical.Validate(); validateErr != nil {
			return validateErr
		}
		result = &canonical
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *OrderEvidenceSubmissionSnapshotRepository) FindLatestByDispute(
	provider string,
	disputeID uint,
) (*orderevidence.OrderEvidenceSubmissionSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" || disputeID == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var snapshot orderevidence.OrderEvidenceSubmissionSnapshot
	err := r.db.Where("provider = ? AND dispute_id = ?", provider, disputeID).
		Order("version DESC, id DESC").
		First(&snapshot).Error
	if err != nil {
		return nil, err
	}
	return &snapshot, nil
}

func (r *OrderEvidenceSubmissionSnapshotRepository) FindByID(
	id uint,
) (*orderevidence.OrderEvidenceSubmissionSnapshot, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if id == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var snapshot orderevidence.OrderEvidenceSubmissionSnapshot
	if err := r.db.First(&snapshot, id).Error; err != nil {
		return nil, err
	}
	return &snapshot, nil
}
