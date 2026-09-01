package service

import (
	"context"
	"errors"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMediaUploadQuotaCountsSoftDeletedAssets(t *testing.T) {
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

	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}))
	existing := media.MediaAsset{
		Filename:         "existing.jpg",
		URL:              "https://media.example.test/uploads/existing.jpg",
		StorageKey:       "existing.jpg",
		MediaType:        "image",
		Size:             6,
		UploaderID:       42,
		Status:           "active",
		Visibility:       "public",
		OriginalFilename: "existing.jpg",
	}
	require.NoError(t, db.Create(&existing).Error)
	require.NoError(t, db.Delete(&existing).Error)

	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: t.TempDir(),
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)

	service := NewMediaService(repository.NewMediaRepository(db), storageService, nil, "", 10)
	_, err = service.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "next.jpg", "image/jpeg", []byte("12345")),
		MediaType:  "image",
		UploaderID: 42,
	})
	require.ErrorIs(t, err, ErrMediaAccountStorageQuotaExceeded)
}

func TestMediaUploadQuotaAllowsUsageAtExactLimit(t *testing.T) {
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

	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}))
	require.NoError(t, db.Create(&media.MediaAsset{
		Filename:         "existing.jpg",
		URL:              "https://media.example.test/uploads/existing.jpg",
		StorageKey:       "existing.jpg",
		MediaType:        "image",
		Size:             6,
		UploaderID:       42,
		Status:           "active",
		Visibility:       "public",
		OriginalFilename: "existing.jpg",
	}).Error)

	storageService, err := storage.NewStorageService(&storage.Config{
		Type:      storage.StorageTypeLocal,
		LocalPath: t.TempDir(),
		BaseURL:   "https://media.example.test",
	})
	require.NoError(t, err)

	payload := mediaUploadTestPNG(t)
	service := NewMediaService(repository.NewMediaRepository(db), storageService, nil, "", 6+int64(len(payload)))
	asset, err := service.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "next.png", "image/png", payload),
		MediaType:  "image",
		UploaderID: 42,
	})
	require.NoError(t, err)
	require.EqualValues(t, len(payload), asset.Size)

	_, err = service.UploadAsset(context.Background(), MediaUploadInput{
		File:       multipartFileHeader(t, "overflow.jpg", "image/jpeg", []byte("1")),
		MediaType:  "image",
		UploaderID: 42,
	})
	require.True(t, errors.Is(err, ErrMediaAccountStorageQuotaExceeded))
}
