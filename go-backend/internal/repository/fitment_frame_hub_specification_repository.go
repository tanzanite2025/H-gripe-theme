package repository

import (
	"errors"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"

	"gorm.io/gorm"
)

type FitmentFrameHubSpecificationRepository struct {
	db *gorm.DB
}

type fitmentFrameReferenceCount struct {
	HubSpecificationID uint
	Count              int64
}

type fitmentFrameHubReferenceCount struct {
	FrameEntryID uint
	Count        int64
}

func NewFitmentFrameHubSpecificationRepository(db *gorm.DB) *FitmentFrameHubSpecificationRepository {
	return &FitmentFrameHubSpecificationRepository{db: db}
}

func (r *FitmentFrameHubSpecificationRepository) WithTx(tx *gorm.DB) *FitmentFrameHubSpecificationRepository {
	return &FitmentFrameHubSpecificationRepository{db: tx}
}

func (r *FitmentFrameHubSpecificationRepository) ListForFrame(frameEntryID uint) ([]fitmentcatalogdomain.HubSpecification, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fitment frame hub specification repository is unavailable")
	}

	var specifications []fitmentcatalogdomain.HubSpecification
	err := r.db.
		Table("fitment_hub_specifications AS hub").
		Select("hub.*").
		Joins("JOIN fitment_frame_hub_specifications AS link ON link.hub_specification_id = hub.id").
		Where("link.frame_entry_id = ?", frameEntryID).
		Where("hub.deleted_at IS NULL").
		Order("hub.sort_order ASC").
		Order("hub.id ASC").
		Find(&specifications).Error
	return specifications, err
}

func (r *FitmentFrameHubSpecificationRepository) ListFrameSpecificationIDs(frameEntryID uint) ([]uint, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fitment frame hub specification repository is unavailable")
	}

	var ids []uint
	err := r.db.
		Table("fitment_frame_hub_specifications").
		Where("frame_entry_id = ?", frameEntryID).
		Order("hub_specification_id ASC").
		Pluck("hub_specification_id", &ids).Error
	return ids, err
}

func (r *FitmentFrameHubSpecificationRepository) ReplaceFrameSpecifications(
	frameEntryID uint,
	specificationIDs []uint,
) error {
	if r == nil || r.db == nil {
		return errors.New("fitment frame hub specification repository is unavailable")
	}

	if err := r.db.
		Where("frame_entry_id = ?", frameEntryID).
		Delete(&fitmentcatalogdomain.FrameHubSpecification{}).Error; err != nil {
		return err
	}
	if len(specificationIDs) == 0 {
		return nil
	}

	links := make([]fitmentcatalogdomain.FrameHubSpecification, 0, len(specificationIDs))
	for _, specificationID := range specificationIDs {
		links = append(links, fitmentcatalogdomain.FrameHubSpecification{
			FrameEntryID:       frameEntryID,
			HubSpecificationID: specificationID,
		})
	}
	return r.db.Create(&links).Error
}

func (r *FitmentFrameHubSpecificationRepository) FrameReferenceCount(id uint) (int64, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("fitment frame hub specification repository is unavailable")
	}

	var count int64
	err := r.db.
		Table("fitment_frame_hub_specifications AS link").
		Joins("JOIN fitment_frame_entries AS frame ON frame.id = link.frame_entry_id").
		Where("link.hub_specification_id = ?", id).
		Where("frame.deleted_at IS NULL").
		Count(&count).Error
	return count, err
}

func (r *FitmentFrameHubSpecificationRepository) CountFrameReferencesByHubSpecificationIDs(
	ids []uint,
) (map[uint]int, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fitment frame hub specification repository is unavailable")
	}

	counts := make(map[uint]int, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}

	var rows []fitmentFrameReferenceCount
	if err := r.db.
		Table("fitment_frame_hub_specifications AS link").
		Select("link.hub_specification_id, COUNT(*) AS count").
		Joins("JOIN fitment_frame_entries AS frame ON frame.id = link.frame_entry_id").
		Where("link.hub_specification_id IN ?", ids).
		Where("frame.deleted_at IS NULL").
		Group("link.hub_specification_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.HubSpecificationID] = int(row.Count)
	}
	return counts, nil
}

func (r *FitmentFrameHubSpecificationRepository) CountFrameReferencesByFrameIDs(
	frameEntryIDs []uint,
) (map[uint]int, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fitment frame hub specification repository is unavailable")
	}

	counts := make(map[uint]int, len(frameEntryIDs))
	if len(frameEntryIDs) == 0 {
		return counts, nil
	}

	var rows []fitmentFrameHubReferenceCount
	if err := r.db.
		Table("fitment_frame_hub_specifications AS link").
		Select("link.frame_entry_id, COUNT(*) AS count").
		Joins("JOIN fitment_hub_specifications AS hub ON hub.id = link.hub_specification_id").
		Where("link.frame_entry_id IN ?", frameEntryIDs).
		Where("hub.deleted_at IS NULL").
		Group("link.frame_entry_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	for _, row := range rows {
		counts[row.FrameEntryID] = int(row.Count)
	}
	return counts, nil
}
