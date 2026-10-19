package service

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func createSiteIdentityOwnerAndSettingMirrorAtomicityTestServices(t *testing.T) (*gorm.DB, *SiteLogoService, *SiteFaviconService, string) {
	t.Helper()
	db := newSiteLogoTestDB(t)
	require.NoError(t, db.AutoMigrate(&sitefavicondomain.Asset{}, &media.MediaAsset{}))
	root := t.TempDir()
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: root, BaseURL: "https://media.example.test"})
	require.NoError(t, err)
	settings := NewSettingService(repository.NewSettingRepository(db), nil, 0)
	logos := NewSiteLogoService(repository.NewSiteLogoRepository(db), provider, "https://shop.example.test")
	logos.ConfigureSettingService(settings)
	logos.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	favicons := NewSiteFaviconService(repository.NewSiteFaviconRepository(db), provider, "https://shop.example.test")
	favicons.ConfigureSettingService(settings)
	favicons.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	return db, logos, favicons, root
}

func TestSiteIdentityOwnerMirrorsSynchronizeAllActiveLocalesPreserveMetadataAndClearWithoutAnOwner(t *testing.T) {
	for _, key := range []string{setting.SiteLogoSettingKey, setting.SiteFaviconSettingKey} {
		t.Run(key, func(t *testing.T) {
			db, logos, favicons, _ := createSiteIdentityOwnerAndSettingMirrorAtomicityTestServices(t)
			for _, locale := range []string{"en", "zh_cn"} {
				row := setting.Setting{Key: key, Locale: locale, Value: "/uploads/legacy.png", Group: "legacy-site", Type: "string", Description: "Keep description", IsPublic: true}
				require.NoError(t, db.Create(&row).Error)
				require.NoError(t, db.Model(&row).Update("is_public", false).Error)
			}
			file := createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "current.webp")
			var currentURL string
			if key == setting.SiteLogoSettingKey {
				owner, err := logos.UploadCurrent(t.Context(), file, 11, "fr")
				require.NoError(t, err)
				currentURL = owner.URL
			} else {
				owner, err := favicons.UploadCurrent(t.Context(), file, 11, "fr")
				require.NoError(t, err)
				currentURL = owner.URL
			}
			var mirrors []setting.Setting
			require.NoError(t, db.Where("key = ?", key).Order("locale").Find(&mirrors).Error)
			require.Len(t, mirrors, 3)
			for _, mirror := range mirrors {
				require.Equal(t, currentURL, mirror.Value)
				if mirror.Locale != "fr" {
					require.False(t, mirror.IsPublic)
					require.Equal(t, "legacy-site", mirror.Group)
					require.Equal(t, "Keep description", mirror.Description)
				} else {
					require.True(t, mirror.IsPublic)
				}
			}
			for iteration := 0; iteration < 2; iteration++ {
				if key == setting.SiteLogoSettingKey {
					require.NoError(t, logos.DeleteCurrent(t.Context(), "fr"))
				} else {
					require.NoError(t, favicons.DeleteCurrent(t.Context(), "fr"))
				}
				mirrors = nil
				require.NoError(t, db.Where("key = ?", key).Find(&mirrors).Error)
				for _, mirror := range mirrors {
					require.Empty(t, mirror.Value)
				}
				if iteration == 0 {
					require.NoError(t, db.Model(&setting.Setting{}).Where("key = ?", key).Update("value", "/uploads/stale-fallback.png").Error)
				}
			}
		})
	}
}

func TestSiteIdentityOwnerMirrorAndOutboxFailuresRollBackMetadataAndOnlyCompensateNewUploads(t *testing.T) {
	for _, key := range []string{setting.SiteLogoSettingKey, setting.SiteFaviconSettingKey} {
		for _, operation := range []string{"first upload", "replace", "delete"} {
			for _, failure := range []string{"mirror write", "outbox write", "deleted locale"} {
				if operation == "first upload" && failure == "outbox write" {
					continue
				}
				t.Run(key+"/"+operation+"/"+failure, func(t *testing.T) {
					db, logos, favicons, root := createSiteIdentityOwnerAndSettingMirrorAtomicityTestServices(t)
					upload := func() error {
						file := createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "current.webp")
						if key == setting.SiteLogoSettingKey {
							_, err := logos.UploadCurrent(t.Context(), file, 11, "en")
							return err
						}
						_, err := favicons.UploadCurrent(t.Context(), file, 11, "en")
						return err
					}
					previousURL := ""
					if operation != "first upload" {
						require.NoError(t, upload())
						if key == setting.SiteLogoSettingKey {
							previousURL = logos.CurrentPublicURL()
						} else {
							previousURL = favicons.CurrentPublicURL()
						}
					}
					switch failure {
					case "deleted locale":
						row := setting.Setting{Key: key, Locale: "en"}
						require.NoError(t, db.Where("key = ? AND locale = ?", key, "en").Delete(&setting.Setting{}).Error)
						if operation == "first upload" {
							require.NoError(t, db.Create(&row).Error)
							require.NoError(t, db.Delete(&row).Error)
						}
					case "mirror write":
						require.NoError(t, db.Callback().Update().Before("gorm:update").Register("reject_site_identity_mirror", func(tx *gorm.DB) {
							if tx.Statement.Table == "settings" {
								tx.AddError(errors.New("injected mirror failure"))
							}
						}))
						t.Cleanup(func() { _ = db.Callback().Update().Remove("reject_site_identity_mirror") })
					default:
						require.NoError(t, db.Callback().Create().Before("gorm:create").Register("reject_site_identity_outbox", func(tx *gorm.DB) {
							if tx.Statement.Table == "outbox_events" {
								tx.AddError(errors.New("injected outbox failure"))
							}
						}))
						t.Cleanup(func() { _ = db.Callback().Create().Remove("reject_site_identity_outbox") })
					}
					var err error
					if operation != "delete" {
						err = upload()
					} else if key == setting.SiteLogoSettingKey {
						err = logos.DeleteCurrent(t.Context(), "en")
					} else {
						err = favicons.DeleteCurrent(t.Context(), "en")
					}
					require.Error(t, err)
					if key == setting.SiteLogoSettingKey {
						require.Equal(t, previousURL, logos.CurrentPublicURL())
					} else {
						require.Equal(t, previousURL, favicons.CurrentPublicURL())
					}
					var events int64
					require.NoError(t, db.Model(&outbox.Event{}).Count(&events).Error)
					if operation != "delete" && failure != "outbox write" {
						require.EqualValues(t, 1, events, "only the failed upload compensation survives the owner rollback")
					} else {
						require.Zero(t, events)
					}
					files := 0
					require.NoError(t, filepath.WalkDir(root, func(_ string, entry fs.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if !entry.IsDir() {
							files++
						}
						return nil
					}))
					expectedFiles := 0
					if operation != "first upload" {
						expectedFiles++
					}
					if operation != "delete" && failure != "outbox write" {
						expectedFiles++
					}
					require.Equal(t, expectedFiles, files, "failed prepared uploads retain bytes for their original receipt; persistence failure prevents any new write")
					var mirrors []setting.Setting
					require.NoError(t, db.Where("key = ?", key).Find(&mirrors).Error)
					for _, mirror := range mirrors {
						require.Equal(t, previousURL, mirror.Value)
					}
					if failure == "deleted locale" {
						require.Empty(t, mirrors, "a deleted unique setting is not silently revived")
					}
				})
			}
		}
	}
}

func TestSiteIdentityGenericWritesCannotOverrideDedicatedOwnersOrImportFaviconValues(t *testing.T) {
	db, logos, favicons, _ := createSiteIdentityOwnerAndSettingMirrorAtomicityTestServices(t)
	settings := NewSettingService(repository.NewSettingRepository(db), nil, 0)
	_, err := logos.UploadCurrent(t.Context(), createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "logo.webp"), 11)
	require.NoError(t, err)
	_, err = favicons.UploadCurrent(t.Context(), createSiteLogoAndFaviconLifecycleMultipartHeaderFromBackendOwnedWebPFixture(t, "favicon.webp"), 11)
	require.NoError(t, err)
	for _, key := range []string{setting.SiteLogoSettingKey, setting.SiteFaviconSettingKey} {
		current, err := settings.Get(key, "en")
		require.NoError(t, err)
		require.NoError(t, settings.BatchSet([]setting.Setting{*current}), "saving the unchanged settings form remains compatible")
		for _, value := range []string{"", "https://external.example.test/new.png"} {
			for _, locale := range []string{"en", "fr"} {
				require.ErrorIs(t, settings.Set(key, value, "string", "site", locale), ErrSettingInvalid)
			}
		}
		items := []setting.Setting{{Key: "site_title", Value: "Must roll back", Locale: "en", Group: "site"}, {Key: key, Value: "/uploads/new.png", Locale: "en", Group: "site"}}
		require.ErrorIs(t, repository.NewSettingRepository(db).SetMultipleValuesPreservingExistingVisibilityAndDescription(items), repository.ErrSettingMediaReferenceInvalid)
		var titleCount int64
		require.NoError(t, db.Model(&setting.Setting{}).Where("key = ?", "site_title").Count(&titleCount).Error)
		require.Zero(t, titleCount)
	}
}
