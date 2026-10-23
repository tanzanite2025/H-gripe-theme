package maintenancegate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductCatalogProjectionRegistrationPreservesOmittedLegacyBoundaryAndRejectsUnknownModes(t *testing.T) {
	legacy := ExternalProjectionBoundary{Version: 1, CustomerServiceSearchMode: "sql_only", RedisDestinationSHA256: strings.Repeat("a", 64), CustomerServiceRealtimeMode: "process_local", MediaDeliveryConfigurationSHA256: strings.Repeat("b", 64)}
	data, err := json.Marshal(legacy)
	require.NoError(t, err)
	require.NotContains(t, string(data), "product_catalog")
	require.NoError(t, legacy.Validate())
	valid := ProductCatalogExternalProjectionBoundary{Version: 1, SearchMode: "sql_only", DetailCacheMode: "redis", DetailCacheContractSHA256: strings.Repeat("c", 64), ViewCountMode: "sql_direct"}
	legacy.ProductCatalog = &valid
	require.NoError(t, legacy.Validate())
	for _, scenario := range []struct {
		name   string
		change func(*ProductCatalogExternalProjectionBoundary)
	}{
		{"unknown_version", func(b *ProductCatalogExternalProjectionBoundary) { b.Version = 2 }},
		{"unknown_search", func(b *ProductCatalogExternalProjectionBoundary) { b.SearchMode = "external" }},
		{"unknown_cache", func(b *ProductCatalogExternalProjectionBoundary) { b.DetailCacheMode = "none" }},
		{"missing_cache_identity", func(b *ProductCatalogExternalProjectionBoundary) { b.DetailCacheContractSHA256 = "" }},
		{"unknown_views", func(b *ProductCatalogExternalProjectionBoundary) { b.ViewCountMode = "" }},
		{"direct_with_delta_keys", func(b *ProductCatalogExternalProjectionBoundary) { b.ViewCountKeysSHA256 = strings.Repeat("d", 64) }},
		{"buffer_without_delta_keys", func(b *ProductCatalogExternalProjectionBoundary) { b.ViewCountMode = "redis_delta_buffer" }},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			changed := valid
			scenario.change(&changed)
			legacy.ProductCatalog = &changed
			require.Error(t, legacy.Validate())
		})
	}
	valid.ViewCountMode, valid.ViewCountKeysSHA256 = "redis_delta_buffer", strings.Repeat("d", 64)
	require.NoError(t, valid.Validate())
}

func TestRestoredProductProjectionGuardRejectsAddedOmittedOrChangedViewPersistenceWithoutOverwritingRegistration(t *testing.T) {
	directory := t.TempDir()
	roots, keys := map[string]string{"uploads": t.TempDir()}, map[string]string{"JWT_SECRET": "synthetic-key"}
	legacy := ExternalProjectionBoundary{Version: 1, CustomerServiceSearchMode: "sql_only", RedisDestinationSHA256: strings.Repeat("a", 64), CustomerServiceRealtimeMode: "process_local", MediaDeliveryConfigurationSHA256: strings.Repeat("b", 64)}
	require.NoError(t, RegisterApplicationBoundary(directory, "storefront", "build-old", roots, keys, legacy))
	registrationPath, guardPath := filepath.Join(directory, "application-boundary.json"), filepath.Join(directory, "restored-data-identity-boundary.json")
	original, err := os.ReadFile(registrationPath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(guardPath, original, 0600))
	current := legacy
	current.ProductCatalog = &ProductCatalogExternalProjectionBoundary{Version: 1, SearchMode: "sql_only", DetailCacheMode: "redis", DetailCacheContractSHA256: strings.Repeat("c", 64), ViewCountMode: "sql_direct"}
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-new", roots, keys, current), "an old reconciled guard must not silently accept newly registered product configuration")
	unchanged, err := os.ReadFile(registrationPath)
	require.NoError(t, err)
	require.Equal(t, original, unchanged)
	var currentGuard ApplicationBoundary
	require.NoError(t, json.Unmarshal(original, &currentGuard))
	currentGuard.ExternalProjections = &current
	guardBytes, err := json.Marshal(currentGuard)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(guardPath, guardBytes, 0600))
	require.NoError(t, RegisterApplicationBoundary(directory, "storefront", "build-new", roots, keys, current))
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-old", roots, keys, legacy), "an older build must not omit the guarded product configuration")
	buffered := current
	changed := *current.ProductCatalog
	changed.ViewCountMode, changed.ViewCountKeysSHA256 = "redis_delta_buffer", strings.Repeat("d", 64)
	buffered.ProductCatalog = &changed
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "build-new", roots, keys, buffered))
}
