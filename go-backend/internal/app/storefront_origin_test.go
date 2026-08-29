package app

import (
	"testing"

	"commerce-platform/internal/pkg/config"
)

func TestResolveStorefrontOriginsDefaultsToDevStorefront(t *testing.T) {
	t.Setenv("STOREFRONT_BASE_URL", "")
	t.Setenv("STOREFRONT_INTERNAL_ORIGIN", "")

	publicOrigin, internalOrigin := resolveStorefrontOrigins(&config.Config{
		Server: config.ServerConfig{
			Mode:    "debug",
			BaseURL: "http://localhost:9200",
		},
	})

	if publicOrigin != defaultDevStorefrontOrigin {
		t.Fatalf("public origin = %q, want %q", publicOrigin, defaultDevStorefrontOrigin)
	}
	if internalOrigin != defaultDevStorefrontOrigin {
		t.Fatalf("internal origin = %q, want %q", internalOrigin, defaultDevStorefrontOrigin)
	}
}

func TestResolveStorefrontOriginsPreservesExplicitOrigins(t *testing.T) {
	t.Setenv("STOREFRONT_BASE_URL", "https://store.example.com/")
	t.Setenv("STOREFRONT_INTERNAL_ORIGIN", "http://storefront:3000/")

	publicOrigin, internalOrigin := resolveStorefrontOrigins(&config.Config{
		Server: config.ServerConfig{Mode: "release", BaseURL: "https://api.example.com"},
	})

	if publicOrigin != "https://store.example.com" {
		t.Fatalf("public origin = %q, want https://store.example.com", publicOrigin)
	}
	if internalOrigin != "http://storefront:3000" {
		t.Fatalf("internal origin = %q, want http://storefront:3000", internalOrigin)
	}
}

func TestResolveStorefrontOriginsDoesNotUsePublicOriginForProductionInternalTraffic(t *testing.T) {
	t.Setenv("STOREFRONT_BASE_URL", "https://store.example.com")
	t.Setenv("STOREFRONT_INTERNAL_ORIGIN", "")

	publicOrigin, internalOrigin := resolveStorefrontOrigins(&config.Config{
		Server: config.ServerConfig{Mode: "release", BaseURL: "https://api.example.com"},
	})

	if publicOrigin != "https://store.example.com" {
		t.Fatalf("public origin = %q, want https://store.example.com", publicOrigin)
	}
	if internalOrigin != "" {
		t.Fatalf("internal origin = %q, want empty when production origin is not configured", internalOrigin)
	}
}

func TestResolveSiteQualityTargetOriginFallsBackToPublicOrigin(t *testing.T) {
	t.Setenv("SITE_QUALITY_TARGET_ORIGIN", "")

	if got := resolveSiteQualityTargetOrigin("http://localhost:9199"); got != "http://localhost:9199" {
		t.Fatalf("Site Quality target origin = %q, want http://localhost:9199", got)
	}
}

func TestResolveSiteQualityTargetOriginTrimsExplicitOrigin(t *testing.T) {
	t.Setenv("SITE_QUALITY_TARGET_ORIGIN", " http://host.docker.internal:9199/// ")

	if got := resolveSiteQualityTargetOrigin("http://localhost:9199"); got != "http://host.docker.internal:9199" {
		t.Fatalf("Site Quality target origin = %q, want http://host.docker.internal:9199", got)
	}
}
