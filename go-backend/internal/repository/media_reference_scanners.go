package repository

import "commerce-platform/internal/domain/media"

type mediaAssetReferenceScanner func(mediaAssetReferenceQuery) ([]media.AssetReference, error)

// assetReferenceScanners is the extension point for persisted domains that
// can consume a media asset. New domains add a scanner in their own file and
// register it here without changing deletion orchestration.
func (r *MediaRepository) assetReferenceScanners() []mediaAssetReferenceScanner {
	return []mediaAssetReferenceScanner{
		r.productCategoryImageReferences,
		r.productBrandLogoReferences,
		r.productMediaReferences,
		r.productRichTextMediaAttributeReferences,
		r.productAndTemplateOptionSwatchReferences,
		r.galleryReferences,
		r.faqReferences,
		r.postReferences,
		r.workbenchFeedImageReferences,
		r.websiteProfileSettingMediaReferences,
		r.siteIdentitySettingMediaReferences,
		r.currentDedicatedSiteIdentityAndVisualShowcaseMediaReferences,
		r.customerServiceAgentAvatarReferences,
		r.showcaseReferences,
		r.reviewReferences,
		r.warrantyClaimReferences,
		r.privateShipmentAttachmentReferences,
		r.privateAfterSalesAttachmentReferences,
		r.suggestionFeedbackReferences,
		r.ticketAutoReplyReferences,
		r.ticketMessageReferences,
		r.customerServiceMessageMediaAssetBindingReferences,
		r.customerServicePendingUploadSessionAssetReferences,
		r.otherRegisteredMediaOriginalAndDerivativeObjectOwnershipReferences,
		r.orderEvidenceImmutableAttachmentAndSnapshotReferences,
	}
}
