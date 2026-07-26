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
		{Key: "site_name", Value: "Tanzanite", Type: "string", Locale: "en", Group: "site", IsPublic: true},
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
