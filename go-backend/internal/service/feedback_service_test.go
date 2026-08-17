package service

import (
	"testing"
	"time"

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

func TestFeedbackListPublicAcceptsPathStyleThreadKeys(t *testing.T) {
	db, feedbackService := newTestFeedbackService(t)

	require.NoError(t, db.Create(&feedback.Feedback{
		ThreadKey: "/support/payment",
		UserID:    1,
		Content:   "approved",
		Status:    "approved",
	}).Error)

	items, total, err := feedbackService.ListPublic("/support/payment", "", 1, 20)
	require.NoError(t, err)
	assert.EqualValues(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "approved", items[0].Content)
}

func TestFeedbackRiskOverviewFlagsSourceBurstAndRateLimitBlocks(t *testing.T) {
	db, feedbackService := newTestFeedbackService(t)
	now := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)

	for index, pagePath := range []string{
		"/support/payment",
		"/support/shipping",
		"/support/faqs",
		"/support/payment",
		"/support/shipping",
	} {
		require.NoError(t, db.Create(&feedback.Feedback{
			ThreadKey:  "page:burst:" + string(rune('a'+index)),
			PagePath:   pagePath,
			PageTitle:  "Support page",
			SourceHash: "source-fingerprint-123456",
			UserID:     7,
			Content:    "Repeated submission",
			Status:     "pending",
			CreatedAt:  now.Add(-time.Duration(index+1) * time.Minute),
		}).Error)
	}

	require.NoError(t, db.Create(&feedback.Feedback{
		ThreadKey: "page:stale",
		PagePath:  "/support/returns",
		Content:   "Old pending feedback",
		Status:    "pending",
		CreatedAt: now.Add(-25 * time.Hour),
	}).Error)

	overview, err := feedbackService.RiskOverview(FeedbackRiskOverviewInput{
		WindowHours: 24,
		GeneratedAt: now,
		RateLimit: FeedbackRateLimitSnapshot{
			WindowHours: 24,
			Total:       1,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, overview)
	assert.Equal(t, "critical", overview.Level)
	assert.EqualValues(t, 6, overview.Totals.PendingTotal)
	assert.EqualValues(t, 1, overview.Totals.PendingOver24Hours)
	assert.EqualValues(t, 5, overview.Totals.WindowTotal)
	assert.Len(t, overview.SourceBursts, 1)
	assert.EqualValues(t, 3, overview.SourceBursts[0].PageCount)
	assert.EqualValues(t, 5, overview.SourceBursts[0].FeedbackCount)
	assert.NotEqual(t, "source-fingerprint-123456", overview.SourceBursts[0].SourceHashPreview)
}

func TestFeedbackRiskOverviewFlagsRedisUnavailable(t *testing.T) {
	_, feedbackService := newTestFeedbackService(t)
	now := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)

	overview, err := feedbackService.RiskOverview(FeedbackRiskOverviewInput{
		WindowHours: 24,
		GeneratedAt: now,
		RateLimit: FeedbackRateLimitSnapshot{
			WindowHours:      24,
			RedisUnavailable: 1,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, overview)
	assert.Equal(t, "warning", overview.Level)

	overview, err = feedbackService.RiskOverview(FeedbackRiskOverviewInput{
		WindowHours: 24,
		GeneratedAt: now,
		RateLimit: FeedbackRateLimitSnapshot{
			WindowHours:      24,
			RedisUnavailable: 20,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, overview)
	assert.Equal(t, "critical", overview.Level)
}

func TestFeedbackRiskOverviewKeepsLegacyThreadKeyFilterable(t *testing.T) {
	db, feedbackService := newTestFeedbackService(t)
	now := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)

	require.NoError(t, db.Create(&feedback.Feedback{
		ThreadKey: "legacy-thread",
		UserID:    7,
		Content:   "Old record without page metadata",
		Status:    "pending",
		CreatedAt: now.Add(-time.Minute),
	}).Error)

	overview, err := feedbackService.RiskOverview(FeedbackRiskOverviewInput{
		WindowHours: 24,
		GeneratedAt: now,
	})
	require.NoError(t, err)
	require.NotNil(t, overview)
	require.Len(t, overview.HotPages, 1)
	assert.Equal(t, "legacy-thread", overview.HotPages[0].PagePath)
	assert.Equal(t, "thread_key", overview.HotPages[0].FilterKind)
	assert.Equal(t, "legacy-thread", overview.HotPages[0].FilterValue)
}

func TestFeedbackRiskOverviewDoesNotMergeLegacyThreadAndPagePath(t *testing.T) {
	db, feedbackService := newTestFeedbackService(t)
	now := time.Date(2026, time.August, 17, 12, 0, 0, 0, time.UTC)

	require.NoError(t, db.Create(&feedback.Feedback{
		ThreadKey: "shared-key",
		UserID:    7,
		Content:   "Legacy thread",
		Status:    "pending",
		CreatedAt: now.Add(-2 * time.Minute),
	}).Error)
	require.NoError(t, db.Create(&feedback.Feedback{
		ThreadKey: "new-thread",
		PagePath:  "shared-key",
		UserID:    8,
		Content:   "New page path",
		Status:    "pending",
		CreatedAt: now.Add(-time.Minute),
	}).Error)

	overview, err := feedbackService.RiskOverview(FeedbackRiskOverviewInput{
		WindowHours: 24,
		GeneratedAt: now,
	})
	require.NoError(t, err)
	require.Len(t, overview.HotPages, 2)

	kinds := map[string]bool{}
	for _, page := range overview.HotPages {
		kinds[page.FilterKind] = true
		assert.Equal(t, "shared-key", page.FilterValue)
	}
	assert.True(t, kinds["page_path"])
	assert.True(t, kinds["thread_key"])
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
