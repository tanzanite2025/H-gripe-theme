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

func TestReviewModerationUpdatesSummaryAtomically(t *testing.T) {
	db, moderationService := newTestReviewModerationService(t)

	pending := review.Review{
		ProductID: 42,
		UserID:    10,
		Rating:    5,
		Title:     "Pending review",
		Content:   "Ready for moderation",
		Status:    review.StatusPending,
	}
	require.NoError(t, db.Create(&pending).Error)

	updated, err := moderationService.UpdateStatus(pending.ID, review.StatusApproved, "Verified purchase", 99)
	require.NoError(t, err)
	assert.Equal(t, review.StatusApproved, updated.Status)
	assert.Equal(t, uint(99), *updated.ModeratedBy)
	assert.Equal(t, "Verified purchase", updated.ModerationReason)
	assert.NotNil(t, updated.ModeratedAt)

	summary, err := moderationService.reviewRepo.FindReviewSummaryByProductID(42)
	require.NoError(t, err)
	assert.Equal(t, 1, summary.TotalReviews)
	assert.InDelta(t, 5.0, summary.AverageRating, 0.001)
	assert.Equal(t, 1, summary.Rating5Count)
}

func TestReviewModerationRejectsAndRequeuesWithoutPublicSummary(t *testing.T) {
	db, moderationService := newTestReviewModerationService(t)

	pending := review.Review{
		ProductID: 42,
		UserID:    10,
		Rating:    2,
		Status:    review.StatusPending,
	}
	require.NoError(t, db.Create(&pending).Error)

	_, err := moderationService.UpdateStatus(pending.ID, review.StatusRejected, "Content policy", 99)
	require.NoError(t, err)

	summary, err := moderationService.reviewRepo.FindReviewSummaryByProductID(42)
	require.NoError(t, err)
	assert.Equal(t, 0, summary.TotalReviews)
	assert.Equal(t, 0, summary.Rating2Count)

	requeued, err := moderationService.UpdateStatus(pending.ID, review.StatusPending, "Needs another pass", 100)
	require.NoError(t, err)
	assert.Equal(t, review.StatusPending, requeued.Status)
	assert.Nil(t, requeued.ModeratedAt)
	assert.Nil(t, requeued.ModeratedBy)

	summary, err = moderationService.reviewRepo.FindReviewSummaryByProductID(42)
	require.NoError(t, err)
	assert.Equal(t, 0, summary.TotalReviews)
}

func TestReviewModerationRejectsInvalidTransitions(t *testing.T) {
	db, moderationService := newTestReviewModerationService(t)

	approved := review.Review{
		ProductID: 42,
		UserID:    10,
		Rating:    5,
		Status:    review.StatusApproved,
	}
	require.NoError(t, db.Create(&approved).Error)

	_, err := moderationService.UpdateStatus(approved.ID, review.StatusRejected, "duplicate", 99)
	assert.True(t, errors.Is(err, ErrInvalidReviewTransition))
}

func TestReviewModerationRequiresReasonForRejection(t *testing.T) {
	db, moderationService := newTestReviewModerationService(t)

	pending := review.Review{
		ProductID: 42,
		UserID:    10,
		Rating:    1,
		Status:    review.StatusPending,
	}
	require.NoError(t, db.Create(&pending).Error)

	_, err := moderationService.UpdateStatus(pending.ID, review.StatusRejected, "  ", 99)
	assert.True(t, errors.Is(err, ErrReviewModerationReason))
}

func newTestReviewModerationService(t *testing.T) (*gorm.DB, *ReviewModerationService) {
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

	require.NoError(t, db.AutoMigrate(&user.User{}, &review.Review{}, &review.ReviewHelpful{}, &review.ReviewSummary{}))
	return db, NewReviewModerationService(repository.NewReviewRepository(db))
}
