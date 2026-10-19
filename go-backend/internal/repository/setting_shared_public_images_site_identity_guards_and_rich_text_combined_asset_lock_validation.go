package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/refundcancellation"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/safehtml"
	"commerce-platform/internal/pkg/websitecontent"
	"gorm.io/gorm"
)

var ErrSettingMediaReferenceInvalid = errors.New("setting media reference is invalid")

func (r *SettingRepository) validateAndCanonicalizeSharedImageAndWebsiteProfileRichTextSettingsWithOneSortedAssetLockPlan(tx *gorm.DB, identities []settingKeyAndLocaleIdentity, previous, requested map[settingKeyAndLocaleIdentity]setting.Setting, newlyReserved map[settingKeyAndLocaleIdentity]bool) error {
	var bindings []publicBusinessMediaReferenceBinding
	imageReferences := make(map[settingKeyAndLocaleIdentity]*string)
	documents := make(map[settingKeyAndLocaleIdentity]*safehtml.RichTextMediaReferenceDocument)
	policies := make(map[settingKeyAndLocaleIdentity]*refundcancellation.Policy)
	for _, identity := range identities {
		value, old := requested[identity], previous[identity]
		retainsPlacement := value.Group == old.Group && (!value.IsPublic || old.IsPublic)
		if value.Value == old.Value && retainsPlacement && !newlyReserved[identity] {
			continue
		}
		if setting.IsDedicatedSiteIdentityOwnerSettingKey(value.Key) {
			if err := validateChangedSiteIdentitySettingUsesItsOwningLifecycle(tx, value); err != nil {
				return err
			}
		}
		if setting.IsSharedPublicMediaImageSettingKey(value.Key) {
			reference := value.Value
			bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &reference, ExpectedMediaType: "image"})
			imageReferences[identity] = &reference
			continue
		}
		if value.Key == refundcancellation.Key && (r.businessMediaReferenceResolver != nil || tx.Migrator().HasTable(&media.MediaAsset{})) {
			var policy, previousPolicy refundcancellation.Policy
			if err := json.Unmarshal([]byte(value.Value), &policy); err != nil {
				return fmt.Errorf("%w: invalid refund policy JSON", ErrSettingMediaReferenceInvalid)
			}
			if retainsPlacement && old.Value != "" {
				if err := json.Unmarshal([]byte(old.Value), &previousPolicy); err != nil {
					return fmt.Errorf("%w: invalid previous refund policy JSON", ErrSettingMediaReferenceInvalid)
				}
			}
			previousImages := make(map[string]string)
			for _, section := range previousPolicy.Sections {
				if section.Image != nil {
					previousImages[section.ID] = section.Image.URL
				}
			}
			for index := range policy.Sections {
				section := &policy.Sections[index]
				if section.Image == nil {
					continue
				}
				bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &section.Image.URL, ExpectedMediaType: "image", PreviousURL: previousImages[section.ID], RetainsExistingPlacement: retainsPlacement})
			}
			if policy.ContactURL != "" && !strings.HasPrefix(strings.ToLower(policy.ContactURL), "mailto:") {
				bindings = append(bindings, publicBusinessMediaReferenceBinding{URL: &policy.ContactURL, PreviousURL: previousPolicy.ContactURL, RetainsExistingPlacement: retainsPlacement})
			}
			policies[identity] = &policy
		}
		if setting.IsWebsiteProfileSharedMediaRichTextSettingKey(value.Key) {
			content, err := websitecontent.NormalizeWebsiteBody(value.Value)
			if err != nil {
				return err
			}
			previousContent := ""
			if retainsPlacement {
				previousContent, err = websitecontent.NormalizeWebsiteBody(old.Value)
				if err != nil {
					return err
				}
			}
			contentBindings, document, err := buildPublicBusinessRichTextMediaAttributeBindingsWithPlacementScopedLegacyCompatibility(content, previousContent)
			if err != nil {
				return err
			}
			bindings = append(bindings, contentBindings...)
			// Preserve historical plain paragraph storage and its existing merge
			// whitespace when no media attribute needs canonical rendering.
			if len(contentBindings) != 0 {
				documents[identity] = document
			}
		}
	}
	references := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		references = append(references, *binding.URL)
	}
	resolved, err := lockAndResolvePublicBusinessMediaReferenceIdentities(tx, r.businessMediaReferenceResolver, references, nil)
	if err != nil {
		return mapPublicBusinessMediaValidationErrorToSettingMediaError(err)
	}
	for _, binding := range bindings {
		if err := resolved.validateAndCanonicalizePublicBusinessMediaReferenceBinding(binding); err != nil {
			return mapPublicBusinessMediaValidationErrorToSettingMediaError(err)
		}
	}
	for _, identity := range identities {
		value := requested[identity]
		if reference := imageReferences[identity]; reference != nil {
			value.Value = *reference
		}
		if document := documents[identity]; document != nil {
			value.Value, err = document.RenderRichTextWithCanonicalMediaReferences()
			if err != nil {
				return err
			}
		}
		if policy := policies[identity]; policy != nil {
			encoded, marshalErr := json.Marshal(policy)
			if marshalErr != nil {
				return marshalErr
			}
			value.Value = string(encoded)
		}
		requested[identity] = value
	}
	return nil
}

func mapPublicBusinessMediaValidationErrorToSettingMediaError(err error) error {
	if errors.Is(err, ErrProductMediaReferenceInvalid) {
		return fmt.Errorf("%w: %v", ErrSettingMediaReferenceInvalid, err)
	}
	return err
}
