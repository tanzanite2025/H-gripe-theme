package service

import (
	"fmt"
	"strings"
	"tanzanite/internal/domain/faq"
)

func (s *FAQService) ListAdminStructure(locale string) ([]FAQPageAdminView, error) {
	pages, err := s.faqRepo.ListPages(locale, true)
	if err != nil {
		return nil, err
	}

	categories, err := s.faqRepo.ListCategories(locale, "", true)
	if err != nil {
		return nil, err
	}

	counts, err := s.faqRepo.ListFAQCounts(locale)
	if err != nil {
		return nil, err
	}

	pageViews := make([]FAQPageAdminView, 0, len(pages))
	pageIndex := make(map[string]int, len(pages))
	for _, page := range pages {
		pageViews = append(pageViews, FAQPageAdminView{
			FAQPage:    page,
			Categories: []FAQCategoryAdminView{},
		})
		pageIndex[page.PageID] = len(pageViews) - 1
	}

	for _, category := range categories {
		idx, ok := pageIndex[category.PageID]
		if !ok {
			continue
		}
		count := counts[faqCountKey(category.PageID, category.CategoryKey, category.Locale)]
		pageViews[idx].FAQCount += count
		pageViews[idx].Categories = append(pageViews[idx].Categories, FAQCategoryAdminView{
			FAQCategory: category,
			FAQCount:    count,
		})
	}

	return pageViews, nil
}

func (s *FAQService) UpsertAdminPage(pageID string, input FAQPageAdminInput) (*faq.FAQPage, error) {
	pageID = strings.TrimSpace(pageID)
	locale := normalizeLocale(input.Locale)
	if pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	existingPage, err := s.faqRepo.FindPageByPageIDLocale(pageID, locale)
	if err != nil {
		if !IsRecordNotFound(err) {
			return nil, err
		}
		existingPage = &faq.FAQPage{
			PageID: pageID,
			Locale: locale,
		}
	}

	existingPage.RoutePath = strings.TrimSpace(input.RoutePath)
	existingPage.Domain = strings.TrimSpace(input.Domain)
	existingPage.Title = strings.TrimSpace(input.Title)
	existingPage.Subtitle = strings.TrimSpace(input.Subtitle)
	existingPage.SortOrder = input.SortOrder
	existingPage.Status = normalizeFAQStatus(input.Status, "active")

	if existingPage.ID == 0 {
		if err := s.faqRepo.CreatePage(existingPage); err != nil {
			return nil, err
		}
	} else if err := s.faqRepo.SavePage(existingPage); err != nil {
		return nil, err
	}

	s.invalidateStorefrontHTMLCache("admin faq page update")
	return existingPage, nil
}

func (s *FAQService) CreateAdminCategory(input FAQCategoryAdminInput) (*faq.FAQCategory, error) {
	category, err := s.buildFAQCategory(input)
	if err != nil {
		return nil, err
	}
	if err := s.faqRepo.CreateCategory(category); err != nil {
		return nil, err
	}
	s.invalidateStorefrontHTMLCache("admin faq category create")
	return category, nil
}

func (s *FAQService) UpdateAdminCategory(id uint, input FAQCategoryAdminInput) (*faq.FAQCategory, error) {
	existingCategory, err := s.faqRepo.FindCategoryByID(id)
	if err != nil {
		return nil, err
	}

	oldPageID := existingCategory.PageID
	oldCategoryKey := existingCategory.CategoryKey
	oldLocale := existingCategory.Locale

	nextCategory, err := s.buildFAQCategory(input)
	if err != nil {
		return nil, err
	}

	existingCategory.PageID = nextCategory.PageID
	existingCategory.CategoryKey = nextCategory.CategoryKey
	existingCategory.Name = nextCategory.Name
	existingCategory.Icon = nextCategory.Icon
	existingCategory.Locale = nextCategory.Locale
	existingCategory.SortOrder = nextCategory.SortOrder
	existingCategory.Status = nextCategory.Status

	if err := s.faqRepo.UpdateCategory(existingCategory, oldPageID, oldCategoryKey, oldLocale); err != nil {
		return nil, err
	}

	s.invalidateStorefrontHTMLCache("admin faq category update")
	return existingCategory, nil
}

func (s *FAQService) DeleteAdminCategory(id uint) error {
	category, err := s.faqRepo.FindCategoryByID(id)
	if err != nil {
		return err
	}
	count, err := s.faqRepo.CountFAQsByCategory(category.PageID, category.CategoryKey, category.Locale)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("category has %d faq items", count)
	}
	if err := s.faqRepo.DeleteCategory(id); err != nil {
		return err
	}
	s.invalidateStorefrontHTMLCache("admin faq category delete")
	return nil
}

func (s *FAQService) buildFAQCategory(input FAQCategoryAdminInput) (*faq.FAQCategory, error) {
	pageID := strings.TrimSpace(input.PageID)
	name := strings.TrimSpace(input.Name)
	if pageID == "" {
		return nil, fmt.Errorf("page_id is required")
	}
	if name == "" {
		return nil, fmt.Errorf("category name is required")
	}

	categoryKey := strings.TrimSpace(input.CategoryKey)
	if categoryKey == "" {
		categoryKey = slugifyFAQKey(name)
	}
	if categoryKey == "" {
		return nil, fmt.Errorf("category_key is required")
	}

	return &faq.FAQCategory{
		PageID:      pageID,
		CategoryKey: categoryKey,
		Name:        name,
		Icon:        strings.TrimSpace(input.Icon),
		Locale:      normalizeLocale(input.Locale),
		SortOrder:   input.SortOrder,
		Status:      normalizeFAQStatus(input.Status, "active"),
	}, nil
}

func (s *FAQService) validateFAQPlacement(pageID, categoryKey, locale string) error {
	pageID = strings.TrimSpace(pageID)
	categoryKey = strings.TrimSpace(categoryKey)
	locale = normalizeLocale(locale)
	if pageID == "" {
		return fmt.Errorf("page_id is required")
	}
	if categoryKey == "" {
		return fmt.Errorf("category is required")
	}
	category, err := s.faqRepo.FindCategoryByPageKeyLocale(pageID, categoryKey, locale)
	if err != nil {
		if IsRecordNotFound(err) {
			return fmt.Errorf("faq category %q does not exist for page %q and locale %q", categoryKey, pageID, locale)
		}
		return err
	}
	if category.Status != "active" {
		return fmt.Errorf("faq category %q is hidden", categoryKey)
	}
	return nil
}

func faqCountKey(pageID, categoryKey, locale string) string {
	return pageID + "\x00" + categoryKey + "\x00" + locale
}
