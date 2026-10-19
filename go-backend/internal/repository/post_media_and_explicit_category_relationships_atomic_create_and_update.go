package repository

import (
	"errors"
	"strings"

	"commerce-platform/internal/domain/post"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrPostCategoryRelationshipInvalid = errors.New("post category relationship is invalid")

func (r *PostRepository) CreateWithCategoryIDs(item *post.Post, categoryIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		categories, err := r.validatePostCategoryIDsWithinPostSaveTransaction(tx, categoryIDs, item.Locale)
		if err != nil {
			return err
		}
		if err := r.validateAndCanonicalizePostContentAndFeaturedImageInOneAssetLockPlan(tx, item, &post.Post{}); err != nil {
			return err
		}
		if err := tx.Omit(clause.Associations).Create(item).Error; err != nil {
			return err
		}
		if err := r.replaceExplicitPostCategoryRelationshipsWithinPostSaveTransaction(tx, item.ID, categories); err != nil {
			return err
		}
		item.Categories = categories
		return nil
	})
}

func (r *PostRepository) UpdateWithCategoryIDs(item *post.Post, categoryIDs []uint, mutations ...PostMediaFieldMutation) error {
	return r.updatePostMediaAndOptionalCategoryRelationshipsInOneTransaction(item, &categoryIDs, mutations...)
}

func (r *PostRepository) updatePostMediaAndOptionalCategoryRelationshipsInOneTransaction(item *post.Post, categoryIDs *[]uint, mutations ...PostMediaFieldMutation) error {
	if item.ID == 0 {
		return gorm.ErrRecordNotFound
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		var previous post.Post
		if err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(&previous, item.ID).Error; err != nil {
			return err
		}
		if len(mutations) > 0 {
			if !mutations[0].UpdateContent {
				item.Content = previous.Content
			}
			if !mutations[0].UpdateFeaturedImage {
				item.FeaturedImg = previous.FeaturedImg
			}
		}
		var categories []post.Category
		if categoryIDs != nil {
			var err error
			categories, err = r.validatePostCategoryIDsWithinPostSaveTransaction(tx, *categoryIDs, item.Locale)
			if err != nil {
				return err
			}
		}
		if err := r.validateAndCanonicalizePostContentAndFeaturedImageInOneAssetLockPlan(tx, item, &previous); err != nil {
			return err
		}
		// Explicit category replacement is the only association write path.
		// Updates cannot upsert a concurrently deleted post.
		if err := tx.Model(&post.Post{}).Where("id = ?", item.ID).Select("*").Omit(clause.Associations, "id", "created_at", "deleted_at").Updates(item).Error; err != nil {
			return err
		}
		if categoryIDs != nil {
			if err := r.replaceExplicitPostCategoryRelationshipsWithinPostSaveTransaction(tx, item.ID, categories); err != nil {
				return err
			}
			item.Categories = categories
		}
		return nil
	})
}

func (r *PostRepository) validatePostCategoryIDsWithinPostSaveTransaction(tx *gorm.DB, categoryIDs []uint, locale string) ([]post.Category, error) {
	if len(categoryIDs) == 0 {
		return nil, nil
	}
	seen := make(map[uint]bool, len(categoryIDs))
	ids := make([]uint, 0, len(categoryIDs))
	for _, id := range categoryIDs {
		if id == 0 {
			return nil, ErrPostCategoryRelationshipInvalid
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	if locale == "" {
		locale = "en"
	} // Match Post.BeforeCreate's existing default.
	var categories []post.Category
	if err := tx.Where("id IN ? AND locale = ?", ids, strings.TrimSpace(locale)).Order("sort_order ASC, name ASC, id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	if len(categories) != len(ids) {
		return nil, ErrPostCategoryRelationshipInvalid
	}
	return categories, nil
}

func (r *PostRepository) replaceExplicitPostCategoryRelationshipsWithinPostSaveTransaction(tx *gorm.DB, postID uint, categories []post.Category) error {
	// Repositories without configured categories also support text-only legacy
	// schemas. Explicit nonempty selections always require the relationship table.
	if len(categories) == 0 && !r.categoriesEnabled && !tx.Migrator().HasTable(&post.PostCategory{}) {
		return nil
	}
	if err := tx.Where("post_id = ?", postID).Delete(&post.PostCategory{}).Error; err != nil {
		return err
	}
	if len(categories) == 0 {
		return nil
	}
	links := make([]post.PostCategory, 0, len(categories))
	for _, category := range categories {
		links = append(links, post.PostCategory{PostID: postID, CategoryID: category.ID})
	}
	return tx.Create(&links).Error
}
