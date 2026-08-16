package repository

import (
	seodomain "commerce-platform/internal/domain/seo"
	"errors"
	"time"

	"gorm.io/gorm"
)

type StorefrontRouteCatalogRepository struct {
	db *gorm.DB
}

func NewStorefrontRouteCatalogRepository(db *gorm.DB) *StorefrontRouteCatalogRepository {
	return &StorefrontRouteCatalogRepository{db: db}
}

func (r *StorefrontRouteCatalogRepository) UpsertSnapshot(entries []seodomain.StorefrontRouteCatalogEntry, seenAt time.Time) error {
	if r == nil || r.db == nil {
		return errors.New("storefront route catalog repository is unavailable")
	}
	if seenAt.IsZero() {
		seenAt = time.Now().UTC()
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&seodomain.StorefrontRouteCatalogEntry{}).
			Where("last_seen_at < ?", seenAt).
			Updates(map[string]interface{}{
				"entry_status": "stale",
				"updated_at":   time.Now().UTC(),
			}).Error; err != nil {
			return err
		}

		for index := range entries {
			entry := entries[index]
			entry.LastSeenAt = seenAt

			var existing seodomain.StorefrontRouteCatalogEntry
			err := tx.Where("route_key = ?", entry.RouteKey).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&entry).Error; err != nil {
					return err
				}
				continue
			}
			if err != nil {
				return err
			}

			if err := tx.Model(&existing).Updates(map[string]interface{}{
				"path":                entry.Path,
				"locale":              entry.Locale,
				"source_type":         entry.SourceType,
				"source_id":           entry.SourceID,
				"source_key":          entry.SourceKey,
				"title":               entry.Title,
				"summary":             entry.Summary,
				"canonical_path":      entry.CanonicalPath,
				"is_alias":            entry.IsAlias,
				"is_searchable":       entry.IsSearchable,
				"is_checkable":        entry.IsCheckable,
				"is_indexable":        entry.IsIndexable,
				"entry_status":        entry.EntryStatus,
				"duplicate_group_key": entry.DuplicateGroupKey,
				"manifest_version":    entry.ManifestVersion,
				"last_seen_at":        entry.LastSeenAt,
				"updated_at":          time.Now().UTC(),
			}).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

type StorefrontRouteCatalogListFilter struct {
	Page         int
	PageSize     int
	Locale       string
	SourceType   string
	EntryStatus  string
	CheckStatus  string
	Search       string
	Searchable   *bool
	ExcludeAlias bool
}

func (r *StorefrontRouteCatalogRepository) List(filter StorefrontRouteCatalogListFilter) ([]seodomain.StorefrontRouteCatalogEntry, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("storefront route catalog repository is unavailable")
	}
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 200 {
		filter.PageSize = 50
	}

	query := r.db.Model(&seodomain.StorefrontRouteCatalogEntry{})
	if filter.Locale != "" {
		query = query.Where("locale = ?", filter.Locale)
	}
	if filter.SourceType != "" {
		query = query.Where("source_type = ?", filter.SourceType)
	}
	if filter.EntryStatus != "" {
		query = query.Where("entry_status = ?", filter.EntryStatus)
	}
	if filter.CheckStatus != "" {
		query = query.Where("last_check_status = ?", filter.CheckStatus)
	}
	if filter.Searchable != nil {
		query = query.Where("is_searchable = ?", *filter.Searchable)
	}
	if filter.ExcludeAlias {
		query = query.Where("is_alias = ?", false)
	}
	if search := filter.Search; search != "" {
		like := "%" + search + "%"
		query = query.Where(
			"LOWER(path) LIKE LOWER(?) OR LOWER(title) LIKE LOWER(?) OR LOWER(summary) LIKE LOWER(?) OR LOWER(source_key) LIKE LOWER(?)",
			like,
			like,
			like,
			like,
		)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var entries []seodomain.StorefrontRouteCatalogEntry
	err := query.
		Order("is_alias ASC").
		Order("locale ASC").
		Order("path ASC").
		Offset((filter.Page - 1) * filter.PageSize).
		Limit(filter.PageSize).
		Find(&entries).Error
	return entries, total, err
}

func (r *StorefrontRouteCatalogRepository) FindByID(id uint) (*seodomain.StorefrontRouteCatalogEntry, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront route catalog repository is unavailable")
	}
	var entry seodomain.StorefrontRouteCatalogEntry
	if err := r.db.First(&entry, id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}
