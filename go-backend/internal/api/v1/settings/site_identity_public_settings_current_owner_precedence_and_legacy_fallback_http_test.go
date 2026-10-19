package settings

import (
	"net/http"
	"net/http/httptest"
	"testing"

	settingdomain "commerce-platform/internal/domain/setting"
	sitefavicondomain "commerce-platform/internal/domain/site_favicon"
	sitelogodomain "commerce-platform/internal/domain/site_logo"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSiteIdentityPublicSettingsHTTPUsesCurrentOwnersAcrossLocalesAndPreservesOnlySharedLegacyLogoFallbacks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}, &sitelogodomain.Asset{}, &sitefavicondomain.Asset{}))
	for _, locale := range []string{"en", "zh_cn"} {
		for _, key := range []string{"site_logo", "site_favicon"} {
			require.NoError(t, db.Create(&settingdomain.Setting{Key: key, Value: "https://old.example.test/stale.png", Locale: locale, Group: "site", IsPublic: true}).Error)
		}
	}
	logo := sitelogodomain.Asset{ID: 1, StorageKey: "site-logo/current.webp", Width: 512, Height: 512}
	favicon := sitefavicondomain.Asset{ID: 1, StorageKey: "site-favicon/current.webp"}
	require.NoError(t, db.Create(&logo).Error)
	require.NoError(t, db.Create(&favicon).Error)
	handler := NewHandler(service.NewSettingService(repository.NewSettingRepository(db), nil, 0))
	handler.ConfigureSiteLogoService(service.NewSiteLogoService(repository.NewSiteLogoRepository(db), nil, "https://shop.example.test"))
	handler.ConfigureSiteFaviconService(service.NewSiteFaviconService(repository.NewSiteFaviconRepository(db), nil, "https://shop.example.test"))
	router := gin.New()
	router.GET("/site", handler.GetSiteSettings)
	router.GET("/group/:group", handler.GetSettingsByGroup)
	router.GET("/single/:key", handler.GetSetting)
	for _, locale := range []string{"en", "zh_cn"} {
		for _, path := range []string{"/site", "/group/site", "/single/site_logo"} {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path+"?locale="+locale, nil))
			require.Equal(t, http.StatusOK, response.Code, response.Body.String())
			require.Contains(t, response.Body.String(), "https://shop.example.test/uploads/site-logo/current.webp")
			require.NotContains(t, response.Body.String(), "stale.png")
			if path != "/single/site_logo" {
				require.Contains(t, response.Body.String(), "https://shop.example.test/uploads/site-favicon/current.webp")
			}
		}
	}
	require.NoError(t, db.Delete(&logo).Error)
	for _, value := range []string{"/uploads/catalog/shared.png", "https://external.example.test/logo.png", "/static/logo.svg"} {
		require.Equal(t, value, handler.publicSetting(&settingdomain.Setting{Key: "site_logo", Value: value}).Value)
	}
	for _, value := range []string{"/uploads/site-logo/old.webp", "https://shop.example.test/uploads/site-logo/old.webp"} {
		require.Empty(t, handler.publicSetting(&settingdomain.Setting{Key: "site_logo", Value: value}).Value)
		require.Empty(t, handler.publicSiteSettings(&settingdomain.SiteSettings{SiteLogo: value}).SiteLogo)
	}
}
