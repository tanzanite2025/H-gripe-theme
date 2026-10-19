package repository

import (
	"fmt"

	"commerce-platform/internal/domain/media"
)

// Legacy JSON placements have no asset ID. Inspect individual values rather
// than substrings, including retained rows and escaped or nested attachment URLs.
func (r *MediaRepository) dedicatedPublicObjectLegacyCommunityAndPrivateAttachmentReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	query = r.includeProviderResolvedLegacyMediaReferenceObjectKeys(query)
	var references []media.AssetReference
	for _, source := range []struct {
		table, resourceType, column, parentColumn string
		category                                  media.ReferenceCategory
		isJSON                                    bool
	}{
		{"showcases", "showcase", "images", "0", media.ReferenceCategoryCommunity, true},
		{"reviews", "review", "images", "product_id", media.ReferenceCategoryCommunity, true},
		{"warranty_claims", "warranty_claim", "images", "0", media.ReferenceCategoryCustomer, true},
		{"warranty_claims", "warranty_claim", "video_url", "0", media.ReferenceCategoryCustomer, false},
		{"suggestion_feedback", "suggestion_feedback", "attachments", "0", media.ReferenceCategoryCustomer, true},
	} {
		if !r.hasTable(source.table) {
			continue
		}
		for afterID := uint(0); ; {
			var rows []struct {
				ID, ParentID uint
				Value        string
			}
			if err := r.db.Table(source.table).Select("id, "+source.parentColumn+" AS parent_id, "+source.column+" AS value").Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
				return nil, err
			}
			for _, row := range rows {
				matches := r.matchesExactRegisteredOrProviderResolvedLegacyMediaReference(query, row.Value)
				if source.isJSON {
					var err error
					matches, err = r.decodedLegacyMediaJSONContainsExactReference(query, row.Value, source.table, source.column, row.ID)
					if err != nil {
						return nil, err
					}
				}
				if matches {
					references = append(references, newMediaReference(source.category, source.resourceType, row.ID, row.ParentID, fmt.Sprintf("历史引用 · %s #%d", source.resourceType, row.ID), source.column))
				}
			}
			if len(rows) < 256 {
				break
			}
			afterID = rows[len(rows)-1].ID
		}
	}
	for _, scan := range []mediaAssetReferenceScanner{r.ticketAutoReplyReferences, r.ticketMessageReferences} {
		items, err := scan(query)
		if err != nil {
			return nil, err
		}
		references = append(references, items...)
	}
	return references, nil
}
