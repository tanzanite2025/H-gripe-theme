package repository

import (
	"commerce-platform/internal/domain/product"

	"gorm.io/gorm"
)

type ProductCategoryRepository struct {
	db *gorm.DB
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

func (r *ProductCategoryRepository) Create(category *product.ProductCategory) error {
	return r.db.Select("*").Create(category).Error
}

func (r *ProductCategoryRepository) Update(category *product.ProductCategory, descendantDepths map[uint]int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
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

func (r *ProductCategoryRepository) ReplaceTranslations(categoryID uint, translations []product.ProductCategoryTranslation) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_category_id = ?", categoryID).
			Delete(&product.ProductCategoryTranslation{}).Error; err != nil {
			return err
		}
		if len(translations) == 0 {
			return nil
		}
		return tx.Create(&translations).Error
	})
}
