package service

import (
	"commerce-platform/internal/domain/media"
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMediaAssetInUse                 = errors.New("media asset is currently in use")
	ErrMediaDeleteConfirmationRequired = errors.New("media delete confirmation is invalid")
)

type MediaAssetInUseError struct {
	References []media.AssetReference
}

func (e *MediaAssetInUseError) Error() string {
	return ErrMediaAssetInUse.Error()
}

func (e *MediaAssetInUseError) Unwrap() error {
	return ErrMediaAssetInUse
}

func MediaAssetDeleteConfirmation(id uint) string {
	return fmt.Sprintf("DELETE %d", id)
}

// DeleteAsset permanently removes an unreferenced asset from storage and the
// database. The reference check is repeated here so the client-side dialog
// cannot bypass safety rules.
func (s *MediaService) DeleteAsset(ctx context.Context, id uint, confirmation string) error {
	asset, err := s.GetAsset(id)
	if err != nil {
		return err
	}
	if !isMediaDeleteConfirmationValid(id, confirmation) {
		return ErrMediaDeleteConfirmationRequired
	}

	references, err := s.repo.FindAssetReferences(asset)
	if err != nil {
		return err
	}
	if len(references) > 0 {
		return &MediaAssetInUseError{References: references}
	}
	if s.storage == nil {
		return ErrMediaStorageUnavailable
	}

	if err := s.storage.Delete(ctx, asset.URL); err != nil {
		return fmt.Errorf("delete media object: %w", err)
	}
	if err := s.repo.HardDeleteAsset(id); err != nil {
		return fmt.Errorf("delete media asset record: %w", err)
	}
	return nil
}

func isMediaDeleteConfirmationValid(id uint, confirmation string) bool {
	return strings.TrimSpace(confirmation) == MediaAssetDeleteConfirmation(id)
}
