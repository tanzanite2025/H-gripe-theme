package repository

import "commerce-platform/internal/domain/media"

func (r *MediaRepository) showcaseReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	return r.scanDecodedLegacyJSONAndDirectMediaReferenceRows(query, legacyJSONAndDirectMediaReferenceSource{
		Table: "showcases", ResourceType: "showcase", Category: media.ReferenceCategoryCommunity,
		JSONColumn: "images", ParentColumn: "0", HasTitle: true,
	})
}

func (r *MediaRepository) reviewReferences(query mediaAssetReferenceQuery) ([]media.AssetReference, error) {
	return r.scanDecodedLegacyJSONAndDirectMediaReferenceRows(query, legacyJSONAndDirectMediaReferenceSource{
		Table: "reviews", ResourceType: "review", Category: media.ReferenceCategoryCommunity,
		JSONColumn: "images", ParentColumn: "product_id", HasSoftDeletion: true,
	})
}
