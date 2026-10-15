package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/product"
)

func (r *ProductRepository) lockValidateAndCanonicalizeTemplateOptionSwatchMediaBindings(tx *gorm.DB, template *product.ProductSpecificationTemplate, updating bool) error {
	var existing []product.ProductSpecOptionItem
	if updating {
		var current product.ProductSpecificationTemplate
		if err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(&current, template.ID).Error; err != nil {
			return err
		}
		if tx.Migrator().HasTable(&product.ProductSpecOptionItem{}) {
			if err := tx.Table("product_spec_option_items AS items").Select("items.*").
				Joins("JOIN product_spec_definitions ON product_spec_definitions.id = items.spec_definition_id").
				Where("product_spec_definitions.product_specification_template_id = ?", template.ID).Find(&existing).Error; err != nil {
				return err
			}
		}
	}
	existingByID := make(map[uint]product.ProductSpecOptionItem, len(existing))
	for _, item := range existing {
		existingByID[item.ID] = item
	}
	var bindings []publicBusinessMediaReferenceBinding
	for definitionIndex := range template.SpecDefinitions {
		definition := &template.SpecDefinitions[definitionIndex]
		for itemIndex := range definition.OptionItems {
			item := &definition.OptionItems[itemIndex]
			old, retained := existingByID[item.ID]
			retained = retained && old.SpecDefinitionID == definition.ID
			bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &item.SwatchURL, AssetID: &item.SwatchMediaAssetID, ExpectedMediaType: "image", PreviousURL: old.SwatchURL, PreviousAssetID: old.SwatchMediaAssetID, RetainsExistingPlacement: retained})
		}
	}
	return r.lockValidateAndCanonicalizeProductMediaAssetBindings(tx, nil, nil, bindings...)
}
