package repository

import "commerce-platform/internal/domain/media"

type mediaAssetReferenceQuery struct {
	AssetID uint
	URLs    []string
}

// FindAssetReferences runs the media reference scanners registered for the
// current application schema. Each business domain owns its scanner so new
// source types do not change deletion orchestration.
func (r *MediaRepository) FindAssetReferences(asset *media.MediaAsset) ([]media.AssetReference, error) {
	query := newMediaAssetReferenceQuery(asset)
	if query.AssetID == 0 || len(query.URLs) == 0 {
		return []media.AssetReference{}, nil
	}

	references := make([]media.AssetReference, 0)
	for _, scan := range r.assetReferenceScanners() {
		items, err := scan(query)
		if err != nil {
			return nil, err
		}
		references = append(references, items...)
	}
	return references, nil
}
