package repository

import (
	"fmt"

	"commerce-platform/internal/domain/media"
)

func (r *MediaRepository) customerServiceMessageMediaAssetBindingReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("ticket_message_media_attachments") {
		return nil, nil
	}
	type bindingReferenceRow struct {
		ID, MessageID, TicketID uint
		ExistingMessageID       *uint
	}
	var rows []bindingReferenceRow
	scan := r.db.Table("ticket_message_media_attachments AS bindings").
		Select("bindings.id, bindings.ticket_message_id AS message_id")
	if r.hasTable("ticket_messages") {
		scan = scan.Select("bindings.id, bindings.ticket_message_id AS message_id, messages.id AS existing_message_id, messages.ticket_id").
			Joins("LEFT JOIN ticket_messages AS messages ON messages.id=bindings.ticket_message_id")
	}
	err := scan.Where("bindings.media_asset_id = ?", query.AssetID).Order("bindings.id ASC").Find(&rows).Error
	if err != nil {
		return nil, err
	}
	references := make([]media.AssetReference, 0, len(rows))
	for _, row := range rows {
		if row.ExistingMessageID == nil {
			// The binding itself remains authoritative even when its message has
			// disappeared. Do not invent a conversation from a different receipt.
			reference := newMediaReference(media.ReferenceCategorySupport, "ticket_message_attachment_binding", row.ID, 0,
				fmt.Sprintf("附件绑定 #%d 的消息记录不存在", row.ID), "media_asset_id")
			reference.UsageState = "missing_parent"
			references = append(references, reference)
			continue
		}
		references = append(references, newMediaReference(media.ReferenceCategorySupport, "ticket_message", row.MessageID, row.TicketID,
			fmt.Sprintf("工单 #%d 的消息 #%d", row.TicketID, row.MessageID), "media_asset_bindings"))
	}
	return r.applyCustomerServiceConversationParentRetentionStates(references)
}

func (r *MediaRepository) customerServicePendingUploadSessionAssetReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("media_upload_sessions") {
		return nil, nil
	}
	query = r.includeProviderResolvedLegacyMediaReferenceObjectKeys(query)
	if query.IncludeAllExactObjectRegistrations {
		return r.customerServiceUploadReceiptExactObjectClaims(query)
	}
	type sessionReferenceRow struct {
		BusinessContextID uint
		Status            string
	}
	var rows []sessionReferenceRow
	scan := r.db.Table("media_upload_sessions AS uploads").Select("uploads.business_context_id, uploads.status").
		Where("uploads.media_asset_id = ? AND uploads.business_context_type = ?", query.AssetID, media.MediaUploadSessionCustomerServiceContext)
	if err := scan.Find(&rows).Error; err != nil {
		return nil, err
	}
	references := make([]media.AssetReference, 0, len(rows))
	for _, row := range rows {
		// A bound receipt also owns its bytes. A dangling or mismatched binding
		// must never hide it, and a bound upload is not available to send again.
		reference := newMediaReference(media.ReferenceCategorySupport, "media_upload_session", query.AssetID, row.BusinessContextID,
			fmt.Sprintf("工单 #%d 的上传收据保留", row.BusinessContextID), "upload_receipt")
		reference.UsageState = "pending_upload"
		if row.Status == "bound" {
			reference.UsageState = "retained_upload_receipt"
		}
		references = append(references, reference)
	}
	references, err := r.applyCustomerServiceConversationParentRetentionStates(references)
	if err != nil {
		return nil, err
	}
	if len(query.ObjectKeys) > 0 {
		claims, err := r.customerServiceUploadReceiptExactObjectClaims(query)
		if err != nil {
			return nil, err
		}
		references = append(references, claims...)
	}
	return references, nil
}

// Failed/unbound receipts can own bytes without any asset ID. Even expired or
// bound receipts retain ownership until their lifecycle removes them atomically.
func (r *MediaRepository) customerServiceUploadReceiptExactObjectClaims(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	var references []media.AssetReference
	for afterID := ""; ; {
		var rows []struct {
			ID, ObjectKey, BusinessContextType string
			BusinessContextID                  uint
			MediaAssetID                       *uint
		}
		if err := r.db.Table("media_upload_sessions").Select("id, object_key, business_context_type, business_context_id, media_asset_id").
			Where("id > ?", afterID).Order("id ASC").Limit(256).Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			if !query.IncludeAllExactObjectRegistrations && row.MediaAssetID != nil && *row.MediaAssetID == query.AssetID && row.BusinessContextType == media.MediaUploadSessionCustomerServiceContext {
				continue // Existing ID-based receipt/message scanners already cover this owner.
			}
			if registeredMediaReferenceQueryContainsExactObjectKey(query, row.ObjectKey) {
				references = append(references, newMediaReference(media.ReferenceCategorySupport, "media_upload_session", 0, row.BusinessContextID,
					fmt.Sprintf("工单 #%d 的上传收据仍占用存储对象", row.BusinessContextID), "object_key"))
			}
		}
		if len(rows) < 256 {
			return references, nil
		}
		afterID = rows[len(rows)-1].ID
	}
}

// Until the shared asset lifecycle can retire these relations atomically,
// retention must hold managed attachments before deleting projections/objects.
func (r *TicketRepository) HasCustomerServiceConversationManagedAttachmentLifecycleHold(ticketID uint) (bool, error) {
	if r.db.Migrator().HasTable("media_upload_sessions") {
		var count int64
		if err := r.db.Table("media_upload_sessions").Where("business_context_type = ? AND business_context_id = ?", media.MediaUploadSessionCustomerServiceContext, ticketID).Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	if r.db.Migrator().HasTable("ticket_message_media_attachments") {
		var count int64
		if err := r.db.Table("ticket_message_media_attachments AS bindings").Joins("JOIN ticket_messages AS messages ON messages.id=bindings.ticket_message_id").
			Where("messages.ticket_id = ?", ticketID).Count(&count).Error; err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}
	return false, nil
}
