package service

import (
	"context"
	"fmt"

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
			canonicalPath := seodomain.BuildProductRoute(item.Locale, item.Slug).Path
			entries = append(entries, seodomain.StorefrontRouteCatalogEntry{
				RouteKey:      fmt.Sprintf("product:%d:%s", item.ID, item.Locale),
				Path:          canonicalPath,
				Locale:        item.Locale,
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
			entries = append(entries, seodomain.StorefrontRouteCatalogEntry{
				RouteKey:      fmt.Sprintf("product-alias:%d:%s", item.ID, item.Locale),
				Path:          seodomain.BuildLegacyProductRoute(item.Locale, item.Slug).Path,
				Locale:        item.Locale,
				SourceType:    seodomain.RouteSourceAlias,
				SourceID:      catalogUintPointer(item.ID),
				SourceKey:     item.Slug,
				Title:         item.Name + " (Legacy Alias)",
				Summary:       "Legacy product route redirected to the canonical shop route",
				CanonicalPath: canonicalPath,
				IsAlias:       true,
				IsSearchable:  false,
				IsCheckable:   true,
				IsIndexable:   false,
				EntryStatus:   seodomain.RouteEntryStatusAlias,
			})
			summary.ProductEntries++
			summary.AliasEntries++
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
