package service

import (
	"errors"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/upload"
)

var ErrUnsupportedMediaLibraryClassification = errors.New("unsupported media library classification")

type trustedMediaAssetUploadClassification struct {
	partition              media.MediaAssetLibraryPartition
	origin                 media.MediaAssetUploadOrigin
	method                 string
	customerServicePurpose string
	originalPrefix         string
}

func resolveMediaAssetClassificationFromTrustedUploadOrigin(input MediaUploadInput, identity media.MediaUploadActorIdentity) (trustedMediaAssetUploadClassification, error) {
	origin := input.UploadOrigin
	if input.PreparedCustomerServiceUploadSessionID != "" {
		if origin != "" && origin != media.MediaUploadOriginCustomerServiceConversation {
			return trustedMediaAssetUploadClassification{}, ErrUnsupportedMediaLibraryClassification
		}
		origin = media.MediaUploadOriginCustomerServiceConversation
	}
	result := trustedMediaAssetUploadClassification{partition: media.MediaLibraryPartitionUnclassified,
		origin: origin, method: media.MediaClassificationRecordedUploadContext}
	switch origin {
	case "", media.MediaUploadOriginLegacyUnknown:
		result.origin, result.method = media.MediaUploadOriginLegacyUnknown, media.MediaClassificationHistoricalUnknown
		return result, nil
	case media.MediaUploadOriginAdministratorGeneral:
		// Compatibility endpoint does not infer ownership from client image_purpose.
	case media.MediaUploadOriginAdministratorCatalog:
		result.partition, result.originalPrefix = media.MediaLibraryPartitionCatalog, "catalog/product-images"
		if input.MediaType == "video" {
			result.originalPrefix = "catalog/product-videos"
		}
		switch input.UploadPurpose {
		case string(upload.SpecProductDescriptionImage):
			result.originalPrefix = "catalog/product-detail-media"
		case string(upload.SpecProductVariantSwatch):
			result.originalPrefix = "catalog/product-variant-swatches"
		case "", string(upload.SpecProductImage), string(upload.SpecMediaLibraryImage):
		default:
			return result, ErrUnsupportedMediaLibraryClassification
		}
	case media.MediaUploadOriginAdministratorContent, media.MediaUploadOriginAdministratorWebsiteProfile:
		result.partition, result.originalPrefix = media.MediaLibraryPartitionContent, "content/page-media"
	case media.MediaUploadOriginAdministratorShared:
		result.partition, result.originalPrefix = media.MediaLibraryPartitionShared, "shared/reusable-brand-and-instruction-media"
	case media.MediaUploadOriginAdministratorReusableCustomerService:
		result.partition, result.customerServicePurpose = media.MediaLibraryPartitionCustomerService, media.CustomerServiceMediaPurposeReusableReply
		result.originalPrefix = "customer-service/reusable-reply-media"
	case media.MediaUploadOriginCustomerServiceConversation:
		result.partition, result.customerServicePurpose = media.MediaLibraryPartitionCustomerService, media.CustomerServiceMediaPurposeConversationAttachment
		// Disabled private rollout keeps public compatibility but a distinct origin.
		result.originalPrefix = "customer-service/public-conversation-attachments"
		return result, nil
	case media.MediaUploadOriginAdministratorWorkbenchFeed:
		result.partition, result.originalPrefix = media.MediaLibraryPartitionContent, "workbench-feed"
	default:
		return result, ErrUnsupportedMediaLibraryClassification
	}
	if identity.ActorType != media.MediaUploadActorAdministrator {
		return result, ErrMediaAssetForbidden
	}
	return result, nil
}

func IsCustomerServiceConversationOwnedMediaAsset(asset *media.MediaAsset) bool {
	return media.IsCustomerServiceConversationOwnedMediaAsset(asset)
}
