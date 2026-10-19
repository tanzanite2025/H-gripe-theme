package repository

import "commerce-platform/internal/domain/media"

func (r *MediaRepository) privateShipmentAttachmentReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	return r.scanDecodedLegacyJSONAndDirectMediaReferenceRows(query, legacyJSONAndDirectMediaReferenceSource{
		Table: "shipment_records", ResourceType: "shipment_record", Category: media.ReferenceCategoryCustomer,
		JSONColumn: "shipping_images", ParentColumn: "order_id",
	})
}

func (r *MediaRepository) privateAfterSalesAttachmentReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	if !r.hasTable("after_sales_case_attachments") {
		return nil, nil
	}
	// Scan attachment rows independently of the case so retained or orphaned
	// historical ownership still blocks cleanup. This table has no JSON field.
	return r.scanDecodedLegacyJSONAndDirectMediaReferenceRows(query, legacyJSONAndDirectMediaReferenceSource{
		Table: "after_sales_case_attachments", ResourceType: "after_sales_attachment", Category: media.ReferenceCategoryCustomer,
		JSONColumn: "'[]'", DirectMediaColumn: "storage_url", ParentColumn: "case_id",
	})
}
