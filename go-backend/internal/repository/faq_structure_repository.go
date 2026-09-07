package repository

import (
	"commerce-platform/internal/domain/faq"

	"gorm.io/gorm/clause"
)

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
