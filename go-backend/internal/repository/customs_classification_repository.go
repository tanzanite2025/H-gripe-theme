package repository

import (
	"commerce-platform/internal/domain/product"
	"strings"

	"gorm.io/gorm"
)

type CustomsClassificationRepository struct {
	db *gorm.DB
}

type CustomsClassificationListFilter struct {
	ProductSpecificationTemplateID uint
	ComponentKind                  string
	Material                       string
	Status                         string
	Search                         string
	IncludePaused                  bool
}

func NewCustomsClassificationRepository(db *gorm.DB) *CustomsClassificationRepository {
	return &CustomsClassificationRepository{db: db}
}

func (r *CustomsClassificationRepository) List(filter CustomsClassificationListFilter) ([]product.CustomsClassificationProfile, error) {
	var profiles []product.CustomsClassificationProfile
	query := r.db.Preload("ProductSpecificationTemplate").Model(&product.CustomsClassificationProfile{})
	if filter.ProductSpecificationTemplateID > 0 {
		query = query.Where("product_specification_template_id = ?", filter.ProductSpecificationTemplateID)
	}
	componentKind := strings.TrimSpace(filter.ComponentKind)
	if componentKind != "" {
		query = query.Where("LOWER(component_kind) = ?", strings.ToLower(componentKind))
	}
	material := strings.TrimSpace(filter.Material)
	if material != "" {
		query = query.Where("LOWER(material) = ?", strings.ToLower(material))
	}
	status := strings.TrimSpace(filter.Status)
	if status != "" {
		query = query.Where("status = ?", status)
	} else if !filter.IncludePaused {
		query = query.Where("status <> ?", product.CustomsClassificationStatusPaused)
	}
	search := strings.TrimSpace(filter.Search)
	if search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where(
			"LOWER(name) LIKE ? OR LOWER(slug) LIKE ? OR LOWER(component_kind) LIKE ? OR LOWER(hs_code) LIKE ? OR LOWER(cn_code) LIKE ?",
			like,
			like,
			like,
			like,
			like,
		)
	}

	err := query.Order("CASE WHEN product_specification_template_id IS NULL THEN 1 ELSE 0 END ASC").Order("name ASC").Order("id ASC").Find(&profiles).Error
	return profiles, err
}

func (r *CustomsClassificationRepository) FindByID(id uint) (*product.CustomsClassificationProfile, error) {
	var profile product.CustomsClassificationProfile
	if err := r.db.Preload("ProductSpecificationTemplate").First(&profile, id).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *CustomsClassificationRepository) SlugExists(slug string, excludeID uint) (bool, error) {
	query := r.db.Model(&product.CustomsClassificationProfile{}).Where("slug = ?", slug)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *CustomsClassificationRepository) Create(profile *product.CustomsClassificationProfile) error {
	return r.db.Create(profile).Error
}

func (r *CustomsClassificationRepository) Update(profile *product.CustomsClassificationProfile) error {
	return r.db.Model(&product.CustomsClassificationProfile{}).Where("id = ?", profile.ID).Updates(map[string]interface{}{
		"product_specification_template_id": profile.ProductSpecificationTemplateID,
		"name":                              profile.Name,
		"slug":                              profile.Slug,
		"component_kind":                    profile.ComponentKind,
		"material":                          profile.Material,
		"hs_code":                           profile.HSCode,
		"cn_code":                           profile.CNCode,
		"country_of_origin":                 profile.CountryOfOrigin,
		"customs_description":               profile.CustomsDescription,
		"source":                            profile.Source,
		"source_code":                       profile.SourceCode,
		"source_url":                        profile.SourceURL,
		"notes":                             profile.Notes,
		"status":                            profile.Status,
	}).Error
}

func (r *CustomsClassificationRepository) Delete(id uint) error {
	return r.db.Delete(&product.CustomsClassificationProfile{}, id).Error
}
