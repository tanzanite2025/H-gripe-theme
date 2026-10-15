package service

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
)

var (
	ErrMediaLibraryClassificationRevisionConflict = errors.New("media library classification changed; reload the ownership review")
	ErrMediaLibraryClassificationReasonRequired   = errors.New("media library classification reason must contain 3 to 1000 characters")
	ErrMediaLibraryClassificationAuditRequired    = errors.New("media library classification requires a transactional audit record")
)

const MediaAssetLibraryClassificationAuditResource = "media_asset_library_classification"

type MediaAssetLibraryClassificationChangeInput struct {
	LibraryPartition            media.MediaAssetLibraryPartition
	CustomerServiceMediaPurpose string
	ExpectedRevision            *int64
	Reason                      string
}

type MediaAssetLibraryClassificationAuditLogFactory func(media.MediaAssetLibraryClassificationSnapshot, media.MediaAssetLibraryClassificationSnapshot) (*audit.AuditLog, error)

func mediaAssetManualOperationalOwnershipReviewRestriction(asset *media.MediaAsset) string {
	if IsCustomerServiceConversationOwnedMediaAsset(asset) {
		return "conversation_attachments_require_conversation_lifecycle"
	}
	if asset.Visibility != "public" {
		return "private_assets_require_their_business_access_policy"
	}
	if asset.Status != "active" {
		return "inactive_assets_cannot_be_reclassified"
	}
	if asset.LibraryPartition == media.MediaLibraryPartitionCustomerSubmissions || asset.UploaderActorType == media.MediaUploadActorCustomer || asset.UploaderActorType == media.MediaUploadActorVisitor {
		return "customer_submissions_require_separate_publication_review"
	}
	return ""
}

func manualOperationalMediaLibraryOwnershipChoices() []media.MediaAssetLibraryOwnershipChoice {
	return []media.MediaAssetLibraryOwnershipChoice{{LibraryPartition: media.MediaLibraryPartitionCatalog},
		{LibraryPartition: media.MediaLibraryPartitionContent}, {LibraryPartition: media.MediaLibraryPartitionShared},
		{LibraryPartition: media.MediaLibraryPartitionCustomerService, CustomerServiceMediaPurpose: media.CustomerServiceMediaPurposeReusableReply},
		{LibraryPartition: media.MediaLibraryPartitionUnclassified}}
}

func (s *MediaService) ReviewMediaAssetLibraryOwnershipForManualClassification(ctx context.Context, id uint) (*media.MediaAssetLibraryOwnershipReview, error) {
	scopedRepository := s.repo.WithContextForMediaLibraryClassification(ctx)
	asset, err := scopedRepository.FindAssetByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrMediaAssetNotFound
	}
	if err != nil {
		return nil, err
	}
	review := &media.MediaAssetLibraryOwnershipReview{Current: media.CreateMediaAssetLibraryClassificationSnapshot(asset),
		RestrictionReason: mediaAssetManualOperationalOwnershipReviewRestriction(asset), ReferenceCoverage: "recognized_business_sources_only",
		SuggestionBasis: "manual_review_required", ReferenceCategoryCounts: map[string]int{}, AllowedChoices: []media.MediaAssetLibraryOwnershipChoice{}}
	review.CanReclassify = review.RestrictionReason == ""
	if !review.CanReclassify {
		return review, nil
	}
	review.AllowedChoices = manualOperationalMediaLibraryOwnershipChoices()
	references, err := scopedRepository.FindAssetReferences(asset)
	if err != nil {
		return nil, err
	}
	review.RecognizedReferenceTotal = len(references)
	for _, reference := range references {
		review.ReferenceCategoryCounts[string(reference.Category)]++
		if reference.UsageState == "retained_deleted_parent" {
			review.RetainedDeletedParentReferences++
		}
		if reference.UsageState == "missing_parent" {
			review.MissingParentReferences++
		}
	}
	// Uses are a review hint, never evidence of the original upload source.
	if asset.LibraryPartition == media.MediaLibraryPartitionUnclassified && len(review.ReferenceCategoryCounts) == 1 && review.MissingParentReferences == 0 {
		var suggested media.MediaAssetLibraryPartition
		if review.ReferenceCategoryCounts[string(media.ReferenceCategoryCatalog)] > 0 {
			suggested = media.MediaLibraryPartitionCatalog
		}
		if review.ReferenceCategoryCounts[string(media.ReferenceCategoryContent)] > 0 {
			suggested = media.MediaLibraryPartitionContent
		}
		if suggested != "" {
			review.SuggestedOwnership = &media.MediaAssetLibraryOwnershipChoice{LibraryPartition: suggested}
			review.SuggestionBasis = "single_recognized_business_category_requires_manual_confirmation"
		}
	}
	return review, nil
}

func (s *MediaService) ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(ctx context.Context, id uint, input MediaAssetLibraryClassificationChangeInput, auditFactory MediaAssetLibraryClassificationAuditLogFactory) (*media.MediaAsset, error) {
	if input.ExpectedRevision == nil || *input.ExpectedRevision < 0 {
		return nil, ErrUnsupportedMediaLibraryClassification
	}
	input.Reason = strings.TrimSpace(input.Reason)
	if reasonLength := utf8.RuneCountInString(input.Reason); reasonLength < 3 || reasonLength > 1000 {
		return nil, ErrMediaLibraryClassificationReasonRequired
	}
	validChoice := false
	for _, choice := range manualOperationalMediaLibraryOwnershipChoices() {
		if choice.LibraryPartition == input.LibraryPartition && choice.CustomerServiceMediaPurpose == input.CustomerServiceMediaPurpose {
			validChoice = true
		}
	}
	if !validChoice {
		return nil, ErrUnsupportedMediaLibraryClassification
	}
	if auditFactory == nil {
		return nil, ErrMediaLibraryClassificationAuditRequired
	}
	var result *media.MediaAsset
	err := s.repo.WithContextForMediaLibraryClassification(ctx).WithTransaction(func(repo *repository.MediaRepository, tx *gorm.DB) error {
		asset, err := repo.FindAssetByIDForUpdate(id)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrMediaAssetNotFound
		}
		if err != nil {
			return err
		}
		if mediaAssetManualOperationalOwnershipReviewRestriction(asset) != "" {
			return ErrMediaAssetForbidden
		}
		if asset.ClassificationRevision != *input.ExpectedRevision {
			return ErrMediaLibraryClassificationRevisionConflict
		}
		if asset.LibraryPartition == input.LibraryPartition && asset.CustomerServiceMediaPurpose == input.CustomerServiceMediaPurpose && asset.ClassificationMethod == media.MediaClassificationManualReview {
			result = asset
			return nil
		}
		if asset.ClassificationRevision == math.MaxInt64 {
			return ErrMediaLibraryClassificationRevisionConflict
		}
		before := media.CreateMediaAssetLibraryClassificationSnapshot(asset)
		asset.LibraryPartition, asset.CustomerServiceMediaPurpose, asset.ClassificationMethod = input.LibraryPartition, input.CustomerServiceMediaPurpose, media.MediaClassificationManualReview
		asset.ClassificationRevision++
		after := media.CreateMediaAssetLibraryClassificationSnapshot(asset)
		log, err := auditFactory(before, after)
		if err != nil {
			return err
		}
		if log == nil || log.UserID == 0 {
			return ErrMediaLibraryClassificationAuditRequired
		}
		log.Action, log.Resource, log.ResourceID, log.Status = "reclassify", MediaAssetLibraryClassificationAuditResource, id, "success"
		log.OldValue, log.NewValue = serviceAuditJSON(before), serviceAuditJSON(after)
		changes, err := json.Marshal(map[string]interface{}{"reason": input.Reason, "classification_revision": after.ClassificationRevision})
		if err != nil {
			return err
		}
		log.Changes = string(changes)
		updated, err := repo.UpdateMediaAssetLibraryClassificationAtExpectedRevision(asset, before.ClassificationRevision)
		if err != nil {
			return err
		}
		if !updated {
			return ErrMediaLibraryClassificationRevisionConflict
		}
		if err := repository.NewAuditRepository(tx).CreateAuditLog(log); err != nil {
			return err
		}
		result = asset
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.hydrateAssetAccessURL(result)
	return result, nil
}
