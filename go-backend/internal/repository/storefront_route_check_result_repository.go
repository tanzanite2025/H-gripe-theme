package repository

import (
	seodomain "commerce-platform/internal/domain/seo"
	"errors"
	"time"

	"gorm.io/gorm"
)

func (r *StorefrontRouteCatalogRepository) ListChecks(routeEntryID uint, page, pageSize int) ([]seodomain.StorefrontRouteCheckResult, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("storefront route catalog repository is unavailable")
	}
	if routeEntryID == 0 {
		return nil, 0, errors.New("route entry ID is required")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}

	query := r.db.Model(&seodomain.StorefrontRouteCheckResult{}).
		Where("route_entry_id = ?", routeEntryID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var results []seodomain.StorefrontRouteCheckResult
	err := query.
		Order("checked_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&results).Error
	return results, total, err
}

func (r *StorefrontRouteCatalogRepository) SaveCheck(result *seodomain.StorefrontRouteCheckResult) error {
	if r == nil || r.db == nil {
		return errors.New("storefront route catalog repository is unavailable")
	}
	if result == nil {
		return errors.New("storefront route check result is nil")
	}
	if result.CheckedAt.IsZero() {
		result.CheckedAt = time.Now().UTC()
	}

	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(result).Error; err != nil {
			return err
		}
		return tx.Model(&seodomain.StorefrontRouteCatalogEntry{}).
			Where("id = ?", result.RouteEntryID).
			Updates(map[string]interface{}{
				"last_check_status":   result.Status,
				"last_http_status":    result.HTTPStatus,
				"last_final_url":      result.FinalURL,
				"last_canonical_url":  result.CanonicalURL,
				"last_response_ms":    result.ResponseMS,
				"last_redirect_count": result.RedirectCount,
				"last_content_hash":   result.ContentHash,
				"last_check_error":    result.ErrorMessage,
				"last_checked_at":     result.CheckedAt,
				"updated_at":          time.Now().UTC(),
			}).Error
	})
}
