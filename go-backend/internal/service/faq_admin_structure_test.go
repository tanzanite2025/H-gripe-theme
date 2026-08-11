package service

import (
	"commerce-platform/internal/domain/faq"
	"commerce-platform/internal/repository"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreateAdminCategoryRequiresExistingLocalizedPage(t *testing.T) {
	t.Parallel()

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

	require.NoError(t, db.AutoMigrate(&faq.FAQPage{}, &faq.FAQCategory{}))

	svc := NewFAQService(repository.NewFAQRepository(db), nil)

	_, err = svc.CreateAdminCategory(FAQCategoryAdminInput{
		PageID:      "support",
		CategoryKey: "billing",
		Name:        "Billing",
		Locale:      "en",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not exist")

	require.NoError(t, db.Create(&faq.FAQPage{
		PageID: "support",
		Locale: "en",
		Title:  "Support",
	}).Error)

	category, err := svc.CreateAdminCategory(FAQCategoryAdminInput{
		PageID:      "support",
		CategoryKey: "billing",
		Name:        "Billing",
		Locale:      "en",
	})
	require.NoError(t, err)
	require.Equal(t, "support", category.PageID)
	require.Equal(t, "en", category.Locale)
}

func TestUpsertAdminPageKeepsExistingRoutePath(t *testing.T) {
	t.Parallel()

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

	require.NoError(t, db.AutoMigrate(&faq.FAQPage{}, &faq.FAQCategory{}))
	require.NoError(t, db.Create(&faq.FAQPage{
		PageID:    "support-payment",
		RoutePath: "/support/payment",
		Domain:    "support",
		Locale:    "en",
		Title:     "Payment FAQs",
		Status:    "active",
	}).Error)

	svc := NewFAQService(repository.NewFAQRepository(db), nil)

	page, err := svc.UpsertAdminPage("support-payment", FAQPageAdminInput{
		RoutePath: "   ",
		Domain:    "support",
		Locale:    "en",
		Title:     "Updated Payment FAQs",
		Status:    "active",
	})

	require.NoError(t, err)
	require.Equal(t, "/support/payment", page.RoutePath)
	require.Equal(t, "Updated Payment FAQs", page.Title)
}
