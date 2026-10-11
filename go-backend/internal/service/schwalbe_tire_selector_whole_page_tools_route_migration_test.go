package service

import (
	"context"
	"os"
	"testing"

	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSchwalbeTireSelectorToolsRouteMigrationPreservesExistingFAQContentAndUpdatesCatalogPaths(t *testing.T) {
	manifestData, err := os.ReadFile("../../../nuxt-i18n/public/storefront-route-manifest.json")
	require.NoError(t, err)
	manifest, err := decodeRouteManifest(manifestData)
	require.NoError(t, err)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&faq.FAQPage{}, &faq.FAQ{}))

	for _, locale := range []string{"en", "zh_cn"} {
		require.NoError(t, db.Create(&faq.FAQPage{
			PageID: "guides-schwalbe-tire-selector", RoutePath: "/guides/tireguides/schwalbe-tire-selector",
			RouteKey: "guides-tireguides-schwalbe-tire-selector", Locale: locale,
			Title: "Existing editorial title", Domain: "guides", Status: "active",
		}).Error)
		require.NoError(t, db.Create(&faq.FAQ{
			PageID: "guides-schwalbe-tire-selector", Locale: locale,
			Question: "Existing tire question", Answer: "<p>Existing tire answer</p>", Status: "published",
		}).Error)
	}

	faqService := NewFAQService(repository.NewFAQRepository(db), nil)
	_, err = faqService.ReconcileStorefrontRoutes(context.Background(), manifest)
	require.NoError(t, err)

	for _, locale := range []string{"en", "zh_cn"} {
		var migratedPage faq.FAQPage
		require.NoError(t, db.Where("page_id = ? AND locale = ?", "guides-schwalbe-tire-selector", locale).First(&migratedPage).Error)
		assert.Equal(t, "/tools/schwalbe-tire-selector", migratedPage.RoutePath)
		assert.Equal(t, "guides", migratedPage.Domain, "Editorial FAQ grouping must be preserved")
		assert.Equal(t, "Existing editorial title", migratedPage.Title)
		pageData, err := faqService.GetPublicPageDataByRoutePath("/tools/schwalbe-tire-selector", locale)
		require.NoError(t, err)
		require.Len(t, pageData.Items, 1)
		assert.Equal(t, "Existing tire question", pageData.Items[0].Question)
		assert.Equal(t, "<p>Existing tire answer</p>", pageData.Items[0].Answer)
	}

	entries := buildManifestRouteCatalogEntries(manifest, &StorefrontRouteCatalogSyncSummary{})
	selectorLocales := 0
	for _, entry := range entries {
		assert.NotContains(t, entry.Path, "/guides/tireguides/schwalbe-tire-selector")
		if entry.SourceKey == "guides-tireguides-schwalbe-tire-selector" {
			selectorLocales++
			assert.Contains(t, entry.Path, "/tools/schwalbe-tire-selector")
			assert.Equal(t, entry.Path, entry.CanonicalPath)
			assert.True(t, entry.IsIndexable)
		}
	}
	assert.Equal(t, 20, selectorLocales)
}
