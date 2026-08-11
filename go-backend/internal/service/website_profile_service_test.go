package service

import (
	"testing"

	settingdomain "tanzanite/internal/domain/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebsiteProfileServiceMergesLocaleAndGlobalSettings(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteProfileService := NewWebsiteProfileService(settingService)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: settingdomain.WebsiteProfileKeyTitle, Value: "English title", Type: "string", Locale: "en", Group: settingdomain.WebsiteProfileGroup, IsPublic: true},
		{Key: settingdomain.WebsiteProfileKeyTitle, Value: "中文标题", Type: "string", Locale: "zh_cn", Group: settingdomain.WebsiteProfileGroup, IsPublic: true},
		{Key: settingdomain.WebsiteProfileKeyAvatarURL, Value: "/uploads/avatar.webp", Type: "string", Locale: "global", Group: settingdomain.WebsiteProfileGroup, IsPublic: true},
		{Key: settingdomain.WebsiteProfileKeyFactoryLink, Value: "/company/about#factory", Type: "string", Locale: "global", Group: settingdomain.WebsiteProfileGroup, IsPublic: true},
	}))

	settings, err := websiteProfileService.Get("zh-CN")
	require.NoError(t, err)
	assert.Equal(t, "zh_cn", settings.Locale)
	assert.Equal(t, "中文标题", settings.Title)
	assert.Equal(t, "/uploads/avatar.webp", settings.AvatarURL)
	assert.Equal(t, "/company/about#factory", settings.FactoryLink)
	assert.Equal(t, "English title", func() string {
		english, getErr := websiteProfileService.Get("fr")
		require.NoError(t, getErr)
		return english.Title
	}())
}

func TestWebsiteProfileServiceUpdateStoresLocaleAndGlobalFields(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteProfileService := NewWebsiteProfileService(settingService)

	settings, err := websiteProfileService.Update(settingdomain.WebsiteProfileUpdateRequest{
		Locale:          "zh-CN",
		Title:           "新的标题",
		AvatarURL:       "/uploads/avatar.webp",
		FactoryImageURL: "/uploads/factory.webp",
		FactoryLink:     "/company/about#factory",
	})
	require.NoError(t, err)
	assert.Equal(t, "zh_cn", settings.Locale)
	assert.Equal(t, "新的标题", settings.Title)
	assert.Equal(t, "/uploads/avatar.webp", settings.AvatarURL)
	assert.Equal(t, "/uploads/factory.webp", settings.FactoryImageURL)

	localeRecord, err := settingService.Get(settingdomain.WebsiteProfileKeyTitle, "zh_cn")
	require.NoError(t, err)
	assert.Equal(t, "新的标题", localeRecord.Value)

	globalRecord, err := settingService.Get(settingdomain.WebsiteProfileKeyAvatarURL, "global")
	require.NoError(t, err)
	assert.Equal(t, "/uploads/avatar.webp", globalRecord.Value)
}

func TestWebsiteProfileServiceIgnoresEmptyOverrides(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteProfileService := NewWebsiteProfileService(settingService)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{
			Key:      settingdomain.WebsiteProfileKeyTitle,
			Value:    "",
			Type:     "string",
			Locale:   "en",
			Group:    settingdomain.WebsiteProfileGroup,
			IsPublic: true,
		},
		{
			Key:      settingdomain.WebsiteProfileKeyFactoryImageURL,
			Value:    "",
			Type:     "string",
			Locale:   "global",
			Group:    settingdomain.WebsiteProfileGroup,
			IsPublic: true,
		},
	}))

	settings, err := websiteProfileService.GetAdmin("en")
	require.NoError(t, err)

	defaults := settingdomain.DefaultWebsiteProfileSettings("en")
	assert.Equal(t, defaults.Title, settings.Title)
	assert.Equal(t, websiteProfileDefaultFactoryImageURL, settings.FactoryImageURL)
}
