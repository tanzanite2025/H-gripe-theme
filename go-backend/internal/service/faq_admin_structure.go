package service

import (
	"commerce-platform/internal/domain/faq"
	"fmt"
	"strings"
)

func (s *FAQService) ListAdminStructure(locale string) ([]FAQPageAdminView, error) {
	if locale != "" {
		locale = normalizeLocale(locale)
	}
	pages, err := s.faqRepo.ListPages(locale, true)
	if err != nil {
		return nil, err
	}

	items, err := s.faqRepo.ListAdminForStructure(locale, "", "", "")
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64, len(items))
	for _, item := range items {
		counts[item.PageID]++
	}

	pageViews := make([]FAQPageAdminView, 0, len(pages))
	for _, page := range pages {
		pageViews = append(pageViews, FAQPageAdminView{
			FAQPage:  page,
			FAQCount: counts[page.PageID],
		})
	}

	return pageViews, nil
}

func (s *FAQService) ListAdminGrouped(locale, pageID, status, search string) ([]FAQPageAdminView, int64, error) {
	if locale != "" {
		locale = normalizeLocale(locale)
	}

	pages, err := s.faqRepo.ListPages(locale, true)
	if err != nil {
		return nil, 0, err
	}

	items, err := s.faqRepo.ListAdminForStructure(locale, pageID, status, search)
	if err != nil {
		return nil, 0, err
	}

	itemsByPage := make(map[string][]faq.FAQ)
	for _, item := range items {
		itemsByPage[item.PageID] = append(itemsByPage[item.PageID], item)
	}

	pageViews := make([]FAQPageAdminView, 0, len(pages))
	for _, page := range pages {
		if pageID != "" && page.PageID != pageID {
			continue
		}
		pageViews = append(pageViews, FAQPageAdminView{
			FAQPage: page,
		})
	}

	var total int64
	for index := range pageViews {
		pageItems := itemsByPage[pageViews[index].PageID]
		pageViews[index].FAQCount = int64(len(pageItems))
		pageViews[index].FAQs = pageItems
		total += int64(len(pageItems))
	}

	return pageViews, total, nil
}

func (s *FAQService) UpsertAdminPage(pageID string, input FAQPageAdminInput) (*faq.FAQPage, error) {
	pageID = strings.TrimSpace(pageID)
	locale, err := requireSupportedLocale(input.Locale)
	if err != nil {
		return nil, err
	}
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
			PageID:    pageID,
			Locale:    locale,
			RoutePath: strings.TrimSpace(input.RoutePath),
		}
	}

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

	s.notifyStorefrontContentChange("admin faq page update")
	return existingPage, nil
}

func (s *FAQService) validateFAQPage(pageID, locale string) error {
	pageID = strings.TrimSpace(pageID)
	locale, err := requireSupportedLocale(locale)
	if err != nil {
		return err
	}
	if pageID == "" {
		return fmt.Errorf("page_id is required")
	}
	page, err := s.faqRepo.FindPageByPageIDLocale(pageID, locale)
	if err != nil {
		if IsRecordNotFound(err) {
			return fmt.Errorf("faq page %q does not exist for locale %q", pageID, locale)
		}
		return err
	}
	if page.Status != "active" {
		return fmt.Errorf("faq page %q is hidden", pageID)
	}
	return nil
}
