package service

import (
	"errors"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"

	"gorm.io/gorm"
)

type MediaAssetListInput struct {
	LibraryPartition               string
	UploadOrigin                   string
	CustomerServiceMediaPurpose    string
	ExcludeConversationAttachments bool
	ReusableAssetsOnly             bool
	Page                           int
	PageSize                       int
	Search                         string
	MediaType                      string
	Status                         string
	Visibility                     string
}

type MediaAssetUpdateInput struct {
	Alt        string
	Caption    string
	Status     string
	Visibility string
}

func (s *MediaService) ListAssets(input MediaAssetListInput) ([]media.MediaAsset, int64, error) {
	filter, err := normalizeMediaAssetListBusinessPartitionAndAccessScope(input)
	if err != nil {
		return nil, 0, err
	}
	filter.Page, filter.PageSize = input.Page, input.PageSize
	assets, total, err := s.repo.ListAssets(filter)
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
	status, err := normalizeMediaStatus(input.Status)
	if err != nil {
		return nil, err
	}
	visibility, err := normalizeVisibility(input.Visibility)
	if err != nil {
		return nil, err
	}
	var asset *media.MediaAsset
	err = s.repo.WithTransaction(func(repo *repository.MediaRepository, _ *gorm.DB) error {
		lockedAsset, err := repo.FindAssetByIDForUpdate(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMediaAssetNotFound
		}
		if err != nil {
			return err
		}
		if IsCustomerServiceConversationOwnedMediaAsset(lockedAsset) {
			return ErrMediaAssetForbidden
		}
		lockedAsset.Alt = strings.TrimSpace(input.Alt)
		lockedAsset.Caption = strings.TrimSpace(input.Caption)
		lockedAsset.Status, lockedAsset.Visibility = status, visibility
		if err := repo.UpdateMediaAssetEditableMetadataWithoutReplacingLibraryClassification(lockedAsset); err != nil {
			return err
		}
		asset = lockedAsset
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.hydrateAssetAccessURL(asset)
	return asset, nil
}
