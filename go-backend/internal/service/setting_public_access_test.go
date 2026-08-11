package service

import (
	"testing"

	settingdomain "tanzanite/internal/domain/setting"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSettingServicePublicAccessFiltersPrivateSettings(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "H-GRIPE", Type: "string", Locale: "en", Group: "site", IsPublic: true},
		{Key: "smtp_password", Value: "secret", Type: "string", Locale: "en", Group: "email", IsPublic: false},
	}))

	_, err := settingService.GetPublic("smtp_password", "en")
	require.Error(t, err)

	settings, err := settingService.GetPublicByGroup("email", "en")
	require.NoError(t, err)
	assert.Empty(t, settings)

	groups, err := settingService.GetPublicGroups()
	require.NoError(t, err)
	assert.Equal(t, []string{"site"}, groups)
}

func TestGetSiteSettingsUsesBrandTitleAsPublicSiteName(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "Legacy Name", Type: "string", Locale: "en", Group: "site", IsPublic: true},
		{Key: "brand_title", Value: "Current Brand", Type: "string", Locale: "en", Group: "site", IsPublic: true},
	}))

	settings, err := settingService.GetSiteSettings("en")
	require.NoError(t, err)
	assert.Equal(t, "Current Brand", settings.BrandTitle)
	assert.Equal(t, "Current Brand", settings.SiteName)
}

func TestGetSiteSettingsFallsBackToLegacySiteName(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "Legacy Name", Type: "string", Locale: "en", Group: "site", IsPublic: true},
	}))

	settings, err := settingService.GetSiteSettings("en")
	require.NoError(t, err)
	assert.Equal(t, "Legacy Name", settings.BrandTitle)
	assert.Equal(t, "Legacy Name", settings.SiteName)
}

func TestGetSiteSettingsFallsBackToEnglishForUnconfiguredLocale(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "brand_title", Value: "Global Brand", Type: "string", Locale: "en", Group: "site", IsPublic: true},
		{Key: "contact_email", Value: "brand@example.test", Type: "string", Locale: "en", Group: "site", IsPublic: true},
	}))

	settings, err := settingService.GetSiteSettings("zh_cn")
	require.NoError(t, err)
	assert.Equal(t, "Global Brand", settings.BrandTitle)
	assert.Equal(t, "Global Brand", settings.SiteName)
	assert.Equal(t, "brand@example.test", settings.ContactEmail)
}

func TestGetSiteSettingsUsesLocaleValueBeforeEnglishFallback(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "brand_title", Value: "Global Brand", Type: "string", Locale: "en", Group: "site", IsPublic: true},
		{Key: "contact_email", Value: "brand@example.test", Type: "string", Locale: "en", Group: "site", IsPublic: true},
		{Key: "brand_title", Value: "中文品牌", Type: "string", Locale: "zh_cn", Group: "site", IsPublic: true},
	}))

	settings, err := settingService.GetSiteSettings("zh-CN")
	require.NoError(t, err)
	assert.Equal(t, "中文品牌", settings.BrandTitle)
	assert.Equal(t, "中文品牌", settings.SiteName)
	assert.Equal(t, "brand@example.test", settings.ContactEmail)
}

func newTestSettingService(t *testing.T) (*gorm.DB, *SettingService) {
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

	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}))

	return db, NewSettingService(repository.NewSettingRepository(db), nil, 0)
}
