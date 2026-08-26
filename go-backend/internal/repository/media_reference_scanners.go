package repository

import "commerce-platform/internal/domain/media"

type mediaAssetReferenceScanner func(mediaAssetReferenceQuery) ([]media.AssetReference, error)

// assetReferenceScanners is the extension point for persisted domains that
// can consume a media asset. New domains add a scanner in their own file and
// register it here without changing deletion orchestration.
func (r *MediaRepository) assetReferenceScanners() []mediaAssetReferenceScanner {
	return []mediaAssetReferenceScanner{
		r.productCategoryImageReferences,
		r.productMediaReferences,
		r.galleryReferences,
		r.giftCardReferences,
		r.faqReferences,
		r.postReferences,
		r.showcaseReferences,
		r.reviewReferences,
		r.warrantyClaimReferences,
		r.suggestionFeedbackReferences,
		r.ticketAutoReplyReferences,
		r.ticketMessageReferences,
	}
}
