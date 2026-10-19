package service

import (
	"bufio"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"strings"
)

func (s *MediaService) validateMediaProviderReturnedCanonicalURLAndExactObjectKey(reference, expectedKey string) error {
	if strings.TrimSpace(reference) == "" {
		return ErrMediaAssetURLUnavailable
	}
	// Require the owning provider canonical URL, not a foreign host with the same suffix.
	if reference != s.storage.GetURL(expectedKey) {
		return ErrMediaAssetURLUnavailable
	}
	key, err := s.storage.ObjectKey(reference)
	if err != nil || key != expectedKey {
		return ErrMediaAssetURLUnavailable
	}
	return nil
}

func (s *MediaService) verifyStoredMediaObjectSourceSizeDigestAndMimeType(ctx context.Context, asset *media.MediaAsset) error {
	object, err := s.storage.(storage.ObjectOpener).Open(ctx, asset.StorageKey)
	if err != nil {
		return fmt.Errorf("%w: stored media source cannot be read", ErrMediaAssetURLUnavailable)
	}
	if object == nil || object.ReadCloser == nil {
		return ErrMediaAssetURLUnavailable
	}
	reader := bufio.NewReader(io.LimitReader(object.ReadCloser, asset.Size+1))
	header, peekError := reader.Peek(512)
	actualMime := upload.DetectContentTypeFromHeaderIncludingVideoFormats(header)
	digest := sha256.New()
	copied, readError := io.Copy(digest, reader)
	closeError := object.ReadCloser.Close()
	if (peekError != nil && !errors.Is(peekError, io.EOF)) || readError != nil || closeError != nil || copied != asset.Size || fmt.Sprintf("%x", digest.Sum(nil)) != asset.ContentSHA256 || actualMime != asset.MimeType {
		return fmt.Errorf("%w: stored media source integrity mismatch", ErrMediaAssetURLUnavailable)
	}
	return ctx.Err()
}
