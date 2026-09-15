package repository

import (
	"commerce-platform/internal/domain/product"

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
	})
	if r.db.Migrator().HasTable(&product.ProductSpecOptionItem{}) {
		query = preloadSpecDefinitionOptionItems(query)
	}
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}

	err := query.Order("sort_order ASC, id ASC").Find(&productSpecificationTemplates).Error
	return productSpecificationTemplates, err
}

func (r *ProductRepository) FindPublicProductSpecificationTemplates(includeDisabled bool) ([]product.ProductSpecificationTemplate, error) {
	var productSpecificationTemplates []product.ProductSpecificationTemplate
	query := r.db.Select("id", "name", "slug", "sort_order", "is_enabled")
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}

	err := query.Order("sort_order ASC, id ASC").Find(&productSpecificationTemplates).Error
	return productSpecificationTemplates, err
}

func (r *ProductRepository) FindProductSpecificationTemplateByID(id uint) (*product.ProductSpecificationTemplate, error) {
	var productSpecificationTemplate product.ProductSpecificationTemplate
	query := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	})
	if r.db.Migrator().HasTable(&product.ProductSpecOptionItem{}) {
		query = preloadSpecDefinitionOptionItems(query)
	}
	err := query.First(&productSpecificationTemplate, id).Error
	if err != nil {
		return nil, err
	}
	return &productSpecificationTemplate, nil
}

func (r *ProductRepository) FindProductSpecificationTemplateBySlug(slug string) (*product.ProductSpecificationTemplate, error) {
	var productSpecificationTemplate product.ProductSpecificationTemplate
	query := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	})
	if r.db.Migrator().HasTable(&product.ProductSpecOptionItem{}) {
		query = preloadSpecDefinitionOptionItems(query)
	}
	err := query.Where("slug = ?", slug).First(&productSpecificationTemplate).Error
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
		isEnabled := productSpecificationTemplate.IsEnabled
		productSpecificationTemplate.SpecDefinitions = nil
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

		return nil
	})
}

func (r *ProductRepository) UpdateProductSpecificationTemplate(productSpecificationTemplate *product.ProductSpecificationTemplate, removedSpecIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&product.ProductSpecificationTemplate{}).Where("id = ?", productSpecificationTemplate.ID).Updates(map[string]interface{}{
			"name":        productSpecificationTemplate.Name,
			"slug":        productSpecificationTemplate.Slug,
			"description": productSpecificationTemplate.Description,
			"sort_order":  productSpecificationTemplate.SortOrder,
			"is_enabled":  productSpecificationTemplate.IsEnabled,
			"revision":    productSpecificationTemplate.Revision,
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
			if err := replaceSpecOptionItems(tx, definition.ID, definition.OptionItems); err != nil {
				return err
			}
		}

		if len(removedSpecIDs) > 0 {
			if err := tx.Where("product_specification_template_id = ? AND id IN ?", productSpecificationTemplate.ID, removedSpecIDs).
				Delete(&product.SpecDefinition{}).Error; err != nil {
				return err
			}
		}

		return nil
	})
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
	optionItems := definition.OptionItems
	definition.OptionItems = nil
	updates := specDefinitionUpdates(definition)
	if err := tx.Create(definition).Error; err != nil {
		return err
	}
	definition.OptionItems = optionItems
	if err := tx.Model(definition).Updates(updates).Error; err != nil {
		return err
	}
	return replaceSpecOptionItems(tx, definition.ID, optionItems)
}

func replaceSpecOptionItems(tx *gorm.DB, definitionID uint, items []product.ProductSpecOptionItem) error {
	if !tx.Migrator().HasTable(&product.ProductSpecOptionItem{}) {
		return nil
	}
	var existing []product.ProductSpecOptionItem
	if err := tx.Where("spec_definition_id = ?", definitionID).Find(&existing).Error; err != nil {
		return err
	}
	existingByID := make(map[uint]struct{}, len(existing))
	for _, item := range existing {
		existingByID[item.ID] = struct{}{}
	}
	keepIDs := make([]uint, 0, len(items))
	for index := range items {
		items[index].SpecDefinitionID = definitionID
		if items[index].ID != 0 {
			if _, ok := existingByID[items[index].ID]; !ok {
				return gorm.ErrRecordNotFound
			}
			if err := tx.Save(&items[index]).Error; err != nil {
				return err
			}
		} else if err := tx.Create(&items[index]).Error; err != nil {
			return err
		}
		keepIDs = append(keepIDs, items[index].ID)
	}
	deleteQuery := tx.Where("spec_definition_id = ?", definitionID)
	if len(keepIDs) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", keepIDs)
	}
	return deleteQuery.Delete(&product.ProductSpecOptionItem{}).Error
}

func specDefinitionUpdates(definition *product.SpecDefinition) map[string]interface{} {
	return map[string]interface{}{
		"group":          definition.Group,
		"name":           definition.Name,
		"slug":           definition.Slug,
		"field_type":     definition.FieldType,
		"role":           definition.Role,
		"selection_mode": definition.SelectionMode,
		"min_selections": definition.MinSelections,
		"max_selections": definition.MaxSelections,
		"presentation":   definition.Presentation,
		"unit":           definition.Unit,
		"is_required":    definition.IsRequired,
		"is_filterable":  definition.IsFilterable,
		"is_visible":     definition.IsVisible,
		"sort_order":     definition.SortOrder,
		"validation":     definition.Validation,
	}
}
