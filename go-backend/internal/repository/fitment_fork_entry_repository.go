package repository

import (
	"errors"
	"strings"
	"time"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"

	"gorm.io/gorm"
)

type ForkFitmentEntryRepository struct {
	db *gorm.DB
}

type ForkFitmentEntryFilter struct {
	Search    string
	IsEnabled *bool
	Year      *int
}

func NewForkFitmentEntryRepository(db *gorm.DB) *ForkFitmentEntryRepository {
	return &ForkFitmentEntryRepository{db: db}
}

func (r *ForkFitmentEntryRepository) List(
	page int,
	pageSize int,
	filter ForkFitmentEntryFilter,
) ([]fitmentcatalogdomain.ForkFitmentEntry, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("fork fitment repository is unavailable")
	}

	var total int64
	query := r.db.Model(&fitmentcatalogdomain.ForkFitmentEntry{})

	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(brand_name) LIKE ? OR LOWER(model_name) LIKE ? OR LOWER(series_name) LIKE ? OR LOWER(generation_name) LIKE ? OR LOWER(market_code) LIKE ?",
			like, like, like, like, like,
		)
	}
	if filter.IsEnabled != nil {
		query = query.Where("is_enabled = ?", *filter.IsEnabled)
	}
	if filter.Year != nil {
		year := *filter.Year
		query = query.Where(
			"(year_mode = ? AND year_from = ?) OR (year_mode = ? AND year_from <= ? AND year_to >= ?) OR year_mode = ?",
			fitmentcatalogdomain.YearModeSingle,
			year,
			fitmentcatalogdomain.YearModeRange,
			year,
			year,
			fitmentcatalogdomain.YearModeAll,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	var entries []fitmentcatalogdomain.ForkFitmentEntry
	if err := query.
		Order("sort_order ASC").
		Order("updated_at DESC").
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&entries).Error; err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (r *ForkFitmentEntryRepository) FindByID(id uint) (*fitmentcatalogdomain.ForkFitmentEntry, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fork fitment repository is unavailable")
	}

	var entry fitmentcatalogdomain.ForkFitmentEntry
	if err := r.db.First(&entry, id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *ForkFitmentEntryRepository) FindDuplicate(
	entry *fitmentcatalogdomain.ForkFitmentEntry,
	excludeID uint,
) (*fitmentcatalogdomain.ForkFitmentEntry, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("fork fitment repository is unavailable")
	}

	query := r.db.
		Where("LOWER(brand_name) = LOWER(?)", entry.BrandName).
		Where("LOWER(model_name) = LOWER(?)", entry.ModelName).
		Where("LOWER(COALESCE(series_name, '')) = LOWER(?)", entry.SeriesName).
		Where("LOWER(COALESCE(generation_name, '')) = LOWER(?)", entry.GenerationName).
		Where("year_mode = ?", entry.YearMode).
		Where("COALESCE(year_from, 0) = COALESCE(?, 0)", entry.YearFrom).
		Where("COALESCE(year_to, 0) = COALESCE(?, 0)", entry.YearTo).
		Where("LOWER(COALESCE(market_code, '')) = LOWER(?)", entry.MarketCode)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var duplicate fitmentcatalogdomain.ForkFitmentEntry
	if err := query.First(&duplicate).Error; err != nil {
		return nil, err
	}
	return &duplicate, nil
}

func (r *ForkFitmentEntryRepository) Create(entry *fitmentcatalogdomain.ForkFitmentEntry) error {
	if r == nil || r.db == nil {
		return errors.New("fork fitment repository is unavailable")
	}
	return r.db.Select("*").Create(entry).Error
}

func (r *ForkFitmentEntryRepository) Update(entry *fitmentcatalogdomain.ForkFitmentEntry) error {
	if r == nil || r.db == nil {
		return errors.New("fork fitment repository is unavailable")
	}
	return r.db.Model(&fitmentcatalogdomain.ForkFitmentEntry{}).
		Where("id = ?", entry.ID).
		Updates(map[string]interface{}{
			"brand_name":      entry.BrandName,
			"model_name":      entry.ModelName,
			"series_name":     entry.SeriesName,
			"generation_name": entry.GenerationName,
			"year_mode":       entry.YearMode,
			"year_from":       entry.YearFrom,
			"year_to":         entry.YearTo,
			"market_code":     entry.MarketCode,
			"notes":           entry.Notes,
			"is_enabled":      entry.IsEnabled,
			"sort_order":      entry.SortOrder,
			"updated_at":      time.Now(),
		}).Error
}

func (r *ForkFitmentEntryRepository) UpdateStatus(id uint, enabled bool) error {
	if r == nil || r.db == nil {
		return errors.New("fork fitment repository is unavailable")
	}
	return r.db.Model(&fitmentcatalogdomain.ForkFitmentEntry{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_enabled": enabled,
			"updated_at": time.Now(),
		}).Error
}

func (r *ForkFitmentEntryRepository) Delete(id uint) error {
	if r == nil || r.db == nil {
		return errors.New("fork fitment repository is unavailable")
	}
	return r.db.Delete(&fitmentcatalogdomain.ForkFitmentEntry{}, id).Error
}

func (r *ForkFitmentEntryRepository) Transaction(fn func(*gorm.DB) error) error {
	if r == nil || r.db == nil {
		return errors.New("fork fitment repository is unavailable")
	}
	return r.db.Transaction(fn)
}

func (r *ForkFitmentEntryRepository) WithTx(tx *gorm.DB) *ForkFitmentEntryRepository {
	return &ForkFitmentEntryRepository{db: tx}
}
