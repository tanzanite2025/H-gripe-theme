package service

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/domain/ticket"
)

func createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator(_, _ media.MediaAssetLibraryClassificationSnapshot) (*audit.AuditLog, error) {
	return &audit.AuditLog{UserID: 42, Username: "reviewer"}, nil
}

func createManualMediaAssetOwnershipChangeForExpectedRevision(partition media.MediaAssetLibraryPartition, revision int64) MediaAssetLibraryClassificationChangeInput {
	return MediaAssetLibraryClassificationChangeInput{LibraryPartition: partition, ExpectedRevision: &revision, Reason: "已核对素材归属"}
}

func TestMediaAssetManualOwnershipAuditFailureRollsBackAndSuccessfulReviewPreservesProvenanceFilesEvidenceAndReferences(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &product.Product{}, &product.ProductMedia{})
	root := t.TempDir()
	s := newMediaDeleteTestService(t, db, root)
	asset, err := s.UploadAsset(t.Context(), MediaUploadInput{File: multipartFileHeader(t, "review.png", "image/png", mediaUploadTestPNG(t)),
		UploadActorIdentity: media.MediaUploadActorIdentity{ActorType: media.MediaUploadActorAdministrator, AccountID: 42}})
	require.NoError(t, err)
	before, err := s.repo.FindAssetByID(asset.ID)
	require.NoError(t, err)
	fileBefore, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(before.StorageKey)))
	require.NoError(t, err)
	item := product.Product{SKU: "manual-review-product", Slug: "manual-review-product", Name: "Review product", PriceMinor: 100}
	require.NoError(t, db.Create(&item).Error)
	relation := product.ProductMedia{ProductID: item.ID, MediaAssetID: &asset.ID, URL: asset.URL, MediaType: "image"}
	require.NoError(t, db.Create(&relation).Error)
	input := createManualMediaAssetOwnershipChangeForExpectedRevision(media.MediaLibraryPartitionCatalog, 0)
	_, err = s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, input, createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator)
	require.Error(t, err, "missing audit table must prevent committing classification")
	unchanged, err := s.repo.FindAssetByID(asset.ID)
	require.NoError(t, err)
	require.Equal(t, media.CreateMediaAssetLibraryClassificationSnapshot(before), media.CreateMediaAssetLibraryClassificationSnapshot(unchanged))
	require.NoError(t, db.AutoMigrate(&audit.AuditLog{}))
	after, err := s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, input, createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator)
	require.NoError(t, err)
	require.Equal(t, media.MediaLibraryPartitionCatalog, after.LibraryPartition)
	require.EqualValues(t, 1, after.ClassificationRevision)
	require.Equal(t, media.MediaClassificationManualReview, after.ClassificationMethod)
	for _, pair := range [][2]string{{before.URL, after.URL}, {before.StorageKey, after.StorageKey}, {string(before.UploadOrigin), string(after.UploadOrigin)},
		{before.CopyrightClaimJSON, after.CopyrightClaimJSON}, {before.ContentSHA256, after.ContentSHA256}, {before.UploadPurpose, after.UploadPurpose}, {before.Visibility, after.Visibility}} {
		require.Equal(t, pair[0], pair[1])
	}
	require.Equal(t, before.UploaderActorType, after.UploaderActorType)
	fileAfter, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(after.StorageKey)))
	require.NoError(t, err)
	require.Equal(t, fileBefore, fileAfter)
	var retainedRelation product.ProductMedia
	require.NoError(t, db.First(&retainedRelation, relation.ID).Error)
	require.Equal(t, relation.MediaAssetID, retainedRelation.MediaAssetID)
	require.Equal(t, relation.URL, retainedRelation.URL)
	var logs []audit.AuditLog
	require.NoError(t, db.Find(&logs).Error)
	require.Len(t, logs, 1)
	require.Equal(t, MediaAssetLibraryClassificationAuditResource, logs[0].Resource)
	require.Equal(t, asset.ID, logs[0].ResourceID)
	var oldSnapshot, newSnapshot media.MediaAssetLibraryClassificationSnapshot
	require.NoError(t, json.Unmarshal([]byte(logs[0].OldValue), &oldSnapshot))
	require.NoError(t, json.Unmarshal([]byte(logs[0].NewValue), &newSnapshot))
	require.Equal(t, media.CreateMediaAssetLibraryClassificationSnapshot(before), oldSnapshot)
	require.Equal(t, media.CreateMediaAssetLibraryClassificationSnapshot(after), newSnapshot)
	require.Contains(t, logs[0].Changes, input.Reason)
	_, err = s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, input, createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator)
	require.ErrorIs(t, err, ErrMediaLibraryClassificationRevisionConflict)
	input.ExpectedRevision = &after.ClassificationRevision
	noop, err := s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, input, createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator)
	require.NoError(t, err)
	require.EqualValues(t, 1, noop.ClassificationRevision)
	var logCount int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Count(&logCount).Error)
	require.EqualValues(t, 1, logCount)
	metadata, err := s.UpdateAsset(asset.ID, MediaAssetUpdateInput{Alt: "updated alt", Caption: "updated caption", Status: "active", Visibility: "public"})
	require.NoError(t, err)
	require.Equal(t, media.CreateMediaAssetLibraryClassificationSnapshot(after), media.CreateMediaAssetLibraryClassificationSnapshot(metadata))
}

func TestMediaAssetManualOwnershipReviewOnlySuggestsRecognizedSingleCategoryAndNeverGuessesUploadOrigin(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &product.Product{}, &product.ProductMedia{}, &ticket.TicketMessage{})
	s := newMediaDeleteTestService(t, db, t.TempDir())
	asset := media.MediaAsset{Filename: "history.png", URL: "https://media.example.test/uploads/history.png"}
	require.NoError(t, db.Create(&asset).Error)
	item := product.Product{SKU: "history", Slug: "history", Name: "Historical product", PriceMinor: 100}
	require.NoError(t, db.Create(&item).Error)
	require.NoError(t, db.Create(&product.ProductMedia{ProductID: item.ID, MediaAssetID: &asset.ID, URL: asset.URL, MediaType: "image"}).Error)
	require.NoError(t, db.Model(&item).Update("deleted_at", time.Now()).Error)
	review, err := s.ReviewMediaAssetLibraryOwnershipForManualClassification(t.Context(), asset.ID)
	require.NoError(t, err)
	require.True(t, review.CanReclassify)
	require.Equal(t, media.MediaLibraryPartitionCatalog, review.SuggestedOwnership.LibraryPartition)
	require.Equal(t, media.MediaUploadOriginLegacyUnknown, review.Current.UploadOrigin)
	require.Equal(t, media.MediaLibraryPartitionUnclassified, review.Current.LibraryPartition)
	require.Equal(t, 1, review.RetainedDeletedParentReferences)
	require.Equal(t, "recognized_business_sources_only", review.ReferenceCoverage)
	require.NoError(t, db.Unscoped().Delete(&item).Error)
	review, err = s.ReviewMediaAssetLibraryOwnershipForManualClassification(t.Context(), asset.ID)
	require.NoError(t, err)
	require.Nil(t, review.SuggestedOwnership)
	require.Equal(t, 1, review.MissingParentReferences)
	attachments, _ := json.Marshal([]string{asset.URL})
	require.NoError(t, db.Create(&ticket.TicketMessage{TicketID: 99, MessageType: "image", Content: "sensitive conversation text", Attachments: string(attachments)}).Error)
	review, err = s.ReviewMediaAssetLibraryOwnershipForManualClassification(t.Context(), asset.ID)
	require.NoError(t, err)
	require.Nil(t, review.SuggestedOwnership)
	require.Len(t, review.ReferenceCategoryCounts, 2)
	encoded, err := json.Marshal(review)
	require.NoError(t, err)
	require.NotContains(t, string(encoded), "sensitive conversation text")
	stored, err := s.repo.FindAssetByID(asset.ID)
	require.NoError(t, err)
	require.Equal(t, media.MediaClassificationHistoricalUnknown, stored.ClassificationMethod)
	require.Zero(t, stored.ClassificationRevision)
}

func TestMediaAssetManualOwnershipReviewRefusesConversationPrivateCustomerAndInactiveAssets(t *testing.T) {
	for _, asset := range []media.MediaAsset{
		{Visibility: "private"}, {Status: "archived"}, {LibraryPartition: media.MediaLibraryPartitionCustomerSubmissions},
		{UploaderActorType: media.MediaUploadActorCustomer}, {UploaderActorType: media.MediaUploadActorVisitor},
		{LibraryPartition: media.MediaLibraryPartitionCustomerService, CustomerServiceMediaPurpose: media.CustomerServiceMediaPurposeConversationAttachment},
		{StorageKey: `customer-service\attachments\originals\history.png`},
		{URL: "https://private.example.test/customer-service/attachments/originals/history.png"},
	} {
		t.Run(asset.Visibility+asset.Status+string(asset.LibraryPartition)+string(asset.UploaderActorType)+asset.StorageKey+asset.URL, func(t *testing.T) {
			db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &audit.AuditLog{})
			s := newMediaDeleteTestService(t, db, t.TempDir())
			asset.Filename = "restricted.png"
			if asset.URL == "" {
				asset.URL = "https://media.example.test/uploads/restricted.png"
			}
			require.NoError(t, db.Create(&asset).Error)
			review, err := s.ReviewMediaAssetLibraryOwnershipForManualClassification(t.Context(), asset.ID)
			require.NoError(t, err)
			require.False(t, review.CanReclassify)
			require.Empty(t, review.AllowedChoices)
			require.NotEmpty(t, review.RestrictionReason)
			_, err = s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, createManualMediaAssetOwnershipChangeForExpectedRevision(media.MediaLibraryPartitionCatalog, 0), createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator)
			require.ErrorIs(t, err, ErrMediaAssetForbidden)
		})
	}
}

func TestMediaAssetManualOwnershipValidationRequiresValidChoiceReasonRevisionAndAuditActor(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &audit.AuditLog{})
	s := newMediaDeleteTestService(t, db, t.TempDir())
	asset := media.MediaAsset{Filename: "validation.png", URL: "https://media.example.test/uploads/validation.png"}
	require.NoError(t, db.Create(&asset).Error)
	for _, choice := range []MediaAssetLibraryClassificationChangeInput{
		{LibraryPartition: media.MediaLibraryPartitionCatalog, Reason: "valid reason"},
		createManualMediaAssetOwnershipChangeForExpectedRevision(media.MediaLibraryPartitionCatalog, -1),
		createManualMediaAssetOwnershipChangeForExpectedRevision(media.MediaLibraryPartitionCustomerSubmissions, 0),
		createManualMediaAssetOwnershipChangeForExpectedRevision(media.MediaLibraryPartitionCustomerService, 0),
		{LibraryPartition: media.MediaLibraryPartitionCatalog, CustomerServiceMediaPurpose: media.CustomerServiceMediaPurposeConversationAttachment, ExpectedRevision: new(int64), Reason: "valid reason"},
	} {
		_, err := s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, choice, createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator)
		require.ErrorIs(t, err, ErrUnsupportedMediaLibraryClassification)
	}
	choice := createManualMediaAssetOwnershipChangeForExpectedRevision(media.MediaLibraryPartitionShared, 0)
	choice.Reason = "  短  "
	_, err := s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, choice, createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator)
	require.ErrorIs(t, err, ErrMediaLibraryClassificationReasonRequired)
	choice.Reason = "人工核对来源与用途"
	_, err = s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, choice, nil)
	require.ErrorIs(t, err, ErrMediaLibraryClassificationAuditRequired)
	_, err = s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, choice, func(_, _ media.MediaAssetLibraryClassificationSnapshot) (*audit.AuditLog, error) {
		return &audit.AuditLog{}, nil
	})
	require.ErrorIs(t, err, ErrMediaLibraryClassificationAuditRequired)
	injected := errors.New("audit identity construction failed")
	_, err = s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, choice, func(_, _ media.MediaAssetLibraryClassificationSnapshot) (*audit.AuditLog, error) {
		return nil, injected
	})
	require.ErrorIs(t, err, injected)
	choice.LibraryPartition = media.MediaLibraryPartitionCustomerService
	choice.CustomerServiceMediaPurpose = media.CustomerServiceMediaPurposeReusableReply
	updated, err := s.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(t.Context(), asset.ID, choice, createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator)
	require.NoError(t, err)
	require.Equal(t, media.MediaUploadOriginLegacyUnknown, updated.UploadOrigin)
}
