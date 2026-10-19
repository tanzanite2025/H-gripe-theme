package repository

import (
	"commerce-platform/internal/domain/media"
)

// Include detached and superseded evidence. These records retain bytes and
// must never grant public access, even if a shared registration collides.
func (r *MediaRepository) orderEvidenceImmutableAttachmentAndSnapshotReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	query = r.includeProviderResolvedLegacyMediaReferenceObjectKeys(query)
	var references []media.AssetReference
	for _, source := range []struct {
		table, column string
		isJSON        bool
	}{
		{"order_evidence_attachments", "storage_key", false},
		{"order_evidence_items", "data_json", true},
		{"order_evidence_snapshots", "snapshot_data", true},
		{"order_evidence_submission_snapshots", "snapshot_data", true},
		{"order_evidence_export_snapshots", "snapshot_data", true},
		{"order_policy_disclosures", "policy_json", true},
		{"paypal_disputes", "evidence_submission_payload", true},
	} {
		if !r.hasTable(source.table) {
			continue
		}
		for afterID := uint(0); ; {
			var rows []struct {
				ID    uint
				Value string
			}
			// Table avoids GORM's active-placement scope, including detached_at.
			if err := r.db.Table(source.table).Select("id, "+source.column+" AS value").Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
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
					references = append(references, newMediaReference(media.ReferenceCategoryCustomer, source.table, row.ID, 0, "订单凭证不可变历史", source.column))
				}
			}
			if len(rows) < 256 {
				break
			}
			afterID = rows[len(rows)-1].ID
		}
	}
	return references, nil
}
