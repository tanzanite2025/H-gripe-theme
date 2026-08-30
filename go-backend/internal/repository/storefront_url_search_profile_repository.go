package repository

import (
	"errors"
	"time"

	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StorefrontURLSearchProfileRepository struct {
	db *gorm.DB
}

func NewStorefrontURLSearchProfileRepository(db *gorm.DB) *StorefrontURLSearchProfileRepository {
	return &StorefrontURLSearchProfileRepository{db: db}
}

func (r *StorefrontURLSearchProfileRepository) List() ([]urlmanagementdomain.StorefrontURLSearchProfile, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront URL search profile repository is unavailable")
	}

	var profiles []urlmanagementdomain.StorefrontURLSearchProfile
	err := r.db.
		Preload("RouteEntry").
		Order("enabled DESC").
		Order("search_weight DESC").
		Order("route_entry_id ASC").
		Find(&profiles).Error
	return profiles, err
}

func (r *StorefrontURLSearchProfileRepository) FindByRouteEntryID(routeEntryID uint) (*urlmanagementdomain.StorefrontURLSearchProfile, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront URL search profile repository is unavailable")
	}
	if routeEntryID == 0 {
		return nil, errors.New("route entry ID is required")
	}

	var profile urlmanagementdomain.StorefrontURLSearchProfile
	if err := r.db.Preload("RouteEntry").Where("route_entry_id = ?", routeEntryID).First(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *StorefrontURLSearchProfileRepository) Upsert(profile *urlmanagementdomain.StorefrontURLSearchProfile) error {
	if r == nil || r.db == nil {
		return errors.New("storefront URL search profile repository is unavailable")
	}
	if profile == nil {
		return errors.New("storefront URL search profile is nil")
	}
	if profile.RouteEntryID == 0 {
		return errors.New("route entry ID is required")
	}
	if profile.Keywords == nil {
		profile.Keywords = datatypes.JSONSlice[string]{}
	}
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = time.Now().UTC()
	}
	profile.UpdatedAt = time.Now().UTC()

	return r.db.Transaction(func(tx *gorm.DB) error {
		var existing urlmanagementdomain.StorefrontURLSearchProfile
		err := tx.Where("route_entry_id = ?", profile.RouteEntryID).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Model(&urlmanagementdomain.StorefrontURLSearchProfile{}).Create(map[string]interface{}{
				"route_entry_id":  profile.RouteEntryID,
				"enabled":         profile.Enabled,
				"search_weight":   profile.SearchWeight,
				"keywords_json":   profile.Keywords,
				"display_title":   profile.DisplayTitle,
				"display_summary": profile.DisplaySummary,
				"created_at":      profile.CreatedAt,
				"updated_at":      profile.UpdatedAt,
			}).Error
		}
		if err != nil {
			return err
		}

		profile.ID = existing.ID
		profile.CreatedAt = existing.CreatedAt
		return tx.Model(&existing).Updates(map[string]interface{}{
			"enabled":         profile.Enabled,
			"search_weight":   profile.SearchWeight,
			"keywords_json":   profile.Keywords,
			"display_title":   profile.DisplayTitle,
			"display_summary": profile.DisplaySummary,
			"updated_at":      profile.UpdatedAt,
		}).Error
	})
}
