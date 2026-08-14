package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/showcase"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShowcaseHandlerListReturnsControlledImageFileReferences(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	require.NoError(t, db.AutoMigrate(&showcase.Showcase{}, &showcase.Comment{}))
	require.NoError(t, db.Create(&showcase.Showcase{
		UserID: 1,
		Kind:   showcase.KindUser,
		Status: showcase.StatusPending,
		Images: datatypes.JSON(`[
			"showcase/pending/2026/08/13/private.webp",
			"http://storage.internal/uploads/showcase/pending/2026/08/13/second.webp"
		]`),
	}).Error)

	handler := NewShowcaseHandler(service.NewShowcaseService(repository.NewShowcaseRepository(db), nil))
	router := gin.New()
	router.GET("/showcase", handler.List)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/showcase?type=user&status=pending", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "showcase/pending")
	require.NotContains(t, recorder.Body.String(), "storage.internal")

	var body struct {
		Data struct {
			Items []struct {
				GalleryImages []string `json:"gallery_images"`
				ImageFiles    []struct {
					Index   int    `json:"index"`
					FileURL string `json:"file_url"`
				} `json:"image_files"`
				ImageCount int `json:"image_count"`
			} `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Len(t, body.Data.Items, 1)
	require.Equal(t, []string{
		"/api/admin/showcase/1/images/0/file",
		"/api/admin/showcase/1/images/1/file",
	}, body.Data.Items[0].GalleryImages)
	require.Equal(t, 2, body.Data.Items[0].ImageCount)
	require.Equal(t, "/api/admin/showcase/1/images/0/file", body.Data.Items[0].ImageFiles[0].FileURL)
}
