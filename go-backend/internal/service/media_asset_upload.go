package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path"
	"strings"

	"commerce-platform/internal/domain/media"
)

type MediaUploadInput struct {
	File       *multipart.FileHeader
	MediaType  string
	Alt        string
	Caption    string
	UploaderID uint
	Width      int
	Height     int
}

func (s *MediaService) UploadAsset(ctx context.Context, input MediaUploadInput) (*media.MediaAsset, error) {
	if input.File == nil {
		return nil, ErrMediaUploadFileRequired
	}

	mediaType, err := normalizeMediaType(input.MediaType)
	if err != nil {
		return nil, err
	}
	if input.UploaderID == 0 {
		return nil, ErrMediaUploadIdentityRequired
	}
	if err := s.ensureUploaderStorageQuota(input.UploaderID, input.File.Size); err != nil {
		return nil, err
	}

	contentSHA256, err := contentSHA256FromMultipartFile(input.File)
	if err != nil {
		return nil, err
	}

	url, err := s.storage.Upload(ctx, input.File)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(url) == "" {
		return nil, ErrMediaAssetURLUnavailable
	}

	copyrightClaim, err := s.buildCopyrightClaim(input, contentSHA256)
	if err != nil {
		_ = s.storage.Delete(ctx, url)
		return nil, err
	}

	asset := &media.MediaAsset{
		Filename:           path.Base(url),
		OriginalFilename:   input.File.Filename,
		URL:                url,
		StorageKey:         extractStorageKey(url),
		MimeType:           input.File.Header.Get("Content-Type"),
		MediaType:          mediaType,
		Size:               input.File.Size,
		Width:              input.Width,
		Height:             input.Height,
		ContentSHA256:      contentSHA256,
		CopyrightClaimJSON: copyrightClaim,
		Alt:                strings.TrimSpace(input.Alt),
		Caption:            strings.TrimSpace(input.Caption),
		UploaderID:         input.UploaderID,
		Status:             "active",
		Visibility:         "public",
	}

	if err := s.repo.CreateAsset(asset); err != nil {
		_ = s.storage.Delete(ctx, url)
		return nil, err
	}

	s.hydrateAssetAccessURL(asset)
	return asset, nil
}

func (s *MediaService) ensureUploaderStorageQuota(uploaderID uint, incomingSize int64) error {
	if s.accountStorageQuotaBytes <= 0 {
		return ErrMediaAccountStorageQuotaExceeded
	}
	if incomingSize <= 0 {
		return ErrMediaUploadFileRequired
	}

	used, err := s.repo.AssetStorageUsageByUploaderID(uploaderID)
	if err != nil {
		return fmt.Errorf("get media storage usage: %w", err)
	}
	if used >= s.accountStorageQuotaBytes || incomingSize > s.accountStorageQuotaBytes-used {
		return ErrMediaAccountStorageQuotaExceeded
	}
	return nil
}
