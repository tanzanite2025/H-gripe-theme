package maintenancegate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProductViewCountReceiptProjectionRequiresProtocolThreeAndRejectsOlderFlusherRegistration(t *testing.T) {
	projection := ExternalProjectionBoundary{Version: 1, CustomerServiceSearchMode: "sql_only", RedisDestinationSHA256: strings.Repeat("a", 64), CustomerServiceRealtimeMode: "process_local", MediaDeliveryConfigurationSHA256: strings.Repeat("b", 64)}
	legacyProduct := ProductCatalogExternalProjectionBoundary{Version: 1, SearchMode: "sql_only", DetailCacheMode: "redis", DetailCacheContractSHA256: strings.Repeat("c", 64), ViewCountMode: "redis_delta_buffer", ViewCountKeysSHA256: strings.Repeat("d", 64)}
	legacyBytes, err := json.Marshal(legacyProduct)
	require.NoError(t, err)
	require.NotContains(t, string(legacyBytes), "view_count_flush_")
	currentProduct := legacyProduct
	currentProduct.Version = 2
	currentProduct.ViewCountFlushProtocol = ProductViewCountImmutableBatchSQLReceiptProtocol
	currentProduct.ViewCountFlushContractSHA256 = strings.Repeat("e", 64)
	projection.ProductCatalog = &currentProduct
	directory := t.TempDir()
	roots, keys := map[string]string{"uploads": t.TempDir()}, map[string]string{"JWT_SECRET": "synthetic-recovery-key"}
	require.NoError(t, RegisterApplicationBoundary(directory, "storefront", "new-buffer-build", roots, keys, projection))
	data, err := os.ReadFile(filepath.Join(directory, "application-boundary.json"))
	require.NoError(t, err)
	var boundary ApplicationBoundary
	require.NoError(t, json.Unmarshal(data, &boundary))
	require.Equal(t, 3, boundary.Protocol)
	require.NoError(t, boundary.ValidateRegisteredRecoveryProjectionProtocol())
	boundary.Protocol = 2
	require.Error(t, boundary.ValidateRegisteredRecoveryProjectionProtocol(), "protocol-2 binaries can ignore product fields; they must reject protocol 3")
	require.NoError(t, os.WriteFile(filepath.Join(directory, "restored-data-identity-boundary.json"), data, 0600))
	older := projection
	older.ProductCatalog = &legacyProduct
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "older-buffer-build", roots, keys, older))
	older.ProductCatalog = nil
	require.Error(t, RegisterApplicationBoundary(directory, "storefront", "pre-product-build", roots, keys, older))
	require.NoError(t, RegisterApplicationBoundary(directory, "storefront", "new-buffer-build-upgrade", roots, keys, projection))
}

func TestProductViewCountReceiptProjectionRejectsContradictoryProtocolClaims(t *testing.T) {
	valid := ProductCatalogExternalProjectionBoundary{Version: 2, SearchMode: "sql_only", DetailCacheMode: "redis", DetailCacheContractSHA256: strings.Repeat("a", 64), ViewCountMode: "redis_delta_buffer", ViewCountKeysSHA256: strings.Repeat("b", 64), ViewCountFlushProtocol: ProductViewCountImmutableBatchSQLReceiptProtocol, ViewCountFlushContractSHA256: strings.Repeat("c", 64)}
	for _, scenario := range []string{"valid", "legacy_with_receipts", "unknown_protocol", "missing_contract", "direct_with_receipts", "future_version"} {
		t.Run(scenario, func(t *testing.T) {
			changed := valid
			switch scenario {
			case "legacy_with_receipts":
				changed.Version = 1
			case "unknown_protocol":
				changed.ViewCountFlushProtocol = "future"
			case "missing_contract":
				changed.ViewCountFlushContractSHA256 = ""
			case "direct_with_receipts":
				changed.ViewCountMode = "sql_direct"
				changed.ViewCountKeysSHA256 = ""
			case "future_version":
				changed.Version = 3
			}
			if scenario == "valid" {
				require.NoError(t, changed.Validate())
			} else {
				require.Error(t, changed.Validate())
			}
		})
	}
}
