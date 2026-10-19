package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"
)

func createWebsiteProfileRegisteredMediaBindingTestServices(t *testing.T) (*gorm.DB, *SettingService, *WebsiteProfileService, *MediaService) {
	t.Helper()
	db, settings := newTestSettingService(t)
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &media.MediaAssetDerivative{}))
	assets := newMediaDeleteTestService(t, db, t.TempDir())
	assets.siteURL = "https://shop.example.test"
	settings.ConfigureMediaService(assets)
	return db, settings, NewWebsiteProfileService(settings), assets
}

func createWebsiteProfileMediaSettingForBindingTest(key, value, locale string) setting.Setting {
	return setting.Setting{Key: key, Value: value, Locale: locale, Type: "string", Group: setting.WebsiteProfileGroup, IsPublic: true}
}

func TestWebsiteProfileCanonicalCrossPartitionAndTranslationReuseProtectsAssetsWithoutDuplicatingRegistrations(t *testing.T) {
	db, settings, profile, assets := createWebsiteProfileRegisteredMediaBindingTestServices(t)
	image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "catalog/profile.png")
	require.NoError(t, db.Model(image).Update("library_partition", media.MediaLibraryPartitionCatalog).Error)
	derivative := media.MediaAssetDerivative{MediaAssetID: image.ID, Preset: "thumbnail", URL: "https://media.example.test/uploads/derivatives/profile.webp", StorageKey: "derivatives/profile.webp"}
	require.NoError(t, db.Create(&derivative).Error)
	for _, locale := range []string{"en", "fr"} {
		result, err := profile.Update(setting.WebsiteProfileUpdateRequest{Locale: locale, Title: "Profile", AvatarURL: "/uploads/" + derivative.StorageKey, FactoryImageURL: "https://shop.example.test/uploads/" + image.StorageKey, StatementBody: `<img src="` + derivative.URL + `"><a href="/uploads/` + image.StorageKey + `#page=2">Manual</a>`})
		require.NoError(t, err)
		require.Equal(t, image.URL, result.AvatarURL)
		require.Equal(t, image.URL, result.FactoryImageURL)
		require.Contains(t, result.StatementBody, image.URL+"#page=2")
	}
	var count int64
	require.NoError(t, db.Model(&media.MediaAsset{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	report, err := assets.GetAssetReferences(image.ID)
	require.NoError(t, err)
	require.Equal(t, 4, report.Total, "two global images and one body per language")
	require.Equal(t, "website_profile_setting", report.References[0].ResourceType)
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)), ErrMediaAssetInUse)
	for _, locale := range []string{"en", "fr"} {
		_, err := profile.Update(setting.WebsiteProfileUpdateRequest{Locale: locale, Title: "Detached", StatementBody: ""})
		require.NoError(t, err)
	}
	stored, err := assets.repo.FindAssetByID(image.ID)
	require.NoError(t, err)
	require.Equal(t, media.MediaLibraryPartitionCatalog, stored.LibraryPartition)
	report, err = assets.GetAssetReferences(image.ID)
	require.NoError(t, err)
	require.Zero(t, report.Total)
	require.NoError(t, settings.BatchSet(nil))
	require.NoError(t, assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)))
}

func TestWebsiteProfileNewImageAndRichTextPlacementsRejectInvalidAssetsAndRollBackWholeForm(t *testing.T) {
	for _, field := range []string{"avatar", "factory", "image", "video", "poster", "file link"} {
		for _, scenario := range []string{"unregistered", "private", "inactive", "customer", "conversation", "retired", "wrong type", "ambiguous"} {
			if field == "file link" && scenario == "wrong type" {
				continue
			}
			t.Run(field+"/"+scenario, func(t *testing.T) {
				db, _, profile, _ := createWebsiteProfileRegisteredMediaBindingTestServices(t)
				image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/rejected-profile.png")
				if field == "video" {
					require.NoError(t, db.Model(image).Update("media_type", "video").Error)
				}
				switch scenario {
				case "unregistered":
					require.NoError(t, db.Unscoped().Delete(image).Error)
				case "private":
					require.NoError(t, db.Model(image).Update("visibility", "private").Error)
				case "inactive":
					require.NoError(t, db.Model(image).Update("status", "inactive").Error)
				case "customer":
					require.NoError(t, db.Model(image).Update("uploader_actor_type", media.MediaUploadActorCustomer).Error)
				case "conversation":
					require.NoError(t, db.Model(image).Update("upload_origin", media.MediaUploadOriginCustomerServiceConversation).Error)
				case "retired":
					require.NoError(t, db.Delete(image).Error)
				case "wrong type":
					mediaType := "video"
					if field == "video" {
						mediaType = "image"
					}
					require.NoError(t, db.Model(image).Update("media_type", mediaType).Error)
				case "ambiguous":
					other := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/other-profile.png")
					require.NoError(t, db.Model(other).Update("storage_key", image.StorageKey).Error)
				}
				request := setting.WebsiteProfileUpdateRequest{Locale: "en", Title: "Must roll back", StatementBody: "Before"}
				reference := "/uploads/" + image.StorageKey
				switch field {
				case "avatar":
					request.AvatarURL = reference
				case "factory":
					request.FactoryImageURL = reference
				default:
					request.StatementBody = createPostMarkupForMediaPlacementTest(field, reference)
				}
				_, err := profile.Update(request)
				require.ErrorIs(t, err, ErrSettingInvalid)
				var count int64
				require.NoError(t, db.Unscoped().Model(&setting.Setting{}).Count(&count).Error)
				require.Zero(t, count, "including first-write reservations and unrelated form fields")
			})
		}
	}
}

func TestWebsiteProfileUnchangedLegacyFieldsRemainReadableButNewPlacementsAndPublicationAreValidated(t *testing.T) {
	db, settings, profile, _ := createWebsiteProfileRegisteredMediaBindingTestServices(t)
	legacyImage := createWebsiteProfileMediaSettingForBindingTest(setting.WebsiteProfileKeyAvatarURL, "/uploads/legacy-avatar.png", "global")
	legacyBody := createWebsiteProfileMediaSettingForBindingTest(setting.WebsiteProfileKeyStatementBody, `<p>Before</p><img src="/uploads/legacy-body.png">`, "en")
	require.NoError(t, db.Create(&legacyImage).Error)
	require.NoError(t, db.Create(&legacyBody).Error)
	_, err := profile.Update(setting.WebsiteProfileUpdateRequest{Locale: "en", Title: "Edited", AvatarURL: legacyImage.Value, StatementBody: `<p>After</p><img src="/uploads/legacy-body.png">`})
	require.NoError(t, err)
	_, err = profile.Update(setting.WebsiteProfileUpdateRequest{Locale: "fr", AvatarURL: legacyImage.Value, StatementBody: legacyBody.Value})
	require.ErrorIs(t, err, ErrSettingInvalid, "copy to a new locale is a new placement")
	_, err = profile.Update(setting.WebsiteProfileUpdateRequest{Locale: "en", FactoryImageURL: legacyImage.Value})
	require.ErrorIs(t, err, ErrSettingInvalid, "moving to another field must be validated")
	legacyImage.Group = "disguised"
	require.ErrorIs(t, settings.BatchSet([]setting.Setting{legacyImage}), ErrSettingInvalid)
	hidden := createWebsiteProfileMediaSettingForBindingTest(setting.WebsiteProfileKeyFactoryImageURL, "/uploads/legacy-hidden.png", "hidden")
	hidden.IsPublic = false
	require.NoError(t, db.Select("*").Create(&hidden).Error)
	require.NoError(t, db.Model(&hidden).Update("is_public", false).Error)
	hidden.IsPublic = true
	require.ErrorIs(t, settings.BatchSet([]setting.Setting{hidden}), ErrSettingInvalid)
	duplicated := legacyBody
	duplicated.Value += `<img src="/uploads/legacy-body.png">`
	require.ErrorIs(t, settings.BatchSet([]setting.Setting{duplicated}), ErrSettingInvalid)
}

func TestWebsiteProfileGenericSingleBatchAndExistingTransactionWritersCannotBypassMediaValidation(t *testing.T) {
	for _, entry := range []string{"service set", "repository set", "batch", "existing transaction", "admin single", "admin batch"} {
		t.Run(entry, func(t *testing.T) {
			db, settings, _, _ := createWebsiteProfileRegisteredMediaBindingTestServices(t)
			value := createWebsiteProfileMediaSettingForBindingTest(setting.WebsiteProfileKeyAvatarURL, "/uploads/missing.png", "global")
			var err error
			switch entry {
			case "service set":
				err = settings.Set(value.Key, value.Value, value.Type, "other", value.Locale)
			case "repository set":
				err = settings.settingRepo.Set(&value)
				require.ErrorIs(t, err, repository.ErrSettingMediaReferenceInvalid)
				return
			case "batch":
				err = settings.BatchSet([]setting.Setting{value})
			case "existing transaction":
				err = db.Transaction(func(tx *gorm.DB) error {
					return settings.BatchSetTx(tx, []setting.Setting{createWebsiteProfileMediaSettingForBindingTest("title", "Rollback", "en"), value})
				})
			case "admin single":
				_, err = NewAdminSettingsService(settings).UpdateSetting(setting.UpdateSettingRequest{Key: value.Key, Value: value.Value, Group: "disguised", Locale: "global", IsPublic: true})
			case "admin batch":
				_, err = NewAdminSettingsService(settings).BatchUpdateSettings(setting.BatchUpdateSettingsRequest{Settings: []setting.UpdateSettingRequest{{Key: value.Key, Value: value.Value, Locale: "global", IsPublic: true}}})
			}
			require.ErrorIs(t, err, ErrSettingInvalid)
			var count int64
			require.NoError(t, db.Model(&setting.Setting{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
}

func TestWebsiteProfileReferenceScannerPagesExactAttributesAndProtectsHiddenAndSoftDeletedSettings(t *testing.T) {
	db, settings, _, assets := createWebsiteProfileRegisteredMediaBindingTestServices(t)
	image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/scanned-profile.png")
	for index := 0; index < 260; index++ {
		row := createWebsiteProfileMediaSettingForBindingTest(setting.WebsiteProfileKeyStatementBody, `<p>`+image.URL+`</p><img src="https://foreign.example.test/uploads/`+image.StorageKey+`"><script src="`+image.URL+`"></script>`, fmt.Sprintf("locale-%03d", index))
		require.NoError(t, db.Create(&row).Error)
	}
	retained := createWebsiteProfileMediaSettingForBindingTest(setting.WebsiteProfileKeyAvatarURL, image.URL, "global")
	require.NoError(t, db.Create(&retained).Error)
	require.NoError(t, settings.Delete(retained.Key, retained.Locale))
	hidden := createWebsiteProfileMediaSettingForBindingTest(setting.WebsiteProfileKeyStatementBody, `<video src="https://external.example.test/demo.webm" poster="/uploads/`+image.StorageKey+`"></video>`, "hidden")
	require.NoError(t, db.Create(&hidden).Error)
	require.NoError(t, db.Model(&hidden).Update("is_public", false).Error)
	report, err := assets.GetAssetReferences(image.ID)
	require.NoError(t, err)
	require.Equal(t, 2, report.Total)
	require.Equal(t, "retained_deleted_parent", report.References[0].UsageState)
	require.Contains(t, report.References[1].Label, "未公开")
	require.ErrorIs(t, assets.DeleteAsset(t.Context(), image.ID, MediaAssetDeleteConfirmation(image.ID)), ErrMediaAssetInUse)
}

func TestWebsiteProfileExternalAndStaticURLsRemainCompatibleAndBatchTransactionKeepsTrustedResolver(t *testing.T) {
	db, settings, profile, assets := createWebsiteProfileRegisteredMediaBindingTestServices(t)
	image := createRegisteredPublicImageForProductMediaBindingTest(t, db, "content/foreign-profile.png")
	foreign := "https://foreign.example.test/uploads/" + image.StorageKey
	result, err := profile.Update(setting.WebsiteProfileUpdateRequest{Locale: "en", AvatarURL: foreign, FactoryImageURL: "/company/factory.webp", StatementBody: `<a href="https://foreign.example.test/manual.pdf">File</a>`})
	require.NoError(t, err)
	require.Equal(t, foreign, result.AvatarURL)
	report, err := assets.GetAssetReferences(image.ID)
	require.NoError(t, err)
	require.Zero(t, report.Total)
	require.NoError(t, db.Transaction(func(tx *gorm.DB) error {
		return settings.BatchSetTx(tx, []setting.Setting{createWebsiteProfileMediaSettingForBindingTest(setting.WebsiteProfileKeyFactoryImageURL, "https://shop.example.test/uploads/"+image.StorageKey, "global")})
	}))
	stored, err := settings.Get(setting.WebsiteProfileKeyFactoryImageURL, "global")
	require.NoError(t, err)
	require.Equal(t, image.URL, stored.Value)
}
