package service

import "commerce-platform/internal/pkg/maintenancegate"

// The application builder always supplies Redis to product detail caching.
// Identify the actual key generators, including read leases and both view hashes.
// These facts describe configuration, never ownership or contents of Redis.
func DescribeProductCatalogRuntimeProjectionBoundaryForRecovery(bufferedViewCounts bool) maintenancegate.ProductCatalogExternalProjectionBoundary {
	idKey, slugKey := productIDCacheKey(1), productSlugCacheKey("recovery-contract-sample", "en")
	boundary := maintenancegate.ProductCatalogExternalProjectionBoundary{
		Version: 1, SearchMode: "sql_only", DetailCacheMode: "redis", ViewCountMode: "sql_direct",
		DetailCacheContractSHA256: maintenancegate.FingerprintExternalProjectionConfiguration([]string{idKey, slugKey, productCacheLockKey(idKey), productCacheLockKey(slugKey)}),
	}
	if bufferedViewCounts {
		boundary.Version = 2
		boundary.ViewCountMode = "redis_delta_buffer"
		boundary.ViewCountKeysSHA256 = maintenancegate.FingerprintExternalProjectionConfiguration(productViewCountActiveProcessingAndBatchIdentityRedisKeys())
		boundary.ViewCountFlushProtocol = maintenancegate.ProductViewCountImmutableBatchSQLReceiptProtocol
		boundary.ViewCountFlushContractSHA256 = maintenancegate.FingerprintExternalProjectionConfiguration([]string{boundary.ViewCountFlushProtocol, "product_view_count_flush_receipts", "batch_id,product_id", "delta", "product_increment_applied", productViewCountClaimImmutableProcessingBatchScript.Hash(), productViewCountAcknowledgeOnlyMatchingCommittedBatchScript.Hash()})
	}
	return boundary
}
