package service

import (
	"context"
	"os"
	"testing"

	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/pkg/locales"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestBrandWheelsetSpokeSpecificationsToolsRouteMigrationPreservesFAQContentAndPublishesOnlyTranslatedLocales(t *testing.T) {
	manifestData, err := os.ReadFile("../../../nuxt-i18n/public/storefront-route-manifest.json")
	require.NoError(t, err)
	manifest, err := decodeRouteManifest(manifestData)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&faq.FAQPage{}, &faq.FAQ{}))
	const contentIdentity = "guides-spokeguides-brand-wheelset-spoke-specs"
	const oldRoutePath = "/guides/spokeguides/brand-wheelset-spoke-specs"
	const newRoutePath = "/tools/brand-wheelset-spoke-specs"

	for _, locale := range locales.EnabledLocaleCodes() {
		require.NoError(t, db.Create(&faq.FAQPage{
			PageID: contentIdentity, RouteKey: contentIdentity, RoutePath: oldRoutePath,
			Locale: locale, Title: "Existing wheelset title", Domain: "guides", Status: "active",
		}).Error)
		require.NoError(t, db.Create(&faq.FAQ{
			PageID: contentIdentity, Locale: locale, Question: "Existing spoke question",
			Answer: "<p>Existing spoke answer</p>", Status: "published",
		}).Error)
	}

	faqService := NewFAQService(repository.NewFAQRepository(db), nil)
	_, err = faqService.ReconcileStorefrontRoutes(context.Background(), manifest)
	require.NoError(t, err)
	for _, locale := range locales.EnabledLocaleCodes() {
		var preservedFAQ faq.FAQ
		require.NoError(t, db.Where("page_id = ? AND locale = ?", contentIdentity, locale).First(&preservedFAQ).Error)
		assert.Equal(t, "Existing spoke question", preservedFAQ.Question)
		assert.Equal(t, "<p>Existing spoke answer</p>", preservedFAQ.Answer)
		var preservedPage faq.FAQPage
		require.NoError(t, db.Where("page_id = ? AND locale = ?", contentIdentity, locale).First(&preservedPage).Error)
		assert.Equal(t, "Existing wheelset title", preservedPage.Title)
		assert.Equal(t, "guides", preservedPage.Domain)
		if locale == "en" || locale == "zh_cn" {
			assert.Equal(t, newRoutePath, preservedPage.RoutePath)
			assert.Equal(t, FAQRouteStatusCurrent, preservedPage.RouteStatus)
			pageData, err := faqService.GetPublicPageDataByRoutePath(newRoutePath, locale)
			require.NoError(t, err)
			require.Len(t, pageData.Items, 1)
			assert.Equal(t, preservedFAQ.Question, pageData.Items[0].Question)
		} else {
			assert.Equal(t, FAQRouteStatusStale, preservedPage.RouteStatus)
		}
		_, err := faqService.GetPublicPageDataByRoutePath(oldRoutePath, locale)
		assert.Error(t, err, "Retired public URLs must not resolve to FAQ pages")
	}

	publishedLocaleCount := 0
	for _, entry := range buildManifestRouteCatalogEntries(manifest, &StorefrontRouteCatalogSyncSummary{}) {
		assert.NotContains(t, entry.Path, oldRoutePath)
		if entry.SourceKey == contentIdentity {
			publishedLocaleCount++
			assert.Contains(t, []string{"en", "zh_cn"}, entry.Locale)
			assert.Contains(t, entry.Path, newRoutePath)
			assert.Equal(t, entry.Path, entry.CanonicalPath)
			assert.True(t, entry.IsIndexable)
		}
	}
	assert.Equal(t, 2, publishedLocaleCount)
}
