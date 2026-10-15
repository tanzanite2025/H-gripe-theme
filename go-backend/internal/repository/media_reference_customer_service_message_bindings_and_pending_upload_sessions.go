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
		ID       uint
		TicketID uint
	}
	var rows []bindingReferenceRow
	err := r.db.Table("ticket_message_media_attachments AS bindings").Select("messages.id,messages.ticket_id").
		Joins("JOIN ticket_messages AS messages ON messages.id=bindings.ticket_message_id").
		Where("bindings.media_asset_id = ?", query.AssetID).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	references := make([]media.AssetReference, 0, len(rows))
	for _, row := range rows {
		references = append(references, newMediaReference(media.ReferenceCategorySupport, "ticket_message", row.ID, row.TicketID,
			fmt.Sprintf("工单 #%d 的消息 #%d", row.TicketID, row.ID), "media_asset_bindings"))
	}
	return references, nil
}

func (r *MediaRepository) customerServicePendingUploadSessionAssetReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("media_upload_sessions") {
		return nil, nil
	}
	type sessionReferenceRow struct{ BusinessContextID uint }
	var rows []sessionReferenceRow
	scan := r.db.Table("media_upload_sessions AS uploads").Select("uploads.business_context_id").
		Where("uploads.media_asset_id = ? AND uploads.business_context_type = ?", query.AssetID, media.MediaUploadSessionCustomerServiceContext)
	if r.hasTable("ticket_message_media_attachments") {
		scan = scan.Where("NOT EXISTS (SELECT 1 FROM ticket_message_media_attachments AS bindings WHERE bindings.upload_session_id=uploads.id)")
	}
	if err := scan.Find(&rows).Error; err != nil {
		return nil, err
	}
	references := make([]media.AssetReference, 0, len(rows))
	for _, row := range rows {
		references = append(references, newMediaReference(media.ReferenceCategorySupport, "media_upload_session", query.AssetID, row.BusinessContextID,
			fmt.Sprintf("工单 #%d 的待发送附件", row.BusinessContextID), "pending_upload"))
	}
	return references, nil
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
