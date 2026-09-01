package service

import (
	"context"
	"fmt"
	"mime/multipart"
	"path"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/upload"
)

var trustedMediaUploadImageRule = upload.FileRule{
	MaxSize:             12 << 20,
	AllowedExtensions:   []string{".jpg", ".jpeg", ".png", ".webp", ".gif"},
	AllowedContentTypes: []string{"image/jpeg", "image/png", "image/webp", "image/gif"},
	MaxWidth:            8000,
	MaxHeight:           8000,
	MaxPixels:           24_000_000,
}

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

	trustedMimeType, err := trustedMediaUploadMimeType(input.File, mediaType)
	if err != nil {
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
		StorageKey:         storageObjectKey(s.storage, url),
		MimeType:           trustedMimeType,
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

	derivatives, err := s.generateAssetDerivatives(ctx, asset, input.File)
	if err != nil {
		_ = s.storage.Delete(ctx, asset.URL)
		_ = s.repo.HardDeleteAsset(asset.ID)
		return nil, err
	}
	if len(derivatives) > 0 {
		if err := s.repo.CreateAssetDerivatives(derivatives); err != nil {
			derivativeURLs := make([]string, 0, len(derivatives))
			for _, derivative := range derivatives {
				derivativeURLs = append(derivativeURLs, derivative.URL)
			}
			deleteUploadedMediaObjectsBestEffort(ctx, s.storage, derivativeURLs)
			_ = s.storage.Delete(ctx, asset.URL)
			_ = s.repo.HardDeleteAsset(asset.ID)
			return nil, err
		}
		asset.Derivatives = derivatives
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

func trustedMediaUploadMimeType(file *multipart.FileHeader, mediaType string) (string, error) {
	var rule upload.FileRule
	switch mediaType {
	case "image":
		rule = trustedMediaUploadImageRule
	case "video":
		rule = upload.ProductVideoRule
	default:
		return "", ErrUnsupportedMediaType
	}

	if err := upload.ValidateFile(file, rule); err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnsupportedMediaType, err)
	}

	mimeType, err := upload.DetectContentType(file)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrUnsupportedMediaType, err)
	}
	if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(mimeType)), mediaType+"/") {
		return "", fmt.Errorf("%w: detected %s for %s upload", ErrUnsupportedMediaType, mimeType, mediaType)
	}
	return mimeType, nil
}
