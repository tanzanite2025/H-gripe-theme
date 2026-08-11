package service

import (
	"errors"
	"testing"

	"commerce-platform/internal/domain/review"
	"commerce-platform/internal/domain/user"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestReviewPublicAccessRequiresApprovedStatus(t *testing.T) {
	db, reviewService := newTestReviewService(t)

	pending := review.Review{
		ProductID: 1,
		UserID:    10,
		Rating:    5,
		Title:     "Pending",
		Content:   "Not public yet",
		Status:    "pending",
	}
	require.NoError(t, db.Create(&pending).Error)

	_, err := reviewService.GetPublicReview(pending.ID)
	assert.True(t, errors.Is(err, ErrReviewNotPublic))

	err = reviewService.MarkHelpful(pending.ID, 20, true)
	assert.True(t, errors.Is(err, ErrReviewNotPublic))

	var helpfulCount int64
	require.NoError(t, db.Model(&review.ReviewHelpful{}).Count(&helpfulCount).Error)
	assert.EqualValues(t, 0, helpfulCount)
}

func TestReviewPublicAccessAllowsApprovedStatus(t *testing.T) {
	db, reviewService := newTestReviewService(t)

	approved := review.Review{
		ProductID: 1,
		UserID:    10,
		Rating:    5,
		Title:     "Approved",
		Content:   "Public",
		Status:    "approved",
	}
	require.NoError(t, db.Create(&approved).Error)

	got, err := reviewService.GetPublicReview(approved.ID)
	require.NoError(t, err)
	assert.Equal(t, approved.ID, got.ID)
}

func newTestReviewService(t *testing.T) (*gorm.DB, *ReviewService) {
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

	require.NoError(t, db.AutoMigrate(&user.User{}, &review.Review{}, &review.ReviewHelpful{}))
	return db, NewReviewService(repository.NewReviewRepository(db))
}
