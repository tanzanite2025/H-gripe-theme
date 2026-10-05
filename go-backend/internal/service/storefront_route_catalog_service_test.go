package service

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestStorefrontRouteCatalogManifestUsesInternalOrigin(t *testing.T) {
	var publicHits atomic.Int32
	publicServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		publicHits.Add(1)
		http.Error(writer, "public edge challenge", http.StatusForbidden)
	}))
	defer publicServer.Close()

	internalServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != routeCatalogManifestPath {
			t.Errorf("unexpected manifest path: %s", request.URL.Path)
		}
		if request.URL.Query().Get("route_catalog_sync") == "" {
			t.Error("manifest request did not include a cache-busting sync marker")
		}
		if request.Header.Get("Cache-Control") != "no-cache, no-store" {
			t.Errorf("manifest Cache-Control = %q", request.Header.Get("Cache-Control"))
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"version":"test-manifest","routes":[]}`))
	}))
	defer internalServer.Close()

	catalog := &StorefrontRouteCatalogService{
		baseURL:         publicServer.URL,
		internalBaseURL: internalServer.URL,
		httpClient:      internalServer.Client(),
	}

	manifest, err := catalog.loadManifest(context.Background())
	if err != nil {
		t.Fatalf("loadManifest returned error: %v", err)
	}
	if manifest.Version != "test-manifest" {
		t.Fatalf("manifest version = %q, want test-manifest", manifest.Version)
	}
	if publicHits.Load() != 0 {
		t.Fatalf("public origin was contacted %d times", publicHits.Load())
	}
}

func TestStorefrontRouteCatalogCheckUsesInternalOrigin(t *testing.T) {
	var publicHits atomic.Int32
	publicServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		publicHits.Add(1)
		http.Error(writer, "public edge challenge", http.StatusForbidden)
	}))
	defer publicServer.Close()

	internalServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/products/example" {
			t.Errorf("unexpected route path: %s", request.URL.Path)
		}
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write([]byte(`<html><head><link rel="canonical" href="https://learn.gripe/products/example"></head><body>ok</body></html>`))
	}))
	defer internalServer.Close()

	catalog := &StorefrontRouteCatalogService{
		baseURL:         publicServer.URL,
		internalBaseURL: internalServer.URL,
		httpClient:      internalServer.Client(),
	}

	result := catalog.checkEntry(context.Background(), seodomain.StorefrontRouteCatalogEntry{
		Path:          "/products/example",
		CanonicalPath: "/products/example",
		IsCheckable:   true,
	})
	if result.Status != seodomain.RouteCheckStatusOK {
		t.Fatalf("route status = %q, want %q (error: %s)", result.Status, seodomain.RouteCheckStatusOK, result.ErrorMessage)
	}
	if publicHits.Load() != 0 {
		t.Fatalf("public origin was contacted %d times", publicHits.Load())
	}
}

func TestStorefrontRouteCatalogCheckClassifiesRedirectBeforeCanonical(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path == "/products/legacy" {
			http.Redirect(writer, request, "/", http.StatusMovedPermanently)
			return
		}
		writer.Header().Set("Content-Type", "text/html")
		_, _ = writer.Write([]byte(`<html><head><link rel="canonical" href="/"></head><body>home</body></html>`))
	}))
	defer server.Close()

	catalog := &StorefrontRouteCatalogService{
		internalBaseURL: server.URL,
		httpClient:      server.Client(),
	}
	result := catalog.checkEntry(context.Background(), seodomain.StorefrontRouteCatalogEntry{
		Path:          "/products/legacy",
		CanonicalPath: "/products/legacy",
		IsCheckable:   true,
	})

	if result.Status != seodomain.RouteCheckStatusRedirect {
		t.Fatalf("route status = %q, want %q (error: %s)", result.Status, seodomain.RouteCheckStatusRedirect, result.ErrorMessage)
	}
	if result.RedirectCount != 1 {
		t.Fatalf("redirect count = %d, want 1", result.RedirectCount)
	}
	if result.CanonicalURL != "" {
		t.Fatalf("canonical URL = %q, want empty for redirected response", result.CanonicalURL)
	}
	if result.FinalURL != server.URL+"/" {
		t.Fatalf("final URL = %q, want redirect destination", result.FinalURL)
	}
}

func TestNewStorefrontRouteCatalogServiceDoesNotFallbackToPublicOrigin(t *testing.T) {
	catalog := NewStorefrontRouteCatalogService(
		nil,
		nil,
		nil,
		"https://public.example.com",
		"",
	)

	if catalog.internalBaseURL != "" {
		t.Fatalf("internal origin = %q, want empty when it is not configured", catalog.internalBaseURL)
	}
}

func TestNewStorefrontRouteCatalogServiceUsesNuxtSSRCheckBudget(t *testing.T) {
	catalog := NewStorefrontRouteCatalogService(
		nil,
		nil,
		nil,
		"https://store.example.com",
		"http://storefront:3000",
	)

	require.Equal(t, 2, routeCheckConcurrency)
	require.Equal(t, 15*time.Second, catalog.httpClient.Timeout)
}

func TestRouteEntryCanBeCheckedIncludesStaleEntriesOnlyWhenRequested(t *testing.T) {
	if !routeEntryCanBeChecked(seodomain.StorefrontRouteCatalogEntry{
		IsCheckable: true,
		EntryStatus: seodomain.RouteEntryStatusActive,
	}, false) {
		t.Fatal("active checkable route should be checkable")
	}
	if routeEntryCanBeChecked(seodomain.StorefrontRouteCatalogEntry{
		IsCheckable: true,
		EntryStatus: seodomain.RouteEntryStatusStale,
	}, false) {
		t.Fatal("stale route should not be included by default")
	}
	if !routeEntryCanBeChecked(seodomain.StorefrontRouteCatalogEntry{
		IsCheckable: true,
		EntryStatus: seodomain.RouteEntryStatusStale,
	}, true) {
		t.Fatal("stale route should be checkable when requested for issue verification")
	}
}

func TestStorefrontRouteCatalogCheckProcessesAllRoutesAcrossDatabasePages(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&seodomain.StorefrontRouteCheckResult{},
	))

	const routeCount = 205
	entries := make([]seodomain.StorefrontRouteCatalogEntry, 0, routeCount)
	for index := 0; index < routeCount; index++ {
		path := fmt.Sprintf("/route/%03d", index)
		entries = append(entries, seodomain.StorefrontRouteCatalogEntry{
			RouteKey:        fmt.Sprintf("manifest:route-%03d:en", index),
			Path:            path,
			Locale:          "en",
			SourceType:      seodomain.RouteSourceStatic,
			CanonicalPath:   path,
			EntryStatus:     seodomain.RouteEntryStatusActive,
			IsCheckable:     true,
			LastCheckStatus: seodomain.RouteCheckStatusServerError,
		})
	}
	require.NoError(t, db.Create(&entries).Error)

	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	routeRepository := repository.NewStorefrontRouteCatalogRepository(db)
	catalog := NewStorefrontRouteCatalogService(routeRepository, nil, nil, server.URL, server.URL)
	summary, err := catalog.Check(context.Background(), repository.StorefrontRouteCatalogListFilter{
		Locale:      "en",
		CheckStatus: seodomain.RouteCheckStatusServerError,
	}, 200)

	require.NoError(t, err)
	require.Equal(t, routeCount, summary.Eligible)
	require.Equal(t, routeCount, summary.Checked)
	require.Zero(t, summary.Remaining)
	require.Equal(t, 200, summary.BatchSize)
	require.Equal(t, 2, summary.TotalBatches)
	require.Equal(t, 2, summary.CurrentBatch)
	require.Equal(t, int32(routeCount), requestCount.Load())

	var checkedCount int64
	require.NoError(t, db.Model(&seodomain.StorefrontRouteCheckResult{}).
		Where("status = ?", seodomain.RouteCheckStatusOK).
		Count(&checkedCount).Error)
	require.Equal(t, int64(routeCount), checkedCount)
}

func TestStorefrontRouteCatalogCheckLeavesUnprocessedRoutesRemainingAfterContextTimeout(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&seodomain.StorefrontRouteCatalogEntry{},
		&seodomain.StorefrontRouteCheckResult{},
	))
	for index := 0; index < 3; index++ {
		path := fmt.Sprintf("/timeout/%d", index)
		require.NoError(t, db.Create(&seodomain.StorefrontRouteCatalogEntry{
			RouteKey:        fmt.Sprintf("manifest:timeout-%d:en", index),
			Path:            path,
			Locale:          "en",
			SourceType:      seodomain.RouteSourceStatic,
			CanonicalPath:   path,
			EntryStatus:     seodomain.RouteEntryStatusActive,
			IsCheckable:     true,
			LastCheckStatus: seodomain.RouteCheckStatusOK,
		}).Error)
	}

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	defer server.Close()

	catalog := NewStorefrontRouteCatalogService(
		repository.NewStorefrontRouteCatalogRepository(db),
		nil,
		nil,
		server.URL,
		server.URL,
	)
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	summary, err := catalog.Check(ctx, repository.StorefrontRouteCatalogListFilter{Locale: "en"}, 2)
	require.Error(t, err)
	require.Equal(t, 3, summary.Eligible)
	require.Greater(t, summary.Remaining, 0)
	require.Less(t, summary.Checked, summary.Eligible)
}

func TestStorefrontRouteCatalogRejectsOverlappingMutations(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	catalog := NewStorefrontRouteCatalogService(
		repository.NewStorefrontRouteCatalogRepository(db),
		nil,
		nil,
		"http://127.0.0.1",
		"http://127.0.0.1",
	)
	releaseOperation, err := catalog.beginCatalogOperation()
	require.NoError(t, err)
	defer releaseOperation()

	_, err = catalog.Sync(context.Background())
	require.ErrorIs(t, err, ErrStorefrontRouteCatalogOperationInProgress)
	_, err = catalog.CheckEntry(context.Background(), 1)
	require.ErrorIs(t, err, ErrStorefrontRouteCatalogOperationInProgress)
	_, err = catalog.Check(context.Background(), repository.StorefrontRouteCatalogListFilter{}, 200)
	require.ErrorIs(t, err, ErrStorefrontRouteCatalogOperationInProgress)
	_, err = catalog.StartCheck(repository.StorefrontRouteCatalogListFilter{}, 200)
	require.ErrorIs(t, err, ErrStorefrontRouteCatalogOperationInProgress)
}
