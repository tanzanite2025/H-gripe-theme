package service

import (
	"testing"

	"commerce-platform/internal/domain/feedback"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestFeedbackListPublicOnlyReturnsApprovedItems(t *testing.T) {
	db, feedbackService := newTestFeedbackService(t)

	require.NoError(t, db.Create(&feedback.Feedback{
		ThreadKey: "product:1",
		UserID:    1,
		Content:   "approved",
		Status:    "approved",
	}).Error)
	require.NoError(t, db.Create(&feedback.Feedback{
		ThreadKey: "product:1",
		UserID:    2,
		Content:   "pending",
		Status:    "pending",
	}).Error)

	items, total, err := feedbackService.ListPublic("product:1", "", 1, 20)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "approved", items[0].Content)
}

func newTestFeedbackService(t *testing.T) (*gorm.DB, *FeedbackService) {
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

	require.NoError(t, db.AutoMigrate(&feedback.Feedback{}))
	return db, NewFeedbackService(repository.NewFeedbackRepository(db))
}
