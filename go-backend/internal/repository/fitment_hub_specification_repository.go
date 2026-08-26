package repository

import (
	"errors"
	"strings"
	"time"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"

	"gorm.io/gorm"
)

type FitmentHubSpecificationRepository struct {
	db                    *gorm.DB
	frameHubSpecification *FitmentFrameHubSpecificationRepository
	forkHubSpecification  *FitmentForkHubSpecificationRepository
}

type FitmentHubSpecificationFilter struct {
	Search    string
	Position  *fitmentcatalogdomain.HubPosition
	AxleType  *fitmentcatalogdomain.HubAxleType
	IsEnabled *bool
}

func NewFitmentHubSpecificationRepository(
	db *gorm.DB,
	frameHubRepositories ...*FitmentFrameHubSpecificationRepository,
) *FitmentHubSpecificationRepository {
	var frameHubRepository *FitmentFrameHubSpecificationRepository
	if len(frameHubRepositories) > 0 {
		frameHubRepository = frameHubRepositories[0]
	}
	if frameHubRepository == nil {
		frameHubRepository = NewFitmentFrameHubSpecificationRepository(db)
	}
	return &FitmentHubSpecificationRepository{
		db:                    db,
		frameHubSpecification: frameHubRepository,
	}
}

func (r *FitmentHubSpecificationRepository) ConfigureForkHubSpecificationRepository(
	forkHubRepository *FitmentForkHubSpecificationRepository,
) {
	if r == nil {
		return
	}
	r.forkHubSpecification = forkHubRepository
}

func (r *FitmentHubSpecificationRepository) WithTx(tx *gorm.DB) *FitmentHubSpecificationRepository {
	var frameHubRepository *FitmentFrameHubSpecificationRepository
	if r != nil && r.frameHubSpecification != nil {
		frameHubRepository = r.frameHubSpecification.WithTx(tx)
	}
	if frameHubRepository == nil {
		frameHubRepository = NewFitmentFrameHubSpecificationRepository(tx)
	}
	var forkHubRepository *FitmentForkHubSpecificationRepository
	if r != nil && r.forkHubSpecification != nil {
		forkHubRepository = r.forkHubSpecification.WithTx(tx)
	}
	return &FitmentHubSpecificationRepository{
		db:                    tx,
		frameHubSpecification: frameHubRepository,
		forkHubSpecification:  forkHubRepository,
	}
}

func (r *FitmentHubSpecificationRepository) FrameHubSpecificationRepository() *FitmentFrameHubSpecificationRepository {
	if r == nil {
		return nil
	}
	return r.frameHubSpecification
}

func (r *FitmentHubSpecificationRepository) List(
	page int,
	pageSize int,
	filter FitmentHubSpecificationFilter,
) ([]fitmentcatalogdomain.HubSpecification, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("fitment hub specification repository is unavailable")
	}

	query := r.db.Model(&fitmentcatalogdomain.HubSpecification{})
	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(spec_code) LIKE ? OR LOWER(display_name) LIKE ? OR LOWER(axle_type) LIKE ?",
			like, like, like,
		)
	}
	if filter.Position != nil {
		query = query.Where("position = ?", *filter.Position)
	}
	if filter.AxleType != nil {
		query = query.Where("axle_type = ?", *filter.AxleType)
	}
	if filter.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *filter.IsEnabled)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var specifications []fitmentcatalogdomain.HubSpecification
	if err := query.
		Order("sort_order ASC").
		Order("updated_at DESC").
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&specifications).Error; err != nil {
		return nil, 0, err
	}

	if err := r.attachFrameReferenceCounts(specifications); err != nil {
		return nil, 0, err
	}

	return specifications, total, nil
}

func (r *FitmentHubSpecificationRepository) FindByID(id uint) (*fitmentcatalogdomain.HubSpecification, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fitment hub specification repository is unavailable")
	}

	var specification fitmentcatalogdomain.HubSpecification
	if err := r.db.First(&specification, id).Error; err != nil {
		return nil, err
	}
	specifications := []fitmentcatalogdomain.HubSpecification{specification}
	if err := r.attachFrameReferenceCounts(specifications); err != nil {
		return nil, err
	}
	return &specifications[0], nil
}

func (r *FitmentHubSpecificationRepository) FindByIDs(ids []uint) ([]fitmentcatalogdomain.HubSpecification, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fitment hub specification repository is unavailable")
	}
	if len(ids) == 0 {
		return []fitmentcatalogdomain.HubSpecification{}, nil
	}

	var specifications []fitmentcatalogdomain.HubSpecification
	if err := r.db.
		Where("id IN ?", ids).
		Order("sort_order ASC").
		Order("id ASC").
		Find(&specifications).Error; err != nil {
		return nil, err
	}
	return specifications, nil
}

func (r *FitmentHubSpecificationRepository) FindDuplicate(
	specification *fitmentcatalogdomain.HubSpecification,
	excludeID uint,
) (*fitmentcatalogdomain.HubSpecification, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fitment hub specification repository is unavailable")
	}

	query := r.db.
		Where("LOWER(TRIM(spec_code)) = LOWER(TRIM(?))", specification.SpecCode)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var duplicate fitmentcatalogdomain.HubSpecification
	if err := query.First(&duplicate).Error; err != nil {
		return nil, err
	}
	return &duplicate, nil
}

func (r *FitmentHubSpecificationRepository) Create(specification *fitmentcatalogdomain.HubSpecification) error {
	if r == nil || r.db == nil {
		return errors.New("fitment hub specification repository is unavailable")
	}
	return r.db.Select("*").Create(specification).Error
}

func (r *FitmentHubSpecificationRepository) Update(specification *fitmentcatalogdomain.HubSpecification) error {
	if r == nil || r.db == nil {
		return errors.New("fitment hub specification repository is unavailable")
	}
	return r.db.Model(&fitmentcatalogdomain.HubSpecification{}).
		Where("id = ?", specification.ID).
		Updates(map[string]interface{}{
			"spec_code":       specification.SpecCode,
			"display_name":    specification.DisplayName,
			"position":        specification.Position,
			"axle_type":       specification.AxleType,
			"axle_spacing_mm": specification.AxleSpacingMM,
			"notes":           specification.Notes,
			"is_enabled":      specification.IsEnabled,
			"sort_order":      specification.SortOrder,
			"updated_at":      time.Now(),
		}).Error
}

func (r *FitmentHubSpecificationRepository) UpdateStatus(id uint, enabled bool) error {
	if r == nil || r.db == nil {
		return errors.New("fitment hub specification repository is unavailable")
	}
	return r.db.Model(&fitmentcatalogdomain.HubSpecification{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_enabled": enabled,
			"updated_at": time.Now(),
		}).Error
}

func (r *FitmentHubSpecificationRepository) Delete(id uint) error {
	if r == nil || r.db == nil {
		return errors.New("fitment hub specification repository is unavailable")
	}
	return r.db.Delete(&fitmentcatalogdomain.HubSpecification{}, id).Error
}

func (r *FitmentHubSpecificationRepository) attachFrameReferenceCounts(
	specifications []fitmentcatalogdomain.HubSpecification,
) error {
	if r == nil || r.frameHubSpecification == nil {
		return errors.New("fitment frame hub specification repository is unavailable")
	}
	if len(specifications) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(specifications))
	for _, specification := range specifications {
		ids = append(ids, specification.ID)
	}

	counts, err := r.frameHubSpecification.CountFrameReferencesByHubSpecificationIDs(ids)
	if err != nil {
		return err
	}
	for index := range specifications {
		specifications[index].FrameReferenceCount = counts[specifications[index].ID]
	}
	if r.forkHubSpecification == nil {
		return nil
	}
	forkCounts, err := r.forkHubSpecification.CountForkReferencesByHubSpecificationIDs(ids)
	if err != nil {
		return err
	}
	for index := range specifications {
		specifications[index].ForkReferenceCount = forkCounts[specifications[index].ID]
	}
	return nil
}
