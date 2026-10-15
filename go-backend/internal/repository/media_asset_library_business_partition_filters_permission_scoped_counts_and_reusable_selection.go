package repository

import (
	"strings"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
)

func (r *MediaRepository) mediaAssetLibraryListQueryWithBusinessPartitionAndAccessScope(filter MediaAssetFilter) *gorm.DB {
	query := r.db.Model(&media.MediaAsset{})
	if filter.LibraryPartition == "operational" {
		query = query.Where("library_partition IN ?", []media.MediaAssetLibraryPartition{media.MediaLibraryPartitionCatalog, media.MediaLibraryPartitionContent, media.MediaLibraryPartitionShared})
	} else if filter.LibraryPartition != "" && filter.LibraryPartition != "all" {
		query = query.Where("library_partition = ?", filter.LibraryPartition)
	}
	if filter.UploadOrigin != "" {
		query = query.Where("upload_origin = ?", filter.UploadOrigin)
	}
	if filter.CustomerServiceMediaPurpose != "" {
		query = query.Where("customer_service_media_purpose = ?", filter.CustomerServiceMediaPurpose)
	}
	if filter.ExcludeConversationAttachments || filter.ReusableAssetsOnly {
		// Cover new classified assets and previously registered private namespaces.
		query = query.Where("customer_service_media_purpose <> ? AND upload_origin <> ? AND REPLACE(COALESCE(storage_key, ''), '\\', '/') NOT LIKE ? AND REPLACE(COALESCE(storage_key, ''), '\\', '/') NOT LIKE ? AND REPLACE(COALESCE(url, ''), '\\', '/') NOT LIKE ? AND REPLACE(COALESCE(url, ''), '\\', '/') NOT LIKE ?",
			media.CustomerServiceMediaPurposeConversationAttachment, media.MediaUploadOriginCustomerServiceConversation,
			"customer-service/attachments/originals/%", "customer-service/public-conversation-attachments/%",
			"%/customer-service/attachments/originals/%", "%/customer-service/public-conversation-attachments/%")
	}
	if filter.ReusableAssetsOnly {
		query = query.Where("visibility = ? AND status = ? AND library_partition <> ?", "public", "active", media.MediaLibraryPartitionCustomerSubmissions)
		query = query.Where("uploader_actor_type NOT IN ?", []media.MediaUploadActorType{media.MediaUploadActorCustomer, media.MediaUploadActorVisitor})
		// Unknown historical files remain selectable for compatibility in this
		// stage. Verified backfill and authoritative business-save policies follow.
	}
	if filter.MediaType != "" {
		query = query.Where("media_type = ?", filter.MediaType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Visibility != "" {
		query = query.Where("visibility = ?", filter.Visibility)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + strings.ToLower(search) + "%"
		query = query.Where("LOWER(COALESCE(filename, '')) LIKE ? OR LOWER(COALESCE(original_filename, '')) LIKE ? OR LOWER(COALESCE(alt, '')) LIKE ? OR LOWER(COALESCE(caption, '')) LIKE ? OR LOWER(COALESCE(url, '')) LIKE ?", like, like, like, like, like)
	}
	return query
}

func (r *MediaRepository) CountMediaAssetsByLibraryPartitionWithinCurrentPermissionScope(filter MediaAssetFilter) (map[string]int64, error) {
	// Sidebar counts share every list predicate except the selected partition/purpose.
	filter.LibraryPartition, filter.CustomerServiceMediaPurpose = "", ""
	var rows []struct {
		LibraryPartition string
		Total            int64
	}
	err := r.mediaAssetLibraryListQueryWithBusinessPartitionAndAccessScope(filter).
		Select("library_partition, COUNT(*) AS total").Group("library_partition").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	counts := map[string]int64{"all": 0, "operational": 0}
	for _, partition := range []media.MediaAssetLibraryPartition{media.MediaLibraryPartitionCatalog, media.MediaLibraryPartitionContent,
		media.MediaLibraryPartitionShared, media.MediaLibraryPartitionCustomerService, media.MediaLibraryPartitionCustomerSubmissions, media.MediaLibraryPartitionUnclassified} {
		counts[string(partition)] = 0
	}
	for _, row := range rows {
		counts[row.LibraryPartition] = row.Total
		counts["all"] += row.Total
		if row.LibraryPartition == string(media.MediaLibraryPartitionCatalog) || row.LibraryPartition == string(media.MediaLibraryPartitionContent) || row.LibraryPartition == string(media.MediaLibraryPartitionShared) {
			counts["operational"] += row.Total
		}
	}
	return counts, nil
}
