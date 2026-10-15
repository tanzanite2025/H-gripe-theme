package service

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/media"
)

func TestPostgreSQLMediaAssetManualClassificationSerializesCompetingReviewsAndPreservesClassificationDuringMetadataEdits(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	for _, name := range []string{"422_add_media_library_business_partitions_trusted_upload_origins_and_customer_service_purposes.up.sql", "423_add_audited_manual_media_library_classification_and_revision_protection.up.sql"} {
		migration, err := os.ReadFile(filepath.Join("..", "..", "migrations", name))
		require.NoError(t, err)
		require.NoError(t, env.db.Exec(string(migration)).Error)
	}
	asset := media.MediaAsset{Filename: "concurrent-history.png", URL: "https://shop.example.test/uploads/concurrent-history.png"}
	require.NoError(t, env.db.Create(&asset).Error)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	firstReviewLocked := make(chan struct{})
	releaseFirstReview := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(releaseFirstReview) })
	firstResult := make(chan error, 1)
	go func() {
		_, err := env.mediaService.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(ctx, asset.ID, createManualMediaAssetOwnershipChangeForExpectedRevision(media.MediaLibraryPartitionCatalog, 0),
			func(before, after media.MediaAssetLibraryClassificationSnapshot) (*audit.AuditLog, error) {
				close(firstReviewLocked)
				select {
				case <-releaseFirstReview:
				case <-ctx.Done():
					return nil, ctx.Err()
				}
				return createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator(before, after)
			})
		firstResult <- err
	}()
	select {
	case <-firstReviewLocked:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	competingReview := make(chan error, 1)
	go func() {
		_, err := env.mediaService.ReclassifyMediaAssetLibraryOwnershipWithRequiredAudit(ctx, asset.ID, createManualMediaAssetOwnershipChangeForExpectedRevision(media.MediaLibraryPartitionContent, 0), createMediaAssetOwnershipReviewAuditFromAuthenticatedTestAdministrator)
		competingReview <- err
	}()
	metadataResult := make(chan error, 1)
	go func() {
		_, err := env.mediaService.UpdateAsset(asset.ID, MediaAssetUpdateInput{Alt: "concurrent metadata", Caption: "preserve reviewed ownership", Status: "active", Visibility: "public"})
		metadataResult <- err
	}()
	// Observe actual PostgreSQL waiters while the first audit callback owns the
	// asset lock. This proves both writes use row locks, beyond final-state checks.
	require.Eventually(t, func() bool {
		var waiting int64
		err := env.db.Raw("SELECT COUNT(*) FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock' AND query LIKE '%media_assets%'").Scan(&waiting).Error
		return err == nil && waiting >= 2
	}, 3*time.Second, 20*time.Millisecond)
	releaseOnce.Do(func() { close(releaseFirstReview) })
	require.NoError(t, <-firstResult)
	require.ErrorIs(t, <-competingReview, ErrMediaLibraryClassificationRevisionConflict)
	require.NoError(t, <-metadataResult)
	stored, err := env.mediaService.repo.FindAssetByID(asset.ID)
	require.NoError(t, err)
	require.Equal(t, media.MediaLibraryPartitionCatalog, stored.LibraryPartition)
	require.EqualValues(t, 1, stored.ClassificationRevision)
	require.Equal(t, media.MediaClassificationManualReview, stored.ClassificationMethod)
	require.Equal(t, media.MediaUploadOriginLegacyUnknown, stored.UploadOrigin)
	require.Equal(t, "concurrent metadata", stored.Alt)
	var count int64
	require.NoError(t, env.db.Model(&audit.AuditLog{}).Where("resource=? AND resource_id=?", MediaAssetLibraryClassificationAuditResource, asset.ID).Count(&count).Error)
	require.EqualValues(t, 1, count)
}
