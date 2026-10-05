package service

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path"
	"strings"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

var (
	ErrHomeVisualTileKeyRequired            = errors.New("visual showcase key is required")
	ErrHomeVisualTileLocaleRequired         = errors.New("visual showcase locale is required")
	ErrHomeVisualTileItemLimit              = errors.New("visual showcase item limit exceeded")
	ErrHomeVisualTileTitleRequired          = errors.New("visual showcase item title is required")
	ErrHomeVisualTileAltTextRequired        = errors.New("visual showcase item alt text is required")
	ErrHomeVisualTileImageRequired          = errors.New("visual showcase item image is required")
	ErrHomeVisualTileImageInvalid           = errors.New("visual showcase item image is not a visual showcase upload")
	ErrHomeVisualTileStorageUnavailable     = errors.New("visual showcase storage is unavailable")
	ErrHomeVisualTileUploadFileRequired     = errors.New("visual showcase upload file is required")
	ErrHomeVisualTileItemCountInvalid       = errors.New("visual showcase item count is invalid")
	ErrHomeVisualTileImageDimensionsInvalid = errors.New("visual showcase image dimensions are invalid")
	ErrHomeVisualTileAspectRatioInvalid     = errors.New("visual showcase image aspect ratio is invalid")
)

const (
	HomeHeroVisualShowcaseTileSetKey       = "home-hero"
	HomeHeroVisualShowcaseMaximumItemCount = 9
	HomeHeroVisualShowcaseImageDimension   = 600
	HomeMainProductCategoriesTileSetKey    = "home-main-product-categories"
	maxHomeVisualTileItems                 = 100
	HomeVisualTileStoragePrefix            = "visual-showcase"
	homeVisualTileImageCacheControl        = "public, max-age=31536000, immutable"
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
	outbox  *repository.OutboxRepository
}

func NewHomeVisualTileService(repo *repository.HomeVisualTileRepository, storageSvc storage.StorageService) *HomeVisualTileService {
	return &HomeVisualTileService{repo: repo, storage: storageSvc}
}

func (s *HomeVisualTileService) ConfigureObjectCleanupOutbox(repo *repository.OutboxRepository) {
	if s == nil {
		return
	}
	s.outbox = repo
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

	publishedOnly := key != HomeHeroVisualShowcaseTileSetKey
	items, err := s.repo.ListItems(key, normalizedLocale, publishedOnly)
	if err != nil {
		return nil, err
	}
	configuredCount, err := s.repo.CountItems(key, normalizedLocale, false)
	if err != nil {
		return nil, err
	}
	if shouldUseRequestedHomeVisualTileLocale(
		key,
		normalizedLocale,
		configuredCount,
		len(items),
	) {
		return &HomeVisualTilePublishedResult{
			Items:           items,
			Locale:          normalizedLocale,
			RequestedLocale: normalizedLocale,
			Fallback:        false,
			ConfiguredCount: configuredCount,
		}, nil
	}

	fallbackItems, err := s.repo.ListItems(key, "en", key != HomeHeroVisualShowcaseTileSetKey)
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

func shouldUseRequestedHomeVisualTileLocale(
	tileSetKey string,
	locale string,
	configuredCount int64,
	publishedCount int,
) bool {
	if strings.TrimSpace(tileSetKey) == HomeHeroVisualShowcaseTileSetKey {
		// The hero is intentionally locale-local. A missing or partial translation
		// must stay empty instead of borrowing another locale's images.
		return true
	}
	if locale == "en" {
		return true
	}
	return configuredCount > 0
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
	switch key {
	case HomeHeroVisualShowcaseTileSetKey:
		specCode = upload.SpecVisualShowcaseHomeHero
	case HomeMainProductCategoriesTileSetKey:
		specCode = upload.SpecVisualShowcaseHomeCategories
	}
	if err := upload.ValidateSpecFile(file, string(specCode)); err != nil {
		return nil, err
	}

	width, height, err := upload.ReadImageDimensions(file)
	if err != nil {
		return nil, err
	}
	if err := validateHomeVisualTileDimensionsForTileSet(key, width, height); err != nil {
		return nil, err
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
	if err := validateHomeVisualTileItemCount(key, len(inputs)); err != nil {
		return nil, err
	}

	retainedStorageKeys := make(map[string]struct{}, len(inputs))
	usedHomeHeroDesktopOrders := make(map[int]struct{}, len(inputs))
	items := make([]homevisualtile.Tile, 0, len(inputs))
	for index, input := range inputs {
		item, err := s.buildHomeVisualTile(key, normalizedLocale, index, input, usedHomeHeroDesktopOrders)
		if err != nil {
			return nil, err
		}
		retainedStorageKeys[item.StorageKey] = struct{}{}
		items = append(items, item)
	}

	if err := s.repo.ReplaceItems(key, normalizedLocale, items, func(tx *gorm.DB, previous []homevisualtile.Tile) error {
		removedKeys := homeVisualTileRemovedStorageKeys(s, previous, retainedStorageKeys)
		if len(removedKeys) == 0 {
			return nil
		}
		if s.outbox == nil {
			return ErrObjectStorageCleanupUnavailable
		}
		aggregateID := key + ":" + normalizedLocale
		event, eventErr := newObjectStorageCleanupEvent(
			objectCleanupResourceHomeVisualTile,
			aggregateID,
			outbox.AggregateTypeHomeVisualTileSet,
			aggregateID,
			removedKeys,
		)
		if eventErr != nil {
			return eventErr
		}
		return s.outbox.WithTx(tx).CreateEvent(event)
	}); err != nil {
		return nil, err
	}
	return s.repo.ListItems(key, normalizedLocale, false)
}

// SaveSingleVisualShowcaseAdminItem persists one showcase slot while leaving every other slot untouched.
func (s *HomeVisualTileService) SaveSingleVisualShowcaseAdminItem(
	ctx context.Context,
	tileSetKey string,
	locale string,
	input HomeVisualTileInput,
) ([]homevisualtile.Tile, error) {
	key := strings.TrimSpace(tileSetKey)
	normalizedLocale := strings.TrimSpace(locale)
	if key == "" {
		return nil, ErrHomeVisualTileKeyRequired
	}
	if normalizedLocale == "" {
		return nil, ErrHomeVisualTileLocaleRequired
	}
	if input.DesktopOrder <= 0 {
		return nil, fmt.Errorf(
			"%w: desktop position must be positive",
			ErrHomeVisualTileItemCountInvalid,
		)
	}

	item, err := s.buildHomeVisualTile(
		key,
		normalizedLocale,
		input.DesktopOrder-1,
		input,
		make(map[int]struct{}),
	)
	if err != nil {
		return nil, err
	}

	retainedStorageKeys := map[string]struct{}{item.StorageKey: {}}
	previousStorageKey := ""
	err = s.repo.SaveSingleVisualShowcaseItem(&item, func(tx *gorm.DB, previous *homevisualtile.Tile) error {
		if previous == nil {
			return nil
		}
		previousStorageKey = previous.StorageKey
		removedKeys := homeVisualTileRemovedStorageKeys(s, []homevisualtile.Tile{*previous}, retainedStorageKeys)
		if len(removedKeys) == 0 {
			return nil
		}
		if s.outbox == nil {
			return ErrObjectStorageCleanupUnavailable
		}
		aggregateID := key + ":" + normalizedLocale
		event, eventErr := newObjectStorageCleanupEvent(
			objectCleanupResourceHomeVisualTile,
			aggregateID,
			outbox.AggregateTypeHomeVisualTileSet,
			aggregateID,
			removedKeys,
		)
		if eventErr != nil {
			return eventErr
		}
		return s.outbox.WithTx(tx).CreateEvent(event)
	})
	if err != nil {
		if s.storage != nil && previousStorageKey != "" && previousStorageKey != item.StorageKey {
			_ = s.storage.Delete(ctx, item.ImageURL)
		}
		return nil, err
	}
	return s.repo.ListItems(key, normalizedLocale, false)
}

func (s *HomeVisualTileService) buildHomeVisualTile(
	tileSetKey string,
	locale string,
	index int,
	input HomeVisualTileInput,
	usedDesktopOrders map[int]struct{},
) (homevisualtile.Tile, error) {
	title := strings.TrimSpace(input.Title)
	altText := strings.TrimSpace(input.AltText)
	imageURL := strings.TrimSpace(input.ImageURL)
	thumbnailURL := strings.TrimSpace(input.ThumbnailURL)
	if thumbnailURL == "" {
		thumbnailURL = imageURL
	}
	if title == "" {
		return homevisualtile.Tile{}, ErrHomeVisualTileTitleRequired
	}
	if altText == "" {
		return homevisualtile.Tile{}, ErrHomeVisualTileAltTextRequired
	}
	if imageURL == "" {
		return homevisualtile.Tile{}, ErrHomeVisualTileImageRequired
	}

	storageKey := s.visualShowcaseStorageKeyFromInput(input.StorageKey, imageURL)
	if storageKey == "" {
		return homevisualtile.Tile{}, ErrHomeVisualTileImageInvalid
	}

	desktopOrder := input.DesktopOrder
	if desktopOrder <= 0 {
		desktopOrder = index + 1
	}
	if tileSetKey == HomeHeroVisualShowcaseTileSetKey {
		if desktopOrder > HomeHeroVisualShowcaseMaximumItemCount {
			return homevisualtile.Tile{}, fmt.Errorf(
				"%w: desktop position must be between 1 and %d",
				ErrHomeVisualTileItemCountInvalid,
				HomeHeroVisualShowcaseMaximumItemCount,
			)
		}
		if _, exists := usedDesktopOrders[desktopOrder]; exists {
			return homevisualtile.Tile{}, fmt.Errorf(
				"%w: duplicate desktop position %d",
				ErrHomeVisualTileItemCountInvalid,
				desktopOrder,
			)
		}
		usedDesktopOrders[desktopOrder] = struct{}{}
	}

	layoutVariant := strings.TrimSpace(input.LayoutVariant)
	if layoutVariant == "" {
		layoutVariant = "standard"
	}
	mobilePairIndex := input.MobilePairIndex
	if mobilePairIndex < 0 {
		mobilePairIndex = 0
	}
	width := input.Width
	height := input.Height
	if tileSetKey != HomeHeroVisualShowcaseTileSetKey {
		defaultWidth, defaultHeight := homeVisualTileDefaultDimensions(tileSetKey)
		if width <= 0 {
			width = defaultWidth
		}
		if height <= 0 {
			height = defaultHeight
		}
	}
	if err := validateHomeVisualTileDimensionsForTileSet(tileSetKey, width, height); err != nil {
		return homevisualtile.Tile{}, err
	}

	targetURL := strings.TrimSpace(input.TargetURL)
	targetLabel := strings.TrimSpace(input.TargetLabel)
	if tileSetKey == HomeHeroVisualShowcaseTileSetKey {
		targetURL = ""
		targetLabel = ""
	}
	isPublished := input.IsPublished
	if tileSetKey == HomeHeroVisualShowcaseTileSetKey {
		isPublished = true
	}

	return homevisualtile.Tile{
		TileSetKey:      tileSetKey,
		Locale:          locale,
		ImageURL:        imageURL,
		ThumbnailURL:    thumbnailURL,
		StorageKey:      storageKey,
		Title:           title,
		Caption:         strings.TrimSpace(input.Caption),
		AltText:         altText,
		DesktopOrder:    desktopOrder,
		MobilePairIndex: mobilePairIndex,
		TargetURL:       targetURL,
		TargetLabel:     targetLabel,
		LayoutVariant:   layoutVariant,
		IsPublished:     isPublished,
		Width:           width,
		Height:          height,
	}, nil
}

func homeVisualTileRemovedStorageKeys(
	s *HomeVisualTileService,
	previousItems []homevisualtile.Tile,
	retainedStorageKeys map[string]struct{},
) []string {
	removed := make([]string, 0, len(previousItems))
	for _, item := range previousItems {
		storageKey := s.visualShowcaseStorageKeyFromInput(item.StorageKey, item.ImageURL)
		if storageKey == "" {
			continue
		}
		if _, retained := retainedStorageKeys[storageKey]; retained {
			continue
		}
		removed = append(removed, storageKey)
	}
	return normalizeObjectCleanupKeys(removed)
}

func (s *HomeVisualTileService) uploadVisualShowcaseImage(ctx context.Context, tileSetKey string, locale string, file *multipart.FileHeader) (string, error) {
	prefix := homeVisualTileStoragePrefix(tileSetKey, locale)
	if cacheControlled, ok := s.storage.(storage.CacheControlledObjectUploader); ok {
		return cacheControlled.UploadWithPrefixAndCacheControl(ctx, file, prefix, homeVisualTileImageCacheControl)
	}
	return s.storage.UploadWithPrefix(ctx, file, prefix)
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

func validateHomeVisualTileItemCount(tileSetKey string, itemCount int) error {
	if itemCount > maxHomeVisualTileItems {
		return ErrHomeVisualTileItemLimit
	}
	if strings.TrimSpace(tileSetKey) == HomeHeroVisualShowcaseTileSetKey && itemCount > HomeHeroVisualShowcaseMaximumItemCount {
		return fmt.Errorf(
			"%w: expected no more than %d items, received %d",
			ErrHomeVisualTileItemCountInvalid,
			HomeHeroVisualShowcaseMaximumItemCount,
			itemCount,
		)
	}
	return nil
}

func validateHomeVisualTileDimensionsForTileSet(tileSetKey string, width, height int) error {
	switch strings.TrimSpace(tileSetKey) {
	case HomeHeroVisualShowcaseTileSetKey:
		if width != HomeHeroVisualShowcaseImageDimension || height != HomeHeroVisualShowcaseImageDimension {
			return fmt.Errorf(
				"%w: expected %dx%d, received %dx%d",
				ErrHomeVisualTileImageDimensionsInvalid,
				HomeHeroVisualShowcaseImageDimension,
				HomeHeroVisualShowcaseImageDimension,
				width,
				height,
			)
		}
	case HomeMainProductCategoriesTileSetKey:
		if !isValidHomeVisualTileAspectRatio(width, height, 16, 9) {
			return homeVisualTileAspectRatioError(tileSetKey, width, height)
		}
	default:
		if !isValidHomeVisualTileAspectRatio(width, height, 3, 4) {
			return homeVisualTileAspectRatioError(tileSetKey, width, height)
		}
	}
	return nil
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
	switch strings.TrimSpace(tileSetKey) {
	case HomeHeroVisualShowcaseTileSetKey:
		return "1:1"
	case HomeMainProductCategoriesTileSetKey:
		return "16:9"
	default:
		return "3:4"
	}
}

func homeVisualTileDefaultDimensions(tileSetKey string) (int, int) {
	switch strings.TrimSpace(tileSetKey) {
	case HomeHeroVisualShowcaseTileSetKey:
		return HomeHeroVisualShowcaseImageDimension, HomeHeroVisualShowcaseImageDimension
	case HomeMainProductCategoriesTileSetKey:
		return 1600, 900
	default:
		return 900, 1200
	}
}

func isValidHomeVisualTileAspectRatio(width, height, ratioWidth, ratioHeight int) bool {
	return width > 0 && height > 0 && ratioWidth > 0 && ratioHeight > 0 && width*ratioHeight == height*ratioWidth
}
