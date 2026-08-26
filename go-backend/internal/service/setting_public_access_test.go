package service

import (
	"testing"

	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSettingServicePublicAccessFiltersPrivateSettings(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "Commerce Platform", Type: "string", Locale: "en", Group: "site", IsPublic: true},
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

func TestGetSiteSettingsUsesSiteNameAndIgnoresLegacyBrandTitle(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "Current Site", Type: "string", Locale: "en", Group: "site", IsPublic: true},
		{Key: "brand_title", Value: "Old Brand", Type: "string", Locale: "en", Group: "site", IsPublic: true},
	}))

	settings, err := settingService.GetSiteSettings("en")
	require.NoError(t, err)
	assert.Equal(t, "Current Site", settings.SiteName)
}

func TestGetSiteSettingsKeepsEmptySiteNameEmpty(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "", Type: "string", Locale: "en", Group: "site", IsPublic: true},
	}))

	settings, err := settingService.GetSiteSettings("en")
	require.NoError(t, err)
	assert.Empty(t, settings.SiteName)
}

func TestGetSiteSettingsFallsBackToEnglishForUnconfiguredLocale(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "Global Site", Type: "string", Locale: "en", Group: "site", IsPublic: true},
		{Key: "contact_email", Value: "site@example.test", Type: "string", Locale: "en", Group: "site", IsPublic: true},
	}))

	settings, err := settingService.GetSiteSettings("zh_cn")
	require.NoError(t, err)
	assert.Equal(t, "Global Site", settings.SiteName)
	assert.Equal(t, "site@example.test", settings.ContactEmail)
}

func TestGetSiteSettingsUsesLocaleValueBeforeEnglishFallback(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "Global Site", Type: "string", Locale: "en", Group: "site", IsPublic: true},
		{Key: "contact_email", Value: "site@example.test", Type: "string", Locale: "en", Group: "site", IsPublic: true},
		{Key: "site_name", Value: "中文站点", Type: "string", Locale: "zh_cn", Group: "site", IsPublic: true},
	}))

	settings, err := settingService.GetSiteSettings("zh-CN")
	require.NoError(t, err)
	assert.Equal(t, "中文站点", settings.SiteName)
	assert.Equal(t, "site@example.test", settings.ContactEmail)
}

func TestGetSiteSettingsRemovesSocialLinkSizeAndUnsupportedPlatforms(t *testing.T) {
	_, settingService := newTestSettingService(t)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{
			Key:      "social_links",
			Value:    `[{"network":"facebook","url":"https://facebook.example","label":"Facebook","size":28},{"network":"linkedin","url":"https://linkedin.example"}]`,
			Type:     "string",
			Locale:   "en",
			Group:    "site",
			IsPublic: true,
		},
	}))

	settings, err := settingService.GetSiteSettings("en")
	require.NoError(t, err)
	assert.JSONEq(t, `[{"network":"facebook","url":"https://facebook.example","label":"Facebook"}]`, settings.SocialLinks)
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
