package repository

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	outboxdomain "commerce-platform/internal/domain/outbox"
	seodomain "commerce-platform/internal/domain/seo"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type StorefrontRouteCatalogRepository struct {
	db     *gorm.DB
	outbox *OutboxRepository
}

func NewStorefrontRouteCatalogRepository(db *gorm.DB) *StorefrontRouteCatalogRepository {
	return &StorefrontRouteCatalogRepository{db: db}
}

func (r *StorefrontRouteCatalogRepository) ConfigureOutbox(outbox *OutboxRepository) {
	if r == nil {
		return
	}
	r.outbox = outbox
}

func (r *StorefrontRouteCatalogRepository) UpsertSnapshot(entries []seodomain.StorefrontRouteCatalogEntry, seenAt time.Time) error {
	if r == nil || r.db == nil {
		return errors.New("storefront route catalog repository is unavailable")
	}
	if seenAt.IsZero() {
		seenAt = time.Now().UTC()
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		var staleIDs []uint
		if err := tx.Model(&seodomain.StorefrontRouteCatalogEntry{}).
			Where("last_seen_at < ?", seenAt).
			Pluck("id", &staleIDs).Error; err != nil {
			return err
		}
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
				if err := r.publishRouteCatalogChanged(tx, entry.ID, entry.ManifestVersion, seenAt, "upsert"); err != nil {
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
			if err := r.publishRouteCatalogChanged(tx, entry.ID, entry.ManifestVersion, seenAt, "upsert"); err != nil {
				return err
			}
		}
		for _, routeEntryID := range staleIDs {
			if err := r.publishRouteCatalogChanged(tx, routeEntryID, "", seenAt, "stale"); err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *StorefrontRouteCatalogRepository) publishRouteCatalogChanged(
	tx *gorm.DB,
	routeEntryID uint,
	marker string,
	seenAt time.Time,
	change string,
) error {
	if r == nil || r.outbox == nil || routeEntryID == 0 {
		return nil
	}
	payload, err := json.Marshal(struct {
		RouteEntryID uint      `json:"route_entry_id"`
		Marker       string    `json:"marker,omitempty"`
		Change       string    `json:"change"`
		SeenAt       time.Time `json:"seen_at"`
	}{
		RouteEntryID: routeEntryID,
		Marker:       marker,
		Change:       change,
		SeenAt:       seenAt,
	})
	if err != nil {
		return err
	}
	return r.outbox.WithTx(tx).CreateEvent(&outboxdomain.Event{
		EventKey:      "storefront-route-catalog:" + strconv.FormatUint(uint64(routeEntryID), 10) + ":" + seenAt.Format(time.RFC3339Nano),
		EventType:     outboxdomain.EventTypeStorefrontRouteCatalogChanged,
		AggregateType: outboxdomain.AggregateTypeStorefrontRouteCatalogEntry,
		AggregateID:   strconv.FormatUint(uint64(routeEntryID), 10),
		Payload:       datatypes.JSON(payload),
		AvailableAt:   seenAt,
	})
}

type StorefrontRouteCatalogListFilter struct {
	Page           int
	PageSize       int
	Locale         string
	SourceType     string
	EntryStatus    string
	CheckStatus    string
	Search         string
	Searchable     *bool
	Indexable      *bool
	NeedsAttention *bool
	ProblemScope   string
	ExcludeAlias   bool
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
	if filter.Indexable != nil {
		query = query.Where("is_indexable = ?", *filter.Indexable)
	}
	if filter.NeedsAttention != nil {
		condition := routeNeedsAttentionCondition()
		if *filter.NeedsAttention {
			query = query.Where(condition)
		} else {
			query = query.Where("NOT (" + condition + ")")
		}
	}
	if filter.ProblemScope == "canonical" {
		query = query.Where(
			"entry_status = ? OR last_check_status = ?",
			seodomain.RouteEntryStatusDuplicate,
			seodomain.RouteCheckStatusCanonicalMisfit,
		)
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

func (r *StorefrontRouteCatalogRepository) ListSitemapEntries(limit int) ([]seodomain.StorefrontRouteCatalogEntry, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront route catalog repository is unavailable")
	}
	if limit < 1 || limit > 50000 {
		limit = 50000
	}

	query := r.db.
		Where("entry_status = ? AND is_alias = ? AND is_indexable = ?", seodomain.RouteEntryStatusActive, false, true).
		// This coarse predicate keeps known legacy product paths out of the
		// bounded query; exact route validation still happens below.
		Where("source_type <> ? OR path NOT LIKE ?", seodomain.RouteSourceProduct, "%/shop/%").
		Order("locale ASC").
		Order("path ASC")

	const pageSize = 1000
	entries := make([]seodomain.StorefrontRouteCatalogEntry, 0, limit)
	offset := 0
	for len(entries) < limit {
		var page []seodomain.StorefrontRouteCatalogEntry
		err := query.Offset(offset).Limit(pageSize).Find(&page).Error
		if err != nil {
			return nil, err
		}
		if len(page) == 0 {
			break
		}

		for _, entry := range page {
			if !sitemapEligibleRouteEntry(entry) {
				continue
			}
			entries = append(entries, entry)
			if len(entries) >= limit {
				break
			}
		}

		offset += len(page)
		if len(page) < pageSize {
			break
		}
	}
	return entries, nil
}

func sitemapEligibleRouteEntry(entry seodomain.StorefrontRouteCatalogEntry) bool {
	if entry.EntryStatus != seodomain.RouteEntryStatusActive ||
		entry.IsAlias ||
		!entry.IsIndexable ||
		entry.Path == "" {
		return false
	}

	if entry.SourceType != seodomain.RouteSourceProduct {
		return true
	}

	return seodomain.IsProductRoute(entry.Locale, entry.Path, entry.SourceKey) &&
		entry.CanonicalPath == entry.Path
}

func routeNeedsAttentionCondition() string {
	return `
		entry_status IN ('duplicate', 'stale')
		OR (
			is_alias = FALSE
			AND last_check_status IN ('redirect', 'not_found', 'server_error', 'canonical_mismatch', 'error')
		)
		OR (
			is_alias = TRUE
			AND last_check_status IN (
				'redirect_chain',
				'redirect_target_mismatch',
				'not_found',
				'server_error',
				'canonical_mismatch',
				'error'
			)
		)
	`
}

func (r *StorefrontRouteCatalogRepository) ListIssueCandidateIDs() ([]uint, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront route catalog repository is unavailable")
	}

	var ids []uint
	err := r.db.Model(&seodomain.StorefrontRouteCatalogEntry{}).
		Where(routeNeedsAttentionCondition()).
		Order("id ASC").
		Pluck("id", &ids).Error
	return ids, err
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

func (r *StorefrontRouteCatalogRepository) FindCurrentByPath(path string) (*seodomain.StorefrontRouteCatalogEntry, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront route catalog repository is unavailable")
	}

	var entry seodomain.StorefrontRouteCatalogEntry
	if err := r.db.
		Where("path = ? AND entry_status <> ?", path, seodomain.RouteEntryStatusStale).
		Order("is_alias DESC").
		First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *StorefrontRouteCatalogRepository) FindCanonicalByPath(path string) (*seodomain.StorefrontRouteCatalogEntry, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("storefront route catalog repository is unavailable")
	}

	var entry seodomain.StorefrontRouteCatalogEntry
	if err := r.db.
		Where(
			"path = ? AND is_alias = ? AND entry_status = ?",
			path,
			false,
			seodomain.RouteEntryStatusActive,
		).
		First(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}
