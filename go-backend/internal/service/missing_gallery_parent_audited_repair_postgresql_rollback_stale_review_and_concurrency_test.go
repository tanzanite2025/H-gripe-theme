package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createIsolatedPostgreSQLMissingGalleryParentRepairFixture(t *testing.T) (*gorm.DB, *MediaService, *media.MediaAsset, *gallery.GalleryImage, *gallery.Gallery) {
	t.Helper()
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	assets := NewMediaService(repository.NewMediaRepository(env.db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
	assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
	asset, image, destination := createMediaAssociationRepairReviewFixture(t, env.db, assets)
	require.NoError(t, os.MkdirAll(filepath.Join(env.publicRoot, "catalog"), 0700))
	require.NoError(t, os.WriteFile(filepath.Join(env.publicRoot, filepath.FromSlash(asset.StorageKey)), []byte("synthetic bytes"), 0600))
	return env.db, assets, asset, image, destination
}

func TestPostgreSQLMissingGalleryParentRepairPreservesMetadataAndCommitsOneAuditWithoutChangingCleanupAuthority(t *testing.T) {
	db, assets, asset, image, destination := createIsolatedPostgreSQLMissingGalleryParentRepairFixture(t)
	preview, err := assets.PreviewMissingGalleryParentReattachment(t.Context(), asset.ID, image.ID, destination.ID, mediaAssociationGalleryOnlyReviewScope())
	require.NoError(t, err)
	_, err = assets.ReattachMissingGalleryParentWithFreshPreviewAndRequiredAudit(t.Context(), asset.ID, image.ID, destination.ID, preview.Fingerprint, "审核历史图片归属", mediaAssociationGalleryOnlyReviewScope(), func() *audit.AuditLog { return &audit.AuditLog{UserID: 42, Username: "reviewer"} })
	require.NoError(t, err)
	var after gallery.GalleryImage
	require.NoError(t, db.First(&after, image.ID).Error)
	require.Equal(t, destination.ID, after.GalleryID)
	references, err := assets.repo.FindAssetReferences(asset)
	require.NoError(t, err)
	require.Len(t, references, 1)
	require.Equal(t, "retained_unpublished_parent", references[0].UsageState)
	beforeComparable := *image
	beforeComparable.GalleryID = after.GalleryID
	beforeComparable.UpdatedAt = after.UpdatedAt
	require.Equal(t, beforeComparable, after)
	var logs []audit.AuditLog
	require.NoError(t, db.Where("resource = ?", "media_gallery_image_association").Find(&logs).Error)
	require.Len(t, logs, 1)
	require.Contains(t, logs[0].OldValue, "987654")
	require.Contains(t, logs[0].Changes, "审核历史图片归属")
	_, err = assets.ReattachMissingGalleryParentWithFreshPreviewAndRequiredAudit(t.Context(), asset.ID, image.ID, destination.ID, preview.Fingerprint, "重复审核提交", mediaAssociationGalleryOnlyReviewScope(), func() *audit.AuditLog { return &audit.AuditLog{UserID: 42} })
	require.ErrorIs(t, err, repository.ErrMediaAssociationRepairConflict)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)), ErrMediaAssetInUse)
}

func TestPostgreSQLMissingGalleryParentRepairRejectsStaleReviewsRestoredParentsAndAuditFailuresAtomically(t *testing.T) {
	for _, scenario := range []string{"image edit", "destination edit", "destination published", "source parent restored", "source parent soft deleted", "asset edit", "missing audit identity", "audit insert failure", "new conflicting registration", "missing object"} {
		t.Run(scenario, func(t *testing.T) {
			db, assets, asset, image, destination := createIsolatedPostgreSQLMissingGalleryParentRepairFixture(t)
			preview, err := assets.PreviewMissingGalleryParentReattachment(t.Context(), asset.ID, image.ID, destination.ID, mediaAssociationGalleryOnlyReviewScope())
			require.NoError(t, err)
			factory := func() *audit.AuditLog { return &audit.AuditLog{UserID: 42} }
			switch scenario {
			case "image edit":
				require.NoError(t, db.Model(image).Update("title", "new title").Error)
			case "destination edit":
				require.NoError(t, db.Model(destination).Update("name", "new draft name").Error)
			case "destination published":
				require.NoError(t, db.Model(destination).Update("status", "published").Error)
			case "source parent restored", "source parent soft deleted":
				parent := gallery.Gallery{ID: image.GalleryID, Name: "Restored", Slug: "restored"}
				require.NoError(t, db.Create(&parent).Error)
				if scenario == "source parent soft deleted" {
					require.NoError(t, db.Delete(&parent).Error)
				}
			case "asset edit":
				require.NoError(t, db.Model(asset).Update("alt", "new alt").Error)
			case "missing audit identity":
				factory = func() *audit.AuditLog { return &audit.AuditLog{} }
			case "audit insert failure":
				require.NoError(t, db.Exec("ALTER TABLE audit_logs ADD CONSTRAINT reject_association_repair_audit CHECK (resource <> 'media_gallery_image_association')").Error)
			case "new conflicting registration":
				other := *asset
				other.ID = 0
				other.URL += "?conflict=1"
				require.NoError(t, db.Create(&other).Error)
			case "missing object":
				require.NoError(t, assets.storage.Delete(t.Context(), asset.StorageKey))
			}
			_, err = assets.ReattachMissingGalleryParentWithFreshPreviewAndRequiredAudit(t.Context(), asset.ID, image.ID, destination.ID, preview.Fingerprint, "核查后修复关联", mediaAssociationGalleryOnlyReviewScope(), factory)
			require.Error(t, err)
			switch scenario {
			case "missing audit identity":
				require.ErrorIs(t, err, ErrMediaLibraryClassificationAuditRequired)
			case "audit insert failure":
				require.Contains(t, err.Error(), "reject_association_repair_audit")
			case "missing object":
				require.ErrorIs(t, err, ErrMediaAssetURLUnavailable)
			default:
				require.ErrorIs(t, err, repository.ErrMediaAssociationRepairConflict)
			}
			var after gallery.GalleryImage
			require.NoError(t, db.First(&after, image.ID).Error)
			require.Equal(t, uint(987654), after.GalleryID)
			var count int64
			require.NoError(t, db.Model(&audit.AuditLog{}).Where("resource = ?", "media_gallery_image_association").Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestPostgreSQLMissingGalleryParentRepairCancellationBeforeAuditRollsBackAssociationAndDestinationTimestamp(t *testing.T) {
	db, assets, asset, image, destination := createIsolatedPostgreSQLMissingGalleryParentRepairFixture(t)
	preview, err := assets.PreviewMissingGalleryParentReattachment(t.Context(), asset.ID, image.ID, destination.ID, mediaAssociationGalleryOnlyReviewScope())
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	callbackName := "cancel_missing_gallery_parent_repair_before_audit"
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table == "audit_logs" {
			cancel()
			tx.AddError(ctx.Err())
		}
	}))
	defer db.Callback().Create().Remove(callbackName)
	_, err = assets.ReattachMissingGalleryParentWithFreshPreviewAndRequiredAudit(ctx, asset.ID, image.ID, destination.ID, preview.Fingerprint, "取消前核查关联", mediaAssociationGalleryOnlyReviewScope(), func() *audit.AuditLog { return &audit.AuditLog{UserID: 42} })
	require.Error(t, err)
	var after gallery.GalleryImage
	require.NoError(t, db.First(&after, image.ID).Error)
	require.Equal(t, image.GalleryID, after.GalleryID)
	var parent gallery.Gallery
	require.NoError(t, db.First(&parent, destination.ID).Error)
	require.True(t, destination.UpdatedAt.Equal(parent.UpdatedAt))
	var count int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Where("resource = ?", "media_gallery_image_association").Count(&count).Error)
	require.Zero(t, count)
}

func TestPostgreSQLMissingGalleryParentRepairUsesBoundedTableLockWaitAndLeavesNoChanges(t *testing.T) {
	db, assets, asset, image, destination := createIsolatedPostgreSQLMissingGalleryParentRepairFixture(t)
	preview, err := assets.PreviewMissingGalleryParentReattachment(t.Context(), asset.ID, image.ID, destination.ID, mediaAssociationGalleryOnlyReviewScope())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 8*time.Second)
	defer cancel()
	locked, release := make(chan struct{}), make(chan struct{})
	writerResult := make(chan error, 1)
	go func() {
		writerResult <- db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec("LOCK TABLE galleries IN ROW EXCLUSIVE MODE").Error; err != nil {
				return err
			}
			close(locked)
			select {
			case <-release:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()
	select {
	case <-locked:
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	_, err = assets.ReattachMissingGalleryParentWithFreshPreviewAndRequiredAudit(ctx, asset.ID, image.ID, destination.ID, preview.Fingerprint, "等待并发图库编辑", mediaAssociationGalleryOnlyReviewScope(), func() *audit.AuditLog { return &audit.AuditLog{UserID: 42} })
	close(release)
	require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, writerResult))
	require.Error(t, err)
	require.Contains(t, err.Error(), "lock timeout")
	var after gallery.GalleryImage
	require.NoError(t, db.First(&after, image.ID).Error)
	require.Equal(t, image.GalleryID, after.GalleryID)
}

func TestPostgreSQLMissingGalleryParentRepairConcurrentSubmissionsCommitExactlyOnce(t *testing.T) {
	db, assets, asset, image, destination := createIsolatedPostgreSQLMissingGalleryParentRepairFixture(t)
	preview, err := assets.PreviewMissingGalleryParentReattachment(t.Context(), asset.ID, image.ID, destination.ID, mediaAssociationGalleryOnlyReviewScope())
	require.NoError(t, err)
	var group sync.WaitGroup
	results := make(chan error, 2)
	start := make(chan struct{})
	for range 2 {
		group.Go(func() {
			<-start
			_, err := assets.ReattachMissingGalleryParentWithFreshPreviewAndRequiredAudit(t.Context(), asset.ID, image.ID, destination.ID, preview.Fingerprint, "并发审核请求", mediaAssociationGalleryOnlyReviewScope(), func() *audit.AuditLog { return &audit.AuditLog{UserID: 42} })
			results <- err
		})
	}
	close(start)
	group.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, repository.ErrMediaAssociationRepairConflict) {
			conflict++
		} else {
			t.Fatal(err)
		}
	}
	require.Equal(t, 1, success)
	require.Equal(t, 1, conflict)
	var count int64
	require.NoError(t, db.Model(&audit.AuditLog{}).Where("resource = ?", "media_gallery_image_association").Count(&count).Error)
	require.EqualValues(t, 1, count)
}

func TestPostgreSQLMissingGalleryParentRepairProtectsAbsentParentAgainstConcurrentInsertion(t *testing.T) {
	db, assets, asset, image, destination := createIsolatedPostgreSQLMissingGalleryParentRepairFixture(t)
	preview, err := assets.PreviewMissingGalleryParentReattachment(t.Context(), asset.ID, image.ID, destination.ID, mediaAssociationGalleryOnlyReviewScope())
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	entered, release := make(chan int, 1), make(chan struct{})
	callbackName := "pause_missing_gallery_parent_repair_before_audit"
	require.NoError(t, db.Callback().Create().Before("gorm:create").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table == "audit_logs" {
			pausePostgreSQLMediaObjectOwnershipTransactionUntilReleased(t, ctx, tx, entered, release)
		}
	}))
	defer db.Callback().Create().Remove(callbackName)
	repaired := make(chan error, 1)
	go func() {
		_, err := assets.ReattachMissingGalleryParentWithFreshPreviewAndRequiredAudit(ctx, asset.ID, image.ID, destination.ID, preview.Fingerprint, "确认缺失父图库", mediaAssociationGalleryOnlyReviewScope(), func() *audit.AuditLog { return &audit.AuditLog{UserID: 42} })
		repaired <- err
	}()
	pid := awaitPostgreSQLMediaObjectOwnershipTransactionPID(t, ctx, entered)
	inserted := make(chan error, 1)
	go func() {
		inserted <- db.WithContext(ctx).Create(&gallery.Gallery{ID: image.GalleryID, Name: "Concurrent restore", Slug: "concurrent-restore"}).Error
	}()
	waitForPostgreSQLGalleryMutationToBlockOnTransaction(t, db, pid, "galleries")
	close(release)
	require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, repaired))
	require.NoError(t, awaitPostgreSQLMediaObjectOwnershipOperationResult(t, ctx, inserted))
	var after gallery.GalleryImage
	require.NoError(t, db.First(&after, image.ID).Error)
	require.Equal(t, destination.ID, after.GalleryID)
}
