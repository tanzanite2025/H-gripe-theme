package repository

import "commerce-platform/internal/pkg/safehtml"

func buildPublicBusinessRichTextMediaAttributeBindingsWithPlacementScopedLegacyCompatibility(requested, previous string) ([]publicBusinessMediaReferenceBinding, *safehtml.RichTextMediaReferenceDocument, error) {
	document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(requested)
	if err != nil {
		return nil, nil, err
	}
	oldDocument, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(previous)
	if err != nil {
		return nil, nil, err
	}
	previousOccurrences := make(map[string]int)
	for _, reference := range oldDocument.References {
		previousOccurrences[reference.ElementAndAttribute+"\x00"+*reference.URL]++
	}
	bindings := make([]publicBusinessMediaReferenceBinding, 0, len(document.References))
	for _, reference := range document.References {
		key := reference.ElementAndAttribute + "\x00" + *reference.URL
		retained := previousOccurrences[key] > 0
		if retained {
			previousOccurrences[key]--
		}
		bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: reference.URL, ExpectedMediaType: reference.ExpectedMediaType, PreviousURL: *reference.URL, RetainsExistingPlacement: retained})
	}
	return bindings, document, nil
}
