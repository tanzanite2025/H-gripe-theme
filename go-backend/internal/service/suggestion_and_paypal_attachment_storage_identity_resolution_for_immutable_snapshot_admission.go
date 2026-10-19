package service

import (
	"context"

	"commerce-platform/internal/pkg/storage"
)

func (s *SuggestionAndPayPalAttachmentUploadService) ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference string) (string, error) {
	return storage.TrustedBusinessMediaReferenceObjectKey(reference, s.storage, s.siteURL)
}

func (s *SuggestionAndPayPalAttachmentUploadService) ResolveRegisteredMediaObjectKeyForCleanupOwnership(reference string) (string, error) {
	if s == nil || s.storage == nil {
		return "", ErrObjectStorageCleanupUnavailable
	}
	if key, err := s.storage.ObjectKey(reference); err == nil {
		if normalized, valid := storage.NormalizeObjectKey(key); valid {
			return normalized, nil
		}
	}
	return storage.TrustedBusinessMediaReferenceObjectKey(reference, s.storage, s.siteURL)
}

func (s *SuggestionAndPayPalAttachmentUploadService) ValidateMediaObjectExistsBeforeRegisteringOwnership(ctx context.Context, key string) error {
	opener, ok := s.storage.(storage.ObjectOpener)
	if !ok {
		return ErrObjectStorageCleanupUnavailable
	}
	object, err := opener.Open(ctx, key)
	if err != nil {
		return err
	}
	return object.ReadCloser.Close()
}
