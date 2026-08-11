package service

import (
	"errors"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

type MediaAssetListInput struct {
	Page       int
	PageSize   int
	Search     string
	MediaType  string
	Status     string
	Visibility string
}

type MediaAssetUpdateInput struct {
	Alt        string
	Caption    string
	Status     string
	Visibility string
}

func (s *MediaService) ListAssets(input MediaAssetListInput) ([]media.MediaAsset, int64, error) {
	mediaType, err := normalizeOptionalMediaType(input.MediaType)
	if err != nil {
		return nil, 0, err
	}
	status, err := normalizeOptionalMediaStatus(input.Status)
	if err != nil {
		return nil, 0, err
	}
	visibility, err := normalizeOptionalVisibility(input.Visibility)
	if err != nil {
		return nil, 0, err
	}

	assets, total, err := s.repo.ListAssets(repository.MediaAssetFilter{
		Page:       input.Page,
		PageSize:   input.PageSize,
		Search:     strings.TrimSpace(input.Search),
		MediaType:  mediaType,
		Status:     status,
		Visibility: visibility,
	})
	if err != nil {
		return nil, 0, err
	}
	s.hydrateAssetListAccessURL(assets)
	return assets, total, nil
}

func (s *MediaService) GetAsset(id uint) (*media.MediaAsset, error) {
	asset, err := s.repo.FindAssetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrMediaAssetNotFound
		}
		return nil, err
	}
	s.hydrateAssetAccessURL(asset)
	return asset, nil
}

func (s *MediaService) UpdateAsset(id uint, input MediaAssetUpdateInput) (*media.MediaAsset, error) {
	asset, err := s.GetAsset(id)
	if err != nil {
		return nil, err
	}

	status, err := normalizeMediaStatus(input.Status)
	if err != nil {
		return nil, err
	}
	visibility, err := normalizeVisibility(input.Visibility)
	if err != nil {
		return nil, err
	}

	asset.Alt = strings.TrimSpace(input.Alt)
	asset.Caption = strings.TrimSpace(input.Caption)
	asset.Status = status
	asset.Visibility = visibility

	if err := s.repo.UpdateAsset(asset); err != nil {
		return nil, err
	}
	s.hydrateAssetAccessURL(asset)
	return asset, nil
}
