package service

import (
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
)

func normalizeMediaAssetListBusinessPartitionAndAccessScope(input MediaAssetListInput) (repository.MediaAssetFilter, error) {
	partition := strings.TrimSpace(input.LibraryPartition)
	if partition != "" && partition != "all" && partition != "operational" && !media.MediaAssetLibraryPartition(partition).IsValidMediaLibraryPartition() {
		return repository.MediaAssetFilter{}, ErrUnsupportedMediaLibraryClassification
	}
	purpose := strings.TrimSpace(input.CustomerServiceMediaPurpose)
	if purpose != "" && purpose != media.CustomerServiceMediaPurposeConversationAttachment && purpose != media.CustomerServiceMediaPurposeReusableReply {
		return repository.MediaAssetFilter{}, ErrUnsupportedMediaLibraryClassification
	}
	if input.ExcludeConversationAttachments && purpose == media.CustomerServiceMediaPurposeConversationAttachment {
		return repository.MediaAssetFilter{}, ErrMediaAssetForbidden
	}
	mediaType, err := normalizeOptionalMediaType(input.MediaType)
	if err != nil {
		return repository.MediaAssetFilter{}, err
	}
	status, err := normalizeOptionalMediaStatus(input.Status)
	if err != nil {
		return repository.MediaAssetFilter{}, err
	}
	visibility, err := normalizeOptionalVisibility(input.Visibility)
	if err != nil {
		return repository.MediaAssetFilter{}, err
	}
	return repository.MediaAssetFilter{LibraryPartition: partition, UploadOrigin: strings.TrimSpace(input.UploadOrigin), CustomerServiceMediaPurpose: purpose,
		MediaType: mediaType, Status: status, Visibility: visibility, Search: strings.TrimSpace(input.Search),
		ExcludeConversationAttachments: input.ExcludeConversationAttachments, ReusableAssetsOnly: input.ReusableAssetsOnly}, nil
}

func (s *MediaService) CountMediaAssetsByLibraryPartition(input MediaAssetListInput) (map[string]int64, error) {
	filter, err := normalizeMediaAssetListBusinessPartitionAndAccessScope(input)
	if err != nil {
		return nil, err
	}
	return s.repo.CountMediaAssetsByLibraryPartitionWithinCurrentPermissionScope(filter)
}
