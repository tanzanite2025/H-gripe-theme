package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/pkg/locales"
)

const routeCatalogManifestPath = "/storefront-route-manifest.json"

func (s *StorefrontRouteCatalogService) Sync(ctx context.Context) (StorefrontRouteCatalogSyncSummary, error) {
	if s == nil || s.repository == nil {
		return StorefrontRouteCatalogSyncSummary{}, errors.New("storefront route catalog service is unavailable")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	manifest, err := s.loadManifest(ctx)
	if err != nil {
		return StorefrontRouteCatalogSyncSummary{}, err
	}

	summary := StorefrontRouteCatalogSyncSummary{ManifestVersion: manifest.Version}
	entries := make([]seodomain.StorefrontRouteCatalogEntry, 0, len(manifest.Routes)*len(locales.EnabledLocaleCodes()))
	entries = append(entries, buildManifestRouteCatalogEntries(manifest, &summary)...)

	productEntries, err := s.buildProductRouteCatalogEntries(ctx, &summary)
	if err != nil {
		return StorefrontRouteCatalogSyncSummary{}, err
	}
	entries = append(entries, productEntries...)

	blogEntries, err := s.buildBlogRouteCatalogEntries(&summary)
	if err != nil {
		return StorefrontRouteCatalogSyncSummary{}, err
	}
	entries = append(entries, blogEntries...)

	markDuplicatePaths(entries)
	for _, entry := range entries {
		if entry.DuplicateGroupKey != "" {
			summary.Duplicates++
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Locale != entries[j].Locale {
			return entries[i].Locale < entries[j].Locale
		}
		return entries[i].Path < entries[j].Path
	})

	if err := s.repository.UpsertSnapshot(entries, time.Now().UTC()); err != nil {
		return StorefrontRouteCatalogSyncSummary{}, fmt.Errorf("save storefront route catalog snapshot: %w", err)
	}
	if s.issueReconciler != nil {
		if err := s.issueReconciler.ReconcileCatalog(ctx); err != nil {
			return StorefrontRouteCatalogSyncSummary{}, fmt.Errorf("reconcile storefront URL issues after sync: %w", err)
		}
	}
	summary.Entries = len(entries)
	return summary, nil
}

func (s *StorefrontRouteCatalogService) loadManifest(ctx context.Context) (seodomain.StorefrontRouteManifest, error) {
	if path := strings.TrimSpace(os.Getenv("STOREFRONT_ROUTE_MANIFEST_PATH")); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return seodomain.StorefrontRouteManifest{}, fmt.Errorf("read storefront route manifest: %w", err)
		}
		return decodeRouteManifest(data)
	}
	if s.baseURL == "" {
		return seodomain.StorefrontRouteManifest{}, errors.New("storefront base URL is not configured")
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+routeCatalogManifestPath, nil)
	if err != nil {
		return seodomain.StorefrontRouteManifest{}, err
	}
	response, err := s.httpClient.Do(request)
	if err != nil {
		return seodomain.StorefrontRouteManifest{}, fmt.Errorf("fetch storefront route manifest: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return seodomain.StorefrontRouteManifest{}, fmt.Errorf("storefront route manifest returned HTTP %d", response.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024))
	if err != nil {
		return seodomain.StorefrontRouteManifest{}, fmt.Errorf("read storefront route manifest: %w", err)
	}
	return decodeRouteManifest(data)
}

func decodeRouteManifest(data []byte) (seodomain.StorefrontRouteManifest, error) {
	var manifest seodomain.StorefrontRouteManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, fmt.Errorf("decode storefront route manifest: %w", err)
	}
	if strings.TrimSpace(manifest.Version) == "" {
		return manifest, errors.New("storefront route manifest version is missing")
	}
	return manifest, nil
}

func markDuplicatePaths(entries []seodomain.StorefrontRouteCatalogEntry) {
	groups := make(map[string][]int)
	for index, entry := range entries {
		key := entry.Locale + "\x00" + entry.Path
		groups[key] = append(groups[key], index)
	}
	for key, indexes := range groups {
		if len(indexes) < 2 {
			continue
		}
		groupKey := "path:" + key
		for _, index := range indexes {
			entries[index].DuplicateGroupKey = groupKey
			if entries[index].EntryStatus == seodomain.RouteEntryStatusActive {
				entries[index].EntryStatus = seodomain.RouteEntryStatusDuplicate
			}
		}
	}
}
