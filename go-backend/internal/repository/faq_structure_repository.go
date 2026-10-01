package repository

import (
	"commerce-platform/internal/domain/faq"
	"context"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReconcileRoutePages updates only route ownership metadata. Editorial FAQ
// titles, subtitles, visibility and answer rows are preserved for existing
// pages; new routes receive an empty page shell for every enabled locale.
func (r *FAQRepository) ReconcileRoutePages(ctx context.Context, pages []faq.FAQPage, seenAt time.Time) (faq.FAQRouteSyncStats, error) {
	var stats faq.FAQRouteSyncStats
	if r == nil || r.db == nil {
		return stats, gorm.ErrInvalidDB
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if seenAt.IsZero() {
		seenAt = time.Now().UTC()
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		desired := make(map[string]struct{}, len(pages))
		for index := range pages {
			candidate := pages[index]
			key := candidate.PageID + "\x00" + candidate.Locale
			desired[key] = struct{}{}
			stats.Total++

			var existing faq.FAQPage
			err := tx.Where("page_id = ? AND locale = ?", candidate.PageID, candidate.Locale).First(&existing).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				err = tx.Unscoped().Where("page_id = ? AND locale = ?", candidate.PageID, candidate.Locale).First(&existing).Error
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// A previous migration may have used a different page_id for the
				// same URL. Reuse that row so existing FAQ content keeps its page
				// identity and deep links remain valid.
				err = tx.Where("route_path = ? AND locale = ?", candidate.RoutePath, candidate.Locale).First(&existing).Error
				if errors.Is(err, gorm.ErrRecordNotFound) {
					err = tx.Unscoped().Where("route_path = ? AND locale = ?", candidate.RoutePath, candidate.Locale).First(&existing).Error
				}
			}
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.Create(&candidate).Error; err != nil {
					return err
				}
				stats.Created++
				continue
			}
			if err != nil {
				return err
			}
			// If a historical/custom page_id owns the same route, keep that
			// content identity instead of marking the reused row stale.
			desired[existing.PageID+"\x00"+existing.Locale] = struct{}{}

			changed := existing.RoutePath != candidate.RoutePath ||
				existing.RouteKey != candidate.RouteKey ||
				existing.ManifestVersion != candidate.ManifestVersion ||
				existing.RouteStatus != candidate.RouteStatus ||
				existing.DeletedAt.Valid
			updates := map[string]interface{}{
				"route_path":       candidate.RoutePath,
				"route_key":        candidate.RouteKey,
				"manifest_version": candidate.ManifestVersion,
				"route_status":     candidate.RouteStatus,
				"last_seen_at":     seenAt,
				"deleted_at":       nil,
				"updated_at":       seenAt,
			}
			if err := tx.Model(&existing).Updates(updates).Error; err != nil {
				return err
			}
			if changed {
				stats.Updated++
			}
		}

		var existingPages []faq.FAQPage
		if err := tx.Where("deleted_at IS NULL").Find(&existingPages).Error; err != nil {
			return err
		}
		for index := range existingPages {
			page := existingPages[index]
			pagePath := strings.TrimSpace(page.RoutePath)
			if _, ok := desired[page.PageID+"\x00"+page.Locale]; ok ||
				(page.RouteStatus == "stale" && pagePath != "") ||
				(page.RouteStatus == "missing" && pagePath == "") {
				continue
			}
			routeStatus := "stale"
			if pagePath == "" {
				routeStatus = "missing"
			}
			if err := tx.Model(&page).Updates(map[string]interface{}{
				"route_status": routeStatus,
				"updated_at":   seenAt,
			}).Error; err != nil {
				return err
			}
			stats.Stale++
		}
		return nil
	})
	return stats, err
}

func (r *FAQRepository) ListPages(locale string, includeHidden bool) ([]faq.FAQPage, error) {
	var pages []faq.FAQPage
	query := r.db.Model(&faq.FAQPage{})
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if !includeHidden {
		query = query.Where("status = ?", "active")
	}
	err := query.
		Order("sort_order ASC").
		Order("page_id ASC").
		Find(&pages).Error
	return pages, err
}

func (r *FAQRepository) FindPageByPageIDLocale(pageID, locale string) (*faq.FAQPage, error) {
	var page faq.FAQPage
	err := r.db.Where("page_id = ? AND locale = ?", pageID, locale).First(&page).Error
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *FAQRepository) FindPageByRoutePathLocale(routePath, locale string) (*faq.FAQPage, error) {
	var page faq.FAQPage
	err := r.db.
		Where("route_path = ? AND locale = ?", routePath, locale).
		First(&page).Error
	if err != nil {
		return nil, err
	}
	return &page, nil
}

func (r *FAQRepository) SavePage(page *faq.FAQPage) error {
	return r.db.Save(page).Error
}

func (r *FAQRepository) CreatePage(page *faq.FAQPage) error {
	return r.db.Create(page).Error
}

func (r *FAQRepository) ListAdminForStructure(locale, pageID, status, search string) ([]faq.FAQ, error) {
	var faqs []faq.FAQ
	query := r.db.Model(&faq.FAQ{})

	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if pageID != "" {
		query = query.Where("page_id = ?", pageID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("question LIKE ? OR answer LIKE ?", searchPattern, searchPattern)
	}

	err := query.
		Order("page_id ASC").
		Order(clause.OrderByColumn{Column: clause.Column{Name: "order"}}).
		Order("created_at DESC").
		Find(&faqs).Error
	return faqs, err
}
