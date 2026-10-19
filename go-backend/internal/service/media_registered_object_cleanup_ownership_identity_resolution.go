package service

import "commerce-platform/internal/pkg/storage"

// Registered cleanup ownership includes public and private objects. The
// provider resolves complete identities; foreign hosts never match by suffix.
func (s *MediaService) ResolveRegisteredMediaObjectKeyForCleanupOwnership(reference string) (string, error) {
	if s == nil || s.storage == nil {
		return "", ErrMediaStorageUnavailable
	}
	if key, err := s.storage.ObjectKey(reference); err == nil {
		if normalized, valid := storage.NormalizeObjectKey(key); valid {
			return normalized, nil
		}
	}
	return storage.TrustedBusinessMediaReferenceObjectKey(reference, s.storage, s.siteURL)
}
