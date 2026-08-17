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

func TestFAQServicePublicResponsesCanonicalizeAnswerImageURL(t *testing.T) {
	db, faqService := newTestFAQService(t)
	faqService.ConfigureMediaService(NewMediaService(nil, nil, nil, "https://shop.example.test", 20<<30))

	publishedFAQ := faqdomain.FAQ{
		Question:       "Image?",
		Answer:         "<p>Yes</p>",
		AnswerImageURL: "http://media.internal:8080/uploads/faq/answer.webp",
		Status:         "published",
		Locale:         "en",
	}
	require.NoError(t, db.Create(&publishedFAQ).Error)

	publicFAQ, err := faqService.GetPublicByID(publishedFAQ.ID)
	require.NoError(t, err)
	assert.Equal(t, "https://shop.example.test/uploads/faq/answer.webp", publicFAQ.AnswerImageURL)

	items, _, err := faqService.List("en", "", "", "published", 1, 20)
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "https://shop.example.test/uploads/faq/answer.webp", items[0].AnswerImageURL)
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
