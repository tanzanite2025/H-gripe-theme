package service

import (
	"fmt"

	seodomain "commerce-platform/internal/domain/seo"
)

func (s *StorefrontRouteCatalogService) buildBlogRouteCatalogEntries(
	summary *StorefrontRouteCatalogSyncSummary,
) ([]seodomain.StorefrontRouteCatalogEntry, error) {
	entries := make([]seodomain.StorefrontRouteCatalogEntry, 0)
	if s.postService == nil {
		return entries, nil
	}

	posts, err := s.postService.GetPublishedPosts()
	if err != nil {
		return nil, fmt.Errorf("list published blog posts: %w", err)
	}
	for _, item := range posts {
		routePath := seodomain.BuildArticleRoute(item.Locale, item.Slug, item.Tags).Path
		entries = append(entries, seodomain.StorefrontRouteCatalogEntry{
			RouteKey:      fmt.Sprintf("blog:%d:%s", item.ID, item.Locale),
			Path:          routePath,
			Locale:        item.Locale,
			SourceType:    seodomain.RouteSourceBlog,
			SourceID:      catalogUintPointer(item.ID),
			SourceKey:     item.Slug,
			Title:         item.Title,
			Summary:       item.Excerpt,
			CanonicalPath: routePath,
			IsSearchable:  true,
			IsCheckable:   true,
			IsIndexable:   true,
			EntryStatus:   seodomain.RouteEntryStatusActive,
		})
		summary.BlogEntries++
	}

	return entries, nil
}
