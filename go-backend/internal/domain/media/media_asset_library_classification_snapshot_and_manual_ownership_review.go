package media

// Immutable source stays distinct from the current management choice and its
// revision. Audit records retain prior classification methods after review.
type MediaAssetLibraryClassificationSnapshot struct {
	MediaAssetID                uint                       `json:"media_asset_id"`
	LibraryPartition            MediaAssetLibraryPartition `json:"library_partition"`
	UploadOrigin                MediaAssetUploadOrigin     `json:"upload_origin"`
	ClassificationMethod        string                     `json:"classification_method"`
	CustomerServiceMediaPurpose string                     `json:"customer_service_media_purpose"`
	ClassificationRevision      int64                      `json:"classification_revision"`
}

type MediaAssetLibraryOwnershipChoice struct {
	LibraryPartition            MediaAssetLibraryPartition `json:"library_partition"`
	CustomerServiceMediaPurpose string                     `json:"customer_service_media_purpose"`
}

type MediaAssetLibraryOwnershipReview struct {
	Current                         MediaAssetLibraryClassificationSnapshot `json:"current"`
	CanReclassify                   bool                                    `json:"can_reclassify"`
	RestrictionReason               string                                  `json:"restriction_reason,omitempty"`
	AllowedChoices                  []MediaAssetLibraryOwnershipChoice      `json:"allowed_choices"`
	SuggestedOwnership              *MediaAssetLibraryOwnershipChoice       `json:"suggested_ownership,omitempty"`
	SuggestionBasis                 string                                  `json:"suggestion_basis"`
	ReferenceCoverage               string                                  `json:"reference_coverage"`
	RecognizedReferenceTotal        int                                     `json:"recognized_reference_total"`
	ReferenceCategoryCounts         map[string]int                          `json:"reference_category_counts"`
	RetainedDeletedParentReferences int                                     `json:"retained_deleted_parent_references"`
	MissingParentReferences         int                                     `json:"missing_parent_references"`
}

func CreateMediaAssetLibraryClassificationSnapshot(asset *MediaAsset) MediaAssetLibraryClassificationSnapshot {
	return MediaAssetLibraryClassificationSnapshot{MediaAssetID: asset.ID, LibraryPartition: asset.LibraryPartition,
		UploadOrigin: asset.UploadOrigin, ClassificationMethod: asset.ClassificationMethod,
		CustomerServiceMediaPurpose: asset.CustomerServiceMediaPurpose, ClassificationRevision: asset.ClassificationRevision}
}
