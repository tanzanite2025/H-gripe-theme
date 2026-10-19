package admin

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"commerce-platform/internal/domain/gallery"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestGalleryMediaHTTPRejectsIneligibleSelectionsAndPreservesOmittedImageFields(t *testing.T) {
	db, _, assets := createPostManagedMediaHTTPContractTestHandler(t)
	require.NoError(t, db.AutoMigrate(&gallery.Gallery{}, &gallery.GalleryImage{}))
	repo := repository.NewGalleryRepository(db)
	galleries := service.NewGalleryService(repo)
	galleries.ConfigureMediaService(assets)
	handler := NewGalleryHandler(galleries)
	item := &gallery.Gallery{Name: "HTTP", Slug: "http"}
	require.NoError(t, repo.CreateGallery(item))
	asset := media.MediaAsset{Filename: "http.png", URL: "https://media.example.test/uploads/content/gallery-http.png", StorageKey: "content/gallery-http.png", MediaType: "image", Status: "active", Visibility: "private"}
	require.NoError(t, db.Create(&asset).Error)
	for _, scenario := range []struct {
		name, body string
		status     int
	}{
		{"missing selection", `{"title":"Image"}`, http.StatusBadRequest},
		{"unknown selection", `{"title":"Image","media_asset_id":999}`, http.StatusNotFound},
		{"private selection", fmt.Sprintf(`{"title":"Image","media_asset_id":%d}`, asset.ID), http.StatusBadRequest},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(recorder)
			context.Params = gin.Params{{Key: "id", Value: fmt.Sprint(item.ID)}}
			context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/galleries/1/images", strings.NewReader(scenario.body))
			context.Request.Header.Set("Content-Type", "application/json")
			handler.CreateImage(context)
			require.Equal(t, scenario.status, recorder.Code, recorder.Body.String())
			var count int64
			require.NoError(t, db.Unscoped().Model(&gallery.GalleryImage{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
	legacy := gallery.GalleryImage{GalleryID: item.ID, URL: "/uploads/historical.png", Thumbnail: "/uploads/historical-thumb.png"}
	require.NoError(t, db.Create(&legacy).Error)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Params = gin.Params{{Key: "id", Value: fmt.Sprint(item.ID)}, {Key: "imageId", Value: fmt.Sprint(legacy.ID)}}
	context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/galleries/1/images/1", strings.NewReader(`{"title":"Ordinary edit"}`))
	context.Request.Header.Set("Content-Type", "application/json")
	handler.UpdateImage(context)
	require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
	var stored gallery.GalleryImage
	require.NoError(t, db.First(&stored, legacy.ID).Error)
	require.Equal(t, legacy.URL, stored.URL)
	require.Equal(t, legacy.Thumbnail, stored.Thumbnail)
}
