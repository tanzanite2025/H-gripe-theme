package repository

import (
	"time"

	"commerce-platform/internal/domain/media"
)

// Parent state changes presentation only. Every returned reference continues to
// retain the object, including deleted conversations and missing message rows.
func (r *MediaRepository) applyCustomerServiceConversationParentRetentionStates(references []media.AssetReference) ([]media.AssetReference, error) {
	if len(references) == 0 || !r.hasTable("tickets") {
		return references, nil
	}
	columns := "id"
	if r.db.Migrator().HasColumn("tickets", "deleted_at") {
		columns += ", deleted_at"
	}
	for start := 0; start < len(references); start += 256 {
		end := min(start+256, len(references))
		ids := make([]uint, 0, end-start)
		for _, reference := range references[start:end] {
			if reference.UsageState != "missing_parent" {
				ids = append(ids, reference.ParentResourceID)
			}
		}
		if len(ids) == 0 {
			continue
		}
		var parents []struct {
			ID        uint
			DeletedAt *time.Time
		}
		// Table intentionally includes tombstones; absence and soft deletion
		// have different diagnosis and neither authorizes cleanup or migration.
		if err := r.db.Table("tickets").Select(columns).Where("id IN ?", ids).Find(&parents).Error; err != nil {
			return nil, err
		}
		states := make(map[uint]string, len(parents))
		for _, parent := range parents {
			states[parent.ID] = "active"
			if parent.DeletedAt != nil {
				states[parent.ID] = "retained_deleted_parent"
			}
		}
		for index := start; index < end; index++ {
			reference := &references[index]
			if reference.UsageState == "missing_parent" {
				continue
			}
			switch states[reference.ParentResourceID] {
			case "":
				reference.UsageState = "missing_parent"
			case "retained_deleted_parent":
				reference.UsageState = "retained_deleted_parent"
			default:
				if reference.UsageState == "" {
					reference.UsageState = "active"
				}
			}
		}
	}
	return references, nil
}
