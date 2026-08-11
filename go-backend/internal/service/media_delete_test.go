package service

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/product"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMediaDeleteAssetPermanentlyRemovesUnreferencedAsset(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{})
	uploadRoot := t.TempDir()
	service := newMediaDeleteTestService(t, db, uploadRoot)

	asset, err := service.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "unused.jpg", "image/jpeg", []byte("unused media bytes")),
		MediaType:  "image",
		UploaderID: 42,
	})
	require.NoError(t, err)

	storedPath := filepath.Join(uploadRoot, filepath.FromSlash(asset.StorageKey))
	_, err = os.Stat(storedPath)
	require.NoError(t, err)

	err = service.DeleteAsset(context.Background(), asset.ID, "DELETE 0")
	require.ErrorIs(t, err, ErrMediaDeleteConfirmationRequired)
	_, err = os.Stat(storedPath)
	require.NoError(t, err)

	require.NoError(t, service.DeleteAsset(context.Background(), asset.ID, MediaAssetDeleteConfirmation(asset.ID)))
	_, err = os.Stat(storedPath)
	require.True(t, errors.Is(err, os.ErrNotExist))

	var assetCount int64
	require.NoError(t, db.Unscoped().Model(&media.MediaAsset{}).Where("id = ?", asset.ID).Count(&assetCount).Error)
	require.Zero(t, assetCount)

	usage, err := repository.NewMediaRepository(db).AssetStorageUsageByUploaderID(42)
	require.NoError(t, err)
	require.Zero(t, usage)
}

func TestMediaDeleteAssetBlocksReferencedProductMedia(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &product.Product{}, &product.ProductMedia{})
	uploadRoot := t.TempDir()
	service := newMediaDeleteTestService(t, db, uploadRoot)

	asset, err := service.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "product.jpg", "image/jpeg", []byte("product media bytes")),
		MediaType:  "image",
		UploaderID: 42,
	})
	require.NoError(t, err)

	item := product.Product{
		SKU:   "MEDIA-DELETE-TEST",
		Name:  "Media delete test product",
		Slug:  "media-delete-test-product",
		Price: 1,
	}
	require.NoError(t, db.Create(&item).Error)
	assetID := asset.ID
	require.NoError(t, db.Create(&product.ProductMedia{
		ProductID:    item.ID,
		MediaAssetID: &assetID,
		MediaType:    "image",
		Role:         "primary",
		URL:          asset.URL,
		IsPrimary:    true,
		IsVisible:    true,
	}).Error)

	report, err := service.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
	require.Equal(t, "product_media", report.References[0].ResourceType)
	require.Contains(t, report.References[0].Field, "media_asset_id")
	require.Contains(t, report.References[0].Field, "url")

	err = service.DeleteAsset(context.Background(), asset.ID, MediaAssetDeleteConfirmation(asset.ID))
	require.ErrorIs(t, err, ErrMediaAssetInUse)
	var inUse *MediaAssetInUseError
	require.ErrorAs(t, err, &inUse)
	require.Len(t, inUse.References, 1)

	storedPath := filepath.Join(uploadRoot, filepath.FromSlash(asset.StorageKey))
	_, err = os.Stat(storedPath)
	require.NoError(t, err)

	var assetCount int64
	require.NoError(t, db.Unscoped().Model(&media.MediaAsset{}).Where("id = ?", asset.ID).Count(&assetCount).Error)
	require.EqualValues(t, 1, assetCount)
}

func TestMediaDeleteAssetBlocksReferencedProductTypeImage(t *testing.T) {
	db := newMediaDeleteTestDB(t, &media.MediaAsset{}, &product.ProductType{})
	uploadRoot := t.TempDir()
	service := newMediaDeleteTestService(t, db, uploadRoot)

	asset, err := service.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "category.webp", "image/webp", []byte("category media bytes")),
		MediaType:  "image",
		UploaderID: 42,
	})
	require.NoError(t, err)

	assetID := asset.ID
	require.NoError(t, db.Create(&product.ProductType{
		Name:              "Category image reference",
		Slug:              "category-image-reference",
		ImageMediaAssetID: &assetID,
		ImageURL:          asset.URL,
		IsEnabled:         true,
	}).Error)

	report, err := service.GetAssetReferences(asset.ID)
	require.NoError(t, err)
	require.Equal(t, 1, report.Total)
	require.Equal(t, "product_type", report.References[0].ResourceType)
	require.Contains(t, report.References[0].Field, "image_media_asset_id")
	require.Contains(t, report.References[0].Field, "image_url")

	err = service.DeleteAsset(context.Background(), asset.ID, MediaAssetDeleteConfirmation(asset.ID))
	require.ErrorIs(t, err, ErrMediaAssetInUse)
}

func newMediaDeleteTestDB(t *testing.T, models ...interface{}) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(models...))
	return db
}

func newMediaDeleteTestService(t *testing.T, db *gorm.DB, uploadRoot string) *MediaService {
	t.Helper()

	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: uploadRoot,
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)

	return NewMediaService(repository.NewMediaRepository(db), storageService, nil, "", 20<<30)
}
