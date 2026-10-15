package repository

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/safehtml"
)

// An omitted HTML field must be reloaded inside the product transaction, so a
// concurrent detach cannot be overwritten by a stale service-layer snapshot.
type ProductRichTextFieldMutation struct {
	UpdateDescription      bool
	UpdateShortDescription bool
}

func buildProductOptionSwatchMediaBindings(values, previous []product.ProductVariantOptionValue) []publicBusinessMediaReferenceBinding {
	existing := make(map[uint]product.ProductVariantOptionValue, len(previous))
	for _, value := range previous {
		existing[value.ID] = value
	}
	bindings := make([]publicBusinessMediaReferenceBinding, 0, len(values))
	for index := range values {
		value := &values[index]
		old, retained := existing[value.ID]
		retained = retained && old.SpecDefinitionID == value.SpecDefinitionID
		bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &value.SwatchURL, AssetID: &value.SwatchMediaAssetID, ExpectedMediaType: "image", PreviousURL: old.SwatchURL, PreviousAssetID: old.SwatchMediaAssetID, RetainsExistingPlacement: retained})
	}
	return bindings
}

// All fields participate in one sorted asset lock acquisition, including the
// product gallery. Do not validate separate field groups in separate lock orders.
func (r *ProductRepository) validateAndCanonicalizeCombinedProductMediaBindings(tx *gorm.DB, item *product.Product, previous *product.Product, mediaItems []product.ProductMedia, replaceMedia bool, optionValues []product.ProductVariantOptionValue, replaceOptions bool) error {
	var bindings []publicBusinessMediaReferenceBinding
	type richTextFieldDocument struct {
		target   *string
		document *safehtml.RichTextMediaReferenceDocument
	}
	var documents []richTextFieldDocument
	for _, field := range []struct {
		target   *string
		previous string
	}{{&item.Description, previous.Description}, {&item.ShortDesc, previous.ShortDesc}} {
		// Exact unchanged HTML retains its existing relationships. A translation
		// passes an empty previous record and must validate every copied field.
		if *field.target == field.previous {
			continue
		}
		fieldBindings, document, err := buildPublicBusinessRichTextMediaAttributeBindingsWithPlacementScopedLegacyCompatibility(*field.target, field.previous)
		if err != nil {
			return err
		}
		bindings = append(bindings, fieldBindings...)
		documents = append(documents, richTextFieldDocument{field.target, document})
	}
	if replaceOptions {
		var existing []product.ProductVariantOptionValue
		if previous.ID != 0 {
			if err := tx.Where("product_id = ?", previous.ID).Find(&existing).Error; err != nil {
				return err
			}
		}
		bindings = append(bindings, buildProductOptionSwatchMediaBindings(optionValues, existing)...)
	}
	var existingMedia []product.ProductMedia
	if replaceMedia && previous.ID != 0 {
		if err := tx.Where("product_id = ?", previous.ID).Find(&existingMedia).Error; err != nil {
			return err
		}
	}
	if !replaceMedia {
		mediaItems = nil
	}
	if err := r.lockValidateAndCanonicalizeProductMediaAssetBindings(tx, mediaItems, existingMedia, bindings...); err != nil {
		return err
	}
	for _, field := range documents {
		canonical, err := field.document.RenderRichTextWithCanonicalMediaReferences()
		if err != nil {
			return err
		}
		*field.target = canonical
	}
	return nil
}

func (r *ProductRepository) lockCurrentProductAndValidateMediaMutation(tx *gorm.DB, item *product.Product, mediaItems []product.ProductMedia, replaceMedia bool, optionValues []product.ProductVariantOptionValue, replaceOptions bool, richTextMutation []ProductRichTextFieldMutation) error {
	var previous product.Product
	if err := tx.Clauses(clause.Locking{Strength: "NO KEY UPDATE"}).First(&previous, item.ID).Error; err != nil {
		return err
	}
	if len(richTextMutation) > 0 {
		if !richTextMutation[0].UpdateDescription {
			item.Description = previous.Description
		}
		if !richTextMutation[0].UpdateShortDescription {
			item.ShortDesc = previous.ShortDesc
		}
	}
	return r.validateAndCanonicalizeCombinedProductMediaBindings(tx, item, &previous, mediaItems, replaceMedia, optionValues, replaceOptions)
}
