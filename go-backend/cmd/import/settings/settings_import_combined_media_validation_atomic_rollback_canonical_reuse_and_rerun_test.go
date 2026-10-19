package main

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func TestSettingsImportValidatesEntireFileAtomicallyAndRerunsWithoutDuplicateAssets(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&setting.Setting{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}))
	files, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	repo := repository.NewSettingRepository(db)
	assets := service.NewMediaService(repository.NewMediaRepository(db), files, nil, "https://shop.example.test", 20<<30)
	repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(assets)
	image := media.MediaAsset{Filename: "profile.png", URL: "https://media.example.test/uploads/content/profile.png", StorageKey: "content/profile.png", Status: "active", Visibility: "public", MediaType: "image"}
	require.NoError(t, db.Create(&image).Error)
	input := []SettingImport{
		{Key: "title", Value: "Imported", Group: "website_profile", Locale: "en", IsPublic: true},
		{Key: "avatar_url", Value: "/uploads/" + image.StorageKey, Group: "website_profile", Locale: "global", IsPublic: true},
		{Key: "statement_body", Value: `<a href="/uploads/missing.pdf">Manual</a>`, Group: "website_profile", Locale: "en", IsPublic: true},
	}
	_, err = importSettingsWithAtomicRegisteredMediaValidation(repo, input)
	require.ErrorIs(t, err, repository.ErrSettingMediaReferenceInvalid)
	var count int64
	require.NoError(t, db.Unscoped().Model(&setting.Setting{}).Count(&count).Error)
	require.Zero(t, count)
	input[2].Value = `<img src="https://shop.example.test/uploads/` + image.StorageKey + `">`
	for iteration := 0; iteration < 2; iteration++ {
		_, err := importSettingsWithAtomicRegisteredMediaValidation(repo, input)
		require.NoError(t, err)
	}
	stored, err := repo.Get("avatar_url", "global")
	require.NoError(t, err)
	require.Equal(t, image.URL, stored.Value)
	require.NoError(t, db.Model(&setting.Setting{}).Count(&count).Error)
	require.EqualValues(t, 3, count)
	require.NoError(t, db.Model(&media.MediaAsset{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	// Preserve the original importer's Set contract for existing metadata.
	require.NoError(t, db.Model(&setting.Setting{}).Where("key = ? AND locale = ?", "title", "en").Updates(map[string]interface{}{"is_public": false, "description": "Keep existing metadata"}).Error)
	input[0].Description = "Do not overwrite"
	_, err = importSettingsWithAtomicRegisteredMediaValidation(repo, input)
	require.NoError(t, err)
	stored, err = repo.Get("title", "en")
	require.NoError(t, err)
	require.False(t, stored.IsPublic)
	require.Equal(t, "Keep existing metadata", stored.Description)
	report, err := assets.GetAssetReferences(image.ID)
	require.NoError(t, err)
	require.Equal(t, 2, report.Total)
	input[0].Value, input[1].Value = "Must roll back", "/uploads/new-missing.png"
	_, err = importSettingsWithAtomicRegisteredMediaValidation(repo, input)
	require.ErrorIs(t, err, repository.ErrSettingMediaReferenceInvalid)
	stored, err = repo.Get("title", "en")
	require.NoError(t, err)
	require.Equal(t, "Imported", stored.Value)
	identityInput := []SettingImport{{Key: "site_logo", Value: "/uploads/" + image.StorageKey, Group: "site", Locale: "en", IsPublic: true}}
	_, err = importSettingsWithAtomicRegisteredMediaValidation(repo, identityInput)
	require.NoError(t, err)
	logo, err := repo.Get("site_logo", "en")
	require.NoError(t, err)
	require.Equal(t, image.URL, logo.Value)
	identityInput[0].Value = "https://external.example.test/replacement.png"
	identityInput = append(identityInput, SettingImport{Key: "site_favicon", Value: "/uploads/" + image.StorageKey, Group: "site", Locale: "en", IsPublic: true})
	_, err = importSettingsWithAtomicRegisteredMediaValidation(repo, identityInput)
	require.ErrorIs(t, err, repository.ErrSettingMediaReferenceInvalid)
	logo, err = repo.Get("site_logo", "en")
	require.NoError(t, err)
	require.Equal(t, image.URL, logo.Value, "a blocked favicon import rolls back the full file")
}
