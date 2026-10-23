package service

import (
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/domain/review"
	"commerce-platform/internal/repository"
)

func TestReadOnlyRecoveryReviewSummaryPreservesExistingCountsAndNeverCreatesMissingRows(t *testing.T) {
	database, ordinary := newTestReviewService(t)
	readOnly := NewReviewServiceForReadOnlyRecoveryVerification(repository.NewReviewRepository(database))
	missing, err := readOnly.GetReviewSummary(901)
	require.NoError(t, err)
	require.Zero(t, missing.TotalReviews)
	require.Zero(t, missing.AverageRating)
	var rows int64
	require.NoError(t, database.Model(&review.ReviewSummary{}).Count(&rows).Error)
	require.Zero(t, rows)
	existing := review.ReviewSummary{ProductID: 902, TotalReviews: 2, AverageRating: 4.5}
	require.NoError(t, database.Create(&existing).Error)
	actual, err := readOnly.GetReviewSummary(902)
	require.NoError(t, err)
	require.Equal(t, existing, *actual)
	_, err = ordinary.GetReviewSummary(901)
	require.NoError(t, err)
	require.NoError(t, database.Model(&review.ReviewSummary{}).Count(&rows).Error)
	require.EqualValues(t, 2, rows, "ordinary read behavior must still create its missing summary")
	closed, err := database.DB()
	require.NoError(t, err)
	require.NoError(t, closed.Close())
	_, err = readOnly.GetReviewSummary(903)
	require.Error(t, err, "a real query failure must not become empty successful counters")
}
