package service

import (
	"errors"
	"testing"

	settingdomain "commerce-platform/internal/domain/setting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebsiteNameServiceFallsBackToEnglish(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteNameService := NewWebsiteNameService(settingService)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{
			Key:      settingdomain.WebsiteNameKeyTitle,
			Value:    "Why this name",
			Type:     "string",
			Locale:   "en",
			Group:    settingdomain.WebsiteNameGroup,
			IsPublic: true,
		},
	}))

	settings, err := websiteNameService.Get("fr")
	require.NoError(t, err)
	assert.Equal(t, "fr", settings.Locale)
	assert.Equal(t, "Why this name", settings.Title)
}

func TestWebsiteNameServiceUsesGeneratedDefaultsWhenLocaleHasNoSettings(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteNameService := NewWebsiteNameService(settingService)

	english, err := websiteNameService.Get("en")
	require.NoError(t, err)
	assert.Equal(t, "Website context", english.Status)
	assert.Equal(t, "Why This Name", english.Title)

	chinese, err := websiteNameService.Get("zh-CN")
	require.NoError(t, err)
	assert.Equal(t, "网站说明", chinese.Status)
	assert.Equal(t, "为什么叫这个名字", chinese.Title)
}

func TestWebsiteNameServiceIgnoresEmptySeedRowsWhenUsingGeneratedDefaults(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteNameService := NewWebsiteNameService(settingService)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{
			Key:      settingdomain.WebsiteNameKeyStatus,
			Value:    "",
			Type:     "string",
			Locale:   "en",
			Group:    settingdomain.WebsiteNameGroup,
			IsPublic: true,
		},
		{
			Key:      settingdomain.WebsiteNameKeyTitle,
			Value:    "",
			Type:     "string",
			Locale:   "en",
			Group:    settingdomain.WebsiteNameGroup,
			IsPublic: true,
		},
	}))

	settings, err := websiteNameService.Get("en")
	require.NoError(t, err)
	assert.Equal(t, "Website context", settings.Status)
	assert.Equal(t, "Why This Name", settings.Title)
}

func TestWebsiteNameServiceReturnsUnavailableErrorWithoutSettingsService(t *testing.T) {
	websiteNameService := NewWebsiteNameService(nil)

	_, err := websiteNameService.Get("en")
	assert.ErrorIs(t, err, ErrWebsiteNameServiceUnavailable)

	_, err = websiteNameService.GetAdmin("en")
	assert.ErrorIs(t, err, ErrWebsiteNameServiceUnavailable)

	_, err = websiteNameService.Update(settingdomain.WebsiteNameUpdateRequest{Locale: "en"})
	assert.ErrorIs(t, err, ErrWebsiteNameServiceUnavailable)
}

func TestWebsiteNameServiceAdminRejectsUnsupportedLocale(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteNameService := NewWebsiteNameService(settingService)

	_, err := websiteNameService.GetAdmin("not-a-language")
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedLocale))
}

func TestWebsiteNameServiceAdminDefaultsEmptyLocaleToEnglish(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteNameService := NewWebsiteNameService(settingService)

	settings, err := websiteNameService.GetAdmin("")
	require.NoError(t, err)
	assert.Equal(t, "en", settings.Locale)
	assert.Equal(t, "Why This Name", settings.Title)
}

func TestWebsiteNameServiceRejectsUnsupportedUpdateLocale(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteNameService := NewWebsiteNameService(settingService)

	_, err := websiteNameService.Update(settingdomain.WebsiteNameUpdateRequest{
		Locale: "not-a-language",
		Title:  "Must not overwrite English",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrUnsupportedLocale))

	english, err := websiteNameService.Get("en")
	require.NoError(t, err)
	assert.Equal(t, "Why This Name", english.Title)
}

func TestWebsiteNameServiceUpdateNormalizesAndStoresSupportedLocale(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteNameService := NewWebsiteNameService(settingService)

	settings, err := websiteNameService.Update(settingdomain.WebsiteNameUpdateRequest{
		Locale: "zh-CN",
		Title:  "为什么叫这个名字",
		Body:   "正文",
	})
	require.NoError(t, err)
	assert.Equal(t, "zh_cn", settings.Locale)
	assert.Equal(t, "为什么叫这个名字", settings.Title)
	assert.Equal(t, "正文", settings.Body)

	record, err := settingService.Get(settingdomain.WebsiteNameKeyTitle, "zh_cn")
	require.NoError(t, err)
	assert.Equal(t, "为什么叫这个名字", record.Value)
	assert.Equal(t, "Why this name page: title", record.Description)
}

func TestWebsiteNameServiceDoesNotOverwriteWebsiteProfileKeys(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteNameService := NewWebsiteNameService(settingService)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{
			Key:      settingdomain.WebsiteProfileKeyEyebrow,
			Value:    "THE PERSON BEHIND THIS WEBSITE",
			Type:     "string",
			Locale:   "en",
			Group:    settingdomain.WebsiteProfileGroup,
			IsPublic: true,
		},
		{
			Key:      settingdomain.WebsiteProfileKeyTitle,
			Value:    "Me & This Website",
			Type:     "string",
			Locale:   "en",
			Group:    settingdomain.WebsiteProfileGroup,
			IsPublic: true,
		},
	}))

	_, err := websiteNameService.Update(settingdomain.WebsiteNameUpdateRequest{
		Locale:  "en",
		Eyebrow: "WHY THIS NAME",
		Title:   "Why This Name",
	})
	require.NoError(t, err)

	profileEyebrow, err := settingService.Get(settingdomain.WebsiteProfileKeyEyebrow, "en")
	require.NoError(t, err)
	assert.Equal(t, settingdomain.WebsiteProfileGroup, profileEyebrow.Group)
	assert.Equal(t, "THE PERSON BEHIND THIS WEBSITE", profileEyebrow.Value)

	profileTitle, err := settingService.Get(settingdomain.WebsiteProfileKeyTitle, "en")
	require.NoError(t, err)
	assert.Equal(t, settingdomain.WebsiteProfileGroup, profileTitle.Group)
	assert.Equal(t, "Me & This Website", profileTitle.Value)

	websiteNameTitle, err := settingService.Get(settingdomain.WebsiteNameKeyTitle, "en")
	require.NoError(t, err)
	assert.Equal(t, "Why This Name", websiteNameTitle.Value)
}

func TestWebsiteNameServiceUsesEnglishForUnsupportedLocale(t *testing.T) {
	_, settingService := newTestSettingService(t)
	websiteNameService := NewWebsiteNameService(settingService)

	settings, err := websiteNameService.Get("pt-BR")
	require.NoError(t, err)
	assert.Equal(t, "pt", settings.Locale)

	settings, err = websiteNameService.Get("not-a-language")
	require.NoError(t, err)
	assert.Equal(t, "en", settings.Locale)
	assert.Equal(t, "Why This Name", settings.Title)
}
