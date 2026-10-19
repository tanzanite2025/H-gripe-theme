package repository

import (
	"errors"
	"fmt"

	"commerce-platform/internal/domain/post"
	"commerce-platform/internal/pkg/safehtml"
	"gorm.io/gorm"
)

var ErrPostMediaReferenceInvalid = errors.New("post media reference is invalid")

type PostMediaFieldMutation struct {
	UpdateContent       bool
	UpdateFeaturedImage bool
}

func (r *PostRepository) validateAndCanonicalizePostContentAndFeaturedImageInOneAssetLockPlan(tx *gorm.DB, item, previous *post.Post) error {
	var bindings []publicBusinessMediaReferenceBinding
	var document *safehtml.RichTextMediaReferenceDocument
	if item.Content != previous.Content {
		content, err := safehtml.Sanitize(item.Content)
		if err != nil {
			return err
		}
		previousContent, err := safehtml.Sanitize(previous.Content)
		if err != nil {
			return err
		}
		item.Content = content
		bindings, document, err = buildPublicBusinessRichTextMediaAttributeBindingsWithPlacementScopedLegacyCompatibility(content, previousContent)
		if err != nil {
			return err
		}
	}
	if item.FeaturedImg != previous.FeaturedImg {
		bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &item.FeaturedImg, ExpectedMediaType: "image"})
	}
	references := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		references = append(references, *binding.URL)
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, r.businessMediaReferenceResolver, references, nil)
	if err != nil {
		return mapPublicBusinessMediaValidationErrorToPostMediaValidationError(err)
	}
	for _, binding := range bindings {
		if err := resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(binding); err != nil {
			return mapPublicBusinessMediaValidationErrorToPostMediaValidationError(err)
		}
	}
	if document != nil {
		canonical, err := document.RenderRichTextWithCanonicalMediaReferences()
		if err != nil {
			return err
		}
		item.Content = canonical
	}
	return nil
}

func mapPublicBusinessMediaValidationErrorToPostMediaValidationError(err error) error {
	if errors.Is(err, ErrProductMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrPostMediaReferenceInvalid, err)
	}
	return err
}
