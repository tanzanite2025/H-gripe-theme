package showcase

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	showcasedomain "commerce-platform/internal/domain/showcase"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestNormalizeShowcasePaginationClampsBounds(t *testing.T) {
	page, perPage := normalizeShowcasePagination(-10, 999)
	if page != 1 || perPage != showcaseMaxPerPage {
		t.Fatalf("normalizeShowcasePagination(-10, 999) = (%d, %d), want (1, %d)", page, perPage, showcaseMaxPerPage)
	}

	page, perPage = normalizeShowcasePagination(showcaseMaxPage+1, 0)
	if page != showcaseMaxPage || perPage != 1 {
		t.Fatalf("normalizeShowcasePagination(max+1, 0) = (%d, %d), want (%d, 1)", page, perPage, showcaseMaxPage)
	}
}

func TestShowcasePublicListReturnsControlledImageURLs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&showcasedomain.Showcase{}, &showcasedomain.Comment{}))
	require.NoError(t, db.Create(&showcasedomain.Showcase{
		UserID: 1,
		Kind:   showcasedomain.KindUser,
		Status: showcasedomain.StatusApproved,
		Images: datatypes.JSON(`["showcase/approved/2026/08/13/approved.webp"]`),
	}).Error)

	handler := NewShowcaseHandler(service.NewShowcaseService(repository.NewShowcaseRepository(db), nil))
	router := gin.New()
	router.GET("/showcase/gallery", handler.List)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/showcase/gallery?type=user", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "showcase/approved")
	require.NotContains(t, recorder.Body.String(), "storage")

	var body []struct {
		GalleryImages []string `json:"gallery_images"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	require.Len(t, body, 1)
	require.Equal(t, []string{
		"/api/v1/showcase/1/images/0/file",
	}, body[0].GalleryImages)
}

func TestShowcasePublicListReturnsEmptyImageListForInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&showcasedomain.Showcase{}, &showcasedomain.Comment{}))
	require.NoError(t, db.Create(&showcasedomain.Showcase{
		UserID: 1,
		Kind:   showcasedomain.KindUser,
		Status: showcasedomain.StatusApproved,
		Images: datatypes.JSON(`{"invalid":true}`),
	}).Error)

	handler := NewShowcaseHandler(service.NewShowcaseService(repository.NewShowcaseRepository(db), nil))
	router := gin.New()
	router.GET("/showcase/gallery", handler.List)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/showcase/gallery?type=user", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.True(t, strings.Contains(recorder.Body.String(), "showcase images are invalid"))
}

func TestShowcasePublicImageHandlerServesApprovedImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&showcasedomain.Showcase{}, &showcasedomain.Comment{}))
	require.NoError(t, db.Create(&showcasedomain.Showcase{
		UserID: 1,
		Kind:   showcasedomain.KindUser,
		Status: showcasedomain.StatusApproved,
		Images: datatypes.JSON(`["showcase/approved/2026/08/13/approved.webp"]`),
	}).Error)

	handler := NewShowcaseHandler(service.NewShowcaseService(
		repository.NewShowcaseRepository(db),
		&publicImageTestStorage{
			objects: map[string][]byte{
				"showcase/approved/2026/08/13/approved.webp": []byte("approved-image"),
			},
		},
	))
	router := gin.New()
	router.GET("/showcase/:id/images/:image_index/file", handler.ServePublicImageFile)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/showcase/1/images/0/file", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "image/webp", recorder.Header().Get("Content-Type"))
	require.Equal(t, "approved-image", recorder.Body.String())
}

func TestShowcasePublicImageHandlerRejectsPendingImageEvenWhenObjectExists(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	require.NoError(t, db.AutoMigrate(&showcasedomain.Showcase{}, &showcasedomain.Comment{}))
	require.NoError(t, db.Create(&showcasedomain.Showcase{
		UserID: 1,
		Kind:   showcasedomain.KindUser,
		Status: showcasedomain.StatusPending,
		Images: datatypes.JSON(`["showcase/pending/2026/08/13/pending.webp"]`),
	}).Error)

	handler := NewShowcaseHandler(service.NewShowcaseService(
		repository.NewShowcaseRepository(db),
		&publicImageTestStorage{
			objects: map[string][]byte{
				"showcase/pending/2026/08/13/pending.webp": []byte("pending-image"),
			},
		},
	))
	router := gin.New()
	router.GET("/showcase/:id/images/:image_index/file", handler.ServePublicImageFile)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/showcase/1/images/0/file", nil)
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.NotContains(t, recorder.Body.String(), "pending-image")
}

type publicImageTestStorage struct {
	objects map[string][]byte
}

func (s *publicImageTestStorage) Upload(context.Context, *multipart.FileHeader) (string, error) {
	return "", errors.New("not implemented")
}

func (s *publicImageTestStorage) UploadWithPrefix(context.Context, *multipart.FileHeader, string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *publicImageTestStorage) UploadFromReader(context.Context, io.Reader, string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *publicImageTestStorage) UploadFromReaderWithPrefix(context.Context, io.Reader, string, string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *publicImageTestStorage) Delete(context.Context, string) error {
	return nil
}

func (s *publicImageTestStorage) GetURL(filename string) string {
	return "/uploads/" + filename
}

func (s *publicImageTestStorage) ObjectKey(reference string) (string, error) {
	key, ok := storage.NormalizeObjectKey(reference)
	if !ok {
		return "", errors.New("invalid object key")
	}
	return key, nil
}

func (s *publicImageTestStorage) CopyObject(context.Context, string, string) error {
	return nil
}

func (s *publicImageTestStorage) Open(_ context.Context, key string) (*storage.StoredObject, error) {
	body, ok := s.objects[key]
	if !ok {
		return nil, errors.New("object not found")
	}
	return &storage.StoredObject{
		ReadCloser: io.NopCloser(bytes.NewReader(body)),
		Name:       "showcase.webp",
		MimeType:   "image/webp",
		Size:       int64(len(body)),
		ModTime:    time.Now(),
	}, nil
}
