package service

import (
	"errors"
	"strings"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
)

type StorefrontSitemapOverview struct {
	PublicPath        string     `json:"public_path"`
	SitemapURL        string     `json:"sitemap_url"`
	Source            string     `json:"source"`
	DynamicSourcePath string     `json:"dynamic_source_path"`
	Entries           int64      `json:"entries"`
	Indexable         int64      `json:"indexable"`
	LastSyncedAt      *time.Time `json:"last_synced_at,omitempty"`
	ManifestVersion   string     `json:"manifest_version"`
}

func (s *StorefrontRouteCatalogService) SitemapOverview() (StorefrontSitemapOverview, error) {
	stats, err := s.Stats()
	if err != nil {
		return StorefrontSitemapOverview{}, err
	}

	publicPath := "/sitemap.xml"
	sitemapURL := publicPath
	if baseURL := strings.TrimRight(strings.TrimSpace(s.baseURL), "/"); baseURL != "" {
		sitemapURL = baseURL + publicPath
	}

	return StorefrontSitemapOverview{
		PublicPath:        publicPath,
		SitemapURL:        sitemapURL,
		Source:            "storefront route catalog",
		DynamicSourcePath: "/api/v1/storefront/sitemap-routes",
		Entries:           stats.SitemapEligible,
		Indexable:         stats.Indexable,
		LastSyncedAt:      stats.LastSyncedAt,
		ManifestVersion:   stats.ManifestVersion,
	}, nil
}

func (s *StorefrontRouteCatalogService) ListSitemapEntries(limit int) ([]seodomain.StorefrontRouteCatalogEntry, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("storefront route catalog service is unavailable")
	}
	return s.repository.ListSitemapEntries(limit)
}
