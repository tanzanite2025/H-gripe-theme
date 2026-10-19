package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
)

func TestPostgreSQLWebsiteProfileImagesAndRichTextCreateAndUpdateSerializeBindingAndRetirement(t *testing.T) {
	for _, field := range []string{"avatar", "factory", "image", "video", "poster", "file link"} {
		for _, mutation := range []string{"create", "update"} {
			for _, firstOperation := range []string{"bind", "retire"} {
				t.Run(field+"/"+mutation+"/"+firstOperation+" first", func(t *testing.T) {
					env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
					require.NoError(t, env.db.AutoMigrate(&setting.Setting{}))
					ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
					defer cancel()
					db := env.db.WithContext(ctx)
					assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
					assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
					settings := NewSettingService(repository.NewSettingRepository(db), nil, 0)
					settings.ConfigureMediaService(assets)
					asset := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/racing-profile.png")
					if field == "video" {
						require.NoError(t, db.Model(asset).Update("media_type", "video").Error)
					}
					key, locale, reference := setting.WebsiteProfileKeyStatementBody, "en", "/uploads/"+asset.StorageKey
					value := createPostMarkupForMediaPlacementTest(field, reference)
					if field == "avatar" {
						key, locale, value = setting.WebsiteProfileKeyAvatarURL, "global", reference
					}
					if field == "factory" {
						key, locale, value = setting.WebsiteProfileKeyFactoryImageURL, "global", reference
					}
					if mutation == "update" {
						require.NoError(t, settings.BatchSet([]setting.Setting{createWebsiteProfileMediaSettingForBindingTest(key, "", locale)}))
					}
					firstLocked, releaseFirst := make(chan int, 1), make(chan struct{})
					var pauseOnce, releaseOnce sync.Once
					release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
					defer release()
					pauseAfterAssetLock := func(tx *gorm.DB) {
						pauseOnce.Do(func() {
							var pid int
							if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
								tx.AddError(err)
								return
							}
							firstLocked <- pid
							select {
							case <-releaseFirst:
							case <-ctx.Done():
								tx.AddError(ctx.Err())
							}
						})
					}
					callbackName := "pause_website_profile_binding_retirement_" + uuid.NewString()
					if firstOperation == "bind" {
						require.NoError(t, db.Callback().Update().After("gorm:update").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == "settings" {
								pauseAfterAssetLock(tx)
							}
						}))
						defer db.Callback().Update().Remove(callbackName)
					} else {
						require.NoError(t, db.Callback().Delete().After("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == "media_assets" {
								pauseAfterAssetLock(tx)
							}
						}))
						defer db.Callback().Delete().Remove(callbackName)
					}
					bindReference := func() error {
						item := createWebsiteProfileMediaSettingForBindingTest(key, value, locale)
						// Exercise both existing public write contracts.
						if field == "avatar" {
							return settings.Set(item.Key, item.Value, item.Type, item.Group, item.Locale)
						}
						if field == "factory" {
							return db.Transaction(func(tx *gorm.DB) error { return settings.BatchSetTx(tx, []setting.Setting{item}) })
						}
						return settings.BatchSet([]setting.Setting{item})
					}
					bindResult, retireResult := make(chan error, 1), make(chan error, 1)
					bind := func() { bindResult <- bindReference() }
					retire := func() { retireResult <- assets.DeleteAsset(ctx, asset.ID, MediaAssetDeleteConfirmation(asset.ID)) }
					if firstOperation == "bind" {
						go bind()
					} else {
						go retire()
					}
					var blockerPID int
					select {
					case blockerPID = <-firstLocked:
					case <-ctx.Done():
						t.Fatal("first operation did not acquire the media lock")
					}
					if firstOperation == "bind" {
						go retire()
					} else {
						go bind()
					}
					require.Eventually(t, func() bool {
						var waiting bool
						err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)) AND query LIKE '%media_assets%')", blockerPID).Scan(&waiting).Error
						return err == nil && waiting
					}, 4*time.Second, 20*time.Millisecond)
					release()
					var bindError, retireError error
					select {
					case bindError = <-bindResult:
					case <-ctx.Done():
						t.Fatal("setting binding timed out")
					}
					select {
					case retireError = <-retireResult:
					case <-ctx.Done():
						t.Fatal("setting retirement timed out")
					}
					var cleanupEvents int64
					require.NoError(t, db.Model(&outbox.Event{}).Where("event_type = ? AND aggregate_id = ?", outbox.EventTypeObjectStorageCleanup, fmt.Sprint(asset.ID)).Count(&cleanupEvents).Error)
					if firstOperation == "bind" {
						require.NoError(t, bindError)
						require.ErrorIs(t, retireError, ErrMediaAssetInUse)
						require.Zero(t, cleanupEvents)
					} else {
						require.NoError(t, retireError)
						require.ErrorIs(t, bindError, ErrSettingInvalid)
						require.EqualValues(t, 1, cleanupEvents)
						var registrations int64
						require.NoError(t, db.Unscoped().Model(&media.MediaAsset{}).Where("id = ?", asset.ID).Count(&registrations).Error)
						require.Zero(t, registrations)
						require.ErrorIs(t, bindReference(), ErrSettingInvalid, "physical cleanup cannot allow URL-only rebinding")
						var rows []setting.Setting
						require.NoError(t, db.Find(&rows).Error)
						if mutation == "create" {
							require.Empty(t, rows)
						} else {
							require.Len(t, rows, 1)
							require.Empty(t, rows[0].Value)
						}
					}
				})
			}
		}
	}
}

func TestPostgreSQLWebsiteProfileImageAndRichTextBatchesAcquireOneCombinedAscendingAssetLockPlan(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, env.db.AutoMigrate(&setting.Setting{}))
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	db := env.db.WithContext(ctx)
	settings := NewSettingService(repository.NewSettingRepository(db), nil, 0)
	settings.ConfigureMediaService(NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30))
	low := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/profile-low.png")
	high := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/profile-high.png")
	firstLocked, releaseFirst := make(chan int, 1), make(chan struct{})
	var pauseOnce, releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
	defer release()
	callbackName := "pause_website_profile_combined_asset_locks_" + uuid.NewString()
	require.NoError(t, db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table != "media_assets" || !strings.Contains(tx.Statement.SQL.String(), "FOR UPDATE") {
			return
		}
		pauseOnce.Do(func() {
			asset, ok := tx.Statement.Dest.(*media.MediaAsset)
			if !ok || asset.ID != low.ID {
				tx.AddError(fmt.Errorf("lowest asset must be locked before higher image and rich text assets"))
				return
			}
			var pid int
			if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
				tx.AddError(err)
				return
			}
			firstLocked <- pid
			select {
			case <-releaseFirst:
			case <-ctx.Done():
				tx.AddError(ctx.Err())
			}
		})
	}))
	defer db.Callback().Query().Remove(callbackName)
	first, second := make(chan error, 1), make(chan error, 1)
	go func() {
		first <- settings.BatchSet([]setting.Setting{
			createWebsiteProfileMediaSettingForBindingTest("avatar_url", high.URL, "first"),
			createWebsiteProfileMediaSettingForBindingTest("factory_image_url", low.URL, "first"),
			createWebsiteProfileMediaSettingForBindingTest("statement_body", `<img src="`+high.URL+`"><a href="`+low.URL+`">File</a>`, "first"),
		})
	}()
	var blockerPID int
	select {
	case blockerPID = <-firstLocked:
	case <-ctx.Done():
		t.Fatal("first batch did not acquire its asset lock")
	}
	go func() {
		second <- settings.BatchSet([]setting.Setting{
			createWebsiteProfileMediaSettingForBindingTest("statement_body", `<img src="`+low.URL+`"><a href="`+high.URL+`">File</a>`, "second"),
			createWebsiteProfileMediaSettingForBindingTest("factory_image_url", high.URL, "second"),
			createWebsiteProfileMediaSettingForBindingTest("avatar_url", low.URL, "second"),
		})
	}()
	require.Eventually(t, func() bool {
		var waiting bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)) AND query LIKE '%media_assets%')", blockerPID).Scan(&waiting).Error
		return err == nil && waiting
	}, 4*time.Second, 20*time.Millisecond)
	release()
	for _, result := range []chan error{first, second} {
		select {
		case err := <-result:
			require.NoError(t, err)
		case <-ctx.Done():
			t.Fatal("combined media plan deadlocked")
		}
	}
}

func TestPostgreSQLWebsiteProfileConcurrentFirstWritesAndReverseInputBatchesSerializeOnSettingKeys(t *testing.T) {
	for _, initialState := range []string{"missing", "existing"} {
		t.Run(initialState, func(t *testing.T) {
			env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
			require.NoError(t, env.db.AutoMigrate(&setting.Setting{}))
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			db := env.db.WithContext(ctx)
			settings := NewSettingService(repository.NewSettingRepository(db), nil, 0)
			settings.ConfigureMediaService(NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30))
			image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/first-write-profile.png")
			if initialState == "existing" {
				require.NoError(t, settings.BatchSet([]setting.Setting{createWebsiteProfileMediaSettingForBindingTest("avatar_url", "", "global"), createWebsiteProfileMediaSettingForBindingTest("factory_image_url", "", "global")}))
			}
			firstLocked, releaseFirst := make(chan int, 1), make(chan struct{})
			var pauseOnce, releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseFirst) }) }
			defer release()
			callbackName := "pause_website_profile_setting_key_locks_" + uuid.NewString()
			require.NoError(t, db.Callback().Query().After("gorm:query").Register(callbackName, func(tx *gorm.DB) {
				if tx.Error != nil || tx.Statement.Table != "settings" || !strings.Contains(tx.Statement.SQL.String(), "FOR NO KEY UPDATE") {
					return
				}
				pauseOnce.Do(func() {
					row, ok := tx.Statement.Dest.(*setting.Setting)
					if !ok || row.Key != "avatar_url" {
						tx.AddError(fmt.Errorf("setting keys must lock in sorted order independently of request order"))
						return
					}
					var pid int
					if err := tx.Session(&gorm.Session{NewDB: true}).Raw("SELECT pg_backend_pid()").Scan(&pid).Error; err != nil {
						tx.AddError(err)
						return
					}
					firstLocked <- pid
					select {
					case <-releaseFirst:
					case <-ctx.Done():
						tx.AddError(ctx.Err())
					}
				})
			}))
			defer db.Callback().Query().Remove(callbackName)
			first, second := make(chan error, 1), make(chan error, 1)
			go func() {
				first <- settings.BatchSet([]setting.Setting{createWebsiteProfileMediaSettingForBindingTest("factory_image_url", image.URL, "global"), createWebsiteProfileMediaSettingForBindingTest("avatar_url", image.URL, "global")})
			}()
			var blockerPID int
			select {
			case blockerPID = <-firstLocked:
			case <-ctx.Done():
				t.Fatal("first batch did not lock its setting")
			}
			go func() {
				second <- settings.BatchSet([]setting.Setting{createWebsiteProfileMediaSettingForBindingTest("avatar_url", image.URL, "global"), createWebsiteProfileMediaSettingForBindingTest("factory_image_url", image.URL, "global")})
			}()
			require.Eventually(t, func() bool {
				var waiting bool
				err := db.Raw("SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE ? = ANY(pg_blocking_pids(pid)) AND query LIKE '%settings%')", blockerPID).Scan(&waiting).Error
				return err == nil && waiting
			}, 4*time.Second, 20*time.Millisecond)
			release()
			for _, result := range []chan error{first, second} {
				select {
				case err := <-result:
					require.NoError(t, err)
				case <-ctx.Done():
					t.Fatal("first writes or reversed setting order deadlocked")
				}
			}
			var rows []setting.Setting
			require.NoError(t, db.Find(&rows).Error)
			require.Len(t, rows, 2)
			for _, row := range rows {
				require.Equal(t, image.URL, row.Value)
			}
		})
	}
}
