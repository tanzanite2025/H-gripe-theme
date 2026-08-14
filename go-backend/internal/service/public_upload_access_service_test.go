package service

import (
	"context"
	"encoding/json"
	"testing"

	"commerce-platform/internal/domain/media"
	showcasedomain "commerce-platform/internal/domain/showcase"
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

	require.NoError(t, db.AutoMigrate(&media.MediaAsset{}, &showcasedomain.Showcase{}, &showcasedomain.Comment{}))

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
	require.NoError(t, db.Create(&showcasedomain.Showcase{
		UserID: 1,
		Kind:   showcasedomain.KindUser,
		Status: showcasedomain.StatusPending,
		Images: imageReferences,
	}).Error)

	mediaService := NewMediaService(repository.NewMediaRepository(db), nil, nil, "", 20<<30)
	showcaseService := NewShowcaseService(repository.NewShowcaseRepository(db), &fakeShowcaseStorage{})
	accessService := NewPublicUploadAccessService(mediaService, showcaseService)

	allowed, err := accessService.CanServePublicUpload(context.Background(), key)
	require.NoError(t, err)
	require.False(t, allowed)
}
