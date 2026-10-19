package admin

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/outbox"
	"commerce-platform/internal/domain/setting"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"commerce-platform/internal/testfixtures"
)

func TestSiteLogoHTTPUploadAndDeleteCommitOwnerAndRequestedLocaleMirrorTogether(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&sitelogodomain.Asset{}, &setting.Setting{}, &outbox.Event{}))
	provider, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://media.example.test"})
	require.NoError(t, err)
	logos := service.NewSiteLogoService(repository.NewSiteLogoRepository(db), provider, "https://shop.example.test")
	logos.ConfigureSettingService(service.NewSettingService(repository.NewSettingRepository(db), nil, 0))
	logos.ConfigureObjectCleanupOutbox(repository.NewOutboxRepository(db))
	handler := NewSiteLogoHandler(logos)
	upload := func() *httptest.ResponseRecorder {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		part, err := writer.CreateFormFile("file", "logo.webp")
		require.NoError(t, err)
		_, err = part.Write(testfixtures.SiteLogoAndFaviconUploadLifecycle512By512LosslessWebPFixtureBytes())
		require.NoError(t, err)
		require.NoError(t, writer.Close())
		response := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(response)
		context.Set("user_id", uint(11))
		context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/logo?locale=fr", &body)
		context.Request.Header.Set("Content-Type", writer.FormDataContentType())
		handler.Upload(context)
		return response
	}
	response := upload()
	require.Equal(t, http.StatusCreated, response.Code, response.Body.String())
	previousURL := logos.CurrentPublicURL()
	mirror, err := repository.NewSettingRepository(db).Get("site_logo", "fr")
	require.NoError(t, err)
	require.Equal(t, previousURL, mirror.Value)
	require.NoError(t, db.Callback().Update().Before("gorm:update").Register("reject_logo_http_mirror", func(tx *gorm.DB) {
		if tx.Statement.Table == "settings" {
			tx.AddError(errors.New("injected mirror failure"))
		}
	}))
	response = upload()
	require.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
	require.Equal(t, previousURL, logos.CurrentPublicURL())
	response = httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/logo?locale=fr", nil)
	handler.Delete(context)
	require.Equal(t, http.StatusInternalServerError, response.Code)
	require.Equal(t, previousURL, logos.CurrentPublicURL())
	require.NoError(t, db.Callback().Update().Remove("reject_logo_http_mirror"))
	response = httptest.NewRecorder()
	context, _ = gin.CreateTestContext(response)
	context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodDelete, "/logo?locale=fr", nil)
	handler.Delete(context)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	require.Empty(t, logos.CurrentPublicURL())
	mirror, err = repository.NewSettingRepository(db).Get("site_logo", "fr")
	require.NoError(t, err)
	require.Empty(t, mirror.Value)
}
