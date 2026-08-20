package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"strings"

	"commerce-platform/internal/domain/visualshowcase"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
)

var (
	ErrVisualShowcaseKeyRequired           = errors.New("visual showcase key is required")
	ErrVisualShowcaseLocaleRequired        = errors.New("visual showcase locale is required")
	ErrVisualShowcaseItemLimit             = errors.New("visual showcase item limit exceeded")
	ErrVisualShowcaseTitleRequired         = errors.New("visual showcase item title is required")
	ErrVisualShowcaseAltTextRequired       = errors.New("visual showcase item alt text is required")
	ErrVisualShowcaseImageRequired         = errors.New("visual showcase item image is required")
	ErrVisualShowcaseImageInvalid          = errors.New("visual showcase item image is not a visual showcase upload")
	ErrVisualShowcaseStorageUnavailable    = errors.New("visual showcase storage is unavailable")
	ErrVisualShowcaseUploadFileRequired    = errors.New("visual showcase upload file is required")
	ErrVisualShowcaseAspectRatioInvalid    = errors.New("visual showcase image aspect ratio is invalid")
	ErrVisualShowcasePreviousDestroyFailed = errors.New("previous visual showcase image could not be destroyed")
)

const (
	HomeMainProductCategoriesShowcaseKey = "home-main-product-categories"
	maxVisualShowcaseItems               = 100
	VisualShowcaseStoragePrefix          = "visual-showcase"
	visualShowcaseImageCacheControl      = "public, max-age=31536000, immutable"
)

type VisualShowcaseItemInput struct {
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

type VisualShowcaseImageUpload struct {
	ImageURL     string `json:"image_url"`
	ThumbnailURL string `json:"thumbnail_url"`
	StorageKey   string `json:"storage_key"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
}

type VisualShowcasePublishedResult struct {
	Items           []visualshowcase.Item
	Locale          string
	RequestedLocale string
	Fallback        bool
	ConfiguredCount int64
}

type VisualShowcaseService struct {
	repo    *repository.VisualShowcaseRepository
	storage storage.StorageService
}

func NewVisualShowcaseService(repo *repository.VisualShowcaseRepository, storageSvc storage.StorageService) *VisualShowcaseService {
	return &VisualShowcaseService{repo: repo, storage: storageSvc}
}

func (s *VisualShowcaseService) GetPublishedItems(showcaseKey, locale string) ([]visualshowcase.Item, error) {
	result, err := s.GetPublishedResult(showcaseKey, locale)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func (s *VisualShowcaseService) GetPublishedResult(showcaseKey, locale string) (*VisualShowcasePublishedResult, error) {
	key := strings.TrimSpace(showcaseKey)
	normalizedLocale := strings.TrimSpace(locale)
	if key == "" {
		return nil, ErrVisualShowcaseKeyRequired
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
		return &VisualShowcasePublishedResult{
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
	return &VisualShowcasePublishedResult{
		Items:           fallbackItems,
		Locale:          "en",
		RequestedLocale: normalizedLocale,
		Fallback:        true,
		ConfiguredCount: fallbackConfiguredCount,
	}, nil
}

func (s *VisualShowcaseService) GetAdminItems(showcaseKey, locale string) ([]visualshowcase.Item, error) {
	key := strings.TrimSpace(showcaseKey)
	normalizedLocale := strings.TrimSpace(locale)
	if key == "" {
		return nil, ErrVisualShowcaseKeyRequired
	}
	if normalizedLocale == "" {
		return nil, ErrVisualShowcaseLocaleRequired
	}
	return s.repo.ListItems(key, normalizedLocale, false)
}

func (s *VisualShowcaseService) UploadAdminImage(
	ctx context.Context,
	showcaseKey string,
	locale string,
	file *multipart.FileHeader,
) (*VisualShowcaseImageUpload, error) {
	key := strings.TrimSpace(showcaseKey)
	normalizedLocale := strings.TrimSpace(locale)
	if key == "" {
		return nil, ErrVisualShowcaseKeyRequired
	}
	if normalizedLocale == "" {
		return nil, ErrVisualShowcaseLocaleRequired
	}
	if s == nil || s.storage == nil {
		return nil, ErrVisualShowcaseStorageUnavailable
	}
	if file == nil {
		return nil, ErrVisualShowcaseUploadFileRequired
	}
	if err := upload.ValidateFile(file, upload.ProductImageRule); err != nil {
		return nil, err
	}

	width, height, err := upload.ReadImageDimensions(file)
	if err != nil {
		return nil, err
	}
	if !isValidVisualShowcaseAspectRatioForShowcase(key, width, height) {
		return nil, visualShowcaseAspectRatioError(key, width, height)
	}

	imageURL, err := s.uploadVisualShowcaseImage(ctx, key, normalizedLocale, file)
	if err != nil {
		return nil, err
	}
	storageKey := s.visualShowcaseStorageKeyFromReference(imageURL)
	if storageKey == "" {
		_ = s.storage.Delete(ctx, imageURL)
		return nil, ErrVisualShowcaseImageInvalid
	}

	return &VisualShowcaseImageUpload{
		ImageURL:     imageURL,
		ThumbnailURL: imageURL,
		StorageKey:   storageKey,
		Width:        width,
		Height:       height,
	}, nil
}

func (s *VisualShowcaseService) ReplaceAdminItems(
	ctx context.Context,
	showcaseKey string,
	locale string,
	inputs []VisualShowcaseItemInput,
) ([]visualshowcase.Item, error) {
	key := strings.TrimSpace(showcaseKey)
	normalizedLocale := strings.TrimSpace(locale)
	if key == "" {
		return nil, ErrVisualShowcaseKeyRequired
	}
	if normalizedLocale == "" {
		return nil, ErrVisualShowcaseLocaleRequired
	}
	if len(inputs) > maxVisualShowcaseItems {
		return nil, ErrVisualShowcaseItemLimit
	}

	previousItems, err := s.repo.ListItems(key, normalizedLocale, false)
	if err != nil {
		return nil, err
	}

	retainedStorageKeys := make(map[string]struct{}, len(inputs))
	items := make([]visualshowcase.Item, 0, len(inputs))
	for index, input := range inputs {
		title := strings.TrimSpace(input.Title)
		altText := strings.TrimSpace(input.AltText)
		imageURL := strings.TrimSpace(input.ImageURL)
		thumbnailURL := strings.TrimSpace(input.ThumbnailURL)
		if thumbnailURL == "" {
			thumbnailURL = imageURL
		}
		if title == "" {
			return nil, ErrVisualShowcaseTitleRequired
		}
		if altText == "" {
			return nil, ErrVisualShowcaseAltTextRequired
		}
		if imageURL == "" {
			return nil, ErrVisualShowcaseImageRequired
		}

		storageKey := s.visualShowcaseStorageKeyFromInput(input.StorageKey, imageURL)
		if storageKey == "" {
			return nil, ErrVisualShowcaseImageInvalid
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
		if !isValidVisualShowcaseAspectRatioForShowcase(key, width, height) {
			return nil, visualShowcaseAspectRatioError(key, width, height)
		}

		items = append(items, visualshowcase.Item{
			ShowcaseKey:     key,
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

func (s *VisualShowcaseService) uploadVisualShowcaseImage(ctx context.Context, showcaseKey string, locale string, file *multipart.FileHeader) (string, error) {
	prefix := visualShowcaseStoragePrefix(showcaseKey, locale)
	if cacheControlled, ok := s.storage.(storage.CacheControlledObjectUploader); ok {
		return cacheControlled.UploadWithPrefixAndCacheControl(ctx, file, prefix, visualShowcaseImageCacheControl)
	}
	return s.storage.UploadWithPrefix(ctx, file, prefix)
}

func (s *VisualShowcaseService) destroyUnreferencedVisualShowcaseImages(
	ctx context.Context,
	previousItems []visualshowcase.Item,
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
			return fmt.Errorf("%w: %v", ErrVisualShowcasePreviousDestroyFailed, err)
		}
	}
	return nil
}

func (s *VisualShowcaseService) visualShowcaseStorageKeyFromInput(inputStorageKey string, imageURL string) string {
	normalizedStorageKey, hasStorageKey := storage.NormalizeObjectKey(inputStorageKey)
	if !hasStorageKey || !IsVisualShowcaseStorageKey(normalizedStorageKey) {
		return s.visualShowcaseStorageKeyFromReference(imageURL)
	}

	referencedStorageKey := s.visualShowcaseStorageKeyFromReference(imageURL)
	if referencedStorageKey == "" || referencedStorageKey != normalizedStorageKey {
		return ""
	}
	return normalizedStorageKey
}

func (s *VisualShowcaseService) visualShowcaseStorageKeyFromReference(reference string) string {
	value := strings.TrimSpace(reference)
	if value == "" {
		return ""
	}
	if s != nil && s.storage != nil {
		if key, err := s.storage.ObjectKey(value); err == nil && IsVisualShowcaseStorageKey(key) {
			return key
		}
	}
	return ""
}

func visualShowcaseStoragePrefix(showcaseKey string, locale string) string {
	prefix, ok := storage.NormalizeObjectKey(path.Join(
		VisualShowcaseStoragePrefix,
		normalizeVisualShowcasePathSegment(showcaseKey, "showcase"),
		normalizeVisualShowcasePathSegment(locale, "global"),
	))
	if !ok {
		return VisualShowcaseStoragePrefix
	}
	return prefix
}

func normalizeVisualShowcasePathSegment(value string, fallback string) string {
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

func IsVisualShowcaseStorageKey(key string) bool {
	normalizedKey, ok := storage.NormalizeObjectKey(key)
	return ok && (normalizedKey == VisualShowcaseStoragePrefix || strings.HasPrefix(normalizedKey, VisualShowcaseStoragePrefix+"/"))
}

func visualShowcaseAspectRatioError(showcaseKey string, width, height int) error {
	return fmt.Errorf(
		"%w: expected %s, received %dx%d",
		ErrVisualShowcaseAspectRatioInvalid,
		visualShowcaseAspectRatioLabel(showcaseKey),
		width,
		height,
	)
}

func visualShowcaseAspectRatioLabel(showcaseKey string) string {
	if strings.TrimSpace(showcaseKey) == HomeMainProductCategoriesShowcaseKey {
		return "16:9"
	}
	return "3:4"
}

func isValidVisualShowcaseAspectRatioForShowcase(showcaseKey string, width, height int) bool {
	if strings.TrimSpace(showcaseKey) == HomeMainProductCategoriesShowcaseKey {
		return isValidVisualShowcaseAspectRatio(width, height, 16, 9)
	}
	return isValidVisualShowcaseAspectRatio(width, height, 3, 4)
}

func isValidVisualShowcaseAspectRatio(width, height, ratioWidth, ratioHeight int) bool {
	return width > 0 && height > 0 && ratioWidth > 0 && ratioHeight > 0 && width*ratioHeight == height*ratioWidth
}
