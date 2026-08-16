package repository

import (
	"commerce-platform/internal/domain/product"
	"strings"

	"gorm.io/gorm"
)

func (r *ProductRepository) FindAttributeByID(id uint) (*product.ProductAttribute, error) {
	var attr product.ProductAttribute
	err := r.db.Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_attribute_values.sort_order ASC")
	}).First(&attr, id).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

func (r *ProductRepository) FindAttributeBySlug(slug string) (*product.ProductAttribute, error) {
	var attr product.ProductAttribute
	err := r.db.Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_attribute_values.sort_order ASC")
	}).Where("slug = ?", slug).First(&attr).Error
	if err != nil {
		return nil, err
	}
	return &attr, nil
}

func (r *ProductRepository) FindAllAttributes(page, pageSize int) ([]product.ProductAttribute, int64, error) {
	var attrs []product.ProductAttribute
	var total int64

	query := r.db.Model(&product.ProductAttribute{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Order("product_attribute_values.sort_order ASC")
	}).Order("sort_order ASC, id ASC").Offset(offset).Limit(pageSize).Find(&attrs).Error

	return attrs, total, err
}

func (r *ProductRepository) CreateAttribute(attr *product.ProductAttribute) error {
	return r.db.Create(attr).Error
}

func (r *ProductRepository) UpdateAttribute(attr *product.ProductAttribute) error {
	return r.db.Save(attr).Error
}

func (r *ProductRepository) DeleteAttribute(id uint) error {
	if err := r.db.Where("attribute_id = ?", id).Delete(&product.AttributeValue{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&product.ProductAttribute{}, id).Error
}

func (r *ProductRepository) FindFilterableAttributes() ([]product.ProductAttribute, error) {
	var attrs []product.ProductAttribute
	err := r.db.Preload("Values", func(db *gorm.DB) *gorm.DB {
		return db.Where("is_enabled = ?", true).Order("product_attribute_values.sort_order ASC")
	}).Where("is_filterable = ? AND is_enabled = ?", true, true).Order("sort_order ASC").Find(&attrs).Error
	return attrs, err
}

func (r *ProductRepository) FindAttributeValueByID(id uint) (*product.AttributeValue, error) {
	var val product.AttributeValue
	err := r.db.First(&val, id).Error
	if err != nil {
		return nil, err
	}
	return &val, nil
}

func (r *ProductRepository) CreateAttributeValue(val *product.AttributeValue) error {
	return r.db.Create(val).Error
}

func (r *ProductRepository) UpdateAttributeValue(val *product.AttributeValue) error {
	return r.db.Save(val).Error
}

func (r *ProductRepository) DeleteAttributeValue(id uint) error {
	return r.db.Delete(&product.AttributeValue{}, id).Error
}

func (r *ProductRepository) FindValuesByAttributeID(attrID uint) ([]product.AttributeValue, error) {
	var values []product.AttributeValue
	err := r.db.Where("attribute_id = ?", attrID).Order("sort_order ASC").Find(&values).Error
	return values, err
}

func (r *ProductRepository) FindAllProductSpecificationTemplates(includeDisabled bool) ([]product.ProductSpecificationTemplate, error) {
	var productSpecificationTemplates []product.ProductSpecificationTemplate
	query := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("Translations", func(db *gorm.DB) *gorm.DB {
		return db.Order("locale ASC, id ASC")
	})
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}

	err := query.Order("sort_order ASC, id ASC").Find(&productSpecificationTemplates).Error
	return productSpecificationTemplates, err
}

func (r *ProductRepository) FindPublicProductSpecificationTemplates(includeDisabled bool) ([]product.ProductSpecificationTemplate, error) {
	var productSpecificationTemplates []product.ProductSpecificationTemplate
	query := r.db.
		Select("id", "name", "slug", "image_url", "sort_order", "is_enabled").
		Preload("Translations", func(db *gorm.DB) *gorm.DB {
			return db.Order("locale ASC, id ASC")
		})
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}

	err := query.Order("sort_order ASC, id ASC").Find(&productSpecificationTemplates).Error
	return productSpecificationTemplates, err
}

func (r *ProductRepository) FindProductSpecificationTemplateByID(id uint) (*product.ProductSpecificationTemplate, error) {
	var productSpecificationTemplate product.ProductSpecificationTemplate
	err := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("Translations", func(db *gorm.DB) *gorm.DB {
		return db.Order("locale ASC, id ASC")
	}).First(&productSpecificationTemplate, id).Error
	if err != nil {
		return nil, err
	}
	return &productSpecificationTemplate, nil
}

func (r *ProductRepository) FindProductSpecificationTemplateBySlug(slug string) (*product.ProductSpecificationTemplate, error) {
	var productSpecificationTemplate product.ProductSpecificationTemplate
	err := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Preload("Translations", func(db *gorm.DB) *gorm.DB {
		return db.Order("locale ASC, id ASC")
	}).Where("slug = ?", slug).First(&productSpecificationTemplate).Error
	if err != nil {
		return nil, err
	}
	return &productSpecificationTemplate, nil
}

func (r *ProductRepository) ProductSpecificationTemplateSlugExists(slug string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Model(&product.ProductSpecificationTemplate{}).Where("slug = ?", slug)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProductRepository) CreateProductSpecificationTemplate(productSpecificationTemplate *product.ProductSpecificationTemplate) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		definitions := productSpecificationTemplate.SpecDefinitions
		translations := productSpecificationTemplate.Translations
		isEnabled := productSpecificationTemplate.IsEnabled
		productSpecificationTemplate.SpecDefinitions = nil
		productSpecificationTemplate.Translations = nil
		if err := tx.Create(productSpecificationTemplate).Error; err != nil {
			return err
		}
		if err := tx.Model(productSpecificationTemplate).Update("is_enabled", isEnabled).Error; err != nil {
			return err
		}
		productSpecificationTemplate.IsEnabled = isEnabled

		for index := range definitions {
			definitions[index].ProductSpecificationTemplateID = productSpecificationTemplate.ID
			if err := createSpecDefinition(tx, &definitions[index]); err != nil {
				return err
			}
		}
		productSpecificationTemplate.SpecDefinitions = definitions

		for index := range translations {
			translations[index].ProductSpecificationTemplateID = productSpecificationTemplate.ID
		}
		if len(translations) > 0 {
			if err := tx.Create(&translations).Error; err != nil {
				return err
			}
		}
		productSpecificationTemplate.Translations = translations
		return nil
	})
}

func (r *ProductRepository) UpdateProductSpecificationTemplate(productSpecificationTemplate *product.ProductSpecificationTemplate, removedSpecIDs []uint, updateTranslations bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&product.ProductSpecificationTemplate{}).Where("id = ?", productSpecificationTemplate.ID).Updates(map[string]interface{}{
			"name":        productSpecificationTemplate.Name,
			"slug":        productSpecificationTemplate.Slug,
			"description": productSpecificationTemplate.Description,
			"sort_order":  productSpecificationTemplate.SortOrder,
			"is_enabled":  productSpecificationTemplate.IsEnabled,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		for index := range productSpecificationTemplate.SpecDefinitions {
			definition := &productSpecificationTemplate.SpecDefinitions[index]
			definition.ProductSpecificationTemplateID = productSpecificationTemplate.ID
			if definition.ID == 0 {
				if err := createSpecDefinition(tx, definition); err != nil {
					return err
				}
				continue
			}

			result = tx.Model(&product.SpecDefinition{}).
				Where("id = ? AND product_specification_template_id = ?", definition.ID, productSpecificationTemplate.ID).
				Updates(specDefinitionUpdates(definition))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}

		if len(removedSpecIDs) > 0 {
			if err := tx.Where("product_specification_template_id = ? AND id IN ?", productSpecificationTemplate.ID, removedSpecIDs).
				Delete(&product.SpecDefinition{}).Error; err != nil {
				return err
			}
		}

		if updateTranslations {
			if err := tx.Where("product_specification_template_id = ?", productSpecificationTemplate.ID).
				Delete(&product.ProductSpecificationTemplateTranslation{}).Error; err != nil {
				return err
			}

			translations := productSpecificationTemplate.Translations
			for index := range translations {
				translations[index].ID = 0
				translations[index].ProductSpecificationTemplateID = productSpecificationTemplate.ID
			}
			if len(translations) > 0 {
				if err := tx.Create(&translations).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *ProductRepository) UpdateProductSpecificationTemplateImage(id uint, mediaAssetID *uint, imageURL string) error {
	updates := map[string]interface{}{
		"image_url": strings.TrimSpace(imageURL),
	}
	if mediaAssetID == nil {
		updates["image_media_asset_id"] = nil
	} else {
		updates["image_media_asset_id"] = *mediaAssetID
	}

	result := r.db.Model(&product.ProductSpecificationTemplate{}).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *ProductRepository) DeleteProductSpecificationTemplate(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_specification_template_id = ?", id).Delete(&product.SpecDefinition{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&product.ProductSpecificationTemplate{}, id)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
}

func createSpecDefinition(tx *gorm.DB, definition *product.SpecDefinition) error {
	updates := specDefinitionUpdates(definition)
	if err := tx.Create(definition).Error; err != nil {
		return err
	}
	return tx.Model(definition).Updates(updates).Error
}

func specDefinitionUpdates(definition *product.SpecDefinition) map[string]interface{} {
	return map[string]interface{}{
		"group":             definition.Group,
		"name":              definition.Name,
		"slug":              definition.Slug,
		"field_type":        definition.FieldType,
		"presentation":      definition.Presentation,
		"unit":              definition.Unit,
		"is_required":       definition.IsRequired,
		"is_filterable":     definition.IsFilterable,
		"is_visible":        definition.IsVisible,
		"is_variant_option": definition.IsVariantOption,
		"sort_order":        definition.SortOrder,
		"options":           definition.Options,
		"validation":        definition.Validation,
	}
}
