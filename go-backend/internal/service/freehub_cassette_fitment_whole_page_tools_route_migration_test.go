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

func TestFreehubCassetteFitmentToolsRouteMigrationPreservesWholePageFAQContent(t *testing.T) {
	manifestData, err := os.ReadFile("../../../nuxt-i18n/public/storefront-route-manifest.json")
	require.NoError(t, err)
	manifest, err := decodeRouteManifest(manifestData)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&faq.FAQPage{}, &faq.FAQ{}))

	const contentIdentity = "guides-wheelset-buyers-choose-freehub"
	const oldRoutePath = "/guides/wheelset-buyers/choose-freehub"
	const newRoutePath = "/tools/freehub-cassette-fitment-engine"

	for _, locale := range locales.EnabledLocaleCodes() {
		require.NoError(t, db.Create(&faq.FAQPage{
			PageID: contentIdentity, RouteKey: contentIdentity, RoutePath: oldRoutePath,
			Locale: locale, Title: "Existing freehub compatibility title", Domain: "guides", Status: "active",
		}).Error)
		require.NoError(t, db.Create(&faq.FAQ{
			PageID: contentIdentity, Locale: locale, Question: "Existing freehub compatibility question",
			Answer: "<p>Existing freehub compatibility answer</p>", Status: "published",
		}).Error)
	}

	faqService := NewFAQService(repository.NewFAQRepository(db))
	_, err = faqService.ReconcileStorefrontRoutes(context.Background(), manifest)
	require.NoError(t, err)

	for _, locale := range locales.EnabledLocaleCodes() {
		var preservedPage faq.FAQPage
		require.NoError(t, db.Where("page_id = ? AND locale = ?", contentIdentity, locale).First(&preservedPage).Error)
		assert.Equal(t, "guides", preservedPage.Domain)
		assert.Equal(t, "Existing freehub compatibility title", preservedPage.Title)
		assert.Equal(t, newRoutePath, preservedPage.RoutePath)
		assert.Equal(t, FAQRouteStatusCurrent, preservedPage.RouteStatus)

		pageData, err := faqService.GetPublicPageDataByRoutePath(newRoutePath, locale)
		require.NoError(t, err)
		require.Len(t, pageData.Items, 1)
		assert.Equal(t, "Existing freehub compatibility question", pageData.Items[0].Question)
		assert.Equal(t, "<p>Existing freehub compatibility answer</p>", pageData.Items[0].Answer)

		_, err = faqService.GetPublicPageDataByRoutePath(oldRoutePath, locale)
		assert.Error(t, err, "Retired freehub virtual-tab URLs must not resolve to FAQ pages")
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
	assert.Equal(t, len(locales.EnabledLocaleCodes()), currentLocaleCount)
}
