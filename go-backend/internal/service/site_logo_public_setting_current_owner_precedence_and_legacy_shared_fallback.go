package service

import (
	"strings"

	"commerce-platform/internal/pkg/storage"
)

// Shared, external and static legacy settings remain usable without a current
// dedicated logo. An obsolete dedicated key cannot pass the current-owner gate.
func (s *SiteLogoService) PublicSettingURLFromCurrentOwnerOrLegacySharedFallback(reference string) string {
	if current := s.CurrentPublicURL(); current != "" {
		return current
	}
	value := strings.TrimSpace(reference)
	if s == nil || value == "" {
		return value
	}
	key, err := storage.TrustedBusinessMediaReferenceObjectKey(value, s.storage, s.siteURL)
	if err == nil && IsSiteLogoStorageKey(key) {
		return ""
	}
	return value
}
