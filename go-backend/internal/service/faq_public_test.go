package service

import (
	"testing"

	faqdomain "commerce-platform/internal/domain/faq"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFAQServicePublicAccessRequiresPublishedFAQ(t *testing.T) {
	db, faqService := newTestFAQService(t)

	publishedFAQ := faqdomain.FAQ{
		Question: "Published?",
		Answer:   "<p>Yes</p>",
		Status:   "published",
		Locale:   "en",
	}
	require.NoError(t, db.Create(&publishedFAQ).Error)

	draftFAQ := faqdomain.FAQ{
		Question: "Draft?",
		Answer:   "<p>No</p>",
		Status:   "draft",
		Locale:   "en",
	}
	require.NoError(t, db.Create(&draftFAQ).Error)

	_, err := faqService.GetPublicByID(draftFAQ.ID)
	require.ErrorIs(t, err, ErrFAQNotFound)

	err = faqService.IncrementPublicViewCount(draftFAQ.ID)
	require.ErrorIs(t, err, ErrFAQNotFound)

	require.NoError(t, faqService.IncrementPublicViewCount(publishedFAQ.ID))

	var refreshed faqdomain.FAQ
	require.NoError(t, db.First(&refreshed, publishedFAQ.ID).Error)
	assert.Equal(t, 1, refreshed.ViewCount)
}

func TestFAQServicePublicPageRejectsHiddenPage(t *testing.T) {
	db, faqService := newTestFAQService(t)

	hiddenPage := faqdomain.FAQPage{
		PageID:    "hidden-page",
		RoutePath: "/hidden",
		Locale:    "en",
		Title:     "Hidden",
		Status:    "hidden",
	}
	require.NoError(t, db.Create(&hiddenPage).Error)

	_, err := faqService.GetPublicPageData(hiddenPage.PageID, hiddenPage.Locale)
	require.ErrorIs(t, err, ErrFAQNotFound)
}

func newTestFAQService(t *testing.T) (*gorm.DB, *FAQService) {
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

	require.NoError(t, db.AutoMigrate(
		&faqdomain.FAQ{},
		&faqdomain.FAQPage{},
		&faqdomain.FAQCategory{},
	))

	return db, NewFAQService(repository.NewFAQRepository(db), nil)
}
