package service

import (
	"context"
	"testing"

	"commerce-platform/internal/domain/faq"
	seodomain "commerce-platform/internal/domain/seo"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestReconcileStorefrontRoutesRepairsPathsCreatesShellsAndPreservesContent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:faq-route-reconciler?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&faq.FAQPage{}, &faq.FAQ{}))

	legacy := faq.FAQPage{
		PageID:    "products-spoke-calculator",
		RoutePath: "/spoke-calculator",
		Locale:    "en",
		Title:     "Custom calculator title",
		Subtitle:  "Keep this editorial text",
		Status:    "active",
	}
	stale := faq.FAQPage{
		PageID:    "support-product-feedback",
		RoutePath: "/support/product-feedback",
		Locale:    "en",
		Title:     "Feedback",
		Status:    "active",
	}
	missing := faq.FAQPage{
		PageID: "missing-route",
		Locale: "en",
		Title:  "Missing route",
		Status: "active",
	}
	require.NoError(t, db.Create(&legacy).Error)
	require.NoError(t, db.Create(&stale).Error)
	require.NoError(t, db.Create(&missing).Error)
	require.NoError(t, db.Create(&faq.FAQ{
		PageID:   legacy.PageID,
		Locale:   "en",
		Question: "Keep answer",
		Answer:   "<p>Existing content</p>",
		Status:   "published",
	}).Error)

	service := NewFAQService(repository.NewFAQRepository(db), nil)
	summary, err := service.ReconcileStorefrontRoutes(context.Background(), seodomain.StorefrontRouteManifest{
		Version: "test-manifest",
		Routes: []seodomain.StorefrontRouteManifestRoute{
			{
				Key:         "resources-spoke-calculator",
				Path:        "/resources/spoke-calculator",
				Label:       "Spoke Calculator",
				Description: "Spoke length calculator",
			},
		},
	})
	require.NoError(t, err)
	require.Greater(t, summary.Created, 0)

	var repaired faq.FAQPage
	require.NoError(t, db.Where("page_id = ? AND locale = ?", legacy.PageID, "en").First(&repaired).Error)
	assert.Equal(t, "/resources/spoke-calculator", repaired.RoutePath)
	assert.Equal(t, "resources-spoke-calculator", repaired.RouteKey)
	assert.Equal(t, FAQRouteStatusCurrent, repaired.RouteStatus)
	assert.Equal(t, "Custom calculator title", repaired.Title)
	assert.Equal(t, "Keep this editorial text", repaired.Subtitle)

	var existingFAQ faq.FAQ
	require.NoError(t, db.Where("page_id = ?", legacy.PageID).First(&existingFAQ).Error)
	assert.Equal(t, "Keep answer", existingFAQ.Question)

	pageData, err := service.GetPublicPageDataByRoutePath("/resources/spoke-calculator", "en")
	require.NoError(t, err)
	require.Len(t, pageData.Items, 1)
	assert.Equal(t, "Keep answer", pageData.Items[0].Question)

	var stalePage faq.FAQPage
	require.NoError(t, db.Where("page_id = ? AND locale = ?", stale.PageID, "en").First(&stalePage).Error)
	assert.Equal(t, FAQRouteStatusStale, stalePage.RouteStatus)
	_, err = service.GetPublicPageDataByRoutePath(stalePage.RoutePath, "en")
	assert.ErrorIs(t, err, ErrFAQNotFound)

	var missingPage faq.FAQPage
	require.NoError(t, db.Where("page_id = ? AND locale = ?", missing.PageID, "en").First(&missingPage).Error)
	assert.Equal(t, FAQRouteStatusMissing, missingPage.RouteStatus)

	secondSummary, err := service.ReconcileStorefrontRoutes(context.Background(), seodomain.StorefrontRouteManifest{
		Version: "test-manifest",
		Routes: []seodomain.StorefrontRouteManifestRoute{
			{Key: "resources-spoke-calculator", Path: "/resources/spoke-calculator", Label: "Spoke Calculator"},
		},
	})
	require.NoError(t, err)
	assert.Zero(t, secondSummary.Created)
	assert.Zero(t, secondSummary.Updated)
	assert.Zero(t, secondSummary.Stale)

	var missingPageAfterSecondSync faq.FAQPage
	require.NoError(t, db.Where("page_id = ? AND locale = ?", missing.PageID, "en").First(&missingPageAfterSecondSync).Error)
	assert.Equal(t, FAQRouteStatusMissing, missingPageAfterSecondSync.RouteStatus)
}
