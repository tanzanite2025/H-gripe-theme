package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"strings"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
)

var (
	ErrHomeVisualTileKeyRequired           = errors.New("visual showcase key is required")
	ErrHomeVisualTileLocaleRequired        = errors.New("visual showcase locale is required")
	ErrHomeVisualTileItemLimit             = errors.New("visual showcase item limit exceeded")
	ErrHomeVisualTileTitleRequired         = errors.New("visual showcase item title is required")
	ErrHomeVisualTileAltTextRequired       = errors.New("visual showcase item alt text is required")
	ErrHomeVisualTileImageRequired         = errors.New("visual showcase item image is required")
	ErrHomeVisualTileImageInvalid          = errors.New("visual showcase item image is not a visual showcase upload")
	ErrHomeVisualTileStorageUnavailable    = errors.New("visual showcase storage is unavailable")
	ErrHomeVisualTileUploadFileRequired    = errors.New("visual showcase upload file is required")
	ErrHomeVisualTileAspectRatioInvalid    = errors.New("visual showcase image aspect ratio is invalid")
	ErrHomeVisualTilePreviousDestroyFailed = errors.New("previous visual showcase image could not be destroyed")
)

const (
	HomeMainProductCategoriesTileSetKey = "home-main-product-categories"
	maxHomeVisualTileItems               = 100
	HomeVisualTileStoragePrefix          = "visual-showcase"
	homeVisualTileImageCacheControl      = "public, max-age=31536000, immutable"
)

type HomeVisualTileInput struct {
	ImageURL        string
	ThumbnailURL    string
	StorageKey      string
	Title           string
	Caption         string
	AltText         string
	DesktopOrder    int
	MobilePairIndex int
	TargetURL       string
	TargetLabel     string
	LayoutVariant   string
	IsPublished     bool
	Width           int
	Height          int
}

type HomeVisualTileImageUpload struct {
	ImageURL     string `json:"image_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	StorageKey   string `json:"storage_key"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

type HomeVisualTilePublishedResult struct {
	Items           []homevisualtile.Tile
	Locale          string
	RequestedLocale string
	Fallback        bool
	ConfiguredCount int64
}

type HomeVisualTileService struct {
	repo    *repository.HomeVisualTileRepository
	storage storage.StorageService
}

func NewHomeVisualTileService(repo *repository.HomeVisualTileRepository, storageSvc storage.StorageService) *HomeVisualTileService {
	return &HomeVisualTileService{repo: repo, storage: storageSvc}
}

func (s *HomeVisualTileService) GetPublishedItems(tileSetKey, locale string) ([]homevisualtile.Tile, error) {
	result, err := s.GetPublishedResult(tileSetKey, locale)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *HomeVisualTileService) GetPublishedResult(tileSetKey, locale string) (*HomeVisualTilePublishedResult, error) {
	key := strings.TrimSpace(tileSetKey)
	normalizedLocale := strings.TrimSpace(locale)
	if key == "" {
		return nil, ErrHomeVisualTileKeyRequired
	}
	if normalizedLocale == "" {
		normalizedLocale = "en"
	}

	items, err := s.repo.ListItems(key, normalizedLocale, true)
	if err != nil {
		return nil, err
	}
	configuredCount, err := s.repo.CountItems(key, normalizedLocale, false)
	if err != nil {
		return nil, err
	}
	if configuredCount > 0 || normalizedLocale == "en" {
		return &HomeVisualTilePublishedResult{
			Items:           items,
			Locale:          normalizedLocale,
			RequestedLocale: normalizedLocale,
			Fallback:        false,
			ConfiguredCount: configuredCount,
		}, nil
	}

	fallbackItems, err := s.repo.ListItems(key, "en", true)
	if err != nil {
		return nil, err
	}
	fallbackConfiguredCount, err := s.repo.CountItems(key, "en", false)
	if err != nil {
		return nil, err
	}
	return &HomeVisualTilePublishedResult{
		Items:           fallbackItems,
		Locale:          "en",
		RequestedLocale: normalizedLocale,
		Fallback:        true,
		ConfiguredCount: fallbackConfiguredCount,
	}, nil
}

func (s *HomeVisualTileService) GetAdminItems(tileSetKey, locale string) ([]homevisualtile.Tile, error) {
	key := strings.TrimSpace(tileSetKey)
	normalizedLocale := strings.TrimSpace(locale)
	if key == "" {
		return nil, ErrHomeVisualTileKeyRequired
	}
	if normalizedLocale == "" {
		return nil, ErrHomeVisualTileLocaleRequired
	}
	return s.repo.ListItems(key, normalizedLocale, false)
}

func (s *HomeVisualTileService) UploadAdminImage(
	ctx context.Context,
	tileSetKey string,
	locale string,
	file *multipart.FileHeader,
) (*HomeVisualTileImageUpload, error) {
	key := strings.TrimSpace(tileSetKey)
	normalizedLocale := strings.TrimSpace(locale)
	if key == "" {
		return nil, ErrHomeVisualTileKeyRequired
	}
	if normalizedLocale == "" {
		return nil, ErrHomeVisualTileLocaleRequired
	}
	if s == nil || s.storage == nil {
		return nil, ErrHomeVisualTileStorageUnavailable
	}
	if file == nil {
		return nil, ErrHomeVisualTileUploadFileRequired
	}
	specCode := upload.SpecVisualShowcaseEditorial
	if key == HomeMainProductCategoriesTileSetKey {
		specCode = upload.SpecVisualShowcaseHomeCategories
	}
	if err := upload.ValidateSpecFile(file, string(specCode)); err != nil {
		return nil, err
	}

	width, height, err := upload.ReadImageDimensions(file)
	if err != nil {
		return nil, err
	}
	if !isValidHomeVisualTileAspectRatioForTileSet(key, width, height) {
		return nil, homeVisualTileAspectRatioError(key, width, height)
	}

	imageURL, err := s.uploadVisualShowcaseImage(ctx, key, normalizedLocale, file)
	if err != nil {
		return nil, err
	}
	storageKey := s.visualShowcaseStorageKeyFromReference(imageURL)
	if storageKey == "" {
		_ = s.storage.Delete(ctx, imageURL)
		return nil, ErrHomeVisualTileImageInvalid
	}

	return &HomeVisualTileImageUpload{
		ImageURL:     imageURL,
		ThumbnailURL: imageURL,
		StorageKey:   storageKey,
		Width:        width,
		Height:       height,
	}, nil
}

func (s *HomeVisualTileService) ReplaceAdminItems(
	ctx context.Context,
	tileSetKey string,
	locale string,
	inputs []HomeVisualTileInput,
) ([]homevisualtile.Tile, error) {
	key := strings.TrimSpace(tileSetKey)
	normalizedLocale := strings.TrimSpace(locale)
	if key == "" {
		return nil, ErrHomeVisualTileKeyRequired
	}
	if normalizedLocale == "" {
		return nil, ErrHomeVisualTileLocaleRequired
	}
	if len(inputs) > maxHomeVisualTileItems {
		return nil, ErrHomeVisualTileItemLimit
	}

	previousItems, err := s.repo.ListItems(key, normalizedLocale, false)
	if err != nil {
		return nil, err
	}

	retainedStorageKeys := make(map[string]struct{}, len(inputs))
	items := make([]homevisualtile.Tile, 0, len(inputs))
	for index, input := range inputs {
		title := strings.TrimSpace(input.Title)
		altText := strings.TrimSpace(input.AltText)
		imageURL := strings.TrimSpace(input.ImageURL)
		thumbnailURL := strings.TrimSpace(input.ThumbnailURL)
		if thumbnailURL == "" {
			thumbnailURL = imageURL
		}
		if title == "" {
			return nil, ErrHomeVisualTileTitleRequired
		}
		if altText == "" {
			return nil, ErrHomeVisualTileAltTextRequired
		}
		if imageURL == "" {
			return nil, ErrHomeVisualTileImageRequired
		}

		storageKey := s.visualShowcaseStorageKeyFromInput(input.StorageKey, imageURL)
		if storageKey == "" {
			return nil, ErrHomeVisualTileImageInvalid
		}
		retainedStorageKeys[storageKey] = struct{}{}

		layoutVariant := strings.TrimSpace(input.LayoutVariant)
		if layoutVariant == "" {
			layoutVariant = "standard"
		}
		desktopOrder := input.DesktopOrder
		if desktopOrder <= 0 {
			desktopOrder = index + 1
		}
		mobilePairIndex := input.MobilePairIndex
		if mobilePairIndex < 0 {
			mobilePairIndex = 0
		}
		width := input.Width
		if width <= 0 {
			width = 900
		}
		height := input.Height
		if height <= 0 {
			height = 1200
		}
		if !isValidHomeVisualTileAspectRatioForTileSet(key, width, height) {
			return nil, homeVisualTileAspectRatioError(key, width, height)
		}

		items = append(items, homevisualtile.Tile{
			TileSetKey:     key,
			Locale:          normalizedLocale,
			ImageURL:        imageURL,
			ThumbnailURL:    thumbnailURL,
			StorageKey:      storageKey,
			Title:           title,
			Caption:         strings.TrimSpace(input.Caption),
			AltText:         altText,
			DesktopOrder:    desktopOrder,
			MobilePairIndex: mobilePairIndex,
			TargetURL:       strings.TrimSpace(input.TargetURL),
			TargetLabel:     strings.TrimSpace(input.TargetLabel),
			LayoutVariant:   layoutVariant,
			IsPublished:     input.IsPublished,
			Width:           width,
			Height:          height,
		})
	}

	if err := s.repo.ReplaceItems(key, normalizedLocale, items); err != nil {
		return nil, err
	}
	if err := s.destroyUnreferencedVisualShowcaseImages(ctx, previousItems, retainedStorageKeys); err != nil {
		return nil, err
	}
	return s.repo.ListItems(key, normalizedLocale, false)
}

func (s *HomeVisualTileService) uploadVisualShowcaseImage(ctx context.Context, tileSetKey string, locale string, file *multipart.FileHeader) (string, error) {
	prefix := homeVisualTileStoragePrefix(tileSetKey, locale)
	if cacheControlled, ok := s.storage.(storage.CacheControlledObjectUploader); ok {
		return cacheControlled.UploadWithPrefixAndCacheControl(ctx, file, prefix, homeVisualTileImageCacheControl)
	}
	return s.storage.UploadWithPrefix(ctx, file, prefix)
}

func (s *HomeVisualTileService) destroyUnreferencedVisualShowcaseImages(
	ctx context.Context,
	previousItems []homevisualtile.Tile,
	retainedStorageKeys map[string]struct{},
) error {
	if s == nil || s.storage == nil {
		return nil
	}

	for _, item := range previousItems {
		storageKey := s.visualShowcaseStorageKeyFromInput(item.StorageKey, item.ImageURL)
		if storageKey == "" {
			continue
		}
		if _, retained := retainedStorageKeys[storageKey]; retained {
			continue
		}
		remainingReferences, err := s.repo.CountItemsByStorageKey(storageKey)
		if err != nil {
			return err
		}
		if remainingReferences > 0 {
			continue
		}
		if err := s.storage.Delete(ctx, storageKey); err != nil {
			return fmt.Errorf("%w: %v", ErrHomeVisualTilePreviousDestroyFailed, err)
		}
	}
	return nil
}

func (s *HomeVisualTileService) visualShowcaseStorageKeyFromInput(inputStorageKey string, imageURL string) string {
	normalizedStorageKey, hasStorageKey := storage.NormalizeObjectKey(inputStorageKey)
	if !hasStorageKey || !IsHomeVisualTileStorageKey(normalizedStorageKey) {
		return s.visualShowcaseStorageKeyFromReference(imageURL)
	}

	referencedStorageKey := s.visualShowcaseStorageKeyFromReference(imageURL)
	if referencedStorageKey == "" || referencedStorageKey != normalizedStorageKey {
		return ""
	}
	return normalizedStorageKey
}

func (s *HomeVisualTileService) visualShowcaseStorageKeyFromReference(reference string) string {
	value := strings.TrimSpace(reference)
	if value == "" {
		return ""
	}
	if s != nil && s.storage != nil {
		if key, err := s.storage.ObjectKey(value); err == nil && IsHomeVisualTileStorageKey(key) {
			return key
		}
	}
	return ""
}

func homeVisualTileStoragePrefix(tileSetKey string, locale string) string {
	prefix, ok := storage.NormalizeObjectKey(path.Join(
		HomeVisualTileStoragePrefix,
		normalizeHomeVisualTilePathSegment(tileSetKey, "showcase"),
		normalizeHomeVisualTilePathSegment(locale, "global"),
	))
	if !ok {
		return HomeVisualTileStoragePrefix
	}
	return prefix
}

func normalizeHomeVisualTilePathSegment(value string, fallback string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	previousDash := false
	for _, r := range value {
		allowed := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if allowed {
			builder.WriteRune(r)
			previousDash = false
			continue
		}
		if r == '-' || r == '_' {
			builder.WriteRune(r)
			previousDash = false
			continue
		}
		if !previousDash {
			builder.WriteByte('-')
			previousDash = true
		}
	}
	normalized := strings.Trim(builder.String(), "-_")
	if normalized == "" {
		return fallback
	}
	return normalized
}

func IsHomeVisualTileStorageKey(key string) bool {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	return ok && (normalizedKey == HomeVisualTileStoragePrefix || strings.HasPrefix(normalizedKey, HomeVisualTileStoragePrefix+"/"))
}

func homeVisualTileAspectRatioError(tileSetKey string, width, height int) error {
	return fmt.Errorf(
		"%w: expected %s, received %dx%d",
		ErrHomeVisualTileAspectRatioInvalid,
		homeVisualTileAspectRatioLabel(tileSetKey),
		width,
		height,
	)
}

func homeVisualTileAspectRatioLabel(tileSetKey string) string {
	if strings.TrimSpace(tileSetKey) == HomeMainProductCategoriesTileSetKey {
		return "16:9"
	}
	return "3:4"
}

func isValidHomeVisualTileAspectRatioForTileSet(tileSetKey string, width, height int) bool {
	if strings.TrimSpace(tileSetKey) == HomeMainProductCategoriesTileSetKey {
		return isValidHomeVisualTileAspectRatio(width, height, 16, 9)
	}
	return isValidHomeVisualTileAspectRatio(width, height, 3, 4)
}

func isValidHomeVisualTileAspectRatio(width, height, ratioWidth, ratioHeight int) bool {
	return width > 0 && height > 0 && ratioWidth > 0 && ratioHeight > 0 && width*ratioHeight == height*ratioWidth
}
