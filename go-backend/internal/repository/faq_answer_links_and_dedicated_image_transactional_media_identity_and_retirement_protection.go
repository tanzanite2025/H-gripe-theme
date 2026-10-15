package repository

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/pkg/faqcontent"
	"commerce-platform/internal/pkg/safehtml"
)

var ErrFAQMediaReferenceInvalid = errors.New("FAQ media reference is invalid")

// Omitted media fields are reloaded under the FAQ row lock before saving.
type FAQMediaFieldMutation struct {
	UpdateAnswer      bool
	UpdateAnswerImage bool
}

func (r *FAQRepository) validateAndCanonicalizeFAQAnswerLinksAndImageInOneAssetLockPlan(tx *gorm.DB, item, previous *faq.FAQ) error {
	var bindings []publicBusinessMediaReferenceBinding
	var document *safehtml.RichTextMediaReferenceDocument
	if item.Answer != previous.Answer {
		// FAQ supports links and text only. Its own sanitizer removes images,
		// video and unsupported containers before media attributes are collected.
		answer, err := faqcontent.SanitizeAnswer(item.Answer)
		if err != nil {
			return err
		}
		previousAnswer, err := faqcontent.SanitizeAnswer(previous.Answer)
		if err != nil {
			return err
		}
		item.Answer = answer
		bindings, document, err = buildPublicBusinessRichTextMediaAttributeBindingsWithPlacementScopedLegacyCompatibility(answer, previousAnswer)
		if err != nil {
			return err
		}
	}
	newAnswerImage := strings.TrimSpace(item.AnswerImageURL) != "" && item.AnswerImageURL != previous.AnswerImageURL
	requestedImageURL := item.AnswerImageURL
	if newAnswerImage {
		bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &item.AnswerImageURL, ExpectedMediaType: "image"})
	}
	references := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		references = append(references, *binding.URL)
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, r.businessMediaReferenceResolver, references, nil)
	if err != nil {
		return mapPublicBusinessMediaValidationErrorToFAQMediaValidationError(err)
	}
	for _, binding := range bindings {
		if err := resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(binding); err != nil {
			return mapPublicBusinessMediaValidationErrorToFAQMediaValidationError(err)
		}
	}
	if newAnswerImage {
		identity := resolved.identities[requestedImageURL]
		asset := resolved.assets[identity.AssetID]
		if asset == nil {
			return fmt.Errorf("%w: answer image must be a registered public asset", ErrFAQMediaReferenceInvalid)
		}
		// The canonical original, rather than a client supplied size or an
		// arbitrary thumbnail, must satisfy the dedicated FAQ image contract.
		if asset.MimeType != "image/webp" || asset.Width != 800 || asset.Height != 800 {
			return fmt.Errorf("%w: answer image must be an 800x800 WebP original", ErrFAQMediaReferenceInvalid)
		}
		item.AnswerImageWidth, item.AnswerImageHeight = asset.Width, asset.Height
	}
	if document != nil {
		canonical, err := document.RenderRichTextWithCanonicalMediaReferences()
		if err != nil {
			return err
		}
		item.Answer = canonical
	}
	return nil
}

func mapPublicBusinessMediaValidationErrorToFAQMediaValidationError(err error) error {
	if errors.Is(err, ErrProductMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrFAQMediaReferenceInvalid, err)
	}
	return err
}
