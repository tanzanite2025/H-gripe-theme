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

var publicMediaAttachmentExtensions = []string{".jpg", ".jpeg", ".png", ".webp", ".gif", ".mp4", ".mov", ".webm"}

type MediaAssetFile struct {
	ReadCloser  io.ReadCloser
	RedirectURL string
	Filename    string
	MimeType    string
	Size        int64
	ModTime     time.Time
}

type PublicUploadAssetAccess struct {
	Found   bool
	Allowed bool
}

func (s *MediaService) OpenAssetFile(ctx context.Context, id uint) (*MediaAssetFile, error) {
	asset, err := s.GetAsset(id)
	if err != nil {
		return nil, err
	}

	key := s.mediaAssetObjectKey(asset)
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
	access, err := s.PublicUploadAssetAccess(key)
	if err != nil {
		return false, err
	}
	if !access.Found {
		return true, nil
	}
	return access.Allowed, nil
}

func (s *MediaService) PublicUploadAssetAccess(key string) (PublicUploadAssetAccess, error) {
	asset, err := s.repo.FindAssetByStorageKey(key)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			derivative, derivativeErr := s.repo.FindDerivativeByStorageKey(key)
			if derivativeErr != nil {
				if errors.Is(derivativeErr, gorm.ErrRecordNotFound) {
					return PublicUploadAssetAccess{}, nil
				}
				return PublicUploadAssetAccess{}, derivativeErr
			}
			if derivative.Asset == nil {
				return PublicUploadAssetAccess{Found: true, Allowed: false}, nil
			}
			return PublicUploadAssetAccess{
				Found:   true,
				Allowed: derivative.Asset.Status == "active" && derivative.Asset.Visibility == "public",
			}, nil
		}
		return PublicUploadAssetAccess{}, err
	}
	return PublicUploadAssetAccess{
		Found:   true,
		Allowed: asset.Status == "active" && asset.Visibility == "public",
	}, nil
}

func (s *MediaService) PublicMediaDimensions(reference string) (int, int, bool) {
	if s == nil || s.repo == nil {
		return 0, 0, false
	}
	key, ok := publicUploadStorageKey(reference)
	if !ok {
		return 0, 0, false
	}
	asset, err := s.repo.FindAssetByStorageKey(key)
	if err != nil || asset == nil {
		return 0, 0, false
	}
	if asset.DeletedAt.Valid || asset.Status != "active" || asset.Visibility != "public" || asset.MediaType != "image" {
		return 0, 0, false
	}
	if asset.Width <= 0 || asset.Height <= 0 {
		return 0, 0, false
	}
	return asset.Width, asset.Height, true
}

// CanonicalPublicMediaURL converts first-party upload references to the
// storefront's public upload URL. Database rows can retain an internal
// storage origin, but public response contracts must not reproduce it.
// Non-upload media remains untouched because it may intentionally point to a
// third-party CDN.
func (s *MediaService) CanonicalPublicMediaURL(reference string) string {
	value := strings.TrimSpace(reference)
	if value == "" {
		return ""
	}

	uploadPath, ok := canonicalPublicUploadPath(value)
	if !ok {
		return value
	}

	if s == nil || s.siteURL == "" {
		return uploadPath
	}
	return strings.TrimRight(s.siteURL, "/") + uploadPath
}

func (s *MediaService) CanonicalPublicImageUploadURL(reference string) (string, error) {
	if s == nil || s.repo == nil {
		return "", ErrMediaAssetURLUnavailable
	}

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
	if strings.ToLower(strings.TrimSpace(asset.MediaType)) != "image" || !allowedPublicImageAttachmentMimeType(asset.MimeType) {
		return "", ErrUnsupportedMediaType
	}
	if strings.TrimSpace(asset.URL) == "" {
		return "", ErrMediaAssetURLUnavailable
	}
	if refs[0].SourceHost != "" && !sameAttachmentReferenceHost(refs[0].SourceHost, asset.URL) {
		return "", ugc.ErrAttachmentInvalidURL
	}
	return s.CanonicalPublicMediaURL(asset.URL), nil
}

func (s *MediaService) CanonicalPublicMediaAttachmentURL(reference string) (string, error) {
	if s == nil || s.repo == nil {
		return "", ErrMediaAssetURLUnavailable
	}

	refs, err := ugc.NormalizeUploadAttachmentReferences([]string{reference}, 1, publicMediaAttachmentExtensions)
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
	switch strings.ToLower(strings.TrimSpace(asset.MediaType)) {
	case "image":
		if !allowedPublicImageAttachmentMimeType(asset.MimeType) {
			return "", ErrUnsupportedMediaType
		}
	case "video":
		if !allowedPublicVideoAttachmentMimeType(asset.MimeType) {
			return "", ErrUnsupportedMediaType
		}
	default:
		return "", ErrUnsupportedMediaType
	}
	if strings.TrimSpace(asset.URL) == "" {
		return "", ErrMediaAssetURLUnavailable
	}
	if refs[0].SourceHost != "" && !sameAttachmentReferenceHost(refs[0].SourceHost, asset.URL) {
		return "", ugc.ErrAttachmentInvalidURL
	}
	return s.CanonicalPublicMediaURL(asset.URL), nil
}

func canonicalPublicUploadPath(reference string) (string, bool) {
	parsed, err := url.Parse(strings.TrimSpace(reference))
	if err != nil {
		return "", false
	}

	candidatePath := parsed.Path
	if candidatePath == "" && strings.HasPrefix(reference, "/uploads/") {
		candidatePath = reference
	}
	if !strings.HasPrefix(candidatePath, "/") {
		candidatePath = "/" + candidatePath
	}

	markerIndex := strings.Index(candidatePath, "/uploads/")
	if markerIndex < 0 {
		return "", false
	}

	key, ok := storage.NormalizeObjectKey(candidatePath[markerIndex+len("/uploads/"):])
	if !ok {
		return "", false
	}

	path := "/uploads/" + key
	if parsed.RawQuery != "" {
		path += "?" + parsed.RawQuery
	}
	if parsed.Fragment != "" {
		path += "#" + parsed.Fragment
	}
	return path, true
}

func publicUploadStorageKey(reference string) (string, bool) {
	value := strings.TrimSpace(reference)
	if value == "" {
		return "", false
	}
	parsed, err := url.Parse(value)
	if err != nil {
		return "", false
	}

	candidatePath := parsed.Path
	if candidatePath == "" && strings.HasPrefix(value, "/uploads/") {
		candidatePath = value
		if queryIndex := strings.Index(candidatePath, "?"); queryIndex >= 0 {
			candidatePath = candidatePath[:queryIndex]
		}
		if fragmentIndex := strings.Index(candidatePath, "#"); fragmentIndex >= 0 {
			candidatePath = candidatePath[:fragmentIndex]
		}
	}
	if !strings.HasPrefix(candidatePath, "/") {
		candidatePath = "/" + candidatePath
	}

	markerIndex := strings.Index(candidatePath, "/uploads/")
	if markerIndex < 0 {
		return "", false
	}
	return storage.NormalizeObjectKey(candidatePath[markerIndex+len("/uploads/"):])
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

func allowedPublicVideoAttachmentMimeType(value string) bool {
	value = strings.ToLower(strings.TrimSpace(strings.SplitN(value, ";", 2)[0]))
	if value == "" {
		return true
	}
	switch value {
	case "video/mp4", "video/quicktime", "video/webm":
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

func (s *MediaService) mediaAssetObjectKey(asset *media.MediaAsset) string {
	if s != nil && s.storage != nil && asset != nil {
		if key, err := s.storage.ObjectKey(asset.URL); err == nil && key != "" {
			return key
		}
	}
	return assetStorageKey(asset)
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
