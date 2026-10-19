package repository

import (
	"fmt"
	"time"

	"commerce-platform/internal/domain/media"
)

func (r *MediaRepository) workbenchFeedImageReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	if !r.hasTable("workbench_feed_media") {
		return references, nil
	}
	query = r.includeProviderResolvedLegacyMediaReferenceObjectKeys(query)
	hasParents := r.hasTable("workbench_feed_entries")
	hasPlacementDeletion := r.db.Migrator().HasColumn("workbench_feed_media", "deleted_at")
	for afterID := uint(0); ; {
		var rows []struct {
			ID, EntryID                uint
			FilePath, ParentStatus     string
			ExistingParentID           *uint
			DeletedAt, ParentDeletedAt *time.Time
		}
		columns := "placements.id, placements.entry_id, placements.file_path"
		scan := r.db.Table("workbench_feed_media AS placements")
		if hasPlacementDeletion {
			columns += ", placements.deleted_at"
		}
		if hasParents {
			columns += ", owners.id AS existing_parent_id, owners.deleted_at AS parent_deleted_at, owners.status AS parent_status"
			scan = scan.Joins("LEFT JOIN workbench_feed_entries AS owners ON owners.id = placements.entry_id")
		}
		if err := scan.Select(columns).Where("placements.id > ?", afterID).Order("placements.id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			if !r.matchesExactRegisteredOrProviderResolvedLegacyMediaReference(query, row.FilePath) {
				continue
			}
			reference := newMediaReference(media.ReferenceCategoryContent, "workbench_feed_media", row.ID, row.EntryID, fmt.Sprintf("工作台动态 #%d 的图片 #%d", row.EntryID, row.ID), "file_path")
			switch {
			case row.ExistingParentID == nil:
				reference.UsageState = "missing_parent"
			case row.ParentDeletedAt != nil:
				reference.UsageState = "retained_deleted_parent"
			case row.DeletedAt != nil:
				reference.UsageState = "retained_deleted_placement"
			case row.ParentStatus == "published":
				reference.UsageState = "active"
			case row.ParentStatus == "draft":
				reference.UsageState = "draft"
			case row.ParentStatus == "archived":
				reference.UsageState = "archived"
			default:
				reference.UsageState = "retained_unpublished_parent"
			}
			references = append(references, reference)
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}
