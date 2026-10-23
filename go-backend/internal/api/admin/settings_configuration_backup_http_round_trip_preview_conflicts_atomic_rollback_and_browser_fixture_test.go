package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/domain/media"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
)

func newSettingsConfigurationBackupHTTPTestDatabaseAndRouter(t *testing.T) (*gorm.DB, *gin.Engine) {
	t.Helper()
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
	handler := NewSettingsHandler(service.NewAdminSettingsService(settings))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		role := c.GetHeader("X-Synthetic-Backup-Test-Role")
		if role == "" {
			role, _ = c.Cookie("synthetic_backup_fixture_role")
		}
		if role == "" {
			role = string(auth.RoleAdmin)
		}
		c.Set("user_role", role)
		c.Set("user_id", uint(42))
	})
	// Register production routes so role tests also verify their real middleware.
	registerSystemRoutes(router.Group("/api/admin"), handler, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	router.GET("/api/admin/auth/profile", func(c *gin.Context) {
		role := auth.Role(c.GetString("user_role"))
		c.JSON(http.StatusOK, gin.H{"user": gin.H{"id": 42, "username": "本地备份验证", "role": role, "permissions": role.GetPermissions()}})
	})
	router.GET("/api/v1/settings/site", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"admin_brand_name": "本地验证商城", "admin_brand_initial": "验", "admin_html_title": "商城后台本地验证"})
	})
	router.GET("/api/admin/customer-service/conversations", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"conversations": []any{}, "pagination": gin.H{"page": 1, "page_size": 1, "total": 0, "total_pages": 0}})
	})
	return db, router
}

func performSettingsConfigurationBackupTestRequest(t *testing.T, router http.Handler, role auth.Role, method, path string, payload any) *httptest.ResponseRecorder {
	t.Helper()
	var body []byte
	if raw, ok := payload.(string); ok {
		body = []byte(raw)
	} else if payload != nil {
		var err error
		body, err = json.Marshal(payload)
		require.NoError(t, err)
	}
	request := httptest.NewRequestWithContext(context.Background(), method, "/api/admin/settings/backup-recovery/"+path, bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Synthetic-Backup-Test-Role", string(role))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func newPortableSettingsConfigurationTestFile(entries ...setting.UpdateSettingRequest) service.SettingsConfigurationBackupFile {
	return service.SettingsConfigurationBackupFile{Format: service.SettingsConfigurationBackupFormat, Version: service.SettingsConfigurationBackupVersion, Settings: entries}
}

func previewSettingsConfigurationTestFile(t *testing.T, router http.Handler, file service.SettingsConfigurationBackupFile) service.SettingsConfigurationImportPreview {
	t.Helper()
	response := performSettingsConfigurationBackupTestRequest(t, router, auth.RoleAdmin, http.MethodPost, "configuration/preview", gin.H{"backup": file})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	var preview service.SettingsConfigurationImportPreview
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &preview))
	return preview
}

func TestSettingsConfigurationBackupExportImportRoundTripPreservesOtherLocalesAndExcludedSecrets(t *testing.T) {
	db, router := newSettingsConfigurationBackupHTTPTestDatabaseAndRouter(t)
	seed := []setting.Setting{
		{Key: "site_name", Value: "Original storefront", Group: "site", Locale: "en"},
		{Key: "site_name", Value: "原站点", Group: "site", Locale: "zh-CN"},
		{Key: "smtp_host", Value: "smtp.example.test", Group: "email", Locale: "global"},
		{Key: "smtp_password", Value: "synthetic-password-stays-private", Group: "email", Locale: "global"},
		{Key: "unregistered_future_setting", Value: "synthetic-future-secret", Group: "site", Locale: "en"},
		{Key: "website_name_title", Value: "Domain-owned title", Group: "website_name", Locale: "en"},
	}
	require.NoError(t, db.Create(&seed).Error)
	response := performSettingsConfigurationBackupTestRequest(t, router, auth.RoleAdmin, http.MethodGet, "configuration/export", nil)
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, "no-store", response.Header().Get("Cache-Control"))
	require.Contains(t, response.Header().Get("Content-Disposition"), "attachment;")
	require.NotContains(t, response.Body.String(), "synthetic-password")
	require.NotContains(t, response.Body.String(), "synthetic-future-secret")
	var file service.SettingsConfigurationBackupFile
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &file))
	require.Len(t, file.Settings, 3)
	require.Equal(t, 3, file.ExcludedSettingCount)
	for _, entry := range file.Settings {
		if entry.Group == "email" {
			require.False(t, entry.IsPublic)
		}
	}
	file.Settings = append(file.Settings, setting.UpdateSettingRequest{Key: "admin_footer_text", Value: "Created from saved configuration", Type: "string", Group: "site", Locale: "en", IsPublic: false})
	require.NoError(t, db.Create(&setting.Setting{Key: "site_description", Value: "An omitted ordinary setting stays unchanged", Group: "site", Locale: "en"}).Error)
	// Change the destination after downloading, then restore its saved state.
	require.NoError(t, db.Model(&setting.Setting{}).Where("key = ? AND locale = ?", "site_name", "en").Update("value", "Changed storefront").Error)
	preview := previewSettingsConfigurationTestFile(t, router, file)
	require.Equal(t, 1, preview.CreateCount)
	require.Equal(t, 2, preview.UpdateCount) // site value plus legacy SMTP visibility.
	require.Equal(t, 1, preview.UnchangedCount)
	response = performSettingsConfigurationBackupTestRequest(t, router, auth.RoleAdmin, http.MethodPost, "configuration/import", gin.H{"backup": file, "expected_target_sha256": preview.TargetSHA256, "confirmation": "IMPORT CONFIGURATION"})
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())
	for _, expected := range seed {
		var actual setting.Setting
		require.NoError(t, db.Where("key = ? AND locale = ?", expected.Key, expected.Locale).First(&actual).Error)
		require.Equal(t, expected.Value, actual.Value)
		if expected.Key == "smtp_host" {
			require.False(t, actual.IsPublic)
		}
	}
	var createdFooter, omittedDescription setting.Setting
	require.NoError(t, db.Where("key = ?", "admin_footer_text").First(&createdFooter).Error)
	require.Equal(t, "Created from saved configuration", createdFooter.Value)
	require.False(t, createdFooter.IsPublic)
	require.NoError(t, db.Where("key = ?", "site_description").First(&omittedDescription).Error)
	require.Equal(t, "An omitted ordinary setting stays unchanged", omittedDescription.Value)
	unchanged := previewSettingsConfigurationTestFile(t, router, file)
	require.Equal(t, 4, unchanged.UnchangedCount)
}

func TestSettingsConfigurationBackupStalePreviewAndChangedFileCannotOverwriteCurrentState(t *testing.T) {
	for _, scenario := range []string{"existing destination changed", "missing destination created", "uploaded file changed"} {
		t.Run(scenario, func(t *testing.T) {
			db, router := newSettingsConfigurationBackupHTTPTestDatabaseAndRouter(t)
			file := newPortableSettingsConfigurationTestFile(setting.UpdateSettingRequest{Key: "site_name", Value: "From saved file", Type: "string", Group: "site", Locale: "en", IsPublic: true})
			if scenario == "existing destination changed" {
				require.NoError(t, db.Create(&setting.Setting{Key: "site_name", Value: "Before preview", Group: "site", Locale: "en"}).Error)
			}
			preview := previewSettingsConfigurationTestFile(t, router, file)
			switch scenario {
			case "existing destination changed":
				require.NoError(t, db.Model(&setting.Setting{}).Where("key = ?", "site_name").Update("value", "Concurrent edit").Error)
			case "missing destination created":
				require.NoError(t, db.Create(&setting.Setting{Key: "site_name", Value: "Concurrent edit", Group: "site", Locale: "en"}).Error)
			default:
				file.Settings[0].Value = "Different file"
			}
			response := performSettingsConfigurationBackupTestRequest(t, router, auth.RoleAdmin, http.MethodPost, "configuration/import", gin.H{"backup": file, "expected_target_sha256": preview.TargetSHA256, "confirmation": "IMPORT CONFIGURATION"})
			require.Equal(t, http.StatusConflict, response.Code, response.Body.String())
			var rows []setting.Setting
			require.NoError(t, db.Unscoped().Find(&rows).Error)
			if scenario == "uploaded file changed" {
				require.Empty(t, rows)
			} else {
				require.Len(t, rows, 1)
				require.Equal(t, "Concurrent edit", rows[0].Value)
			}
		})
	}
}

func TestSettingsConfigurationBackupInvalidMediaRollsBackEverySettingAndReservation(t *testing.T) {
	db, router := newSettingsConfigurationBackupHTTPTestDatabaseAndRouter(t)
	require.NoError(t, db.Create(&setting.Setting{Key: "admin_brand_name", Value: "Keep this brand", Group: "site", Locale: "en"}).Error)
	file := newPortableSettingsConfigurationTestFile(
		setting.UpdateSettingRequest{Key: "admin_brand_name", Value: "Must not persist", Type: "string", Group: "site", Locale: "en", IsPublic: true},
		setting.UpdateSettingRequest{Key: "site_description", Value: "Must not be created", Type: "string", Group: "site", Locale: "en", IsPublic: true},
		setting.UpdateSettingRequest{Key: "site_logo", Value: "/uploads/missing.png", Type: "string", Group: "site", Locale: "en", IsPublic: true},
	)
	preview := previewSettingsConfigurationTestFile(t, router, file)
	response := performSettingsConfigurationBackupTestRequest(t, router, auth.RoleAdmin, http.MethodPost, "configuration/import", gin.H{"backup": file, "expected_target_sha256": preview.TargetSHA256, "confirmation": "IMPORT CONFIGURATION"})
	require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
	var rows []setting.Setting
	require.NoError(t, db.Unscoped().Find(&rows).Error)
	require.Len(t, rows, 1)
	require.Equal(t, "Keep this brand", rows[0].Value)
}

func TestSettingsConfigurationBackupRejectsUnsupportedFilesWithoutLeakingProtectedTarget(t *testing.T) {
	db, router := newSettingsConfigurationBackupHTTPTestDatabaseAndRouter(t)
	entry := setting.UpdateSettingRequest{Key: "site_name", Value: "Imported", Type: "string", Group: "site", Locale: "en", IsPublic: true}
	for _, scenario := range []string{"duplicate", "secret", "domain owned", "type", "locale", "version", "quoted number", "null number", "public email", "wrong target group"} {
		t.Run(scenario, func(t *testing.T) {
			file := newPortableSettingsConfigurationTestFile(entry)
			switch scenario {
			case "duplicate":
				file.Settings = append(file.Settings, entry)
			case "secret":
				file.Settings[0].Key = "smtp_password"
				file.Settings[0].Group = "email"
			case "domain owned":
				file.Settings[0].Key = "website_name_title"
				file.Settings[0].Group = "website_name"
			case "type":
				file.Settings[0].Type = "json"
			case "locale":
				file.Settings[0].Locale = "../../en"
			case "version":
				file.Version++
			case "quoted number", "null number":
				file.Settings[0] = setting.UpdateSettingRequest{Key: "smtp_port", Group: "email", Type: "number", Locale: "global", Value: `"587"`}
				if scenario == "null number" {
					file.Settings[0].Value = "null"
				}
			case "public email":
				file.Settings[0] = setting.UpdateSettingRequest{Key: "smtp_host", Group: "email", Type: "string", Locale: "global", IsPublic: true}
			case "wrong target group":
				require.NoError(t, db.Create(&setting.Setting{Key: "site_name", Value: "protected-target-value", Group: "payment_secret", Locale: "en"}).Error)
			}
			response := performSettingsConfigurationBackupTestRequest(t, router, auth.RoleAdmin, http.MethodPost, "configuration/preview", gin.H{"backup": file})
			require.Equal(t, http.StatusBadRequest, response.Code, response.Body.String())
			require.NotContains(t, response.Body.String(), "protected-target-value")
		})
	}
	for _, raw := range []string{`{"backup":{},"unknown":true}`, `{"backup":{"unknown":true}}`, `{"backup":{}} {}`} {
		response := performSettingsConfigurationBackupTestRequest(t, router, auth.RoleAdmin, http.MethodPost, "configuration/preview", raw)
		require.Equal(t, http.StatusBadRequest, response.Code)
	}
	response := performSettingsConfigurationBackupTestRequest(t, router, auth.RoleAdmin, http.MethodPost, "configuration/preview", `{"backup":{"format":"`+strings.Repeat("a", service.MaximumSettingsConfigurationFileBytes+(65<<10)))
	require.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
}

func TestSettingsConfigurationBackupProductionRoutesEnforceReadAndEditPermissions(t *testing.T) {
	_, router := newSettingsConfigurationBackupHTTPTestDatabaseAndRouter(t)
	file := newPortableSettingsConfigurationTestFile(setting.UpdateSettingRequest{Key: "site_name", Type: "string", Group: "site", Locale: "en"})
	for _, role := range []auth.Role{auth.RoleAdmin, auth.RoleManager, auth.RoleViewer, auth.RoleSupport} {
		for _, path := range []string{"overview", "configuration/export", "configuration/preview", "configuration/import"} {
			method := http.MethodGet
			if strings.Contains(path, "preview") || strings.Contains(path, "import") {
				method = http.MethodPost
			}
			response := performSettingsConfigurationBackupTestRequest(t, router, role, method, path, gin.H{"backup": file})
			if !role.HasPermission(auth.PermSettingsView) || (strings.Contains(path, "import") && !role.HasPermission(auth.PermSettingsEdit)) {
				require.Equal(t, http.StatusForbidden, response.Code, "%s %s", role, path)
			} else if strings.Contains(path, "import") {
				require.Equal(t, http.StatusBadRequest, response.Code)
			} else {
				require.Equal(t, http.StatusOK, response.Code, "%s %s %s", role, path, response.Body.String())
			}
		}
	}
}

// Only compiled into the Go test binary. Its synthetic profile and SQLite data
// never enter the application server or connect to a real storefront database.
func TestServeLoopbackOnlySyntheticSettingsConfigurationBackupBrowserFixture(t *testing.T) {
	port := os.Getenv("SETTINGS_CONFIGURATION_BACKUP_BROWSER_FIXTURE_PORT")
	if port == "" {
		t.Skip("opt-in loopback browser fixture")
	}
	db, router := newSettingsConfigurationBackupHTTPTestDatabaseAndRouter(t)
	require.NoError(t, db.Create(&[]setting.Setting{
		{Key: "site_name", Value: "本地验证商城", Group: "site", Locale: "zh-CN"},
		{Key: "site_name", Value: "Local verification storefront", Group: "site", Locale: "en"},
		{Key: "contact_email", Value: "support@example.test", Group: "site", Locale: "global"},
		{Key: "instagram", Value: "https://www.instagram.com/example", Group: "social", Locale: "global"},
		{Key: "title", Value: "本地测试网站介绍", Group: "website_profile", Locale: "zh-CN"},
		{Key: "smtp_password", Value: "synthetic-excluded-value", Group: "email", Locale: "global"},
	}).Error)
	listener, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", net.JoinHostPort("127.0.0.1", port))
	require.NoError(t, err)
	defer listener.Close()
	t.Logf("Synthetic browser fixture listening on %s", listener.Addr())
	require.NoError(t, http.Serve(listener, router))
}
