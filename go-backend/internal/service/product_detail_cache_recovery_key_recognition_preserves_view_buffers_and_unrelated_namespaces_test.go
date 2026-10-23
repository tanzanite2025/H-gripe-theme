package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStoppedRecoveryRecognizesActualProductReadKeysWithoutClaimingViewBuffersOrOtherNamespaces(t *testing.T) {
	for _, key := range []string{productIDCacheKey(0), productIDCacheKey(123), productSlugCacheKey("legacy-slug", ""), productSlugCacheKey("old:slug", "EN-us"), productSlugCacheKey("old*slug?", "legacy:locale")} {
		detail, lease := RecognizeProductDetailCacheAndReadLeaseKeyForStoppedRecovery(key)
		require.True(t, detail, key)
		require.False(t, lease, key)
		detail, lease = RecognizeProductDetailCacheAndReadLeaseKeyForStoppedRecovery(productCacheLockKey(key))
		require.False(t, detail, key)
		require.True(t, lease, key)
	}
	for _, key := range []string{"product:view_counts", "product:view_counts:processing", "product:view_counts:processing_batch_id", "product:view_counts:recovery_ownership", "lock:product:view_counts", "product:01", "product:-1", "product:+1", "product:18446744073709551616", "product:slug:missing-locale-separator", "product:123:other", "lock:lock:product:1", "customer_service:stream", "outbox:pending", "idempotency:product:1"} {
		detail, lease := RecognizeProductDetailCacheAndReadLeaseKeyForStoppedRecovery(key)
		require.False(t, detail, key)
		require.False(t, lease, key)
	}
}
