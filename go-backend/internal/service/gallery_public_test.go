package service

import (
	"testing"

	gallerydomain "commerce-platform/internal/domain/gallery"
	mediadomain "commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"

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
	require.NoError(t, db.Create(&gallerydomain.GalleryImage{
		GalleryID: publishedGallery.ID,
		URL:       "/published-detail.jpg",
		Title:     "Published detail",
		Order:     1,
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
	assert.EqualValues(t, 2, galleries[0].ImageCount)
	require.Len(t, galleries[0].Images, 1)
	assert.Equal(t, "/published.jpg", galleries[0].Images[0].URL)

	_, err = galleryService.GetPublicImagesByGalleryID(draftGallery.ID)
	require.ErrorIs(t, err, ErrGalleryNotFound)

	images, err := galleryService.GetPublicImagesByGalleryID(publishedGallery.ID)
	require.NoError(t, err)
	require.Len(t, images, 2)
	assert.Equal(t, "/published.jpg", images[0].URL)
}

func TestGalleryServiceAdminImageMutationsRequireGalleryOwnership(t *testing.T) {
	db, galleryService := newTestGalleryService(t)

	galleryA := gallerydomain.Gallery{Name: "Gallery A", Slug: "gallery-a", Status: "published"}
	galleryB := gallerydomain.Gallery{Name: "Gallery B", Slug: "gallery-b", Status: "published"}
	require.NoError(t, db.Create(&galleryA).Error)
	require.NoError(t, db.Create(&galleryB).Error)

	imageA := gallerydomain.GalleryImage{
		GalleryID: galleryA.ID,
		URL:       "/gallery-a.jpg",
		Title:     "Original",
	}
	imageB := gallerydomain.GalleryImage{
		GalleryID: galleryB.ID,
		URL:       "/gallery-b.jpg",
		Title:     "Other",
	}
	require.NoError(t, db.Create(&imageA).Error)
	require.NoError(t, db.Create(&imageB).Error)

	updatedTitle := "Wrong gallery update"
	_, err := galleryService.UpdateAdminGalleryImageForGallery(galleryB.ID, imageA.ID, GalleryImageAdminUpdateInput{
		Title: &updatedTitle,
	})
	require.True(t, IsRecordNotFound(err))

	var unchanged gallerydomain.GalleryImage
	require.NoError(t, db.First(&unchanged, imageA.ID).Error)
	assert.Equal(t, "Original", unchanged.Title)

	err = galleryService.DeleteGalleryImageForGallery(galleryB.ID, imageA.ID)
	require.True(t, IsRecordNotFound(err))
	require.NoError(t, db.First(&unchanged, imageA.ID).Error)

	deleted, err := galleryService.BatchDeleteGalleryImages(galleryB.ID, []uint{imageB.ID, imageA.ID})
	require.True(t, IsRecordNotFound(err))
	assert.EqualValues(t, 0, deleted)
	require.NoError(t, db.First(&unchanged, imageA.ID).Error)
	require.NoError(t, db.First(&gallerydomain.GalleryImage{}, imageB.ID).Error)

	deleted, err = galleryService.BatchDeleteGalleryImages(galleryB.ID, []uint{imageB.ID})
	require.NoError(t, err)
	assert.EqualValues(t, 1, deleted)
	require.Error(t, db.First(&gallerydomain.GalleryImage{}, imageB.ID).Error)
}

func TestGalleryServiceAdminProductLinksOnlyAllowActiveProducts(t *testing.T) {
	db, galleryService := newTestGalleryService(t)
	enableGalleryProductLinkTestTables(t, db)

	require.NoError(t, db.Exec(`
		INSERT INTO products (id, sku, name, slug, locale, status, price, currency, display_prices, deleted_at)
		VALUES
			(101, 'ACTIVE-101', 'Active Product', 'active-product', 'en', 'active', 100, 'USD', '[]', NULL),
			(202, 'INACTIVE-202', 'Inactive Product', 'inactive-product', 'en', 'inactive', 100, 'USD', '[]', NULL)
	`).Error)

	created, err := galleryService.CreateAdminGallery(GalleryAdminCreateInput{
		Title:      "Linked Gallery",
		Slug:       "linked-gallery",
		ProductIDs: []uint{101},
	})
	require.NoError(t, err)
	require.Len(t, created.ProductLinks, 1)
	assert.EqualValues(t, 101, created.ProductLinks[0].ProductID)
	require.NotNil(t, created.ProductLinks[0].Product)
	assert.Equal(t, "active", created.ProductLinks[0].Product.Status)

	_, err = galleryService.CreateAdminGallery(GalleryAdminCreateInput{
		Title:      "Inactive Linked Gallery",
		Slug:       "inactive-linked-gallery",
		ProductIDs: []uint{202},
	})
	require.True(t, IsRecordNotFound(err))

	productIDs := []uint{202}
	_, err = galleryService.UpdateAdminGallery(created.ID, GalleryAdminUpdateInput{
		ProductIDs: &productIDs,
	})
	require.True(t, IsRecordNotFound(err))
}

func TestGalleryServiceAdminCreationIncludesMediaImages(t *testing.T) {
	db, galleryService := newTestGalleryService(t)
	enableGalleryProductLinkTestTables(t, db)

	asset := mediadomain.MediaAsset{
		Filename:         "brand-gallery.jpg",
		OriginalFilename: "brand-gallery.jpg",
		URL:              "https://cdn.example.test/uploads/brand-gallery.jpg",
		StorageKey:       "brand-gallery.jpg",
		MediaType:        "image",
		Status:           "active",
		Visibility:       "public",
		Alt:              "Brand gallery",
	}
	require.NoError(t, db.Create(&asset).Error)

	created, err := galleryService.CreateAdminGallery(GalleryAdminCreateInput{
		Title: "Brand Gallery",
		Slug:  "brand-gallery",
		Images: []GalleryImageAdminCreateInput{{
			MediaAssetID: asset.ID,
			Title:        "Brand gallery image",
			Order:        2,
		}},
	})
	require.NoError(t, err)
	require.Len(t, created.Images, 1)
	require.NotNil(t, created.Images[0].MediaAssetID)
	assert.Equal(t, asset.ID, *created.Images[0].MediaAssetID)
	assert.Equal(t, asset.URL, created.Images[0].URL)
	assert.Equal(t, "Brand gallery image", created.Images[0].Title)
	assert.Equal(t, 2, created.Images[0].Order)
}

func TestGalleryServiceAdminCreationRejectsMissingMediaBeforePersisting(t *testing.T) {
	db, galleryService := newTestGalleryService(t)
	enableGalleryProductLinkTestTables(t, db)

	_, err := galleryService.CreateAdminGallery(GalleryAdminCreateInput{
		Title: "Invalid Gallery",
		Slug:  "invalid-gallery",
		Images: []GalleryImageAdminCreateInput{{
			MediaAssetID: 999,
			Title:        "Missing image",
		}},
	})
	require.ErrorIs(t, err, ErrGalleryMediaAssetNotFound)

	var galleryCount int64
	require.NoError(t, db.Model(&gallerydomain.Gallery{}).Count(&galleryCount).Error)
	assert.Zero(t, galleryCount)
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

	require.NoError(t, db.AutoMigrate(
		&gallerydomain.Gallery{},
		&gallerydomain.GalleryImage{},
		&mediadomain.MediaAsset{},
	))

	return db, NewGalleryService(
		repository.NewGalleryRepository(db),
		repository.NewMediaRepository(db),
	)
}

func enableGalleryProductLinkTestTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	require.NoError(t, db.Exec(`
		CREATE TABLE products (
			id INTEGER PRIMARY KEY,
			sku TEXT,
			name TEXT,
			slug TEXT,
			locale TEXT,
			status TEXT,
			price REAL NOT NULL DEFAULT 0,
			currency TEXT NOT NULL DEFAULT 'USD',
			display_prices TEXT NOT NULL DEFAULT '[]',
			deleted_at DATETIME
		)
	`).Error)
	require.NoError(t, db.AutoMigrate(&gallerydomain.GalleryProductLink{}))
}
