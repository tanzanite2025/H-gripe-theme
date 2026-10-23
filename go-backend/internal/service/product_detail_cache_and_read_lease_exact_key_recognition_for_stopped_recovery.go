package service

import (
	"strconv"
	"strings"
)

// RecognizeProductDetailCacheAndReadLeaseKeyForStoppedRecovery uses the same
// generators as live reads. It includes obsolete slug/locale variants without
// interpreting product view buffers, ownership markers or other product keys as
// caches. Callers must independently prove exclusive stopped runtime ownership.
func RecognizeProductDetailCacheAndReadLeaseKeyForStoppedRecovery(key string) (detailCache, readLease bool) {
	cacheKey := key
	lockPrefix := productCacheLockKey("")
	if strings.HasPrefix(cacheKey, lockPrefix) {
		readLease = true
		cacheKey = strings.TrimPrefix(cacheKey, lockPrefix)
	}
	idPrefix := strings.TrimSuffix(productIDCacheKey(0), "0")
	if idText, ok := strings.CutPrefix(cacheKey, idPrefix); ok {
		if id, err := strconv.ParseUint(idText, 10, strconv.IntSize); err == nil && productIDCacheKey(uint(id)) == cacheKey {
			return !readLease, readLease
		}
	}
	slugPrefix := strings.TrimSuffix(productSlugCacheKey("", ""), ":")
	if suffix, ok := strings.CutPrefix(cacheKey, slugPrefix); ok {
		// Slugs and legacy locale arguments can contain colons. A round trip
		// recognizes their key, without guessing any supported language list.
		if separator := strings.LastIndexByte(suffix, ':'); separator >= 0 && productSlugCacheKey(suffix[:separator], suffix[separator+1:]) == cacheKey {
			return !readLease, readLease
		}
	}
	return false, false
}
