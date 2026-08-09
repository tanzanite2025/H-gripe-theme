package service

import (
	"errors"
	"testing"

	"tanzanite/internal/domain/media"
	"tanzanite/internal/pkg/ugc"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCanonicalPublicImageUploadURLAcceptsMatchingPublicImageAsset(t *testing.T) {
	service := newMediaAssetAccessTestService(t, media.MediaAsset{
		Filename:         "photo.jpg",
		URL:              "https://media.example.test/uploads/2026/08/02/photo.jpg",
		StorageKey:       "2026/08/02/photo.jpg",
		MimeType:         "image/jpeg",
		MediaType:        "image",
		UploaderID:       42,
		Status:           "active",
		Visibility:       "public",
		OriginalFilename: "photo.jpg",
	})

	got, err := service.CanonicalPublicImageUploadURL("https://media.example.test/uploads/2026/08/02/photo.jpg?cache=1")
	require.NoError(t, err)
	require.Equal(t, "https://media.example.test/uploads/2026/08/02/photo.jpg", got)
}

func TestCanonicalPublicImageUploadURLRejectsUntrustedHost(t *testing.T) {
	service := newMediaAssetAccessTestService(t, media.MediaAsset{
		Filename:         "photo.jpg",
		URL:              "https://media.example.test/uploads/photo.jpg",
		StorageKey:       "photo.jpg",
		MimeType:         "image/jpeg",
		MediaType:        "image",
		UploaderID:       42,
		Status:           "active",
		Visibility:       "public",
		OriginalFilename: "photo.jpg",
	})

	_, err := service.CanonicalPublicImageUploadURL("https://evil.example.test/uploads/photo.jpg")
	require.True(t, errors.Is(err, ugc.ErrAttachmentInvalidURL))
}

func TestCanonicalPublicImageUploadURLRejectsInactiveOrNonImageAsset(t *testing.T) {
	inactive := newMediaAssetAccessTestService(t, media.MediaAsset{
		Filename:         "photo.jpg",
		URL:              "https://media.example.test/uploads/photo.jpg",
		StorageKey:       "photo.jpg",
		MimeType:         "image/jpeg",
		MediaType:        "image",
		UploaderID:       42,
		Status:           "archived",
		Visibility:       "public",
		OriginalFilename: "photo.jpg",
	})
	_, err := inactive.CanonicalPublicImageUploadURL("/uploads/photo.jpg")
	require.ErrorIs(t, err, ErrMediaAssetForbidden)

	video := newMediaAssetAccessTestService(t, media.MediaAsset{
		Filename:         "clip.jpg",
		URL:              "https://media.example.test/uploads/clip.jpg",
		StorageKey:       "clip.jpg",
		MimeType:         "video/mp4",
		MediaType:        "video",
		UploaderID:       42,
		Status:           "active",
		Visibility:       "public",
		OriginalFilename: "clip.jpg",
	})
	_, err = video.CanonicalPublicImageUploadURL("/uploads/clip.jpg")
	require.ErrorIs(t, err, ErrUnsupportedMediaType)
}

func newMediaAssetAccessTestService(t *testing.T, assets ...media.MediaAsset) *MediaService {
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

	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}))
	for _, asset := range assets {
		require.NoError(t, db.Create(&asset).Error)
	}
	return NewMediaService(repository.NewMediaRepository(db), nil, nil, "", 20<<30)
}
