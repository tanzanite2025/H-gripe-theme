package repository

import (
	"errors"
	"fmt"

	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProductCategoryRepository struct {
	db                             *gorm.DB
	businessMediaReferenceResolver PublicBusinessMediaReferenceObjectKeyResolver
}

func (r *ProductCategoryRepository) ConfigurePublicBusinessMediaReferenceIdentityResolver(resolver PublicBusinessMediaReferenceObjectKeyResolver) {
	r.businessMediaReferenceResolver = resolver
}

func NewProductCategoryRepository(db *gorm.DB) *ProductCategoryRepository {
	return &ProductCategoryRepository{db: db}
}

func (r *ProductCategoryRepository) List(includeDisabled bool) ([]product.ProductCategory, error) {
	var categories []product.ProductCategory
	query := r.db.Model(&product.ProductCategory{}).Preload("Translations", func(db *gorm.DB) *gorm.DB {
		return db.Order("locale ASC, id ASC")
	})
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}
	if err := query.Order("depth ASC").Order("sort_order ASC").Order("name ASC").Order("id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

// FindPublished returns enabled categories in route-construction order.  The
// storefront exposes the same category tree for every locale (translations
// affect labels/SEO metadata, not the slug path), so sitemap generation can
// safely reuse this complete enabled set.
func (r *ProductCategoryRepository) FindPublished() ([]product.ProductCategory, error) {
	var categories []product.ProductCategory
	if err := r.db.Model(&product.ProductCategory{}).
		Where("is_enabled = ?", true).
		Order("depth ASC").Order("sort_order ASC").Order("name ASC").Order("id ASC").
		Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *ProductCategoryRepository) ListWithTranslations(includeDisabled bool, locales []string) ([]product.ProductCategory, error) {
	var categories []product.ProductCategory
	query := r.db.Model(&product.ProductCategory{}).
		Preload("Translations", func(db *gorm.DB) *gorm.DB {
			if len(locales) == 0 {
				return db
			}
			return db.Where("locale IN ?", locales)
		})
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}
	if err := query.Order("depth ASC").Order("sort_order ASC").Order("name ASC").Order("id ASC").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (r *ProductCategoryRepository) FindByID(id uint) (*product.ProductCategory, error) {
	var category product.ProductCategory
	if err := r.db.First(&category, id).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *ProductCategoryRepository) FindBySlug(slug string, includeDisabled bool) (*product.ProductCategory, error) {
	var category product.ProductCategory
	query := r.db.Model(&product.ProductCategory{}).
		Preload("Translations", func(db *gorm.DB) *gorm.DB {
			return db.Order("locale ASC, id ASC")
		}).
		Where("slug = ?", slug)
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}
	if err := query.First(&category).Error; err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *ProductCategoryRepository) ExistsBySlug(slug string, excludeID uint) (bool, error) {
	query := r.db.Model(&product.ProductCategory{}).Where("slug = ?", slug)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProductCategoryRepository) CountChildren(parentID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&product.ProductCategory{}).Where("parent_id = ?", parentID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProductCategoryRepository) CountProducts(categoryID uint) (int64, error) {
	var count int64
	if err := r.db.Model(&product.Product{}).Where("product_category_id = ?", categoryID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProductCategoryRepository) CountQuickBuyReferences(categoryID uint) (int64, error) {
	if !r.db.Migrator().HasTable("quick_buy_step_product_categories") {
		return 0, nil
	}
	var count int64
	if err := r.db.Table("quick_buy_step_product_categories").
		Where("product_category_id = ?", categoryID).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProductCategoryRepository) ListProductIDs(categoryIDs []uint) ([]uint, error) {
	if len(categoryIDs) == 0 {
		return []uint{}, nil
	}
	var ids []uint
	if err := r.db.Model(&product.Product{}).
		Where("product_category_id IN ?", categoryIDs).
		Pluck("id", &ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}

func (r *ProductCategoryRepository) Create(category *product.ProductCategory) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var assetIDs []uint
		if category.ImageMediaAssetID != nil {
			assetIDs = append(assetIDs, *category.ImageMediaAssetID)
		}
		fields := []templateOrCategoryRichTextMediaFieldMutation{{Requested: &category.SEOIntro}}
		for index := range category.Translations {
			fields = append(fields, templateOrCategoryRichTextMediaFieldMutation{Requested: &category.Translations[index].SEOIntro})
		}
		resolved, err := lockValidateAndCanonicalizeTemplateOrCategoryRichTextFieldsWithOneAssetLockPlan(tx, r.businessMediaReferenceResolver, fields, assetIDs)
		if errors.Is(err, ErrProductMediaReferenceInvalid) {
			return fmt.Errorf("%w: %v", ErrProductCategoryImageReferenceInvalid, err)
		}
		if err != nil {
			return err
		}
		if err := validateAndResolveCategoryImageFromLockedRegisteredAsset(tx, category, resolved); err != nil {
			return err
		}
		return tx.Select("*").Create(category).Error
	})
}

func (r *ProductCategoryRepository) Update(category *product.ProductCategory, descendantDepths map[uint]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var current product.ProductCategory
		if err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(&current, category.ID).Error; err != nil {
			return err
		}
		if err := r.lockValidateAndResolveCategoryImageFromRegisteredAsset(tx, category, current); err != nil {
			return err
		}
		if err := tx.Model(&product.ProductCategory{}).Where("id = ?", category.ID).Updates(map[string]interface{}{
			"parent_id":            category.ParentID,
			"name":                 category.Name,
			"slug":                 category.Slug,
			"description":          category.Description,
			"image_media_asset_id": category.ImageMediaAssetID,
			"image_url":            category.ImageURL,
			"depth":                category.Depth,
			"sort_order":           category.SortOrder,
			"is_enabled":           category.IsEnabled,
		}).Error; err != nil {
			return err
		}

		for id, depth := range descendantDepths {
			if id == category.ID {
				continue
			}
			if err := tx.Model(&product.ProductCategory{}).Where("id = ?", id).Update("depth", depth).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProductCategoryRepository) UpdateSEO(categoryID uint, locale string, metaTitle, metaDescription, intro *string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// The parent also serializes first writes for a previously absent locale
		// with name/description replacement and category deletion.
		var category product.ProductCategory
		if err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(&category, categoryID).Error; err != nil {
			return err
		}
		translation := product.ProductCategoryTranslation{ProductCategoryID: categoryID, Locale: locale}
		currentTitle, currentDescription, currentIntro := category.MetaTitle, category.MetaDesc, category.SEOIntro
		if locale != "en" {
			err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).Where("product_category_id = ? AND locale = ?", categoryID, locale).First(&translation).Error
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			currentTitle, currentDescription, currentIntro = translation.MetaTitle, translation.MetaDesc, translation.SEOIntro
		}
		if metaTitle != nil {
			currentTitle = *metaTitle
		}
		if metaDescription != nil {
			currentDescription = *metaDescription
		}
		nextIntro := currentIntro
		if intro != nil {
			nextIntro = *intro
		}
		canonicalIntro, _, err := lockValidateAndCanonicalizeTemplateOrCategoryRichTextMediaWithinTransaction(tx, r.businessMediaReferenceResolver, nextIntro, currentIntro, nil)
		if err != nil {
			return err
		}
		updates := map[string]interface{}{"meta_title": currentTitle, "meta_description": currentDescription, "seo_intro": canonicalIntro}
		if locale == "en" {
			return tx.Model(&category).Updates(updates).Error
		}
		if translation.ID == 0 {
			translation.MetaTitle, translation.MetaDesc, translation.SEOIntro = currentTitle, currentDescription, canonicalIntro
			return tx.Omit(clause.Associations).Create(&translation).Error
		}
		return tx.Model(&translation).Updates(updates).Error
	})
}

func (r *ProductCategoryRepository) Delete(id uint) error {
	return r.db.Delete(&product.ProductCategory{}, id).Error
}

func (r *ProductCategoryRepository) ListTranslations(categoryID uint) ([]product.ProductCategoryTranslation, error) {
	var translations []product.ProductCategoryTranslation
	err := r.db.
		Where("product_category_id = ?", categoryID).
		Order("locale ASC, id ASC").
		Find(&translations).Error
	return translations, err
}

func (r *ProductCategoryRepository) ReplaceTranslationNamesAndDescriptionsPreservingLockedSEO(categoryID uint, translations []product.ProductCategoryTranslation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var category product.ProductCategory
		if err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(&category, categoryID).Error; err != nil {
			return err
		}
		retainedLocales := make([]string, 0, len(translations))
		for _, requested := range translations {
			retainedLocales = append(retainedLocales, requested.Locale)
			var current product.ProductCategoryTranslation
			err := tx.Where("product_category_id = ? AND locale = ?", categoryID, requested.Locale).First(&current).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				created := product.ProductCategoryTranslation{ProductCategoryID: categoryID, Locale: requested.Locale, Name: requested.Name, Description: requested.Description}
				if err := tx.Omit(clause.Associations).Create(&created).Error; err != nil {
					return err
				}
			} else if err != nil {
				return err
			} else if err := tx.Model(&current).Updates(map[string]interface{}{"name": requested.Name, "description": requested.Description}).Error; err != nil {
				return err
			}
		}
		removed := tx.Where("product_category_id = ?", categoryID)
		if len(retainedLocales) > 0 {
			removed = removed.Where("locale NOT IN ?", retainedLocales)
		}
		return removed.Delete(&product.ProductCategoryTranslation{}).Error
	})
}
