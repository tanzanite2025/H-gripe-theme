package repository

import (
	"strings"
	"tanzanite/internal/domain/spoke"

	"gorm.io/gorm"
)

type SpokeRepository struct {
	db *gorm.DB
}

func NewSpokeRepository(db *gorm.DB) *SpokeRepository {
	return &SpokeRepository{db: db}
}

func (r *SpokeRepository) ListHistory(search string, page, pageSize int) ([]spoke.History, int64, error) {
	var items []spoke.History
	var total int64

	query := r.db.Model(&spoke.History{})
	search = strings.ToLower(strings.TrimSpace(search))
	if search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"LOWER(COALESCE(rim_brand, '')) LIKE ? OR LOWER(COALESCE(rim_model, '')) LIKE ? OR LOWER(COALESCE(hub_brand, '')) LIKE ? OR LOWER(COALESCE(hub_model, '')) LIKE ?",
			like,
			like,
			like,
			like,
		)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}
