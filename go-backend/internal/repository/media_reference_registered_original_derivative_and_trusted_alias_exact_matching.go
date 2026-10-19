package repository

import (
	"net/url"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

func (r *MediaRepository) includeRegisteredDerivativeURLsAndObjectKeysInMediaReferenceQuery(query mediaAssetReferenceQuery, asset *media.MediaAsset) (mediaAssetReferenceQuery, error) {
	if asset.StorageKey != "" {
		query.ObjectKeys = append(query.ObjectKeys, strings.ReplaceAll(asset.StorageKey, "\\", "/"))
	}
	if !r.hasAssetDerivativesTable() {
		return query, nil
	}
	var derivatives []media.MediaAssetDerivative
	// Older HTML may retain the URL of a replaced derivative. Its owner remains
	// relevant to retirement while that derivative registration still exists.
	if err := r.db.Unscoped().Select("url", "storage_key").Where("media_asset_id = ?", query.AssetID).Find(&derivatives).Error; err != nil {
		return query, err
	}
	for _, derivative := range derivatives {
		if derivative.URL != "" {
			query.URLs = append(query.URLs, derivative.URL)
		}
		if derivative.StorageKey != "" {
			key := strings.ReplaceAll(derivative.StorageKey, "\\", "/")
			query.ObjectKeys = append(query.ObjectKeys, key)
			query.URLs = append(query.URLs, "/uploads/"+key, "uploads/"+key)
		}
	}
	return query, nil
}

func mediaReferenceURLWithoutQueryOrFragment(reference string) string {
	parsed, err := url.Parse(strings.TrimSpace(reference))
	if err != nil {
		return strings.TrimSpace(reference)
	}
	parsed.RawQuery, parsed.Fragment, parsed.RawFragment, parsed.ForceQuery = "", "", "", false
	return parsed.String()
}

func (r *MediaRepository) matchesExactRegisteredOrTrustedMediaReference(query mediaAssetReferenceQuery, reference string) bool {
	if strings.TrimSpace(reference) == "" {
		return false
	}
	normalized := mediaReferenceURLWithoutQueryOrFragment(reference)
	for _, candidate := range query.URLs {
		if normalized == mediaReferenceURLWithoutQueryOrFragment(candidate) {
			return true
		}
	}
	if query.IncludeAllExactObjectRegistrations {
		// Physical retention recognizes configured private origins too. This
		// branch does not authorize public reuse or expose attachment bytes.
		key := r.resolveRegisteredMediaObjectKeyForCleanupOwnership(reference)
		if key != "" && registeredMediaReferenceQueryContainsExactObjectKey(query, key) {
			return true
		}
	}
	key, err := storage.TrustedBusinessMediaReferenceObjectKey(reference, nil, "")
	if err == nil && r.businessMediaReferenceResolver != nil {
		key, err = r.businessMediaReferenceResolver.ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference)
	}
	if err != nil || key == "" {
		return false
	}
	for _, candidate := range query.ObjectKeys {
		if key == candidate {
			return true
		}
	}
	return false
}
