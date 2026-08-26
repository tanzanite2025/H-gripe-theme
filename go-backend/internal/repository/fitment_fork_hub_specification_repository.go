package repository

import (
	"errors"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"

	"gorm.io/gorm"
)

type FitmentForkHubSpecificationRepository struct {
	db *gorm.DB
}

type fitmentForkReferenceCount struct {
	HubSpecificationID uint
	Count              int64
}

type fitmentForkHubReferenceCount struct {
	ForkEntryID uint
	Count       int64
}

func NewFitmentForkHubSpecificationRepository(db *gorm.DB) *FitmentForkHubSpecificationRepository {
	return &FitmentForkHubSpecificationRepository{db: db}
}

func (r *FitmentForkHubSpecificationRepository) WithTx(tx *gorm.DB) *FitmentForkHubSpecificationRepository {
	return &FitmentForkHubSpecificationRepository{db: tx}
}

func (r *FitmentForkHubSpecificationRepository) ListForFork(forkEntryID uint) ([]fitmentcatalogdomain.HubSpecification, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fitment fork hub specification repository is unavailable")
	}

	var specifications []fitmentcatalogdomain.HubSpecification
	err := r.db.
		Table("fitment_hub_specifications AS hub").
		Select("hub.*").
		Joins("JOIN fitment_fork_hub_specifications AS link ON link.hub_specification_id = hub.id").
		Where("link.fork_entry_id = ?", forkEntryID).
		Where("hub.deleted_at IS NULL").
		Order("hub.sort_order ASC").
		Order("hub.id ASC").
		Find(&specifications).Error
	return specifications, err
}

func (r *FitmentForkHubSpecificationRepository) ReplaceForkSpecifications(
	forkEntryID uint,
	specificationIDs []uint,
) error {
	if r == nil || r.db == nil {
		return errors.New("fitment fork hub specification repository is unavailable")
	}

	if err := r.db.
		Where("fork_entry_id = ?", forkEntryID).
		Delete(&fitmentcatalogdomain.ForkHubSpecification{}).Error; err != nil {
		return err
	}
	if len(specificationIDs) == 0 {
		return nil
	}

	links := make([]fitmentcatalogdomain.ForkHubSpecification, 0, len(specificationIDs))
	for _, specificationID := range specificationIDs {
		links = append(links, fitmentcatalogdomain.ForkHubSpecification{
			ForkEntryID:        forkEntryID,
			HubSpecificationID: specificationID,
		})
	}
	return r.db.Create(&links).Error
}

func (r *FitmentForkHubSpecificationRepository) ForkReferenceCount(id uint) (int64, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("fitment fork hub specification repository is unavailable")
	}

	var count int64
	err := r.db.
		Table("fitment_fork_hub_specifications AS link").
		Joins("JOIN fitment_fork_entries AS fork ON fork.id = link.fork_entry_id").
		Where("link.hub_specification_id = ?", id).
		Where("fork.deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

func (r *FitmentForkHubSpecificationRepository) CountForkReferencesByHubSpecificationIDs(
	ids []uint,
) (map[uint]int, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fitment fork hub specification repository is unavailable")
	}

	counts := make(map[uint]int, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}

	var rows []fitmentForkReferenceCount
	if err := r.db.
		Table("fitment_fork_hub_specifications AS link").
		Select("link.hub_specification_id, COUNT(*) AS count").
		Joins("JOIN fitment_fork_entries AS fork ON fork.id = link.fork_entry_id").
		Where("link.hub_specification_id IN ?", ids).
		Where("fork.deleted_at IS NULL").
		Group("link.hub_specification_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.HubSpecificationID] = int(row.Count)
	}
	return counts, nil
}

func (r *FitmentForkHubSpecificationRepository) CountForkReferencesByForkIDs(
	forkEntryIDs []uint,
) (map[uint]int, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fitment fork hub specification repository is unavailable")
	}

	counts := make(map[uint]int, len(forkEntryIDs))
	if len(forkEntryIDs) == 0 {
		return counts, nil
	}

	var rows []fitmentForkHubReferenceCount
	if err := r.db.
		Table("fitment_fork_hub_specifications AS link").
		Select("link.fork_entry_id, COUNT(*) AS count").
		Joins("JOIN fitment_hub_specifications AS hub ON hub.id = link.hub_specification_id").
		Where("link.fork_entry_id IN ?", forkEntryIDs).
		Where("hub.deleted_at IS NULL").
		Group("link.fork_entry_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.ForkEntryID] = int(row.Count)
	}
	return counts, nil
}
