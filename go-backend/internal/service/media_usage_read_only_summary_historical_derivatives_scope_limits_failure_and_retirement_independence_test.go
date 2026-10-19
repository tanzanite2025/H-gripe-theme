package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/repository"
)

func TestReadOnlyMediaUsageSummaryIncludesRetainedDerivativeReferencesCrossBusinessAndMissingParentsWithoutChangingRows(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &product.Product{}, &product.ProductMedia{})
	s := newMediaDeleteTestService(t, db, t.TempDir())
	asset := media.MediaAsset{Filename: "shared.png", URL: "https://media.example.test/uploads/shared.png", StorageKey: "shared.png", LibraryPartition: media.MediaLibraryPartitionShared}
	unused := media.MediaAsset{Filename: "unused.png", URL: "https://media.example.test/uploads/unused.png", StorageKey: "unused.png"}
	require.NoError(t, db.Create(&asset).Error)
	require.NoError(t, db.Create(&unused).Error)
	derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "historical", URL: "https://media.example.test/uploads/derived.png", StorageKey: "derived.png"}
	require.NoError(t, db.Create(&derivative).Error)
	require.NoError(t, db.Delete(&derivative).Error)
	parent := product.Product{SKU: "retained-usage", Slug: "retained-usage", Name: "retained", PriceMinor: 1}
	require.NoError(t, db.Create(&parent).Error)
	placement := product.ProductMedia{ProductID: parent.ID, MediaAssetID: &asset.ID, URL: asset.URL, MediaType: "image"}
	require.NoError(t, db.Create(&placement).Error)
	require.NoError(t, db.Delete(&parent).Error)
	require.NoError(t, db.Create(&product.ProductMedia{ProductID: 99999, MediaAssetID: &asset.ID, URL: asset.URL, MediaType: "image"}).Error)
	require.NoError(t, db.Exec("CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, content TEXT, featured_img TEXT, deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO posts(id,title,content,featured_img) VALUES(1,'authorized article',?,''),(2,'external same basename',?, '')", `<p><img src="`+derivative.URL+`"></p>`, `<img src="https://external.example.test/uploads/unused.png">`).Error)
	scope := MediaUsageSummaryAccessScope{AllowedReferenceTypes: map[string]bool{"product_media": true, "post": true}}
	report, err := s.RebuildReadOnlyMediaUsageSummary(t.Context(), MediaAssetListInput{}, scope)
	require.NoError(t, err)
	require.Equal(t, "recognized_business_sources_only", report.Coverage)
	require.Len(t, report.Rows, 2)
	require.Equal(t, 1, report.Counts["no_recognized_references"])
	require.True(t, report.Rows[0].CrossBusiness)
	require.True(t, report.Rows[0].Retained)
	require.True(t, report.Rows[0].Anomalous)
	require.Equal(t, 3, report.Rows[0].VisibleReferenceTotal)
	require.True(t, report.Rows[1].NoRecognizedReferences)
	var after media.MediaAsset
	require.NoError(t, db.First(&after, asset.ID).Error)
	require.True(t, after.UpdatedAt.Equal(asset.UpdatedAt))
	err = s.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID))
	require.ErrorIs(t, err, ErrMediaAssetInUse, "summary must not weaken authoritative retirement")
	// Rebuilding observes a later business edit instead of retaining stale data.
	require.NoError(t, db.Exec("UPDATE posts SET featured_img = ? WHERE id = 2", unused.URL).Error)
	later, err := s.RebuildReadOnlyMediaUsageSummary(t.Context(), MediaAssetListInput{}, scope)
	require.NoError(t, err)
	require.False(t, later.Rows[1].NoRecognizedReferences)
	require.Equal(t, 0, later.Counts["no_recognized_references"])
}

func TestReadOnlyMediaUsageSummaryNeverTreatsRestrictedReferencesAsUnusedOrExposesTheirMetadata(t *testing.T) {
	asset := media.MediaAsset{Filename: "safe.png"}
	asset.ID = 1
	refs := []media.AssetReference{{Category: media.ReferenceCategoryCatalog, ResourceType: "product_media", ResourceID: 3, Label: "visible"},
		{Category: media.ReferenceCategorySupport, ResourceType: "ticket_message", ResourceID: 999, ParentResourceID: 444, Label: "private customer identity", Field: "attachments", UsageState: "missing_parent"}}
	row := summarizeMediaUsageWithPermissionScopedReferenceStates(asset, refs, MediaUsageSummaryAccessScope{AllowedReferenceTypes: map[string]bool{"product_media": true}})
	require.False(t, row.NoRecognizedReferences)
	require.True(t, row.HasRestrictedReferences)
	require.Equal(t, 1, row.VisibleReferenceTotal)
	require.False(t, row.CrossBusiness)
	require.False(t, row.Anomalous)
	encoded, err := json.Marshal(row)
	require.NoError(t, err)
	for _, private := range []string{"private customer identity", "ticket_message", "999", "444", "support", "attachments", "missing_parent"} {
		require.NotContains(t, string(encoded), private)
	}
	row = summarizeMediaUsageWithPermissionScopedReferenceStates(asset, refs[1:], MediaUsageSummaryAccessScope{})
	require.False(t, row.NoRecognizedReferences)
	require.Zero(t, row.VisibleReferenceTotal)
	require.True(t, row.HasRestrictedReferences)
}

func TestReadOnlyMediaUsageSummaryClassifiesPendingImmutableDisabledAndRegistrationConflicts(t *testing.T) {
	for _, source := range []struct{ resourceType, state, field string }{
		{"media_upload_session", "", "pending_upload"}, {"order_evidence_snapshots", "", "snapshot_data"}, {"ticket_auto_reply", "disabled", "metadata"},
		{"post", "retained_deleted_parent", "content"}, {"post", "retained_unpublished_parent", "content"},
	} {
		t.Run(source.resourceType+source.state, func(t *testing.T) {
			row := summarizeMediaUsageWithPermissionScopedReferenceStates(media.MediaAsset{}, []media.AssetReference{{ResourceType: source.resourceType, UsageState: source.state, Field: source.field}},
				MediaUsageSummaryAccessScope{AllowedReferenceTypes: map[string]bool{source.resourceType: true}})
			require.True(t, row.Retained)
			require.False(t, row.NoRecognizedReferences)
		})
	}
	row := summarizeMediaUsageWithPermissionScopedReferenceStates(media.MediaAsset{}, []media.AssetReference{{Category: media.ReferenceCategoryRegistration,
		ResourceType: "media_asset_registration", ResourceID: 3456, ParentResourceID: 7890, Label: "private owner #7890"}},
		MediaUsageSummaryAccessScope{AllowedReferenceTypes: map[string]bool{"media_asset_registration": true}})
	require.True(t, row.Anomalous)
	require.False(t, row.CrossBusiness)
	require.Empty(t, row.Categories)
	require.Equal(t, "0", row.References[0].ResourceID)
	require.Equal(t, "0", row.References[0].ParentResourceID)
	require.NotContains(t, row.References[0].Label, "7890")
}

func TestReadOnlyMediaUsageSummaryReferencePreviewDoesNotTruncateCountsOrStates(t *testing.T) {
	refs := make([]media.AssetReference, 21)
	for i := range refs {
		refs[i] = media.AssetReference{Category: media.ReferenceCategoryCatalog, ResourceType: "product_media"}
	}
	refs[20] = media.AssetReference{Category: media.ReferenceCategoryContent, ResourceType: "post", UsageState: "missing_parent"}
	row := summarizeMediaUsageWithPermissionScopedReferenceStates(media.MediaAsset{}, refs, MediaUsageSummaryAccessScope{AllowedReferenceTypes: map[string]bool{"product_media": true, "post": true}})
	require.Len(t, row.References, 20)
	require.Equal(t, 21, row.VisibleReferenceTotal)
	require.True(t, row.ReferencePreviewLimited)
	require.True(t, row.CrossBusiness)
	require.True(t, row.Anomalous)
}

func TestReadOnlyMediaUsageSummaryRejectsOversizeScopeAndCanRebuildNarrowedScope(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	s := newMediaDeleteTestService(t, db, t.TempDir())
	for i := 0; i <= MediaUsageSummaryCandidateLimit; i++ {
		require.NoError(t, db.Create(&media.MediaAsset{Filename: fmt.Sprintf("scope-%03d.png", i), URL: fmt.Sprintf("https://media.example.test/%d.png", i)}).Error)
	}
	report, err := s.RebuildReadOnlyMediaUsageSummary(t.Context(), MediaAssetListInput{}, MediaUsageSummaryAccessScope{})
	require.ErrorIs(t, err, repository.ErrMediaUsageSnapshotLimitExceeded)
	require.Nil(t, report.Rows)
	report, err = s.RebuildReadOnlyMediaUsageSummary(t.Context(), MediaAssetListInput{Search: "scope-000"}, MediaUsageSummaryAccessScope{})
	require.NoError(t, err)
	require.Len(t, report.Rows, 1)
	require.True(t, report.Rows[0].NoRecognizedReferences)
}

func TestReadOnlyMediaUsageSummaryFailureCancellationAndConcurrentRequestNeverDeliverPartialConclusions(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	s := newMediaDeleteTestService(t, db, t.TempDir())
	require.NoError(t, db.Create(&media.MediaAsset{Filename: "safe.png", URL: "https://media.example.test/safe.png"}).Error)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	report, err := s.RebuildReadOnlyMediaUsageSummary(ctx, MediaAssetListInput{}, MediaUsageSummaryAccessScope{})
	require.Error(t, err)
	require.Nil(t, report.Rows)
	s.usageSummaryRunning.Store(true)
	_, err = s.RebuildReadOnlyMediaUsageSummary(t.Context(), MediaAssetListInput{}, MediaUsageSummaryAccessScope{})
	require.ErrorIs(t, err, ErrMediaUsageSummaryAlreadyRunning)
	s.usageSummaryRunning.Store(false)
	// A malformed recognized source must fail the entire snapshot, not appear unused.
	require.NoError(t, db.Exec("CREATE TABLE suggestion_feedback(id INTEGER PRIMARY KEY,attachments TEXT,deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO suggestion_feedback(id,attachments) VALUES(1,'{broken')").Error)
	report, err = s.RebuildReadOnlyMediaUsageSummary(t.Context(), MediaAssetListInput{}, MediaUsageSummaryAccessScope{})
	require.Error(t, err)
	require.Nil(t, report.Rows)
	require.Nil(t, report.Counts)
	require.False(t, s.usageSummaryRunning.Load())
}

func TestReadOnlyMediaUsageSummaryKeepsUnboundAndExpiredBusinessDraftReceiptsOutOfUnusedUntilConsumed(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{}, &media.BusinessDraftMediaUploadReceipt{})
	s := newMediaDeleteTestService(t, db, t.TempDir())
	for index, status := range []string{"uploaded", "bound"} {
		asset := media.MediaAsset{Filename: fmt.Sprintf("draft-%d.png", index), URL: fmt.Sprintf("https://media.example.test/draft-%d.png", index)}
		require.NoError(t, db.Create(&asset).Error)
		require.NoError(t, db.Create(&media.BusinessDraftMediaUploadReceipt{ID: fmt.Sprintf("00000000-0000-0000-0000-%012d", index+1), AssetID: &asset.ID,
			OriginalObjectKey: fmt.Sprintf("draft-%d.png", index), ObjectKeysJSON: datatypes.JSON(`[]`), Status: status, ExpiresAt: time.Now().Add(-time.Hour)}).Error)
	}
	report, err := s.RebuildReadOnlyMediaUsageSummary(t.Context(), MediaAssetListInput{}, MediaUsageSummaryAccessScope{})
	require.NoError(t, err)
	require.True(t, report.Rows[0].PendingBusinessDraft)
	require.True(t, report.Rows[0].Retained)
	require.False(t, report.Rows[0].NoRecognizedReferences)
	require.False(t, report.Rows[1].PendingBusinessDraft)
	require.True(t, report.Rows[1].NoRecognizedReferences)
	require.Equal(t, 1, report.Counts["retained"])
	require.Equal(t, 1, report.Counts["no_recognized_references"])
	require.NoError(t, db.Model(&media.BusinessDraftMediaUploadReceipt{}).Where("status = ?", "uploaded").Update("status", "bound").Error)
	rebuilt, err := s.RebuildReadOnlyMediaUsageSummary(t.Context(), MediaAssetListInput{}, MediaUsageSummaryAccessScope{})
	require.NoError(t, err)
	require.Equal(t, 2, rebuilt.Counts["no_recognized_references"])
	require.Zero(t, rebuilt.Counts["retained"])
}

func TestReadOnlyMediaUsageSummaryOwnershipReviewSuppressesSuggestionsDerivedFromPartialReferenceVisibility(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &media.MediaAssetDerivative{})
	s := newMediaDeleteTestService(t, db, t.TempDir())
	asset := media.MediaAsset{Filename: "historical.png", URL: "https://media.example.test/historical.png"}
	require.NoError(t, db.Create(&asset).Error)
	require.NoError(t, db.Exec("CREATE TABLE posts(id INTEGER PRIMARY KEY,title TEXT,content TEXT,featured_img TEXT,deleted_at DATETIME)").Error)
	require.NoError(t, db.Exec("INSERT INTO posts(id,title,content,featured_img) VALUES(1,'visible','',?)", asset.URL).Error)
	require.NoError(t, db.Exec("CREATE TABLE ticket_messages(id INTEGER PRIMARY KEY,ticket_id INTEGER,attachments TEXT)").Error)
	require.NoError(t, db.Exec("INSERT INTO ticket_messages(id,ticket_id,attachments) VALUES(12345,54321,?)", `["`+asset.URL+`"]`).Error)
	scope := MediaUsageSummaryAccessScope{AllowedReferenceTypes: map[string]bool{"post": true}}
	review, err := s.ReviewMediaAssetLibraryOwnershipWithPermissionScopedReferencePresentation(t.Context(), asset.ID, scope)
	require.NoError(t, err)
	require.True(t, review.HasRestrictedReferences)
	require.Equal(t, 1, review.RecognizedReferenceTotal)
	require.Equal(t, map[string]int{"content": 1}, review.ReferenceCategoryCounts)
	require.Nil(t, review.SuggestedOwnership)
	require.NoError(t, db.Exec("UPDATE ticket_messages SET attachments = '[]'").Error)
	review, err = s.ReviewMediaAssetLibraryOwnershipWithPermissionScopedReferencePresentation(t.Context(), asset.ID, scope)
	require.NoError(t, err)
	require.False(t, review.HasRestrictedReferences)
	require.NotNil(t, review.SuggestedOwnership)
	require.Equal(t, media.MediaLibraryPartitionContent, review.SuggestedOwnership.LibraryPartition)
}
