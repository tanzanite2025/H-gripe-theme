package repository

import (
	"errors"
	"strings"
	"time"

	fitmentcatalogdomain "commerce-platform/internal/domain/fitmentcatalog"

	"gorm.io/gorm"
)

type FrameFitmentEntryRepository struct {
	db *gorm.DB
}

type FrameFitmentEntryFilter struct {
	Search    string
	IsEnabled *bool
	Year      *int
}

func NewFrameFitmentEntryRepository(db *gorm.DB) *FrameFitmentEntryRepository {
	return &FrameFitmentEntryRepository{db: db}
}

func (r *FrameFitmentEntryRepository) List(
	page int,
	pageSize int,
	filter FrameFitmentEntryFilter,
) ([]fitmentcatalogdomain.FrameFitmentEntry, int64, error) {
	var total int64
	query := r.db.Model(&fitmentcatalogdomain.FrameFitmentEntry{})

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

	var entries []fitmentcatalogdomain.FrameFitmentEntry
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

func (r *FrameFitmentEntryRepository) FindByID(id uint) (*fitmentcatalogdomain.FrameFitmentEntry, error) {
	var entry fitmentcatalogdomain.FrameFitmentEntry
	if err := r.db.First(&entry, id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *FrameFitmentEntryRepository) FindDuplicate(
	entry *fitmentcatalogdomain.FrameFitmentEntry,
	excludeID uint,
) (*fitmentcatalogdomain.FrameFitmentEntry, error) {
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

	var duplicate fitmentcatalogdomain.FrameFitmentEntry
	if err := query.First(&duplicate).Error; err != nil {
		return nil, err
	}
	return &duplicate, nil
}

func (r *FrameFitmentEntryRepository) Create(entry *fitmentcatalogdomain.FrameFitmentEntry) error {
	return r.db.Select("*").Create(entry).Error
}

func (r *FrameFitmentEntryRepository) Update(entry *fitmentcatalogdomain.FrameFitmentEntry) error {
	return r.db.Model(&fitmentcatalogdomain.FrameFitmentEntry{}).
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

func (r *FrameFitmentEntryRepository) UpdateStatus(id uint, enabled bool) error {
	return r.db.Model(&fitmentcatalogdomain.FrameFitmentEntry{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_enabled": enabled,
			"updated_at": time.Now(),
		}).Error
}

func (r *FrameFitmentEntryRepository) Delete(id uint) error {
	return r.db.Delete(&fitmentcatalogdomain.FrameFitmentEntry{}, id).Error
}

func (r *FrameFitmentEntryRepository) Transaction(fn func(*gorm.DB) error) error {
	if r == nil || r.db == nil {
		return errors.New("frame fitment repository is unavailable")
	}
	return r.db.Transaction(fn)
}

func (r *FrameFitmentEntryRepository) WithTx(tx *gorm.DB) *FrameFitmentEntryRepository {
	return &FrameFitmentEntryRepository{db: tx}
}
