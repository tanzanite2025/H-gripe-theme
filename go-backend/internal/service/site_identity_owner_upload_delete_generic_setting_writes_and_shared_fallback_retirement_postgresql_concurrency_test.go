package service

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/repository"
)

func TestPostgreSQLSiteIdentityDedicatedMutationsAndGenericStaleSettingWritesSerializeInBothOrders(t *testing.T) {
	for _, key := range []string{setting.SiteLogoSettingKey, setting.SiteFaviconSettingKey} {
		for _, mutation := range []string{"upload", "delete"} {
			for _, first := range []string{"dedicated", "generic"} {
				t.Run(key+"/"+mutation+"/"+first+" first", func(t *testing.T) {
					env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
					require.NoError(t, env.db.AutoMigrate(&setting.Setting{}, &sitelogodomain.Asset{}, &sitefavicondomain.Asset{}))
					ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
					defer cancel()
					db := env.db.WithContext(ctx)
					settings := NewSettingService(repository.NewSettingRepository(db), nil, 0)
					settings.ConfigureMediaService(NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30))
					logos := NewSiteLogoService(repository.NewSiteLogoRepository(db), env.storage.StorageService, "https://shop.example.test")
					logos.ConfigureSettingService(settings)
					logos.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
					favicons := NewSiteFaviconService(repository.NewSiteFaviconRepository(db), env.storage.StorageService, "https://shop.example.test")
					favicons.ConfigureSettingService(settings)
					favicons.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
					file := createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "identity.webp")
					upload := func() error {
						if key == setting.SiteLogoSettingKey {
							_, err := logos.UploadCurrent(ctx, file, 11)
							return err
						}
						_, err := favicons.UploadCurrent(ctx, file, 11)
						return err
					}
					require.NoError(t, upload())
					previous, err := settings.Get(key, "en")
					require.NoError(t, err)
					locked, releaseChannel := make(chan int, 1), make(chan struct{})
					var pauseOnce, releaseOnce sync.Once
					release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
					defer release()
					pause := func(tx *gorm.DB) {
						pauseOnce.Do(func() {
							pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
						})
					}
					callbackName := "pause_site_identity_mirror_race_" + uuid.NewString()
					if first == "generic" {
						require.NoError(t, db.Callback().Update().After("gorm:update").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == "settings" {
								pause(tx)
							}
						}))
						defer db.Callback().Update().Remove(callbackName)
					} else if mutation == "upload" {
						require.NoError(t, db.Callback().Create().After("gorm:create").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == key+"_assets" {
								pause(tx)
							}
						}))
						defer db.Callback().Create().Remove(callbackName)
					} else {
						require.NoError(t, db.Callback().Delete().After("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
							if tx.Statement.Table == key+"_assets" {
								pause(tx)
							}
						}))
						defer db.Callback().Delete().Remove(callbackName)
					}
					dedicatedResult, genericResult := make(chan error, 1), make(chan error, 1)
					dedicated := func() {
						if mutation == "upload" {
							dedicatedResult <- upload()
						} else if key == setting.SiteLogoSettingKey {
							dedicatedResult <- logos.DeleteCurrent(ctx)
						} else {
							dedicatedResult <- favicons.DeleteCurrent(ctx)
						}
					}
					generic := func() { genericResult <- settings.BatchSet([]setting.Setting{*previous}) }
					if first == "dedicated" {
						go dedicated()
					} else {
						go generic()
					}
					var blocker int
					select {
					case blocker = <-locked:
					case <-ctx.Done():
						t.Fatal("first site identity transaction did not reach its locked write")
					}
					if first == "dedicated" {
						go generic()
					} else {
						go dedicated()
					}
					requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, db, blocker)
					release()
					require.NoError(t, <-dedicatedResult)
					genericError := <-genericResult
					if first == "dedicated" {
						require.ErrorIs(t, genericError, ErrSettingInvalid)
					} else {
						require.NoError(t, genericError)
					}
					mirror, err := settings.Get(key, "en")
					require.NoError(t, err)
					if mutation == "delete" {
						require.Empty(t, mirror.Value)
					} else {
						require.NotEqual(t, previous.Value, mirror.Value)
						if key == setting.SiteLogoSettingKey {
							require.Equal(t, logos.CurrentPublicURL(), mirror.Value)
						} else {
							require.Equal(t, favicons.CurrentPublicURL(), mirror.Value)
						}
					}
				})
			}
		}
	}
}

func TestPostgreSQLSiteLogoSharedFallbackBindingAndRetirementUseTheSameAssetLock(t *testing.T) {
	for _, first := range []string{"bind", "retire"} {
		t.Run(first+" first", func(t *testing.T) {
			env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
			require.NoError(t, env.db.AutoMigrate(&setting.Setting{}))
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			db := env.db.WithContext(ctx)
			assets := NewMediaService(repository.NewMediaRepository(db), env.storage.StorageService, nil, "https://shop.example.test", 20<<30)
			assets.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
			settings := NewSettingService(repository.NewSettingRepository(db), nil, 0)
			settings.ConfigureMediaService(assets)
			image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/fallback-race.png")
			locked, releaseChannel := make(chan int, 1), make(chan struct{})
			var pauseOnce, releaseOnce sync.Once
			release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
			defer release()
			pause := func(tx *gorm.DB) {
				pauseOnce.Do(func() {
					pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
				})
			}
			callbackName := "pause_logo_fallback_retirement_" + uuid.NewString()
			if first == "bind" {
				require.NoError(t, db.Callback().Update().After("gorm:update").Register(callbackName, func(tx *gorm.DB) {
					if tx.Statement.Table == "settings" {
						pause(tx)
					}
				}))
				defer db.Callback().Update().Remove(callbackName)
			} else {
				require.NoError(t, db.Callback().Delete().After("gorm:delete").Register(callbackName, func(tx *gorm.DB) {
					if tx.Statement.Table == "media_assets" {
						pause(tx)
					}
				}))
				defer db.Callback().Delete().Remove(callbackName)
			}
			bindResult, retireResult := make(chan error, 1), make(chan error, 1)
			bind := func() {
				bindResult <- settings.Set(setting.SiteLogoSettingKey, "/uploads/"+image.StorageKey, "string", "site", "en")
			}
			retire := func() { retireResult <- assets.DeleteAsset(ctx, image.ID, MediaAssetDeleteConfirmation(image.ID)) }
			if first == "bind" {
				go bind()
			} else {
				go retire()
			}
			var blocker int
			select {
			case blocker = <-locked:
			case <-ctx.Done():
				t.Fatal("first fallback transaction did not reach its asset lock")
			}
			if first == "bind" {
				go retire()
			} else {
				go bind()
			}
			requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, db, blocker)
			release()
			bindError, retireError := <-bindResult, <-retireResult
			if first == "bind" {
				require.NoError(t, bindError)
				require.ErrorIs(t, retireError, ErrMediaAssetInUse)
			} else {
				require.NoError(t, retireError)
				require.ErrorIs(t, bindError, ErrSettingInvalid)
			}
		})
	}
}

func TestPostgreSQLSiteIdentitySettingBatchesWithReversedInputAcquireNamespacesAndLocaleRowsInTheSameOrder(t *testing.T) {
	env, _, _ := createIsolatedPostgreSQLCustomerServiceAttachmentRetirementTestEnvironment(t)
	require.NoError(t, env.db.AutoMigrate(&setting.Setting{}))
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	db := env.db.WithContext(ctx)
	firstValues := []setting.Setting{{Key: setting.SiteLogoSettingKey, Locale: "en", Group: "site", IsPublic: true}, {Key: setting.SiteFaviconSettingKey, Locale: "en", Group: "site", IsPublic: true}}
	require.NoError(t, db.Create(&firstValues).Error)
	secondValues := []setting.Setting{firstValues[1], firstValues[0]}
	for index := range firstValues {
		firstValues[index].Description = "First"
		secondValues[index].Description = "Second"
	}
	locked, releaseChannel := make(chan int, 1), make(chan struct{})
	var pauseOnce, releaseOnce sync.Once
	release := func() { releaseOnce.Do(func() { close(releaseChannel) }) }
	defer release()
	callbackName := "pause_reversed_site_identity_batches_" + uuid.NewString()
	require.NoError(t, db.Callback().Update().After("gorm:update").Register(callbackName, func(tx *gorm.DB) {
		if tx.Statement.Table == "settings" {
			pauseOnce.Do(func() {
				pauseDedicatedObjectLifecyclePostgreSQLTransactionAndReportBackendPID(ctx, tx, locked, releaseChannel)
			})
		}
	}))
	defer db.Callback().Update().Remove(callbackName)
	settings := NewSettingService(repository.NewSettingRepository(db), nil, 0)
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	go func() { firstResult <- settings.BatchSet(firstValues) }()
	var blocker int
	select {
	case blocker = <-locked:
	case <-ctx.Done():
		t.Fatal("first identity batch did not acquire its locks")
	}
	go func() { secondResult <- settings.BatchSet(secondValues) }()
	requireDedicatedObjectLifecyclePostgreSQLTransactionActuallyBlocks(ctx, t, db, blocker)
	release()
	require.NoError(t, <-firstResult)
	require.NoError(t, <-secondResult)
	for _, key := range []string{setting.SiteLogoSettingKey, setting.SiteFaviconSettingKey} {
		row, err := settings.Get(key, "en")
		require.NoError(t, err)
		require.Equal(t, "Second", row.Description)
	}
}
