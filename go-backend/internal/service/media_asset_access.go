package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/ugc"

	"gorm.io/gorm"
)

const protectedMediaURLPrefix = "/api/admin/media/assets"

const protectedMediaSignedURLTTL = 10 * time.Minute

type MediaAssetFile struct {
	ReadCloser  io.ReadCloser
	RedirectURL string
	Filename    string
	MimeType    string
	Size        int64
	ModTime     time.Time
}

func (s *MediaService) OpenAssetFile(ctx context.Context, id uint) (*MediaAssetFile, error) {
	asset, err := s.GetAsset(id)
	if err != nil {
		return nil, err
	}

	key := assetStorageKey(asset)
	if key == "" {
		return nil, ErrMediaAssetURLUnavailable
	}

	if opener, ok := s.storage.(storage.ObjectOpener); ok {
		object, err := opener.Open(ctx, key)
		if err != nil {
			return nil, err
		}
		filename := object.Name
		if filename == "" {
			filename = asset.Filename
		}
		mimeType := object.MimeType
		if mimeType == "" {
			mimeType = asset.MimeType
		}
		return &MediaAssetFile{
			ReadCloser: object.ReadCloser,
			Filename:   filename,
			MimeType:   mimeType,
			Size:       object.Size,
			ModTime:    object.ModTime,
		}, nil
	}

	if signer, ok := s.storage.(storage.PresignedURLProvider); ok {
		url, err := signer.GetPresignedURL(ctx, key, protectedMediaSignedURLTTL)
		if err != nil {
			return nil, err
		}
		return &MediaAssetFile{RedirectURL: url, Filename: asset.Filename, MimeType: asset.MimeType, Size: asset.Size}, nil
	}

	return nil, ErrMediaStorageUnavailable
}

func (s *MediaService) CanServePublicUpload(key string) (bool, error) {
	asset, err := s.repo.FindAssetByStorageKey(key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}
	return asset.Status == "active" && asset.Visibility == "public", nil
}

func (s *MediaService) CanonicalPublicImageUploadURL(reference string) (string, error) {
	refs, err := ugc.NormalizeUploadImageAttachmentReferences([]string{reference}, 1)
	if err != nil {
		return "", err
	}
	if len(refs) == 0 {
		return "", ErrMediaAssetURLUnavailable
	}

	asset, err := s.repo.FindAssetByStorageKey(refs[0].StorageKey)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", ErrMediaAssetNotFound
		}
		return "", err
	}
	if asset.DeletedAt.Valid || asset.Status != "active" || asset.Visibility != "public" {
		return "", ErrMediaAssetForbidden
	}
	if asset.MediaType != "image" || !allowedPublicImageAttachmentMimeType(asset.MimeType) {
		return "", ErrUnsupportedMediaType
	}
	if strings.TrimSpace(asset.URL) == "" {
		return "", ErrMediaAssetURLUnavailable
	}
	if refs[0].SourceHost != "" && !sameAttachmentReferenceHost(refs[0].SourceHost, asset.URL) {
		return "", ugc.ErrAttachmentInvalidURL
	}
	return strings.TrimSpace(asset.URL), nil
}

func sameAttachmentReferenceHost(sourceHost, assetURL string) bool {
	parsedAsset, err := url.Parse(strings.TrimSpace(assetURL))
	if err != nil || parsedAsset.Host == "" {
		return false
	}
	return normalizeAttachmentHost(sourceHost) == normalizeAttachmentHost(parsedAsset.Host)
}

func normalizeAttachmentHost(value string) string {
	parsed, err := url.Parse("//" + strings.TrimSpace(value))
	if err != nil || parsed.Host == "" {
		return strings.ToLower(strings.TrimSpace(value))
	}
	host := strings.ToLower(parsed.Hostname())
	port := parsed.Port()
	if port == "" || port == "80" || port == "443" {
		return host
	}
	return host + ":" + port
}

func allowedPublicImageAttachmentMimeType(value string) bool {
	value = strings.ToLower(strings.TrimSpace(strings.SplitN(value, ";", 2)[0]))
	if value == "" {
		return true
	}
	switch value {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return true
	default:
		return false
	}
}

func (s *MediaService) hydrateAssetListAccessURL(assets []media.MediaAsset) {
	for index := range assets {
		s.hydrateAssetAccessURL(&assets[index])
	}
}

func (s *MediaService) hydrateAssetAccessURL(asset *media.MediaAsset) {
	if asset == nil {
		return
	}
	if asset.Visibility == "private" {
		asset.AccessURL = fmt.Sprintf("%s/%d/file", protectedMediaURLPrefix, asset.ID)
		return
	}
	asset.AccessURL = asset.URL
}

func assetStorageKey(asset *media.MediaAsset) string {
	if asset == nil {
		return ""
	}
	if key := strings.Trim(strings.ReplaceAll(asset.StorageKey, "\\", "/"), "/"); key != "" {
		return key
	}
	return extractStorageKey(asset.URL)
}

func extractStorageKey(url string) string {
	cleanURL := strings.TrimSpace(url)
	if cleanURL == "" {
		return ""
	}

	if markerIndex := strings.Index(cleanURL, "/uploads/"); markerIndex >= 0 {
		return strings.Trim(strings.ReplaceAll(cleanURL[markerIndex+len("/uploads/"):], "\\", "/"), "/")
	}

	return strings.Trim(strings.ReplaceAll(path.Base(cleanURL), "\\", "/"), "/")
}
