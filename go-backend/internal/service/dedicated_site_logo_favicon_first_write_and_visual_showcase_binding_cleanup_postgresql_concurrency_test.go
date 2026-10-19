package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"commerce-platform/internal/domain/homevisualtile"
	"commerce-platform/internal/domain/outbox"
	settingdomain "commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx context.Context, tx *gorm.DB, locked chan<- int, release <-chan struct{}) {
	var pid int
	if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
		tx.AddError(err)
		return
	}
	locked <- pid
	select {
	case <-release:
	case <-ctx.Done():
		tx.AddError(ctx.Err())
	}
}

func requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t *testing.T, ctx context.Context, db *gorm.DB, blockerPID int) {
	t.Helper()
	require.Eventually(t, func() bool {
		var blocked bool
		err := db.WithContext(ctx).Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)))", blockerPID).Scan(&blocked).Error
		return err == nil && blocked
	}, 5*time.Second, 20*time.Millisecond, "the second operation must wait on the real database lock")
}

func TestPostgreSQLDedicatedSiteLogoAndFaviconConcurrentFirstUploadsRetainCleanupOfTheReplacedObject(t *testing.T) {
	for _, resource := range []string{"site_logo", "site_favicon"} {
		t.Run(resource, func(t *testing.T) {
			env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
			require.NoError(t, env.db.AutoMigrate(&sitelogodomain.Asset{}, &sitefavicondomain.Asset{}, &settingdomain.Setting{}))
			provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://media.example.test"})
			require.NoError(t, err)
			logos := NewSiteLogoService(repository.NewSiteLogoRepository(env.db), provider, "https://shop.example.test")
			favicons := NewSiteFaviconService(repository.NewSiteFaviconRepository(env.db), provider, "https://shop.example.test")
			logos.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
			logos.ConfigureSettingService(NewSettingService(repository.NewSettingRepository(env.db), nil, 60))
			favicons.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
			favicons.ConfigureSettingService(NewSettingService(repository.NewSettingRepository(env.db), nil, 60))
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			locked := make(chan int, 1)
			release := make(chan struct{})
			var releaseOnce, pauseOnce sync.Once
			releaseFirst := func() { releaseOnce.Do(func() { close(release) }) }
			t.Cleanup(releaseFirst)
			callbackName := "pause_dedicated_first_upload_" + uuid.NewString()
			require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
				if tx.Statement.Table == resource+"_assets" {
					pauseOnce.Do(func() {
						pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, release)
					})
				}
			}))
			t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callbackName) })
			type uploadResult struct {
				key string
				err error
			}
			first, second := make(chan uploadResult, 1), make(chan uploadResult, 1)
			firstFile := createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "first.webp")
			secondFile := createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "second.webp")
			uploadCurrent := func(firstUpload bool) uploadResult {
				file := secondFile
				if firstUpload {
					file = firstFile
				}
				if resource == "site_logo" {
					asset, err := logos.UploadCurrent(ctx, file, 11)
					if err != nil {
						return uploadResult{err: err}
					}
					return uploadResult{key: asset.StorageKey}
				}
				asset, err := favicons.UploadCurrent(ctx, file, 11)
				if err != nil {
					return uploadResult{err: err}
				}
				return uploadResult{key: asset.StorageKey}
			}
			go func() { first <- uploadCurrent(true) }()
			var blocker int
			select {
			case blocker = <-locked:
			case <-ctx.Done():
				t.Fatal("first upload did not reach the locked write")
			}
			go func() { second <- uploadCurrent(false) }()
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, env.db, blocker)
			releaseFirst()
			firstResult, secondResult := <-first, <-second
			require.NoError(t, firstResult.err)
			require.NoError(t, secondResult.err)
			require.NotEqual(t, firstResult.key, secondResult.key)
			var events []outbox.Event
			require.NoError(t, env.db.Where("aggregate_type = ?", resource).Find(&events).Error)
			require.Len(t, events, 1, "the second first-write contender must see the first committed owner")
			handler := NewObjectStorageCleanupOutboxHandler(nil, logos, nil, nil)
			handler.ConfigureSiteFaviconService(favicons)
			require.NoError(t, handler.Handle(ctx, events[0]))
			object, err := provider.(storage.ObjectOpener).Open(ctx, firstResult.key)
			require.Error(t, err)
			require.Nil(t, object)
			object, err = provider.(storage.ObjectOpener).Open(ctx, secondResult.key)
			require.NoError(t, err)
			require.NoError(t, object.ReadCloser.Close())
			if resource == "site_favicon" {
				current, err := favicons.repo.Current()
				require.NoError(t, err)
				require.Equal(t, secondResult.key, current.StorageKey)
				var setting settingdomain.Setting
				require.NoError(t, env.db.Where("key = ? AND locale = ?", "site_favicon", "en").First(&setting).Error)
				require.Equal(t, favicons.CanonicalPublicURL(current.URL), setting.Value)
			}
		})
	}
}

type dedicatedVisualShowcaseStorageDeletionPauseForPostgreSQLRace struct {
	storage.StorageService
	beforeDelete func() error
}

func (s *dedicatedVisualShowcaseStorageDeletionPauseForPostgreSQLRace) Open(ctx context.Context, key string) (*storage.StoredObject, error) {
	return s.StorageService.(storage.ObjectOpener).Open(ctx, key)
}

func (s *dedicatedVisualShowcaseStorageDeletionPauseForPostgreSQLRace) Delete(ctx context.Context, key string) error {
	if err := s.beforeDelete(); err != nil {
		return err
	}
	return s.StorageService.Delete(ctx, key)
}

func TestPostgreSQLVisualShowcaseOriginalAndThumbnailRebindingSerializeWithCleanupInBothOrders(t *testing.T) {
	for _, binding := range []string{"original", "thumbnail"} {
		for _, saveMode := range []string{"create slot", "update slot", "replace group"} {
			for _, firstOperation := range []string{"bind", "cleanup"} {
				t.Run(binding+"/"+saveMode+"/"+firstOperation+" first", func(t *testing.T) {
					env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
					require.NoError(t, env.db.AutoMigrate(&homevisualtile.Tile{}))
					service := NewHomeVisualTileService(repository.NewHomeVisualTileRepository(env.db), env.storage.StorageService)
					service.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(env.db))
					retired := uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
					input := retired
					if binding == "thumbnail" {
						input = uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
						input.ThumbnailURL = retired.ImageURL
					}
					var initial HomeVisualTileInput
					if saveMode == "update slot" {
						initial = uploadDedicatedVisualShowcaseImageAndCreateSlotInputForLifecycleTest(t, service, 1)
						_, err := service.SaveSingleVisualShowcaseAdminItem(t.Context(), "home-hero", "zh_cn", initial)
						require.NoError(t, err)
					}
					ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
					defer cancel()
					locked := make(chan int, 1)
					release := make(chan struct{})
					var releaseOnce, pauseOnce sync.Once
					releaseFirst := func() { releaseOnce.Do(func() { close(release) }) }
					t.Cleanup(releaseFirst)
					if firstOperation == "bind" {
						callbackName := "pause_dedicated_showcase_binding_" + uuid.NewString()
						pauseBinding := func(tx *gorm.DB) {
							if tx.Statement.Table == "visual_showcase_items" {
								pauseOnce.Do(func() {
									pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, release)
								})
							}
						}
						if saveMode == "update slot" {
							require.NoError(t, env.db.Callback().Update().After("gorm:update").Register(callbackName, pauseBinding))
							t.Cleanup(func() { _ = env.db.Callback().Update().Remove(callbackName) })
						} else {
							require.NoError(t, env.db.Callback().Create().After("gorm:create").Register(callbackName, pauseBinding))
							t.Cleanup(func() { _ = env.db.Callback().Create().Remove(callbackName) })
						}
					} else {
						cleanupBackendPID := make(chan int, 1)
						callbackName := "observe_dedicated_showcase_cleanup_backend_" + uuid.NewString()
						require.NoError(t, env.db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == "visual_showcase_items" {
								pauseOnce.Do(func() {
									var pid int
									if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
										tx.AddError(err)
										return
									}
									cleanupBackendPID <- pid
								})
							}
						}))
						t.Cleanup(func() { _ = env.db.Callback().Query().Remove(callbackName) })
						service.storage = &dedicatedVisualShowcaseStorageDeletionPauseForPostgreSQLRace{StorageService: service.storage, beforeDelete: func() error {
							// Cleanup is paused inside Delete, after its reference check, while
							// still holding the actual repository namespace transaction lock.
							select {
							case pid := <-cleanupBackendPID:
								locked <- pid
							case <-ctx.Done():
								return ctx.Err()
							}
							select {
							case <-release:
								return nil
							case <-ctx.Done():
								return ctx.Err()
							}
						}}
					}
					bindResult, cleanupResult := make(chan error, 1), make(chan error, 1)
					bind := func() {
						var err error
						if saveMode != "replace group" {
							_, err = service.SaveSingleVisualShowcaseAdminItem(ctx, "home-hero", "zh_cn", input)
						} else {
							_, err = service.ReplaceAdminItems(ctx, "home-hero", "zh_cn", []HomeVisualTileInput{input})
						}
						bindResult <- err
					}
					cleanup := func() {
						cleanupResult <- NewObjectStorageCleanupOutboxHandler(nil, nil, service, nil).cleanupHomeVisualTiles(ctx, []string{retired.StorageKey})
					}
					if firstOperation == "bind" {
						go bind()
					} else {
						go cleanup()
					}
					var blocker int
					select {
					case blocker = <-locked:
					case <-ctx.Done():
						t.Fatal("first operation did not acquire its lifecycle lock")
					}
					if firstOperation == "bind" {
						go cleanup()
					} else {
						go bind()
					}
					requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(t, ctx, env.db, blocker)
					releaseFirst()
					bindErr, cleanupErr := <-bindResult, <-cleanupResult
					require.NoError(t, cleanupErr)
					if firstOperation == "bind" {
						require.NoError(t, bindErr)
						requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, retired.StorageKey, true)
					} else {
						require.ErrorIs(t, bindErr, ErrHomeVisualTileImageInvalid)
						requireDedicatedVisualShowcaseObjectPhysicalExistence(t, service, retired.StorageKey, false)
						items, err := service.GetAdminItems("home-hero", "zh_cn")
						require.NoError(t, err)
						if saveMode == "update slot" {
							require.Len(t, items, 1)
							require.Equal(t, initial.StorageKey, items[0].StorageKey)
							require.Equal(t, initial.ThumbnailURL, items[0].ThumbnailURL)
						} else {
							require.Empty(t, items)
						}
					}
				})
			}
		}
	}
}
