package repository

import (
	seodomain "commerce-platform/internal/domain/seo"
	"errors"
)

func (r *StorefrontRouteCatalogRepository) Stats() (seodomain.StorefrontRouteCatalogStats, error) {
	if r == nil || r.db == nil {
		return seodomain.StorefrontRouteCatalogStats{}, errors.New("storefront route catalog repository is unavailable")
	}

	var stats seodomain.StorefrontRouteCatalogStats
	err := r.db.Model(&seodomain.StorefrontRouteCatalogEntry{}).
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
			COALESCE(SUM(CASE WHEN last_check_status IS NOT NULL AND last_check_status <> '' THEN 1 ELSE 0 END), 0) AS checked,
			COALESCE(SUM(CASE WHEN last_check_status IS NULL OR last_check_status = '' THEN 1 ELSE 0 END), 0) AS unchecked,
			COALESCE(SUM(CASE WHEN last_check_status = 'ok' THEN 1 ELSE 0 END), 0) AS ok,
			COALESCE(SUM(CASE WHEN last_check_status = 'redirect' THEN 1 ELSE 0 END), 0) AS redirects,
			COALESCE(SUM(CASE WHEN last_check_status = 'not_found' THEN 1 ELSE 0 END), 0) AS not_found,
			COALESCE(SUM(CASE WHEN last_check_status = 'server_error' THEN 1 ELSE 0 END), 0) AS server_errors,
			COALESCE(SUM(CASE WHEN last_check_status = 'canonical_mismatch' THEN 1 ELSE 0 END), 0) AS canonical_mismatch,
			COALESCE(SUM(CASE WHEN last_check_status = 'error' THEN 1 ELSE 0 END), 0) AS errors,
			COALESCE(SUM(CASE WHEN is_searchable = TRUE THEN 1 ELSE 0 END), 0) AS searchable,
			COALESCE(SUM(CASE WHEN is_checkable = TRUE THEN 1 ELSE 0 END), 0) AS checkable,
			COALESCE(SUM(CASE WHEN is_indexable = TRUE THEN 1 ELSE 0 END), 0) AS indexable,
			COALESCE(SUM(CASE
				WHEN entry_status = 'active'
					AND is_alias = FALSE
					AND is_indexable = TRUE
				THEN 1 ELSE 0
			END), 0) AS sitemap_eligible,
			MAX(last_seen_at) AS last_synced_at,
			COALESCE(MAX(manifest_version), '') AS manifest_version
		`).
		Scan(&stats).Error
	return stats, err
}
