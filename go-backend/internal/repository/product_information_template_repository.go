package repository

import (
	"tanzanite/internal/domain/product"

	"gorm.io/gorm"
)

type ProductInformationTemplateRepository struct {
	db *gorm.DB
}

func NewProductInformationTemplateRepository(db *gorm.DB) *ProductInformationTemplateRepository {
	return &ProductInformationTemplateRepository{db: db}
}

func (r *ProductInformationTemplateRepository) List(kind, locale string, includeDisabled bool) ([]product.ProductInformationTemplate, error) {
	var templates []product.ProductInformationTemplate
	query := r.db.Model(&product.ProductInformationTemplate{})
	if kind != "" {
		query = query.Where("kind = ?", kind)
	}
	if locale != "" {
		query = query.Where("locale = ?", locale)
	}
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}
	err := query.Order("kind ASC").Order("sort_order ASC").Order("id ASC").Find(&templates).Error
	return templates, err
}

func (r *ProductInformationTemplateRepository) FindByID(id uint) (*product.ProductInformationTemplate, error) {
	var template product.ProductInformationTemplate
	if err := r.db.First(&template, id).Error; err != nil {
		return nil, err
	}
	return &template, nil
}

func (r *ProductInformationTemplateRepository) ExistsByKindSlugLocale(kind, slug, locale string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Model(&product.ProductInformationTemplate{}).
		Where("kind = ? AND slug = ? AND locale = ?", kind, slug, locale)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProductInformationTemplateRepository) Create(template *product.ProductInformationTemplate) error {
	return r.db.Select("*").Create(template).Error
}

func (r *ProductInformationTemplateRepository) Update(template *product.ProductInformationTemplate) error {
	return r.db.Model(&product.ProductInformationTemplate{}).Where("id = ?", template.ID).Updates(map[string]interface{}{
		"name":       template.Name,
		"slug":       template.Slug,
		"content":    template.Content,
		"locale":     template.Locale,
		"is_enabled": template.IsEnabled,
		"sort_order": template.SortOrder,
	}).Error
}

func (r *ProductInformationTemplateRepository) Delete(id uint) error {
	return r.db.Unscoped().Delete(&product.ProductInformationTemplate{}, id).Error
}
