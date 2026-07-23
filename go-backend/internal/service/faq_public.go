package service

import (
	"fmt"
	"tanzanite/internal/domain/faq"
	"tanzanite/internal/pkg/faqcontent"
)

func (s *FAQService) GetPublicByID(id uint) (*faq.FAQ, error) {
	item, err := s.faqRepo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if sanitized, sanitizeErr := faqcontent.SanitizeAnswer(item.Answer); sanitizeErr == nil {
		item.Answer = sanitized
	}
	return item, nil
}

func (s *FAQService) GetPublicPageData(pageID, locale string) (*FAQPublicPageData, error) {
	page, err := s.faqRepo.FindPageByPageIDLocale(pageID, normalizeLocale(locale))
	if err != nil {
		return nil, err
	}
	if page.Status != "active" {
		return &FAQPublicPageData{PageID: page.PageID, Title: page.Title, Subtitle: page.Subtitle, Categories: []FAQPublicCategory{}}, nil
	}

	categories, err := s.faqRepo.ListCategories(page.Locale, page.PageID, false)
	if err != nil {
		return nil, err
	}

	faqItems, err := s.faqRepo.ListForPage(page.Locale, page.PageID, "published")
	if err != nil {
		return nil, err
	}
	faqItems = sanitizeFAQSliceForPublic(faqItems)

	itemsByCategory := make(map[string][]FAQPublicItem, len(categories))
	for _, item := range faqItems {
		itemsByCategory[item.Category] = append(itemsByCategory[item.Category], FAQPublicItem{
			ID:                fmt.Sprintf("%d", item.ID),
			Question:          item.Question,
			Answer:            item.Answer,
			AnswerImageURL:    item.AnswerImageURL,
			AnswerImageAlt:    item.AnswerImageAlt,
			AnswerImageWidth:  item.AnswerImageWidth,
			AnswerImageHeight: item.AnswerImageHeight,
			Tags:              []string{},
		})
	}

	publicPage := &FAQPublicPageData{
		PageID:     page.PageID,
		Title:      page.Title,
		Subtitle:   page.Subtitle,
		Categories: []FAQPublicCategory{},
	}
	for _, category := range categories {
		items := itemsByCategory[category.CategoryKey]
		if len(items) == 0 {
			continue
		}
		publicPage.Categories = append(publicPage.Categories, FAQPublicCategory{
			ID:    category.CategoryKey,
			Name:  category.Name,
			Icon:  category.Icon,
			Items: items,
		})
	}

	return publicPage, nil
}

func (s *FAQService) GetPublicPageDataByRoutePath(routePath, locale string) (*FAQPublicPageData, error) {
	page, err := s.faqRepo.FindPageByRoutePathLocale(normalizeRoutePath(routePath), normalizeLocale(locale))
	if err != nil {
		return nil, err
	}
	return s.GetPublicPageData(page.PageID, page.Locale)
}

func (s *FAQService) ListPublicPageData(locale string) ([]FAQPublicPageData, error) {
	pages, err := s.faqRepo.ListPages(normalizeLocale(locale), false)
	if err != nil {
		return nil, err
	}

	result := make([]FAQPublicPageData, 0, len(pages))
	for _, page := range pages {
		pageData, err := s.GetPublicPageData(page.PageID, page.Locale)
		if err != nil {
			return nil, err
		}
		if len(pageData.Categories) == 0 {
			continue
		}
		result = append(result, *pageData)
	}
	return result, nil
}

func sanitizeFAQSliceForPublic(items []faq.FAQ) []faq.FAQ {
	for index := range items {
		if sanitized, err := faqcontent.SanitizeAnswer(items[index].Answer); err == nil {
			items[index].Answer = sanitized
		}
	}
	return items
}
