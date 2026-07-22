package repository

import (
	"tanzanite/internal/domain/media"

	"gorm.io/gorm"
)

type MediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) CreateAsset(asset *media.MediaAsset) error {
	return r.db.Create(asset).Error
}

func (r *MediaRepository) FindAssetByID(id uint) (*media.MediaAsset, error) {
	var asset media.MediaAsset
	if err := r.db.First(&asset, id).Error; err != nil {
		return nil, err
	}
	return &asset, nil
}
