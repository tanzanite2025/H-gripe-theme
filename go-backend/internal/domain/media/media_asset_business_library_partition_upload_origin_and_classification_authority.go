package media

// Library ownership is independent from visibility, storage keys and consumers.
type MediaAssetLibraryPartition string
type MediaAssetUploadOrigin string

const (
	MediaLibraryPartitionCatalog             MediaAssetLibraryPartition = "catalog"
	MediaLibraryPartitionContent             MediaAssetLibraryPartition = "content"
	MediaLibraryPartitionShared              MediaAssetLibraryPartition = "shared"
	MediaLibraryPartitionCustomerService     MediaAssetLibraryPartition = "customer_service"
	MediaLibraryPartitionCustomerSubmissions MediaAssetLibraryPartition = "customer_submissions"
	MediaLibraryPartitionUnclassified        MediaAssetLibraryPartition = "unclassified"

	MediaUploadOriginLegacyUnknown                        MediaAssetUploadOrigin = "legacy_unknown"
	MediaUploadOriginAdministratorGeneral                 MediaAssetUploadOrigin = "administrator_general_upload"
	MediaUploadOriginAdministratorCatalog                 MediaAssetUploadOrigin = "administrator_catalog_upload"
	MediaUploadOriginAdministratorContent                 MediaAssetUploadOrigin = "administrator_content_upload"
	MediaUploadOriginAdministratorWebsiteProfile          MediaAssetUploadOrigin = "administrator_website_profile_upload"
	MediaUploadOriginAdministratorShared                  MediaAssetUploadOrigin = "administrator_shared_upload"
	MediaUploadOriginAdministratorReusableCustomerService MediaAssetUploadOrigin = "administrator_customer_service_reusable_upload"
	MediaUploadOriginCustomerServiceConversation          MediaAssetUploadOrigin = "customer_service_conversation_upload"
	MediaUploadOriginAdministratorWorkbenchFeed           MediaAssetUploadOrigin = "administrator_workbench_feed_upload"

	MediaClassificationHistoricalUnknown              = "historical_unknown"
	MediaClassificationRecordedUploadContext          = "recorded_upload_context"
	MediaClassificationManualReview                   = "manual_reviewed"
	CustomerServiceMediaPurposeConversationAttachment = "conversation_attachment"
	CustomerServiceMediaPurposeReusableReply          = "reusable_reply"
)

func (partition MediaAssetLibraryPartition) IsValidMediaLibraryPartition() bool {
	switch partition {
	case MediaLibraryPartitionCatalog, MediaLibraryPartitionContent, MediaLibraryPartitionShared,
		MediaLibraryPartitionCustomerService, MediaLibraryPartitionCustomerSubmissions, MediaLibraryPartitionUnclassified:
		return true
	default:
		return false
	}
}
