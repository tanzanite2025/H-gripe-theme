package repository

import (
	"commerce-platform/internal/domain/product"
	spokedomain "commerce-platform/internal/domain/spoke"

	"gorm.io/gorm"
)

type ProductBrandRepository struct {
	db *gorm.DB
}

func NewProductBrandRepository(db *gorm.DB) *ProductBrandRepository {
	return &ProductBrandRepository{db: db}
}

func (r *ProductBrandRepository) WithTx(tx *gorm.DB) *ProductBrandRepository {
	return &ProductBrandRepository{db: tx}
}

func (r *ProductBrandRepository) List(includeDisabled bool) ([]product.ProductBrand, error) {
	var brands []product.ProductBrand
	query := r.db.Model(&product.ProductBrand{})
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}
	if err := query.Order("sort_order ASC").Order("name ASC").Order("id ASC").Find(&brands).Error; err != nil {
		return nil, err
	}
	return brands, nil
}

func (r *ProductBrandRepository) FindByID(id uint) (*product.ProductBrand, error) {
	var brand product.ProductBrand
	if err := r.db.First(&brand, id).Error; err != nil {
		return nil, err
	}
	return &brand, nil
}

func (r *ProductBrandRepository) ExistsBySlug(slug string, excludeID uint) (bool, error) {
	query := r.db.Model(&product.ProductBrand{}).Where("slug = ?", slug)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProductBrandRepository) CountProducts(id uint) (int64, error) {
	var count int64
	if err := r.db.Unscoped().Model(&product.Product{}).Where("brand_id = ?", id).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProductBrandRepository) CountSpokeRimBrands(id uint) (int64, error) {
	var count int64
	if err := r.db.Unscoped().Model(&spokedomain.CatalogRimBrand{}).Where("product_brand_id = ?", id).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProductBrandRepository) Create(brand *product.ProductBrand) error {
	return r.db.Select("*").Create(brand).Error
}

func (r *ProductBrandRepository) Update(brand *product.ProductBrand) error {
	return r.db.Model(&product.ProductBrand{}).Where("id = ?", brand.ID).Updates(map[string]interface{}{
		"name":        brand.Name,
		"slug":        brand.Slug,
		"description": brand.Description,
		"logo_url":    brand.LogoURL,
		"website_url": brand.WebsiteURL,
		"is_enabled":  brand.IsEnabled,
		"sort_order":  brand.SortOrder,
	}).Error
}

func (r *ProductBrandRepository) Delete(id uint) error {
	return r.db.Delete(&product.ProductBrand{}, id).Error
}
