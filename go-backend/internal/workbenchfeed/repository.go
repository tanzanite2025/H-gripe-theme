package workbenchfeed

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/repository"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db                             *gorm.DB
	businessMediaReferenceResolver repository.PublicBusinessMediaReferenceObjectKeyResolver
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx, businessMediaReferenceResolver: r.businessMediaReferenceResolver}
}

func (r *Repository) List(input ListInput) ([]FeedEntry, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, errors.New("workbench feed repository is unavailable")
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

	query := r.db.Model(&FeedEntry{})
	if !input.Admin {
		query = query.Where("status = ?", StatusPublished)
	}
	if locale := strings.TrimSpace(input.Locale); locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if status := strings.TrimSpace(input.Status); status != "" {
		query = query.Where("status = ?", status)
	}
	if search := strings.TrimSpace(input.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(content) LIKE ? OR LOWER(entry_number) LIKE ?", like, like)
	}
	if tag := strings.TrimSpace(input.Tag); tag != "" {
		query = query.Where("EXISTS (SELECT 1 FROM jsonb_array_elements_text(tags) AS feed_tag WHERE LOWER(feed_tag) = LOWER(?))", tag)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var entries []FeedEntry
	if err := query.
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("TaggedProducts", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Order("published_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&entries).Error; err != nil {
		return nil, 0, err
	}
	if entries == nil {
		entries = []FeedEntry{}
	}
	return entries, total, nil
}

func (r *Repository) FindByID(id uint) (*FeedEntry, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("workbench feed repository is unavailable")
	}
	var entry FeedEntry
	err := r.db.
		Preload("Media", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		Preload("TaggedProducts", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC, id ASC")
		}).
		First(&entry, id).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *Repository) NextEntryNumber() (string, error) {
	var latest FeedEntry
	err := r.db.Unscoped().
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Order("id DESC").
		First(&latest).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "LOG-001", nil
	}
	if err != nil {
		return "", err
	}

	raw := strings.TrimSpace(latest.EntryNumber)
	if len(raw) > 4 && strings.EqualFold(raw[:4], "LOG-") {
		if number, parseErr := strconv.Atoi(raw[4:]); parseErr == nil && number >= 0 {
			return fmt.Sprintf("LOG-%03d", number+1), nil
		}
	}
	return fmt.Sprintf("LOG-%03d", latest.ID+1), nil
}

func (r *Repository) createEntryFieldsWithoutAssociations(entry *FeedEntry) error {
	return r.db.Omit(clause.Associations).Create(entry).Error
}

// Service owns the transaction and locks the active parent before calling this.
func (r *Repository) replaceChildrenWithValidatedMediaWithinTransaction(entryID uint, media []FeedMedia, products []TaggedProduct) error {
	var previous []FeedMedia
	if err := r.db.Where("entry_id = ?", entryID).Find(&previous).Error; err != nil {
		return err
	}
	paths, previousPaths := make([]string, len(media)), make([]string, len(previous))
	for index := range media {
		paths[index] = media[index].FilePath
	}
	for index := range previous {
		previousPaths[index] = previous[index].FilePath
	}
	if err := repository.ValidateAndCanonicalizeWorkbenchFeedImagePathsWithinTransaction(r.db, r.businessMediaReferenceResolver, paths, previousPaths); err != nil {
		if errors.Is(err, repository.ErrProductMediaReferenceInvalid) {
			return fmt.Errorf("%w: %v", ErrMediaReferenceInvalid, err)
		}
		return err
	}
	for index := range media {
		media[index].FilePath = paths[index]
	}
	if err := r.db.Where("entry_id = ?", entryID).Delete(&FeedMedia{}).Error; err != nil {
		return err
	}
	if err := r.db.Where("entry_id = ?", entryID).Delete(&TaggedProduct{}).Error; err != nil {
		return err
	}
	for index := range media {
		media[index].ID = 0
		media[index].EntryID = entryID
	}
	for index := range products {
		products[index].ID = 0
		products[index].EntryID = entryID
	}
	if len(media) > 0 {
		if err := r.db.Create(&media).Error; err != nil {
			return err
		}
	}
	if len(products) > 0 {
		if err := r.db.Create(&products).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) updateEntryFieldsWithoutAssociations(entry *FeedEntry) error {
	return r.db.Model(&FeedEntry{}).Where("id = ?", entry.ID).
		Select("*").Omit(clause.Associations, "id", "created_at", "created_by", "deleted_at").Updates(entry).Error
}

func (r *Repository) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if _, err := r.WithTx(tx).findActiveEntryWithMediaForUpdate(id); err != nil {
			return err
		}
		if err := tx.Where("id = ?", id).Delete(&FeedEntry{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) TouchPublishedAt(entry *FeedEntry) {
	if entry != nil && entry.PublishedAt.IsZero() {
		entry.PublishedAt = time.Now()
	}
}
