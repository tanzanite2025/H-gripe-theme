package repository

import (
	sitelogodomain "commerce-platform/internal/domain/site_logo"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SiteLogoRepository struct {
	db *gorm.DB
}

func NewSiteLogoRepository(db *gorm.DB) *SiteLogoRepository {
	return &SiteLogoRepository{db: db}
}

func (r *SiteLogoRepository) Current() (*sitelogodomain.Asset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}

	var asset sitelogodomain.Asset
	err := r.db.Where("id = ?", sitelogodomain.CurrentAssetID).First(&asset).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &asset, nil
}

func (r *SiteLogoRepository) ReplaceCurrent(asset *sitelogodomain.Asset) (*sitelogodomain.Asset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}
	if asset == nil {
		return nil, gorm.ErrInvalidData
	}

	asset.ID = sitelogodomain.CurrentAssetID
	var previous *sitelogodomain.Asset
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current sitelogodomain.Asset
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", sitelogodomain.CurrentAssetID).
			First(&current).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}
		if err == nil {
			previousCopy := current
			previous = &previousCopy
		}

		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			UpdateAll: true,
		}).Create(asset).Error
	})
	if err != nil {
		return nil, err
	}
	return previous, nil
}

func (r *SiteLogoRepository) DeleteCurrent() (*sitelogodomain.Asset, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrInvalidDB
	}

	var previous *sitelogodomain.Asset
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var current sitelogodomain.Asset
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", sitelogodomain.CurrentAssetID).
			First(&current).Error
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		if err != nil {
			return err
		}

		previousCopy := current
		previous = &previousCopy
		return tx.Delete(&current).Error
	})
	if err != nil {
		return nil, err
	}
	return previous, nil
}
