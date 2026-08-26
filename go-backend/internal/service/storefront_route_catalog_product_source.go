package service

import (
	"context"
	"fmt"
	"strings"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/pkg/locales"
)

const (
	routeCatalogMaxPages = 100
	routeCatalogPageSize = 100
)

func (s *StorefrontRouteCatalogService) buildProductRouteCatalogEntries(
	ctx context.Context,
	summary *StorefrontRouteCatalogSyncSummary,
) ([]seodomain.StorefrontRouteCatalogEntry, error) {
	entries := make([]seodomain.StorefrontRouteCatalogEntry, 0)
	if s.productService == nil {
		return entries, nil
	}

	for _, locale := range locales.EnabledLocaleCodes() {
		products, err := s.listAllProducts(ctx, locale)
		if err != nil {
			return nil, fmt.Errorf("list products for %s: %w", locale, err)
		}
		for _, item := range products {
			itemLocale := locales.Normalize(item.Locale)
			if itemLocale != locales.Normalize(locale) || strings.TrimSpace(item.Slug) == "" {
				continue
			}
			canonicalPath := seodomain.BuildProductRoute(item.Locale, item.Slug).Path
			if !seodomain.IsProductRoute(itemLocale, canonicalPath, item.Slug) {
				continue
			}
			entries = append(entries, seodomain.StorefrontRouteCatalogEntry{
				RouteKey:      fmt.Sprintf("product:%d:%s", item.ID, itemLocale),
				Path:          canonicalPath,
				Locale:        itemLocale,
				SourceType:    seodomain.RouteSourceProduct,
				SourceID:      catalogUintPointer(item.ID),
				SourceKey:     item.Slug,
				Title:         item.Name,
				Summary:       item.ShortDesc,
				CanonicalPath: canonicalPath,
				IsSearchable:  true,
				IsCheckable:   true,
				IsIndexable:   true,
				EntryStatus:   seodomain.RouteEntryStatusActive,
			})
			summary.ProductEntries++
		}
	}

	return entries, nil
}

func (s *StorefrontRouteCatalogService) listAllProducts(ctx context.Context, locale string) ([]productSummary, error) {
	products := make([]productSummary, 0)
	if s.productService == nil {
		return products, nil
	}
	for page := 1; page <= routeCatalogMaxPages; page++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		items, total, err := s.productService.ListPublic(locale, false, page, routeCatalogPageSize)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			products = append(products, productSummary{
				ID:        item.ID,
				Locale:    item.Locale,
				Slug:      item.Slug,
				Name:      item.Name,
				ShortDesc: item.ShortDesc,
			})
		}
		if len(items) == 0 || len(items) < routeCatalogPageSize || int64(page*routeCatalogPageSize) >= total {
			break
		}
	}
	return products, nil
}

type productSummary struct {
	ID        uint
	Locale    string
	Slug      string
	Name      string
	ShortDesc string
}
