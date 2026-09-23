package workbenchfeed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"strings"
	"time"
	"unicode/utf8"

	"commerce-platform/internal/pkg/upload"
	mediaservice "commerce-platform/internal/service"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrInvalidEntry        = errors.New("workbench feed entry is invalid")
	ErrEntryNotFound       = errors.New("workbench feed entry not found")
	ErrMediaLimit          = errors.New("workbench feed media limit exceeded")
	ErrProductLimit        = errors.New("workbench feed tagged product limit exceeded")
	ErrProductNotFound     = errors.New("workbench feed product not found")
	ErrInvalidStatus       = errors.New("workbench feed status is invalid")
	ErrInvalidDirectAction = errors.New("workbench feed direct action must be detail_drawer")
)

type Service struct {
	repo    *Repository
	catalog *Catalog
	media   *mediaservice.MediaService
	db      *gorm.DB
}

func NewService(db *gorm.DB, repo *Repository, catalog *Catalog, media *mediaservice.MediaService) *Service {
	return &Service{db: db, repo: repo, catalog: catalog, media: media}
}

func (s *Service) List(input ListInput) (*ListResult, error) {
	entries, total, err := s.repo.List(input)
	if err != nil {
		return nil, err
	}
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return &ListResult{
		Entries: entries,
		Pagination: Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: (int(total) + pageSize - 1) / pageSize,
		},
	}, nil
}

func (s *Service) Get(id uint) (*FeedEntry, error) {
	entry, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEntryNotFound
	}
	return entry, err
}

func (s *Service) Create(input EntryInput) (*FeedEntry, error) {
	if err := s.validateInput(input); err != nil {
		return nil, err
	}
	resolvedProducts, err := s.resolveTaggedProducts(input.TaggedProducts)
	if err != nil {
		return nil, err
	}
	var result *FeedEntry
	err = s.db.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		entryNumber, err := repo.NextEntryNumber()
		if err != nil {
			return err
		}
		entry := &FeedEntry{
			EntryNumber: entryNumber,
			PublishedAt: normalizedPublishedAt(input),
			Locale:      normalizeLocale(input.Locale),
			Content:     strings.TrimSpace(input.Content),
			Tags:        mustJSON(normalizeTags(input.Tags)),
			Status:      normalizeStatus(input.Status),
			CreatedBy:   input.ActorID,
			UpdatedBy:   input.ActorID,
		}
		if err := repo.Create(entry); err != nil {
			return err
		}
		if err := repo.ReplaceChildren(entry.ID, mediaModels(input.Media), taggedProductModels(resolvedProducts)); err != nil {
			return err
		}
		result, err = repo.FindByID(entry.ID)
		return err
	})
	return result, err
}

func (s *Service) Update(id uint, input EntryInput) (*FeedEntry, error) {
	if err := s.validateInput(input); err != nil {
		return nil, err
	}
	resolvedProducts, err := s.resolveTaggedProducts(input.TaggedProducts)
	if err != nil {
		return nil, err
	}
	var result *FeedEntry
	err = s.db.Transaction(func(tx *gorm.DB) error {
		repo := s.repo.WithTx(tx)
		entry, err := repo.FindByID(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrEntryNotFound
		}
		if err != nil {
			return err
		}
		entry.PublishedAt = normalizedPublishedAt(input)
		entry.Locale = normalizeLocale(input.Locale)
		entry.Content = strings.TrimSpace(input.Content)
		entry.Tags = mustJSON(normalizeTags(input.Tags))
		entry.Status = normalizeStatus(input.Status)
		entry.UpdatedBy = input.ActorID
		if err := repo.Update(entry); err != nil {
			return err
		}
		if err := repo.ReplaceChildren(entry.ID, mediaModels(input.Media), taggedProductModels(resolvedProducts)); err != nil {
			return err
		}
		result, err = repo.FindByID(entry.ID)
		return err
	})
	return result, err
}

func (s *Service) UpdateStatus(id uint, status string, actorID uint) (*FeedEntry, error) {
	normalized := normalizeStatus(status)
	if !isValidStatus(normalized) {
		return nil, ErrInvalidStatus
	}
	entry, err := s.Get(id)
	if err != nil {
		return nil, err
	}
	entry.Status = normalized
	entry.UpdatedBy = actorID
	if normalized == StatusPublished && entry.PublishedAt.IsZero() {
		entry.PublishedAt = time.Now()
	}
	if err := s.repo.Update(entry); err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *Service) Delete(id uint) error {
	if _, err := s.Get(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *Service) ListProductOptions(input CatalogQuery) ([]ProductCatalogOption, int64, error) {
	return s.catalog.ListOptions(input)
}

func (s *Service) UploadMedia(ctx context.Context, uploaderID uint, file *multipart.FileHeader, caption string) (*FeedMedia, error) {
	if s.media == nil {
		return nil, mediaservice.ErrMediaStorageUnavailable
	}
	if file == nil {
		return nil, mediaservice.ErrMediaUploadFileRequired
	}
	width, height, err := upload.ReadImageDimensions(file)
	if err != nil {
		return nil, err
	}
	asset, err := s.media.UploadAsset(ctx, mediaservice.MediaUploadInput{
		File:          file,
		MediaType:     "image",
		Caption:       strings.TrimSpace(caption),
		UploaderID:    uploaderID,
		Width:         width,
		Height:        height,
		StoragePrefix: "workbench-feed",
	})
	if err != nil {
		return nil, err
	}
	sizeKB := int((file.Size + 1023) / 1024)
	return &FeedMedia{
		FilePath:   asset.URL,
		Width:      width,
		Height:     height,
		FileSizeKB: sizeKB,
		Caption:    strings.TrimSpace(caption),
	}, nil
}

func (s *Service) validateInput(input EntryInput) error {
	if s == nil || s.repo == nil || s.catalog == nil || s.db == nil {
		return fmt.Errorf("%w: service is unavailable", ErrInvalidEntry)
	}
	content := strings.TrimSpace(input.Content)
	if content == "" || utf8.RuneCountInString(content) > MaxContentRunes {
		return fmt.Errorf("%w: content must contain 1-%d characters", ErrInvalidEntry, MaxContentRunes)
	}
	if !isValidStatus(normalizeStatus(input.Status)) {
		return ErrInvalidStatus
	}
	tags := normalizeTags(input.Tags)
	if len(tags) > MaxTags {
		return fmt.Errorf("%w: too many tags", ErrInvalidEntry)
	}
	if len(input.Media) > MaxMedia {
		return ErrMediaLimit
	}
	if len(input.TaggedProducts) > MaxTaggedProducts {
		return ErrProductLimit
	}
	for _, item := range input.Media {
		if strings.TrimSpace(item.FilePath) == "" || item.Width <= 0 || item.Height <= 0 || item.FileSizeKB < 0 {
			return fmt.Errorf("%w: media metadata is invalid", ErrInvalidEntry)
		}
	}
	for _, item := range input.TaggedProducts {
		if item.ProductID == 0 {
			return fmt.Errorf("%w: tagged product is invalid", ErrInvalidEntry)
		}
		if normalizeDirectAction(item.DirectAction) != DirectActionDetailDrawer {
			return ErrInvalidDirectAction
		}
		if _, err := s.catalog.Resolve(item.ProductID, item.VariantID); errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		} else if err != nil {
			return err
		}
	}
	return nil
}

func mediaModels(items []MediaInput) []FeedMedia {
	result := make([]FeedMedia, 0, len(items))
	for index, item := range items {
		sortOrder := item.SortOrder
		if sortOrder < 0 {
			sortOrder = index
		}
		result = append(result, FeedMedia{
			FilePath:   strings.TrimSpace(item.FilePath),
			Width:      item.Width,
			Height:     item.Height,
			FileSizeKB: item.FileSizeKB,
			Caption:    strings.TrimSpace(item.Caption),
			SortOrder:  sortOrder,
		})
	}
	return result
}

func taggedProductModels(items []resolvedTaggedProduct) []TaggedProduct {
	result := make([]TaggedProduct, 0, len(items))
	for index, item := range items {
		sortOrder := item.Input.SortOrder
		if sortOrder < 0 {
			sortOrder = index
		}
		result = append(result, TaggedProduct{
			ProductID:    item.Option.ProductID,
			VariantID:    item.Option.VariantID,
			ProductSlug:  strings.TrimSpace(item.Option.ProductSlug),
			DisplayTitle: resolvedProductDisplayTitle(item.Option),
			PriceMinor:   item.Option.PriceMinor,
			Currency:     normalizeCurrency(item.Option.Currency),
			DirectAction: DirectActionDetailDrawer,
			Available:    item.Option.Available,
			SortOrder:    sortOrder,
		})
	}
	return result
}

type resolvedTaggedProduct struct {
	Input  TaggedProductInput
	Option ProductCatalogOption
}

func (s *Service) resolveTaggedProducts(items []TaggedProductInput) ([]resolvedTaggedProduct, error) {
	result := make([]resolvedTaggedProduct, 0, len(items))
	for _, item := range items {
		option, err := s.catalog.Resolve(item.ProductID, item.VariantID)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		if err != nil {
			return nil, err
		}
		result = append(result, resolvedTaggedProduct{Input: item, Option: *option})
	}
	return result, nil
}

func resolvedProductDisplayTitle(option ProductCatalogOption) string {
	productName := strings.TrimSpace(option.ProductName)
	variantTitle := strings.TrimSpace(option.VariantTitle)
	if variantTitle == "" || strings.EqualFold(variantTitle, productName) {
		return productName
	}
	return productName + " · " + variantTitle
}

func normalizeLocale(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return DefaultLocale
	}
	return value
}

func normalizeStatus(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return StatusDraft
	}
	return value
}

func isValidStatus(value string) bool {
	return value == StatusDraft || value == StatusPublished || value == StatusArchived
}

func normalizeDirectAction(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return DirectActionDetailDrawer
	}
	return value
}

func normalizeCurrency(value string) string {
	value = strings.TrimSpace(strings.ToUpper(value))
	if len(value) != 3 {
		return "USD"
	}
	return value
}

func normalizeTags(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if !strings.HasPrefix(normalized, "#") {
			normalized = "#" + normalized
		}
		key := strings.ToLower(normalized)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func mustJSON(value []string) datatypes.JSON {
	payload, err := json.Marshal(value)
	if err != nil {
		return datatypes.JSON([]byte("[]"))
	}
	return datatypes.JSON(payload)
}

func normalizedPublishedAt(input EntryInput) time.Time {
	if input.PublishedAt != nil && !input.PublishedAt.IsZero() {
		return input.PublishedAt.UTC()
	}
	return time.Now().UTC()
}
