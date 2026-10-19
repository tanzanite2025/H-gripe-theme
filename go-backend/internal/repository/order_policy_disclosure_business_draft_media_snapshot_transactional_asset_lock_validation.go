package repository

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/refundcancellation"
	"commerce-platform/internal/pkg/storage"
	"gorm.io/gorm"
)

// Snapshot bytes and their hash remain immutable. Only the new draft namespace
// joins the asset lock protocol; legacy snapshot formats remain unchanged.
func lockBusinessDraftMediaBeforeCapturingImmutableOrderPolicyDisclosure(tx *gorm.DB, disclosure *order.PolicyDisclosure) error {
	if disclosure == nil {
		return gorm.ErrInvalidData
	}
	var policy refundcancellation.Policy
	if err := json.Unmarshal([]byte(disclosure.PolicyJSON), &policy); err != nil {
		if strings.Contains(disclosure.PolicyJSON, storage.BusinessDraftMediaObjectPrefix) {
			return fmt.Errorf("decode order policy snapshot media: %w", err)
		}
		return nil
	}
	var references []string
	imageReferences := make(map[string]bool)
	for _, section := range policy.Sections {
		if section.Image == nil {
			continue
		}
		parsed, err := url.Parse(section.Image.URL)
		if err != nil {
			return err
		}
		// Detection also covers configured base-URL subpaths. It grants no object
		// authority: the exact registration below must still resolve and be active.
		if strings.Contains("/"+strings.TrimPrefix(parsed.Path, "/"), "/"+storage.BusinessDraftMediaObjectPrefix+"/") {
			references = append(references, section.Image.URL)
			imageReferences[section.Image.URL] = true
		}
	}
	if parsed, err := url.Parse(policy.ContactURL); err == nil && strings.Contains("/"+strings.TrimPrefix(parsed.Path, "/"), "/"+storage.BusinessDraftMediaObjectPrefix+"/") {
		references = append(references, policy.ContactURL)
	}
	if len(references) == 0 {
		return nil
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, nil, references, nil)
	if err != nil {
		return err
	}
	for _, reference := range references {
		if resolved.identities[reference].AssetID == 0 {
			return fmt.Errorf("%w: draft policy snapshot media is unavailable", ErrProductMediaReferenceInvalid)
		}
		unchangedSnapshotReference := reference
		expectedMediaType := ""
		if imageReferences[reference] {
			expectedMediaType = "image"
		}
		if err := resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(publicBusinessMediaReferenceBinding{URL: &unchangedSnapshotReference, ExpectedMediaType: expectedMediaType}); err != nil {
			return err
		}
	}
	return nil
}
