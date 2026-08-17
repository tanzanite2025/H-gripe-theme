package repository

import (
	"errors"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"

	"gorm.io/gorm"
)

type SiteQualityRunRepository struct {
	db *gorm.DB
}

type SiteQualityRunListFilter struct {
	Page      int
	PageSize  int
	TargetURL string
	Strategy  string
}

func NewSiteQualityRunRepository(db *gorm.DB) *SiteQualityRunRepository {
	return &SiteQualityRunRepository{db: db}
}

func (r *SiteQualityRunRepository) WithTx(tx *gorm.DB) *SiteQualityRunRepository {
	return &SiteQualityRunRepository{db: tx}
}

func (r *SiteQualityRunRepository) Create(run *sitequalitydomain.SiteQualityRun) error {
	if r == nil || r.db == nil {
		return errors.New("SiteQuality run repository is unavailable")
	}
	if run == nil {
		return errors.New("SiteQuality run is required")
	}
	return r.db.Create(run).Error
}

func (r *SiteQualityRunRepository) List(
	filter SiteQualityRunListFilter,
) ([]sitequalitydomain.SiteQualityRun, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("SiteQuality run repository is unavailable")
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	query := r.db.Model(&sitequalitydomain.SiteQualityRun{})
	if filter.TargetURL != "" {
		query = query.Where("target_url = ?", filter.TargetURL)
	}
	if filter.Strategy != "" {
		query = query.Where("strategy = ?", filter.Strategy)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var runs []sitequalitydomain.SiteQualityRun
	err := query.
		Order("created_at DESC").
		Order("id DESC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&runs).Error
	return runs, total, err
}

func (r *SiteQualityRunRepository) LatestSuccessfulAt() (*time.Time, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("SiteQuality run repository is unavailable")
	}
	var run sitequalitydomain.SiteQualityRun
	err := r.db.
		Where("status = ?", sitequalitydomain.SiteQualityRunStatusSuccess).
		Order("created_at DESC").
		Order("id DESC").
		First(&run).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	latestSuccessAt := run.CreatedAt.UTC()
	return &latestSuccessAt, nil
}
