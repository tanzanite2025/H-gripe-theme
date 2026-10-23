package main

import (
	"encoding/json"
	"testing"

	"commerce-platform/internal/pkg/config"
	"commerce-platform/internal/pkg/maintenancegate"
	"commerce-platform/internal/pkg/storage"
	"github.com/stretchr/testify/require"
)

func TestProductCatalogProjectionRegistrationMatchesRuntimeViewBufferAndExcludesPrivateConfiguration(t *testing.T) {
	t.Setenv("CUSTOMER_SERVICE_SEARCH_INDEX_DELETE_URL", "")
	cfg := &config.Config{Redis: config.RedisConfig{Host: "private-product-redis.invalid", Port: 6379, DB: 7, Password: "private-product-redis-password"}}
	provider := &storage.Config{BaseURL: "https://private-product-media-config.invalid"}
	direct, err := collectRegisteredRecoveryExternalProjectionConfiguration(cfg, provider, provider)
	require.NoError(t, err)
	require.Equal(t, "sql_only", direct.ProductCatalog.SearchMode)
	require.Equal(t, "redis", direct.ProductCatalog.DetailCacheMode)
	require.Equal(t, "sql_direct", direct.ProductCatalog.ViewCountMode)
	require.Empty(t, direct.ProductCatalog.ViewCountKeysSHA256)
	cfg.Worker.ProductViewCountFlushEnabled = true
	buffered, err := collectRegisteredRecoveryExternalProjectionConfiguration(cfg, provider, provider)
	require.NoError(t, err)
	require.Equal(t, "redis_delta_buffer", buffered.ProductCatalog.ViewCountMode)
	require.Equal(t, 2, buffered.ProductCatalog.Version)
	require.Equal(t, maintenancegate.ProductViewCountImmutableBatchSQLReceiptProtocol, buffered.ProductCatalog.ViewCountFlushProtocol)
	require.Len(t, buffered.ProductCatalog.ViewCountFlushContractSHA256, 64)
	require.Equal(t, maintenancegate.FingerprintExternalProjectionConfiguration([]string{"product:view_counts", "product:view_counts:processing", "product:view_counts:processing_batch_id"}), buffered.ProductCatalog.ViewCountKeysSHA256)
	require.Equal(t, direct.ProductCatalog.DetailCacheContractSHA256, buffered.ProductCatalog.DetailCacheContractSHA256)
	require.NotEqual(t, maintenancegate.FingerprintExternalProjectionConfiguration(direct), maintenancegate.FingerprintExternalProjectionConfiguration(buffered))
	encoded, err := json.Marshal(buffered)
	require.NoError(t, err)
	for _, privateValue := range []string{cfg.Redis.Host, cfg.Redis.Password, provider.BaseURL, "product:view_counts"} {
		require.NotContains(t, string(encoded), privateValue)
	}
}
