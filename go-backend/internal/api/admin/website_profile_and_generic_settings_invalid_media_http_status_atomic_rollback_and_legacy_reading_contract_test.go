package admin

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestWebsiteProfileAndGenericSettingsHTTPRejectInvalidMediaWith400AndNoPartialWrites(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&setting.Setting{}, &media.MediaAsset{}, &media.MediaAssetDerivative{}))
	files, err := storage.NewStorageService(&storage.Config{Type: storage.StorageTypeLocal, LocalPath: t.TempDir(), BaseURL: "https://shop.example.test"})
	require.NoError(t, err)
	settings := service.NewSettingService(repository.NewSettingRepository(db), nil, 0)
	settings.ConfigureMediaService(service.NewMediaService(repository.NewMediaRepository(db), files, nil, "https://shop.example.test", 20<<30))
	profile := NewWebsiteProfileHandler(service.NewWebsiteProfileService(settings))
	generic := NewSettingsHandler(service.NewAdminSettingsService(settings))
	for _, entry := range []string{"profile image", "profile rich text", "generic single", "generic batch"} {
		t.Run(entry, func(t *testing.T) {
			payload := map[string]interface{}{"locale": "en", "title": "Must roll back", "avatar_url": "/uploads/missing.png"}
			switch entry {
			case "profile rich text":
				payload = map[string]interface{}{"locale": "en", "title": "Must roll back", "statement_body": `<a href="/uploads/missing.pdf">File</a>`}
			case "generic single":
				payload = map[string]interface{}{"key": "factory_image_url", "value": "/uploads/missing.png", "group": "disguised", "locale": "global", "is_public": true}
			case "generic batch":
				payload = map[string]interface{}{"settings": []map[string]interface{}{{"key": "title", "value": "Must roll back", "locale": "en"}, {"key": "avatar_url", "value": "/uploads/missing.png", "locale": "global"}}}
			}
			body, err := json.Marshal(payload)
			require.NoError(t, err)
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/settings", bytes.NewReader(body))
			context.Request.Header.Set("Content-Type", "application/json")
			switch entry {
			case "generic single":
				generic.UpdateSetting(context)
			case "generic batch":
				generic.BatchUpdateSettings(context)
			default:
				profile.Update(context)
			}
			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
			var count int64
			require.NoError(t, db.Unscoped().Model(&setting.Setting{}).Count(&count).Error)
			require.Zero(t, count)
		})
	}
	legacy := setting.Setting{Key: "avatar_url", Value: "/uploads/legacy.png", Group: setting.WebsiteProfileGroup, Locale: "global", IsPublic: true}
	require.NoError(t, db.Create(&legacy).Error)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)
	context.Request = httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/settings/website-profile?locale=en", nil)
	profile.Get(context)
	require.Equal(t, http.StatusOK, response.Code)
	require.Contains(t, response.Body.String(), legacy.Value)
}
