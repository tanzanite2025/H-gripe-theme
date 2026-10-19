package repository

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
)

type PublicBusinessMediaReferenceObjectKeyResolver interface {
	ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference string) (string, error)
}

// This is a transaction-local validation plan, never another reference ledger.
// The owning business table remains the authority for the persisted placement.
type publicBusinessMediaReferenceBinding struct {
	URL                      *string
	AssetID                  **uint
	ExpectedMediaType        string
	PreviousURL              string
	PreviousAssetID          *uint
	RetainsExistingPlacement bool
}

type lockedPublicBusinessMediaReferenceIdentities struct {
	tx                 *gorm.DB
	identities         map[string]registeredBusinessMediaObjectIdentity
	keys               map[string]string
	assets             map[uint]*media.MediaAsset
	assetURLObjectKeys map[uint]string
}

func lockAndResolvePublicBusinessMediaReferenceIdentities(tx *gorm.DB, resolver PublicBusinessMediaReferenceObjectKeyResolver, references []string, requestedAssetIDs []uint) (*lockedPublicBusinessMediaReferenceIdentities, error) {
	return lockAndResolvePublicBusinessMediaReferenceIdentitiesWithAdditionalPrivateAttachmentLocks(tx, resolver, references, requestedAssetIDs, nil)
}

// Additional private IDs join the same lock order without granting public reuse.
// The owning conversation binding protocol validates their actor/session policy.
func lockAndResolvePublicBusinessMediaReferenceIdentitiesWithAdditionalPrivateAttachmentLocks(tx *gorm.DB, resolver PublicBusinessMediaReferenceObjectKeyResolver, references []string, requestedAssetIDs, privateAssetIDs []uint) (*lockedPublicBusinessMediaReferenceIdentities, error) {
	resolved := &lockedPublicBusinessMediaReferenceIdentities{
		tx:         tx,
		identities: make(map[string]registeredBusinessMediaObjectIdentity),
		keys:       make(map[string]string), assets: make(map[uint]*media.MediaAsset),
		assetURLObjectKeys: make(map[uint]string),
	}
	if len(references) == 0 && len(requestedAssetIDs) == 0 && len(privateAssetIDs) == 0 {
		return resolved, nil
	}
	repo := NewMediaRepository(tx)
	repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(resolver)
	hasAssets := repo.hasTable("media_assets")
	hasDerivatives := hasAssets && repo.hasAssetDerivativesTable()
	if !hasAssets && resolver != nil {
		return nil, fmt.Errorf("%w: registered media storage is unavailable", ErrProductMediaReferenceInvalid)
	}
	assetIDs := make(map[uint]struct{}, len(requestedAssetIDs))
	for _, id := range requestedAssetIDs {
		if id == 0 {
			return nil, fmt.Errorf("%w: media asset ID must be positive", ErrProductMediaReferenceInvalid)
		}
		assetIDs[id] = struct{}{}
	}
	for _, reference := range references {
		if strings.TrimSpace(reference) == "" {
			continue
		}
		if _, exists := resolved.identities[reference]; exists {
			continue
		}
		key, err := storage.TrustedBusinessMediaReferenceObjectKey(reference, nil, "")
		if err == nil && resolver != nil {
			key, err = resolver.ResolveTrustedMediaObjectKeyForPublicBusinessReference(reference)
		}
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrProductMediaReferenceInvalid, err)
		}
		if storage.IsPrivateObjectKey(key) || media.IsCustomerServiceConversationOwnedMediaAsset(&media.MediaAsset{StorageKey: key, URL: reference}) {
			return nil, fmt.Errorf("%w: private or conversation files cannot be published", ErrProductMediaReferenceInvalid)
		}
		identity, err := repo.resolveRegisteredBusinessMediaObjectIdentity(reference, key, hasAssets, hasDerivatives)
		if err != nil {
			return nil, mapRegisteredMediaIdentityResolutionErrorToInvalidProductMediaReference(err)
		}
		resolved.identities[reference], resolved.keys[reference] = identity, key
		if identity.AssetID != 0 {
			assetIDs[identity.AssetID] = struct{}{}
		}
	}
	if len(assetIDs) > 0 && !hasAssets {
		return nil, fmt.Errorf("%w: registered media storage is unavailable", ErrProductMediaReferenceInvalid)
	}
	publicAssetIDs := make(map[uint]bool, len(assetIDs))
	for id := range assetIDs {
		publicAssetIDs[id] = true
	}
	for _, id := range privateAssetIDs {
		if id == 0 {
			return nil, ErrCustomerServiceAttachmentBindingUnavailable
		}
		assetIDs[id] = struct{}{}
	}
	ordered := make([]uint, 0, len(assetIDs))
	for id := range assetIDs {
		ordered = append(ordered, id)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i] < ordered[j] })
	for _, id := range ordered {
		asset, err := repo.FindAssetByIDForUpdate(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: media asset is missing or retired; reload media selection", ErrProductMediaReferenceInvalid)
		}
		if err != nil {
			return nil, err
		}
		if !publicAssetIDs[id] {
			continue
		}
		asset.RequiresConversationAccess, err = repo.HistoricalUnclassifiedMediaRequiresConversationAccess(asset)
		if err != nil {
			return nil, err
		}
		if !media.CanMediaAssetBeReusedByPublicBusinessContent(asset) || storage.IsPrivateObjectKey(asset.StorageKey) || strings.TrimSpace(asset.URL) == "" {
			return nil, fmt.Errorf("%w: selected asset is not eligible for public business reuse", ErrProductMediaReferenceInvalid)
		}
		resolved.assets[id] = asset
		// ID-only fields can select historical registrations without StorageKey.
		// Their URL belongs to a dedicated namespace only if its origin is trusted.
		key, keyErr := storage.TrustedBusinessMediaReferenceObjectKey(asset.URL, nil, "")
		if keyErr == nil && resolver != nil {
			key, keyErr = resolver.ResolveTrustedMediaObjectKeyForPublicBusinessReference(asset.URL)
		}
		if keyErr == nil {
			resolved.assetURLObjectKeys[id] = key
		}
	}
	for reference, identity := range resolved.identities {
		current, err := repo.resolveRegisteredBusinessMediaObjectIdentity(reference, resolved.keys[reference], hasAssets, hasDerivatives)
		if err != nil {
			return nil, mapRegisteredMediaIdentityResolutionErrorToInvalidProductMediaReference(err)
		}
		if current != identity || current.Retired {
			return nil, fmt.Errorf("%w: media identity changed or retired; reload media selection", ErrProductMediaReferenceInvalid)
		}
	}
	return resolved, nil
}

func (resolved *lockedPublicBusinessMediaReferenceIdentities) validateAndCanonicalizePublicBusinessMediaReferenceBinding(binding publicBusinessMediaReferenceBinding) error {
	if binding.URL == nil {
		return gorm.ErrInvalidData
	}
	reference := *binding.URL
	identity := resolved.identities[reference]
	var requestedID *uint
	if binding.AssetID != nil {
		requestedID = *binding.AssetID
	}
	preservePreviousAlias := false
	if requestedID != nil && identity.AssetID != *requestedID {
		if strings.TrimSpace(reference) == "" {
			identity.AssetID = *requestedID // ID-only selection; URL comes from the locked asset.
		} else if identity.AssetID == 0 && resolved.keys[reference] == "" && binding.RetainsExistingPlacement && binding.PreviousAssetID != nil && *binding.PreviousAssetID == *requestedID && binding.PreviousURL == reference {
			identity.AssetID, preservePreviousAlias = *requestedID, true
		} else {
			return fmt.Errorf("%w: media asset ID and URL identify different files", ErrProductMediaReferenceInvalid)
		}
	}
	if resolved.keys[reference] != "" && identity.AssetID == 0 && (!binding.RetainsExistingPlacement || reference != binding.PreviousURL) {
		return fmt.Errorf("%w: managed media file is not registered; select or upload an active asset", ErrProductMediaReferenceInvalid)
	}
	dedicatedNamespace := resolved.dedicatedObjectNamespaceForResolvedPublicMediaReference(reference, identity.AssetID)
	if dedicatedNamespace != "" {
		if !binding.RetainsExistingPlacement || binding.PreviousURL != reference || (requestedID != nil && (binding.PreviousAssetID == nil || *binding.PreviousAssetID != *requestedID)) {
			return fmt.Errorf("%w: %s objects belong to their dedicated lifecycle and cannot be newly attached through a media registration", ErrProductMediaReferenceInvalid, dedicatedNamespace)
		}
		if asset := resolved.assets[identity.AssetID]; asset != nil && binding.ExpectedMediaType != "" && asset.MediaType != binding.ExpectedMediaType {
			return fmt.Errorf("%w: media type does not match the business field", ErrProductMediaReferenceInvalid)
		}
		return nil // Keep the exact historical placement without canonicalizing it into a new alias.
	}
	if identity.AssetID == 0 {
		return nil
	}
	asset := resolved.assets[identity.AssetID]
	if asset == nil {
		return fmt.Errorf("%w: media identity was not locked", ErrProductMediaReferenceInvalid)
	}
	if binding.ExpectedMediaType != "" && asset.MediaType != binding.ExpectedMediaType {
		return fmt.Errorf("%w: media type does not match the business field", ErrProductMediaReferenceInvalid)
	}
	if !preservePreviousAlias {
		*binding.URL = asset.URL
		// Client-side fragments such as PDF #page=2 and video #t=10 retain
		// their meaning when a configured storage alias becomes canonical.
		if original, err := url.Parse(reference); err == nil && original.Fragment != "" {
			if canonical, err := url.Parse(asset.URL); err == nil {
				canonical.Fragment, canonical.RawFragment = original.Fragment, original.RawFragment
				*binding.URL = canonical.String()
			}
		}
	}
	if binding.AssetID != nil {
		id := asset.ID
		*binding.AssetID = &id
	}
	return NewMediaRepository(resolved.tx).BindBusinessDraftMediaUploadReceiptWithinTransaction(asset)
}

func (resolved *lockedPublicBusinessMediaReferenceIdentities) dedicatedObjectNamespaceForResolvedPublicMediaReference(reference string, assetID uint) string {
	if namespace := dedicatedSiteIdentityOrVisualShowcaseNamespaceForObjectKey(resolved.keys[reference]); namespace != "" {
		return namespace
	}
	if asset := resolved.assets[assetID]; asset != nil {
		if namespace := dedicatedSiteIdentityOrVisualShowcaseNamespaceForObjectKey(asset.StorageKey); namespace != "" {
			return namespace
		}
	}
	return dedicatedSiteIdentityOrVisualShowcaseNamespaceForObjectKey(resolved.assetURLObjectKeys[assetID])
}
