package service

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/post"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type historicalDedicatedObjectCleanupStorageWithDeletionPause struct {
	storage.StorageService
	beforeDelete func() error
}

func (s *historicalDedicatedObjectCleanupStorageWithDeletionPause) Delete(ctx context.Context, key string) error {
	if err := s.beforeDelete(); err != nil {
		return err
	}
	return s.StorageService.Delete(ctx, key)
}

func TestPostgreSQLHistoricalDedicatedObjectCleanupRetainsUncommittedDetachmentAndRejectsStaleRestoration(t *testing.T) {
	for _, namespace := range []string{"site-logo", "site-favicon", "visual-showcase"} {
		for _, race := range []string{"uncommitted detachment", "stale restoration during cleanup"} {
			t.Run(namespace+"/"+race, func(t *testing.T) {
				env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
				ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
				defer cancel()
				db := env.db.WithContext(ctx)
				require.NoError(t, db.AutoMigrate(&post.Post{}, &sitelogodomain.Asset{}, &sitefavicondomain.Asset{}, &homevisualtile.Tile{}))
				root := t.TempDir()
				provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://dedicated.example.test"})
				require.NoError(t, err)
				key := namespace + "/historical/concurrent.png"
				objectPath := filepath.Join(root, filepath.FromSlash(key))
				require.NoError(t, os.MkdirAll(filepath.Dir(objectPath), 0o755))
				require.NoError(t, os.WriteFile(objectPath, []byte("historical object"), 0o644))
				article := post.Post{Title: "Published article", Slug: "historical-concurrent", Status: "published", Content: `<img src="/uploads/` + key + `">`}
				require.NoError(t, db.Create(&article).Error)
				posts := repository.NewPostRepository(db)
				posts.ConfigurePublicBusinessMediaReferenceIdentityResolver(&MediaService{storage: provider, siteURL: "https://shop.example.test"})
				logos := NewSiteLogoService(repository.NewSiteLogoRepository(db), provider, "https://shop.example.test")
				favicons := NewSiteFaviconService(repository.NewSiteFaviconRepository(db), provider, "https://shop.example.test")
				visuals := NewHomeVisualTileService(repository.NewHomeVisualTileRepository(db), provider, "https://shop.example.test")
				access := NewPublicUploadAccessService(nil, nil)
				access.ConfigureSiteLogoService(logos)
				access.ConfigureSiteFaviconService(favicons)
				access.ConfigureHomeVisualTileService(visuals)
				handler := NewObjectStorageCleanupOutboxHandler(nil, logos, visuals, nil)
				handler.ConfigureSiteFaviconService(favicons)
				resource := map[string]string{"site-logo": objectCleanupResourceSiteLogo, "site-favicon": objectCleanupResourceSiteFavicon, "visual-showcase": objectCleanupResourceHomeVisualTile}[namespace]
				event, err := newObjectStorageCleanupEvent(resource, "1", resource, "1", []string{key})
				require.NoError(t, err)
				detached := article
				detached.Content = "<p>Detached</p>"
				releaseChannel := make(chan struct{})
				var releaseOnce sync.Once
				release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
				defer release()
				if race == "uncommitted detachment" {
					locked := make(chan int, 1)
					var pauseOnce sync.Once
					callbackName := "pause_historical_post_detachment_" + uuid.NewString()
					require.NoError(t, db.Callback().Update().After("gorm:update").Register(callbackName, func(tx *gorm.DB) {
						if tx.Statement.Table == "posts" {
							pauseOnce.Do(func() {
								pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
							})
						}
					}))
					t.Cleanup(func() { _ = db.Callback().Update().Remove(callbackName) })
					result := make(chan error, 1)
					go func() { result <- posts.Update(&detached) }()
					select {
					case <-locked:
					case <-ctx.Done():
						t.Fatal("detachment did not reach its uncommitted database write")
					}
					allowed, err := access.CanServePublicUpload(ctx, key)
					require.NoError(t, err)
					require.True(t, allowed, "public reads observe the last committed consumer")
					require.ErrorIs(t, handler.Handle(ctx, *event), ErrMediaAssetInUse)
					requireDedicatedPublicObjectStillReadableAfterCleanupAttempt(t, provider, key)
					release()
					require.NoError(t, <-result)
					require.NoError(t, handler.Handle(ctx, *event), "the original cleanup succeeds after detachment commits")
				} else {
					require.NoError(t, posts.Update(&detached))
					deleting := make(chan struct{}, 1)
					pausedStorage := &historicalDedicatedObjectCleanupStorageWithDeletionPause{StorageService: provider, beforeDelete: func() error {
						deleting <- struct{}{}
						select {
						case <-releaseChannel:
							return nil
						case <-ctx.Done():
							return ctx.Err()
						}
					}}
					logos.storage, favicons.storage, visuals.storage = pausedStorage, pausedStorage, pausedStorage
					result := make(chan error, 1)
					go func() { result <- handler.Handle(ctx, *event) }()
					select {
					case <-deleting:
					case <-ctx.Done():
						t.Fatal("cleanup did not reach physical deletion under its namespace lock")
					}
					article.Title = "Stale editor snapshot"
					require.ErrorIs(t, posts.Update(&article), repository.ErrPostMediaReferenceInvalid, "an old editor cannot restore a detached dedicated URL")
					release()
					require.NoError(t, <-result)
				}
				allowed, err := access.CanServePublicUpload(ctx, key)
				require.NoError(t, err)
				require.False(t, allowed)
				_, err = os.Stat(objectPath)
				require.True(t, os.IsNotExist(err))
			})
		}
	}
}
