package repository

import (
	"fmt"
	"strings"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/ticket"
)

func (r *MediaRepository) ticketStructuredMetadataReferences(query mediaAssetReferenceQuery, table, resourceType string) ([]media.AssetReference, error) {
	references := make([]media.AssetReference, 0)
	if !r.hasTable(table) || !r.db.Migrator().HasColumn(table, "metadata") {
		return references, nil
	}
	query = r.includeProviderResolvedLegacyMediaReferenceObjectKeys(query)
	columns := "id, metadata"
	hasRuleActiveState := table == "ticket_auto_replies" && r.db.Migrator().HasColumn(table, "is_active")
	if hasRuleActiveState {
		columns += ", is_active"
	}
	hasMessageParent := table == "ticket_messages" && r.db.Migrator().HasColumn(table, "ticket_id")
	if hasMessageParent {
		columns += ", ticket_id AS parent_id"
	}
	if r.db.Migrator().HasColumn(table, "message_type") {
		columns += ", message_type"
	}
	for afterID := uint(0); ; {
		var rows []struct {
			ID, ParentID          uint
			Metadata, MessageType string
			IsActive              bool
		}
		if err := r.db.Table(table).Select(columns).Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			document, err := ticket.ParseStructuredMessageMetadataKnownMediaFields(row.MessageType, row.Metadata)
			if err != nil {
				return nil, fmt.Errorf("inspect retained %s.metadata #%d: invalid structured media fields", table, row.ID)
			}
			var fields []string
			for _, field := range document.Fields {
				if r.matchesExactRegisteredOrProviderResolvedLegacyMediaReference(query, field.URL) {
					fields = append(fields, "metadata."+field.Path)
				}
			}
			if len(fields) == 0 {
				continue
			}
			label := "客服自动回复媒体"
			if table == "ticket_messages" {
				label = "工单消息媒体"
			}
			reference := newMediaReference(media.ReferenceCategorySupport, resourceType, row.ID, row.ParentID, fmt.Sprintf("%s #%d", label, row.ID), strings.Join(fields, ", "))
			reference.UsageState = "active"
			if hasRuleActiveState && !row.IsActive {
				reference.UsageState = "disabled"
			}
			references = append(references, reference)
		}
		if len(rows) < 256 {
			if hasMessageParent {
				return r.applyCustomerServiceConversationParentRetentionStates(references)
			}
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}

func mergeTicketAttachmentAndStructuredMetadataReferences(attachments, metadata []media.AssetReference) []media.AssetReference {
	positions := make(map[uint]int, len(attachments))
	for index, reference := range attachments {
		positions[reference.ResourceID] = index
	}
	for _, reference := range metadata {
		if index, exists := positions[reference.ResourceID]; exists {
			attachments[index].Field += ", " + reference.Field
			attachments[index].UsageState = reference.UsageState
		} else {
			attachments = append(attachments, reference)
		}
	}
	return attachments
}
