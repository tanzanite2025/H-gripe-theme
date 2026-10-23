package repository

import (
	"commerce-platform/internal/pkg/safehtml"
	"gorm.io/gorm"
)

type templateOrCategoryRichTextMediaFieldMutation struct {
	Requested *string
	Previous  string
}

// The business row is locked by the caller before this plan is built. Every
// requested asset joins one sorted lock order, shared with retirement. Exact
// unchanged HTML keeps its existing relationships without adding a placement.
func lockValidateAndCanonicalizeTemplateOrCategoryRichTextMediaWithinTransaction(tx *gorm.DB, resolver PublicBusinessMediaReferenceObjectKeyResolver, requested, previous string, additionalAssetIDs []uint) (string, *lockedPublicBusinessMediaReferenceIdentities, error) {
	resolved, err := lockValidateAndCanonicalizeTemplateOrCategoryRichTextFieldsWithOneAssetLockPlan(tx, resolver, []templateOrCategoryRichTextMediaFieldMutation{{Requested: &requested, Previous: previous}}, additionalAssetIDs)
	return requested, resolved, err
}

func lockValidateAndCanonicalizeTemplateOrCategoryRichTextFieldsWithOneAssetLockPlan(tx *gorm.DB, resolver PublicBusinessMediaReferenceObjectKeyResolver, fields []templateOrCategoryRichTextMediaFieldMutation, additionalAssetIDs []uint) (*lockedPublicBusinessMediaReferenceIdentities, error) {
	var bindings []publicBusinessMediaReferenceBinding
	var documents []struct {
		target   *string
		document *safehtml.RichTextMediaReferenceDocument
	}
	for _, field := range fields {
		if *field.Requested == field.Previous {
			continue
		}
		fieldBindings, document, err := buildPublicBusinessRichTextMediaAttributeBindingsWithPlacementScopedLegacyCompatibility(*field.Requested, field.Previous)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, fieldBindings...)
		documents = append(documents, struct {
			target   *string
			document *safehtml.RichTextMediaReferenceDocument
		}{field.Requested, document})
	}
	references := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		references = append(references, *binding.URL)
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, resolver, references, additionalAssetIDs)
	if err != nil {
		return nil, err
	}
	for _, binding := range bindings {
		if err := resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(binding); err != nil {
			return nil, err
		}
	}
	for _, field := range documents {
		canonical, err := field.document.RenderRichTextWithCanonicalMediaReferences()
		if err != nil {
			return nil, err
		}
		*field.target = canonical
	}
	return resolved, nil
}
