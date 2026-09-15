package service

import (
	"context"
	"encoding/json"
	"testing"

	"commerce-platform/internal/domain/media"
	ugcshowcasedomain "commerce-platform/internal/domain/ugcshowcase"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestPublicUploadAccessShowcaseNamespaceCannotBeOverriddenByMediaAsset(t *testing.T) {
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

	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &ugcshowcasedomain.UGCShowcase{}, &ugcshowcasedomain.UGCShowcaseComment{}))

	const key = "showcase/pending/2026/08/13/colliding.webp"
	imageReferences, err := json.Marshal([]string{key})
	require.NoError(t, err)

	require.NoError(t, db.Create(&media.MediaAsset{
		Filename:   "colliding.webp",
		URL:        "http://localhost:9200/uploads/" + key,
		StorageKey: key,
		MediaType:  "image",
		MimeType:   "image/webp",
		Status:     "active",
		Visibility: "public",
	}).Error)
	require.NoError(t, db.Create(&ugcshowcasedomain.UGCShowcase{
		UserID: 1,
		Kind:   ugcshowcasedomain.KindUser,
		Status: ugcshowcasedomain.StatusPending,
		Images: imageReferences,
	}).Error)

	mediaService := NewMediaService(repository.NewMediaRepository(db), nil, nil, "", 20<<30)
	ugcShowcaseService := NewUGCShowcaseService(repository.NewUGCShowcaseRepository(db), &fakeShowcaseStorage{})
	accessService := NewPublicUploadAccessService(mediaService, ugcShowcaseService)

	allowed, err := accessService.CanServePublicUpload(context.Background(), key)
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestPublicUploadAccessDeniesUnknownObject(t *testing.T) {
	accessService := NewPublicUploadAccessService(nil, nil)
	allowed, err := accessService.CanServePublicUpload(context.Background(), "misc/unregistered.png")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestPublicUploadAccessDeniesPrivateNamespacesEvenWhenRegistered(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}))
	key := "warranty/2026/09/14/claim.pdf"
	require.NoError(t, db.Create(&media.MediaAsset{
		Filename: "claim.pdf", URL: "https://private.example.test/" + key, StorageKey: key,
		MediaType: "video", MimeType: "application/pdf", Status: "active", Visibility: "public",
	}).Error)
	accessService := NewPublicUploadAccessService(NewMediaService(repository.NewMediaRepository(db), nil, nil, "", 20<<30), nil)
	allowed, err := accessService.CanServePublicUpload(context.Background(), key)
	require.NoError(t, err)
	require.False(t, allowed)
}
