package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/safehtml"
)

type legacyJSONAndDirectMediaReferenceSource struct {
	Table, ResourceType, JSONColumn, DirectMediaColumn, ParentColumn string
	Category                                                         media.ReferenceCategory
	HasSoftDeletion, HasTitle                                        bool
}

type decodedLegacyMediaReferenceRow struct {
	ID, ParentID                  uint
	JSONValue, DirectValue, Title string
	RetainedDeleted               bool
}

// Escaped JSON and configured URL aliases cannot be safely prefiltered by a
// substring. Read only media fields in bounded pages, including retained rows.
func (r *MediaRepository) scanDecodedLegacyJSONAndDirectMediaReferenceRows(query mediaAssetReferenceQuery, source legacyJSONAndDirectMediaReferenceSource) ([]media.AssetReference, error) {
	if !r.hasTable(source.Table) {
		return []media.AssetReference{}, nil
	}
	query = r.includeProviderResolvedLegacyMediaReferenceObjectKeys(query)
	columns := []string{"id", source.ParentColumn + " AS parent_id", source.JSONColumn + " AS json_value"}
	if source.DirectMediaColumn != "" {
		columns = append(columns, source.DirectMediaColumn+" AS direct_value")
	}
	if source.HasSoftDeletion {
		columns = append(columns, "(deleted_at IS NOT NULL) AS retained_deleted")
	}
	if source.HasTitle {
		columns = append(columns, "title")
	}
	references := make([]media.AssetReference, 0)
	for afterID := uint(0); ; {
		var rows []decodedLegacyMediaReferenceRow
		if err := r.db.Table(source.Table).Select(strings.Join(columns, ", ")).
			Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			matchesJSON, err := r.decodedLegacyMediaJSONContainsExactReference(query, row.JSONValue, source.Table, source.JSONColumn, row.ID)
			if err != nil {
				return nil, err
			}
			fields := make([]string, 0, 2)
			if source.DirectMediaColumn != "" && r.matchesExactRegisteredOrProviderResolvedLegacyMediaReference(query, row.DirectValue) {
				fields = append(fields, source.DirectMediaColumn)
			}
			if matchesJSON {
				fields = append(fields, source.JSONColumn)
			}
			if len(fields) == 0 {
				continue
			}
			reference := newMediaReference(source.Category, source.ResourceType, row.ID, row.ParentID,
				legacyMediaReferenceLabelForSourceRow(source.ResourceType, row), strings.Join(fields, ", "))
			reference.UsageState = "active"
			if row.RetainedDeleted {
				reference.UsageState = "retained_deleted_parent"
			}
			references = append(references, reference)
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}

func legacyMediaReferenceLabelForSourceRow(resourceType string, row decodedLegacyMediaReferenceRow) string {
	switch resourceType {
	case "showcase":
		return namedMediaReferenceLabel("买家秀", row.ID, row.Title)
	case "review":
		return fmt.Sprintf("商品 #%d 的评价 #%d", row.ParentID, row.ID)
	case "warranty_claim":
		return fmt.Sprintf("保修申请 #%d", row.ID)
	case "suggestion_feedback":
		return fmt.Sprintf("建议反馈 #%d", row.ID)
	case "ticket_auto_reply":
		return fmt.Sprintf("客服自动回复 #%d", row.ID)
	case "ticket_message":
		return fmt.Sprintf("工单 #%d 的消息 #%d", row.ParentID, row.ID)
	default:
		return fmt.Sprintf("历史媒体引用 #%d", row.ID)
	}
}

func (r *MediaRepository) includeProviderResolvedLegacyMediaReferenceObjectKeys(query mediaAssetReferenceQuery) mediaAssetReferenceQuery {
	for _, reference := range query.URLs {
		if key := r.resolveRegisteredMediaObjectKeyForCleanupOwnership(reference); key != "" {
			query.ObjectKeys = append(query.ObjectKeys, key)
		}
	}
	return query
}

// Retention must recognize configured private addresses as well as public
// ones. This resolver does not authorize reuse or publish any attachment.
func (r *MediaRepository) matchesExactRegisteredOrProviderResolvedLegacyMediaReference(query mediaAssetReferenceQuery, reference string) bool {
	if r.matchesExactRegisteredOrTrustedMediaReference(query, reference) {
		return true
	}
	key := r.resolveRegisteredMediaObjectKeyForCleanupOwnership(reference)
	return key != "" && registeredMediaReferenceQueryContainsExactObjectKey(query, key)
}

func (r *MediaRepository) decodedLegacyMediaJSONContainsExactReference(query mediaAssetReferenceQuery, rawJSON, table, column string, recordID uint) (bool, error) {
	if strings.TrimSpace(rawJSON) == "" {
		return false, nil
	}
	var value any
	if err := json.Unmarshal([]byte(rawJSON), &value); err != nil {
		// Never include private attachment values in errors or cleanup logs.
		return false, fmt.Errorf("decode retained %s.%s #%d: invalid JSON", table, column, recordID)
	}
	return r.decodedLegacyMediaJSONValueContainsExactReference(query, value), nil
}

func (r *MediaRepository) decodedLegacyMediaJSONValueContainsExactReference(query mediaAssetReferenceQuery, value any) bool {
	switch value := value.(type) {
	case string:
		if r.matchesExactRegisteredOrProviderResolvedLegacyMediaReference(query, value) {
			return true
		}
		if strings.Contains(value, "<") {
			if document, err := safehtml.ParseRichTextMediaSourceAndFileLinkAttributes(value); err == nil {
				for _, reference := range document.References {
					if r.matchesExactRegisteredOrProviderResolvedLegacyMediaReference(query, *reference.URL) {
						return true
					}
				}
			}
		}
		return false
	case []any:
		for _, item := range value {
			if r.decodedLegacyMediaJSONValueContainsExactReference(query, item) {
				return true
			}
		}
	case map[string]any:
		for _, item := range value {
			if r.decodedLegacyMediaJSONValueContainsExactReference(query, item) {
				return true
			}
		}
	}
	return false
}
