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

func TestInnerTubeSelectionToolsRouteMigrationPreservesWholePageFAQContent(t *testing.T) {
	manifestData, err := os.ReadFile("../../../nuxt-i18n/public/storefront-route-manifest.json")
	require.NoError(t, err)
	manifest, err := decodeRouteManifest(manifestData)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&faq.FAQPage{}, &faq.FAQ{}))

	const contentIdentity = "guides-tireguides-choose-inner-tube"
	const oldRoutePath = "/guides/tireguides/choose-inner-tube"
	const newRoutePath = "/tools/valve-length-fitment-calculator"

	for _, locale := range locales.EnabledLocaleCodes() {
		require.NoError(t, db.Create(&faq.FAQPage{
			PageID: contentIdentity, RouteKey: contentIdentity, RoutePath: oldRoutePath,
			Locale: locale, Title: "Existing inner-tube title", Domain: "guides", Status: "active",
		}).Error)
		require.NoError(t, db.Create(&faq.FAQ{
			PageID: contentIdentity, Locale: locale, Question: "Existing inner-tube question",
			Answer: "<p>Existing inner-tube answer</p>", Status: "published",
		}).Error)
	}

	faqService := NewFAQService(repository.NewFAQRepository(db))
	_, err = faqService.ReconcileStorefrontRoutes(context.Background(), manifest)
	require.NoError(t, err)

	for _, locale := range locales.EnabledLocaleCodes() {
		var preservedPage faq.FAQPage
		require.NoError(t, db.Where("page_id = ? AND locale = ?", contentIdentity, locale).First(&preservedPage).Error)
		assert.Equal(t, "guides", preservedPage.Domain)
		assert.Equal(t, "Existing inner-tube title", preservedPage.Title)

		if locale == "en" || locale == "zh_cn" {
			assert.Equal(t, newRoutePath, preservedPage.RoutePath)
			assert.Equal(t, FAQRouteStatusCurrent, preservedPage.RouteStatus)
			pageData, err := faqService.GetPublicPageDataByRoutePath(newRoutePath, locale)
			require.NoError(t, err)
			require.Len(t, pageData.Items, 1)
			assert.Equal(t, "Existing inner-tube question", pageData.Items[0].Question)
			assert.Equal(t, "<p>Existing inner-tube answer</p>", pageData.Items[0].Answer)
		} else {
			assert.Equal(t, FAQRouteStatusStale, preservedPage.RouteStatus)
		}

		_, err = faqService.GetPublicPageDataByRoutePath(oldRoutePath, locale)
		assert.Error(t, err, "Retired inner-tube guide URLs must not resolve to FAQ pages")
	}

	currentLocaleCount := 0
	for _, entry := range buildManifestRouteCatalogEntries(manifest, &StorefrontRouteCatalogSyncSummary{}) {
		assert.NotContains(t, entry.Path, oldRoutePath)
		if entry.SourceKey != contentIdentity {
			continue
		}
		currentLocaleCount++
		assert.Contains(t, entry.Path, newRoutePath)
		assert.Equal(t, entry.Path, entry.CanonicalPath)
		assert.True(t, entry.IsIndexable)
	}
	assert.Equal(t, 2, currentLocaleCount)
}
