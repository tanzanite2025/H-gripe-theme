package service

import (
	"context"
	"fmt"

	"commerce-platform/internal/pkg/storage"
)

// Production providers inspect the exact object at its origin without fetching
// its body. Adapters with only Open are closed immediately without reading.
func (s *MediaService) ValidateMediaObjectExistsBeforeRegisteringOwnership(ctx context.Context, key string) error {
	if s == nil || s.storage == nil {
		return ErrMediaStorageUnavailable
	}
	if verifier, available := s.storage.(storage.ObjectExistenceVerifier); available {
		if err := verifier.VerifyStoredObjectExists(ctx, key); err != nil {
			return fmt.Errorf("%w: object cannot be verified before registration", ErrMediaAssetURLUnavailable)
		}
		return nil
	}
	opener, available := s.storage.(storage.ObjectOpener)
	if !available {
		return ErrMediaStorageUnavailable
	}
	object, err := opener.Open(ctx, key)
	if err != nil {
		// Provider errors can contain private URLs or credentials. Preserve the
		// public classification, without exposing a provider error to the API.
		return fmt.Errorf("%w: object cannot be verified before registration", ErrMediaAssetURLUnavailable)
	}
	if object == nil || object.ReadCloser == nil {
		return ErrMediaAssetURLUnavailable
	}
	if err := object.ReadCloser.Close(); err != nil {
		return fmt.Errorf("%w: object verification did not complete", ErrMediaAssetURLUnavailable)
	}
	return nil
}
