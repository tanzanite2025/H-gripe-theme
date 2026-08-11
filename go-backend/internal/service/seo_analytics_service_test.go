package service

import (
	"errors"
	"strings"
	"testing"

	analyticsdomain "commerce-platform/internal/domain/analytics"
	seodomain "commerce-platform/internal/domain/seo"
	settingdomain "commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/require"
)

type recordingSEOHTMLCacheInvalidator struct {
	reasons []string
}

func (r *recordingSEOHTMLCacheInvalidator) PurgeAllAsync(reason string) {
	r.reasons = append(r.reasons, reason)
}

func TestSEOServiceUsesLocaleValuesBeforeEnglishFallback(t *testing.T) {
	_, settingService := newTestSettingService(t)
	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: seodomain.HomeKeys.MetaTitle, Value: "Global title", Locale: "en", Group: seodomain.Group, IsPublic: true},
		{Key: seodomain.HomeKeys.MetaDescription, Value: "Global description", Locale: "en", Group: seodomain.Group, IsPublic: true},
		{Key: seodomain.HomeKeys.MetaTitle, Value: "本地标题", Locale: "zh_cn", Group: seodomain.Group, IsPublic: true},
	}))

	result, err := NewSEOService(settingService).GetHome("zh-CN")

	require.NoError(t, err)
	require.Equal(t, "本地标题", result.MetaTitle)
	require.Equal(t, "Global description", result.MetaDescription)
}

func TestSEOServiceRejectsOverlongUnicodeValues(t *testing.T) {
	_, settingService := newTestSettingService(t)
	description := strings.Repeat("界", seoMetaDescriptionLimit+1)

	_, err := NewSEOService(settingService).UpdateHome(seodomain.UpdateRequest{
		MetaDescription: &description,
	})

	require.Error(t, err)
	require.True(t, errors.Is(err, ErrInvalidSEOSettings))
}

func TestSEOServiceUpdatesTrimmedValues(t *testing.T) {
	db, settingService := newTestSettingService(t)
	title := "  New title  "

	result, err := NewSEOService(settingService).UpdateHome(seodomain.UpdateRequest{
		MetaTitle: &title,
		Locale:    "en",
	})

	require.NoError(t, err)
	require.Equal(t, "New title", result.MetaTitle)

	stored, err := repository.NewSettingRepository(db).Get(seodomain.HomeKeys.MetaTitle, "en")
	require.NoError(t, err)
	require.Equal(t, seodomain.Group, stored.Group)
	require.True(t, stored.IsPublic)
}

func TestSEOServicePurgesStorefrontHTMLCacheAfterUpdate(t *testing.T) {
	_, settingService := newTestSettingService(t)
	invalidator := &recordingSEOHTMLCacheInvalidator{}
	seoService := NewSEOService(settingService)
	seoService.SetStorefrontHTMLCacheInvalidator(invalidator)
	title := "Updated title"

	_, err := seoService.UpdateHome(seodomain.UpdateRequest{
		MetaTitle: &title,
		Locale:    "en",
	})

	require.NoError(t, err)
	require.Equal(t, []string{"SEO settings updated"}, invalidator.reasons)
}

func TestAnalyticsServiceWritesToAnalyticsGroup(t *testing.T) {
	db, settingService := newTestSettingService(t)
	analyticsID := " G-123 "

	result, err := NewAnalyticsService(settingService).Update(analyticsdomain.UpdateRequest{
		GoogleAnalytics: &analyticsID,
		Locale:          "en",
	})

	require.NoError(t, err)
	require.Equal(t, "G-123", result.GoogleAnalytics)

	stored, err := repository.NewSettingRepository(db).Get("google_analytics", "en")
	require.NoError(t, err)
	require.Equal(t, analyticsdomain.Group, stored.Group)
}
