package service

import (
	"tanzanite/internal/domain/faq"
	"tanzanite/internal/repository"
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
