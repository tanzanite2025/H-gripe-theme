package media

import "strings"

// IsCustomerServiceConversationOwnedMediaAsset also recognizes historical keys
// and cloud URLs, which may predate the explicit source and purpose columns.
func IsCustomerServiceConversationOwnedMediaAsset(asset *MediaAsset) bool {
	if asset == nil {
		return false
	}
	key := strings.Trim(strings.ReplaceAll(strings.TrimSpace(asset.StorageKey), "\\", "/"), "/")
	address := strings.ReplaceAll(asset.URL, "\\", "/")
	return asset.RequiresConversationAccess || asset.CustomerServiceMediaPurpose == CustomerServiceMediaPurposeConversationAttachment ||
		asset.UploadOrigin == MediaUploadOriginCustomerServiceConversation ||
		strings.HasPrefix(key, "customer-service/attachments/originals/") ||
		strings.HasPrefix(key, "customer-service/public-conversation-attachments/") ||
		strings.Contains("/"+strings.TrimLeft(address, "/"), "/customer-service/attachments/originals/") ||
		strings.Contains("/"+strings.TrimLeft(address, "/"), "/customer-service/public-conversation-attachments/")
}

// Historical reference checks must not overwrite an audited public ownership
// decision or invent upload provenance from a file's present use.
func IsUnclassifiedHistoricalMediaAssetWithUnknownUploadProvenance(asset *MediaAsset) bool {
	return asset != nil && (asset.UploadOrigin == "" || asset.UploadOrigin == MediaUploadOriginLegacyUnknown) &&
		(asset.LibraryPartition == "" || asset.LibraryPartition == MediaLibraryPartitionUnclassified) &&
		(asset.ClassificationMethod == "" || asset.ClassificationMethod == MediaClassificationHistoricalUnknown) &&
		asset.ClassificationRevision == 0 &&
		(asset.UploaderActorType == "" || asset.UploaderActorType == MediaUploadActorLegacyAccount)
}

// CanMediaAssetBeReusedByPublicBusinessContent does not infer upload provenance
// or require catalog ownership: public operational files can be shared by domains.
func CanMediaAssetBeReusedByPublicBusinessContent(asset *MediaAsset) bool {
	return asset != nil && !asset.DeletedAt.Valid && asset.Status == "active" && asset.Visibility == "public" &&
		!IsCustomerServiceConversationOwnedMediaAsset(asset) &&
		asset.LibraryPartition != MediaLibraryPartitionCustomerSubmissions &&
		asset.UploaderActorType != MediaUploadActorCustomer && asset.UploaderActorType != MediaUploadActorVisitor
}
