package service

import "tanzanite/internal/domain/media"

func (s *MediaService) GetAssetReferences(id uint) (media.AssetReferenceReport, error) {
	asset, err := s.GetAsset(id)
	if err != nil {
		return media.AssetReferenceReport{}, err
	}

	references, err := s.repo.FindAssetReferences(asset)
	if err != nil {
		return media.AssetReferenceReport{}, err
	}
	return media.AssetReferenceReport{
		Total:      len(references),
		References: references,
	}, nil
}
