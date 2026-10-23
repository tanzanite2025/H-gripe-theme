package repository

import "commerce-platform/internal/domain/media"

func (r *MediaRepository) warrantyClaimReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	videoColumn := ""
	// The original formal warranty schema has images only. Scan video_url
	// whenever the later application-created column is actually present.
	if r.hasTable("warranty_claims") && r.db.Migrator().HasColumn("warranty_claims", "video_url") {
		videoColumn = "video_url"
	}
	return r.scanDecodedLegacyJSONAndDirectMediaReferenceRows(query, legacyJSONAndDirectMediaReferenceSource{
		Table: "warranty_claims", ResourceType: "warranty_claim", Category: media.ReferenceCategoryCustomer,
		JSONColumn: "images", DirectMediaColumn: videoColumn, ParentColumn: "0", HasSoftDeletion: true,
	})
}

func (r *MediaRepository) suggestionFeedbackReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	return r.scanDecodedLegacyJSONAndDirectMediaReferenceRows(query, legacyJSONAndDirectMediaReferenceSource{
		Table: "suggestion_feedback", ResourceType: "suggestion_feedback", Category: media.ReferenceCategoryCustomer,
		JSONColumn: "attachments", ParentColumn: "0", HasSoftDeletion: true,
	})
}

func (r *MediaRepository) ticketAutoReplyReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	var attachments []media.AssetReference
	var err error
	if r.hasTable("ticket_auto_replies") && r.db.Migrator().HasColumn("ticket_auto_replies", "attachments") {
		attachments, err = r.scanDecodedLegacyJSONAndDirectMediaReferenceRows(query, legacyJSONAndDirectMediaReferenceSource{
			Table: "ticket_auto_replies", ResourceType: "ticket_auto_reply", Category: media.ReferenceCategorySupport,
			JSONColumn: "attachments", ParentColumn: "0",
		})
	}
	if err != nil {
		return nil, err
	}
	metadata, err := r.ticketStructuredMetadataReferences(query, "ticket_auto_replies", "ticket_auto_reply")
	return mergeTicketAttachmentAndStructuredMetadataReferences(attachments, metadata), err
}

func (r *MediaRepository) ticketMessageReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	var attachments []media.AssetReference
	var err error
	if r.hasTable("ticket_messages") && r.db.Migrator().HasColumn("ticket_messages", "attachments") {
		parentColumn := "0"
		if r.db.Migrator().HasColumn("ticket_messages", "ticket_id") {
			parentColumn = "ticket_id"
		}
		attachments, err = r.scanDecodedLegacyJSONAndDirectMediaReferenceRows(query, legacyJSONAndDirectMediaReferenceSource{
			Table: "ticket_messages", ResourceType: "ticket_message", Category: media.ReferenceCategorySupport,
			JSONColumn: "attachments", ParentColumn: parentColumn,
		})
	}
	if err != nil {
		return nil, err
	}
	if r.hasTable("ticket_messages") && r.db.Migrator().HasColumn("ticket_messages", "ticket_id") {
		attachments, err = r.applyCustomerServiceConversationParentRetentionStates(attachments)
		if err != nil {
			return nil, err
		}
	}
	metadata, err := r.ticketStructuredMetadataReferences(query, "ticket_messages", "ticket_message")
	return mergeTicketAttachmentAndStructuredMetadataReferences(attachments, metadata), err
}
