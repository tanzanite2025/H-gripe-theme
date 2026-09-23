package repository

import (
	"errors"

	seodomain "commerce-platform/internal/domain/seo"

	"gorm.io/gorm"
)

func (r *StorefrontRouteCatalogRepository) Stats() (seodomain.StorefrontRouteCatalogStats, error) {
	return r.StatsForLocale("")
}

func (r *StorefrontRouteCatalogRepository) StatsForLocale(locale string) (seodomain.StorefrontRouteCatalogStats, error) {
	return r.StatsForLocaleAndScope(locale, "")
}

// StatsForLocaleAndScope returns aggregate metrics for the same logical scope
// used by the route list. Keeping the problem scope in the aggregate query
// prevents the Canonical view from displaying whole-catalog numbers.
func (r *StorefrontRouteCatalogRepository) StatsForLocaleAndScope(locale, problemScope string) (seodomain.StorefrontRouteCatalogStats, error) {
	if r == nil || r.db == nil {
		return seodomain.StorefrontRouteCatalogStats{}, errors.New("storefront route catalog repository is unavailable")
	}

	query := r.db.Model(&seodomain.StorefrontRouteCatalogEntry{})
	if locale != "" && locale != "all" {
		query = query.Where("locale = ?", locale)
	}
	if problemScope == "canonical" {
		query = query.Where(
			"(entry_status = ? OR last_check_status = ?) AND is_alias = ?",
			seodomain.RouteEntryStatusDuplicate,
			seodomain.RouteCheckStatusCanonicalMisfit,
			false,
		)
	}

	var latestEntry seodomain.StorefrontRouteCatalogEntry
	latestQuery := r.db.Model(&seodomain.StorefrontRouteCatalogEntry{})
	if locale != "" && locale != "all" {
		latestQuery = latestQuery.Where("locale = ?", locale)
	}
	if err := latestQuery.Order("last_seen_at DESC").First(&latestEntry).Error; err != nil &&
		!errors.Is(err, gorm.ErrRecordNotFound) {
		return seodomain.StorefrontRouteCatalogStats{}, err
	}

	var stats seodomain.StorefrontRouteCatalogStats
	err := query.
		Select(`
			COUNT(*) AS total,
			COALESCE(SUM(CASE WHEN entry_status = 'active' THEN 1 ELSE 0 END), 0) AS active,
			COALESCE(SUM(CASE WHEN entry_status = 'alias' OR is_alias = TRUE THEN 1 ELSE 0 END), 0) AS alias,
			COALESCE(SUM(CASE WHEN entry_status = 'duplicate' THEN 1 ELSE 0 END), 0) AS duplicate,
			COALESCE(SUM(CASE WHEN entry_status = 'stale' THEN 1 ELSE 0 END), 0) AS stale,
			COALESCE(SUM(CASE
				WHEN entry_status IN ('duplicate', 'stale')
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
				THEN 1 ELSE 0
			END), 0) AS needs_attention,
			COALESCE(SUM(CASE WHEN entry_status <> 'stale' AND last_check_status IS NOT NULL AND last_check_status <> '' THEN 1 ELSE 0 END), 0) AS checked,
			COALESCE(SUM(CASE WHEN entry_status <> 'stale' AND (last_check_status IS NULL OR last_check_status = '') THEN 1 ELSE 0 END), 0) AS unchecked,
			COALESCE(SUM(CASE WHEN entry_status <> 'stale' AND last_check_status = 'ok' THEN 1 ELSE 0 END), 0) AS ok,
			COALESCE(SUM(CASE WHEN entry_status <> 'stale' AND last_check_status = 'redirect' THEN 1 ELSE 0 END), 0) AS redirects,
			COALESCE(SUM(CASE WHEN entry_status <> 'stale' AND last_check_status = 'not_found' THEN 1 ELSE 0 END), 0) AS not_found,
			COALESCE(SUM(CASE WHEN entry_status <> 'stale' AND last_check_status = 'server_error' THEN 1 ELSE 0 END), 0) AS server_errors,
			COALESCE(SUM(CASE WHEN entry_status <> 'stale' AND last_check_status = 'canonical_mismatch' THEN 1 ELSE 0 END), 0) AS canonical_mismatch,
			COALESCE(SUM(CASE WHEN entry_status <> 'stale' AND last_check_status = 'error' THEN 1 ELSE 0 END), 0) AS errors,
			COALESCE(SUM(CASE WHEN is_searchable = TRUE THEN 1 ELSE 0 END), 0) AS searchable,
			COALESCE(SUM(CASE WHEN entry_status <> 'stale' AND is_checkable = TRUE THEN 1 ELSE 0 END), 0) AS checkable,
			COALESCE(SUM(CASE WHEN is_indexable = TRUE THEN 1 ELSE 0 END), 0) AS indexable,
			COALESCE(SUM(CASE
				WHEN entry_status = 'active'
					AND is_alias = FALSE
					AND is_indexable = TRUE
				THEN 1 ELSE 0
			END), 0) AS sitemap_eligible,
			COALESCE(MAX(manifest_version), '') AS manifest_version
		`).
		Scan(&stats).Error
	if err != nil {
		return stats, err
	}
	// Sitemap export applies the exact route predicate in
	// sitemapEligibleRouteEntry (including product canonical-path rules). The
	// aggregate SQL above intentionally remains cheap for the other counters,
	// while this metric reuses the export predicate so the card cannot drift.
	stats.SitemapEligible, err = r.countSitemapEligible(locale, problemScope)
	if err != nil {
		return stats, err
	}
	if latestEntry.ID != 0 {
		stats.LastSyncedAt = &latestEntry.LastSeenAt
	}
	return stats, err
}

func (r *StorefrontRouteCatalogRepository) countSitemapEligible(locale, problemScope string) (int64, error) {
	query := r.db.Model(&seodomain.StorefrontRouteCatalogEntry{}).
		Where("entry_status = ? AND is_alias = ? AND is_indexable = ?", seodomain.RouteEntryStatusActive, false, true).
		Where("source_type <> ? OR path NOT LIKE ?", seodomain.RouteSourceProduct, "%/shop/%")
	if locale != "" && locale != "all" {
		query = query.Where("locale = ?", locale)
	}
	if problemScope == "canonical" {
		query = query.Where("(entry_status = ? OR last_check_status = ?) AND is_alias = ?", seodomain.RouteEntryStatusDuplicate, seodomain.RouteCheckStatusCanonicalMisfit, false)
	}

	var entries []seodomain.StorefrontRouteCatalogEntry
	if err := query.Find(&entries).Error; err != nil {
		return 0, err
	}
	var count int64
	for _, entry := range entries {
		if sitemapEligibleRouteEntry(entry) {
			count++
		}
	}
	return count, nil
}
