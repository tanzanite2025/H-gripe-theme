package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	seodomain "commerce-platform/internal/domain/seo"
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
