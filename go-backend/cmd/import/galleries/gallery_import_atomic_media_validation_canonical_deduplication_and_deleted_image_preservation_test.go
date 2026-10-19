package main

import (
	"encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func createGalleryImportTransactionalMediaTestRepository(t *testing.T) (*gorm.DB, *repository.GalleryRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&gallery.Gallery{}, &gallery.GalleryImage{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}))
	files, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	repo := repository.NewGalleryRepository(db)
	repo.ConfigurePublicBusinessMediaReferenceIdentityResolver(service.NewMediaService(repository.NewMediaRepository(db), files, nil, "https://shop.example.test", 20<<30))
	return db, repo
}

func decodeGalleryImportFixture(t *testing.T, body string) *GalleryExportData {
	t.Helper()
	var data GalleryExportData
	require.NoError(t, json.Unmarshal([]byte(body), &data))
	return &data
}

func TestGalleryImportCanonicalIdentityDeduplicatesAliasesAndPreservesDeletedImages(t *testing.T) {
	db, repo := createGalleryImportTransactionalMediaTestRepository(t)
	asset := media.MediaAsset{Filename: "photo.png", URL: "https://media.example.test/uploads/content/import-photo.png", StorageKey: "content/import-photo.png", MediaType: "image", Status: "active", Visibility: "public"}
	require.NoError(t, db.Create(&asset).Error)
	derivative := media.MediaAssetDerivative{MediaAssetID: asset.ID, Preset: "thumbnail", StorageKey: "derivatives/import-photo.webp", URL: "https://media.example.test/uploads/derivatives/import-photo.webp"}
	require.NoError(t, db.Create(&derivative).Error)
	data := decodeGalleryImportFixture(t, `{"galleries":[{"id":1,"name":"Imported","slug":"imported","cover_image":"/uploads/content/import-photo.png"}],"images":[{"gallery_id":1,"url":"/uploads/derivatives/import-photo.webp?v=1","title":"First"},{"gallery_id":1,"url":"https://shop.example.test/uploads/content/import-photo.png","title":"Same asset"}]}`)
	stats := importGalleriesAndAppendImagesThroughTransactionalMediaRepository(db, repo, data)
	require.Equal(t, 1, stats.GalleriesCreated)
	require.Equal(t, 1, stats.ImagesCreated)
	require.Equal(t, 1, stats.ImagesSkipped)
	var item gallery.Gallery
	require.NoError(t, db.First(&item).Error)
	require.Equal(t, asset.URL, item.CoverImage)
	var image gallery.GalleryImage
	require.NoError(t, db.First(&image).Error)
	require.Equal(t, &asset.ID, image.MediaAssetID)
	require.Equal(t, asset.URL, image.URL)
	stats = importGalleriesAndAppendImagesThroughTransactionalMediaRepository(db, repo, data)
	require.Zero(t, stats.ImagesCreated)
	require.Equal(t, 2, stats.ImagesSkipped)
	require.NoError(t, repo.DeleteGalleryImage(image.ID))
	stats = importGalleriesAndAppendImagesThroughTransactionalMediaRepository(db, repo, data)
	require.Zero(t, stats.ImagesCreated)
	var count int64
	require.NoError(t, db.Unscoped().Model(&gallery.GalleryImage{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
	// Old URL-only placements also deduplicate against a different registered alias.
	legacy := gallery.Gallery{Name: "Legacy", Slug: "legacy"}
	require.NoError(t, db.Create(&legacy).Error)
	require.NoError(t, db.Create(&gallery.GalleryImage{GalleryID: legacy.ID, URL: derivative.URL}).Error)
	legacyData := decodeGalleryImportFixture(t, `{"galleries":[{"id":2,"name":"Legacy","slug":"legacy"}],"images":[{"gallery_id":2,"url":"/uploads/content/import-photo.png"}]}`)
	stats = importGalleriesAndAppendImagesThroughTransactionalMediaRepository(db, repo, legacyData)
	require.Zero(t, stats.ImagesCreated)
	require.Equal(t, 1, stats.ImagesSkipped)
	var assets int64
	require.NoError(t, db.Model(&media.MediaAsset{}).Count(&assets).Error)
	require.EqualValues(t, 1, assets)
}

func TestGalleryImportInvalidNewImageRollsBackGalleryCoverMetadataAndAllAppendedImages(t *testing.T) {
	db, repo := createGalleryImportTransactionalMediaTestRepository(t)
	item := gallery.Gallery{Name: "Original", Slug: "rollback", CoverImage: "/uploads/legacy-cover.png"}
	require.NoError(t, db.Create(&item).Error)
	data := decodeGalleryImportFixture(t, `{"galleries":[{"id":1,"name":"Must roll back","slug":"rollback","cover_image":"https://external.example.test/new.png"}],"images":[{"gallery_id":1,"url":"https://external.example.test/valid.png"},{"gallery_id":1,"url":"/uploads/unregistered.png"}]}`)
	stats := importGalleriesAndAppendImagesThroughTransactionalMediaRepository(db, repo, data)
	require.Equal(t, 1, stats.GalleriesSkipped)
	require.Equal(t, 2, stats.ImagesSkipped)
	var stored gallery.Gallery
	require.NoError(t, db.First(&stored, item.ID).Error)
	require.Equal(t, item.Name, stored.Name)
	require.Equal(t, item.CoverImage, stored.CoverImage)
	var count int64
	require.NoError(t, db.Model(&gallery.GalleryImage{}).Count(&count).Error)
	require.Zero(t, count)
	data.Galleries[0].Slug = "new-rollback"
	stats = importGalleriesAndAppendImagesThroughTransactionalMediaRepository(db, repo, data)
	require.Equal(t, 1, stats.GalleriesSkipped)
	require.NoError(t, db.Unscoped().Model(&gallery.Gallery{}).Count(&count).Error)
	require.EqualValues(t, 1, count)
}
