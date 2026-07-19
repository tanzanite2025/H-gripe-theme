package repository

import (
	"tanzanite/internal/domain/product"

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

func (r *ProductRepository) FindAllProductTypes(includeDisabled bool) ([]product.ProductType, error) {
	var productTypes []product.ProductType
	query := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	})
	if !includeDisabled {
		query = query.Where("is_enabled = ?", true)
	}

	err := query.Order("sort_order ASC, id ASC").Find(&productTypes).Error
	return productTypes, err
}

func (r *ProductRepository) FindProductTypeByID(id uint) (*product.ProductType, error) {
	var productType product.ProductType
	err := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).First(&productType, id).Error
	if err != nil {
		return nil, err
	}
	return &productType, nil
}

func (r *ProductRepository) FindProductTypeBySlug(slug string) (*product.ProductType, error) {
	var productType product.ProductType
	err := r.db.Preload("SpecDefinitions", func(db *gorm.DB) *gorm.DB {
		return orderSpecDefinitions(db)
	}).Where("slug = ?", slug).First(&productType).Error
	if err != nil {
		return nil, err
	}
	return &productType, nil
}

func (r *ProductRepository) ProductTypeSlugExists(slug string, excludeID uint) (bool, error) {
	var count int64
	query := r.db.Model(&product.ProductType{}).Where("slug = ?", slug)
	if excludeID > 0 {
		query = query.Where("id <> ?", excludeID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProductRepository) CreateProductType(productType *product.ProductType) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		definitions := productType.SpecDefinitions
		isEnabled := productType.IsEnabled
		productType.SpecDefinitions = nil
		if err := tx.Create(productType).Error; err != nil {
			return err
		}
		if err := tx.Model(productType).Update("is_enabled", isEnabled).Error; err != nil {
			return err
		}
		productType.IsEnabled = isEnabled

		for index := range definitions {
			definitions[index].ProductTypeID = productType.ID
			if err := createSpecDefinition(tx, &definitions[index]); err != nil {
				return err
			}
		}
		productType.SpecDefinitions = definitions
		return nil
	})
}

func (r *ProductRepository) UpdateProductType(productType *product.ProductType, removedSpecIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&product.ProductType{}).Where("id = ?", productType.ID).Updates(map[string]interface{}{
			"name":        productType.Name,
			"slug":        productType.Slug,
			"description": productType.Description,
			"sort_order":  productType.SortOrder,
			"is_enabled":  productType.IsEnabled,
		})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		for index := range productType.SpecDefinitions {
			definition := &productType.SpecDefinitions[index]
			definition.ProductTypeID = productType.ID
			if definition.ID == 0 {
				if err := createSpecDefinition(tx, definition); err != nil {
					return err
				}
				continue
			}

			result = tx.Model(&product.SpecDefinition{}).
				Where("id = ? AND product_type_id = ?", definition.ID, productType.ID).
				Updates(specDefinitionUpdates(definition))
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return gorm.ErrRecordNotFound
			}
		}

		if len(removedSpecIDs) > 0 {
			if err := tx.Where("product_type_id = ? AND id IN ?", productType.ID, removedSpecIDs).
				Delete(&product.SpecDefinition{}).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *ProductRepository) DeleteProductType(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("product_type_id = ?", id).Delete(&product.SpecDefinition{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&product.ProductType{}, id)
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
