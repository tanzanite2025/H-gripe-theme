package service

import (
	"testing"

	gallerydomain "tanzanite/internal/domain/gallery"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestGalleryServicePublicAccessOnlyReturnsPublishedGalleries(t *testing.T) {
	db, galleryService := newTestGalleryService(t)

	publishedGallery := gallerydomain.Gallery{Name: "Published", Slug: "published", Status: "published"}
	require.NoError(t, db.Create(&publishedGallery).Error)
	require.NoError(t, db.Create(&gallerydomain.GalleryImage{
		GalleryID: publishedGallery.ID,
		URL:       "/published.jpg",
		Title:     "Published",
	}).Error)

	draftGallery := gallerydomain.Gallery{Name: "Draft", Slug: "draft", Status: "draft"}
	require.NoError(t, db.Create(&draftGallery).Error)
	require.NoError(t, db.Create(&gallerydomain.GalleryImage{
		GalleryID: draftGallery.ID,
		URL:       "/draft.jpg",
		Title:     "Draft",
	}).Error)

	_, err := galleryService.GetPublicGalleryByID(draftGallery.ID)
	require.ErrorIs(t, err, ErrGalleryNotFound)

	galleries, total, err := galleryService.GetPublicGalleries(1, 20)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, galleries, 1)
	assert.Equal(t, publishedGallery.ID, galleries[0].ID)

	_, err = galleryService.GetPublicImagesByGalleryID(draftGallery.ID)
	require.ErrorIs(t, err, ErrGalleryNotFound)

	images, err := galleryService.GetPublicImagesByGalleryID(publishedGallery.ID)
	require.NoError(t, err)
	require.Len(t, images, 1)
	assert.Equal(t, "/published.jpg", images[0].URL)
}

func newTestGalleryService(t *testing.T) (*gorm.DB, *GalleryService) {
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

	require.NoError(t, db.AutoMigrate(&gallerydomain.Gallery{}, &gallerydomain.GalleryImage{}))

	return db, NewGalleryService(repository.NewGalleryRepository(db))
}
