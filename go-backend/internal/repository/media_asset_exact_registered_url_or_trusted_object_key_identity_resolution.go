package repository

import (
	"errors"
	"net/url"
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
)

var ErrRegisteredMediaObjectOwnershipAmbiguous = errors.New("media address has ambiguous registered ownership")

// Exact registered legacy origins remain usable without trusting another host
// with the same /uploads suffix. Tombstones and duplicate identities fail closed.
func (r *MediaRepository) FindUniquelyRegisteredMediaObjectByExactURLOrTrustedKey(reference, trustedKey string) (*media.MediaAsset, *media.MediaAssetDerivative, error) {
	return r.findUniqueRegisteredMediaObjectWithExplicitRetiredDerivativeReadPolicy(reference, trustedKey, false)
}

// For existing consumers only. New bindings must continue using the strict
// method above. Active parents and unambiguous ownership remain mandatory.
func (r *MediaRepository) FindUniquelyRegisteredMediaObjectForExistingConsumerReadIncludingRetiredDerivative(reference, trustedKey string) (*media.MediaAsset, *media.MediaAssetDerivative, error) {
	return r.findUniqueRegisteredMediaObjectWithExplicitRetiredDerivativeReadPolicy(reference, trustedKey, true)
}

func (r *MediaRepository) findUniqueRegisteredMediaObjectWithExplicitRetiredDerivativeReadPolicy(reference, trustedKey string, allowRetiredDerivative bool) (*media.MediaAsset, *media.MediaAssetDerivative, error) {
	identity, err := r.resolveRegisteredBusinessMediaObjectIdentity(reference, trustedKey, r.hasTable("media_assets"), r.hasAssetDerivativesTable())
	if err != nil {
		return nil, nil, err
	}
	if identity.AssetID == 0 || (identity.Retired && (!allowRetiredDerivative || identity.DerivativeID == 0)) {
		return nil, nil, gorm.ErrRecordNotFound
	}
	asset, err := r.FindAssetByID(identity.AssetID)
	if err != nil {
		return nil, nil, err
	}
	if identity.DerivativeID == 0 {
		return asset, nil, nil
	}
	var derivative media.MediaAssetDerivative
	query := r.db
	if allowRetiredDerivative {
		query = query.Unscoped()
	}
	if err := query.First(&derivative, identity.DerivativeID).Error; err != nil {
		return nil, nil, err
	}
	return asset, &derivative, nil
}

type registeredBusinessMediaObjectIdentity struct {
	AssetID      uint
	DerivativeID uint
	URL          string
	Retired      bool
}

// Include tombstones so a retired file cannot fall through as an unmanaged URL.
// Never match a basename, a URL suffix, or the first of several possible owners.
func (r *MediaRepository) resolveRegisteredBusinessMediaObjectIdentity(reference, trustedKey string, hasAssetsTable, hasDerivativesTable bool) (registeredBusinessMediaObjectIdentity, error) {
	identity := registeredBusinessMediaObjectIdentity{}
	if !hasAssetsTable {
		return identity, nil
	}
	addresses := []string{strings.TrimSpace(reference)}
	if parsed, err := url.Parse(addresses[0]); err == nil {
		parsed.RawQuery, parsed.Fragment, parsed.RawFragment, parsed.ForceQuery = "", "", "", false
		addresses = append(addresses, parsed.String())
	}
	assetsQuery := r.db.Unscoped().Where("url IN ?", addresses)
	if trustedKey != "" {
		assetsQuery = r.db.Unscoped().Where("url IN ? OR REPLACE(storage_key, '\\', '/') = ?", addresses, trustedKey)
	}
	var assets []media.MediaAsset
	if err := assetsQuery.Select("id", "url", "deleted_at").Limit(2).Find(&assets).Error; err != nil {
		return identity, err
	}
	identities := make([]registeredBusinessMediaObjectIdentity, 0, len(assets)+2)
	for _, asset := range assets {
		identities = append(identities, registeredBusinessMediaObjectIdentity{AssetID: asset.ID, URL: asset.URL, Retired: asset.DeletedAt.Valid})
	}
	if hasDerivativesTable {
		derivativeQuery := r.db.Unscoped().Where("url IN ?", addresses)
		if trustedKey != "" {
			derivativeQuery = r.db.Unscoped().Where("url IN ? OR REPLACE(storage_key, '\\', '/') = ?", addresses, trustedKey)
		}
		var derivatives []media.MediaAssetDerivative
		if err := derivativeQuery.Select("id", "media_asset_id", "url", "deleted_at").Limit(2).Find(&derivatives).Error; err != nil {
			return identity, err
		}
		for _, derivative := range derivatives {
			identities = append(identities, registeredBusinessMediaObjectIdentity{AssetID: derivative.MediaAssetID, DerivativeID: derivative.ID, URL: derivative.URL, Retired: derivative.DeletedAt.Valid})
		}
	}
	if len(identities) > 1 {
		return identity, ErrRegisteredMediaObjectOwnershipAmbiguous
	}
	if len(identities) == 1 {
		return identities[0], nil
	}
	return identity, nil
}
