package shipping

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClassifyYanwenTrackingStatusCoversFiveStagesAndExceptionOverlay(t *testing.T) {
	tests := []struct {
		name      string
		status    string
		expected  YanwenTrackingStage
		rank      int
		exception bool
	}{
		{name: "collected", status: "OR10", expected: YanwenTrackingStageCollected, rank: 1},
		{name: "air transit", status: "LH20", expected: YanwenTrackingStageInTransitAir, rank: 2},
		{name: "customs", status: "IC60", expected: YanwenTrackingStageCustoms, rank: 3},
		{name: "last mile", status: "LM25", expected: YanwenTrackingStageLastMile, rank: 4},
		{name: "delivered", status: "LM40", expected: YanwenTrackingStageDelivered, rank: 5},
		{name: "customs exception does not guess a stage", status: "IC51", expected: YanwenTrackingStageUnknown, exception: true},
		{name: "returned does not guess a stage", status: "LM90", expected: YanwenTrackingStageUnknown, exception: true},
		{name: "unknown status is not guessed", status: "NEW_STATUS", expected: YanwenTrackingStageUnknown},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			classification := ClassifyYanwenTrackingStatus(test.status)
			require.Equal(t, test.expected, classification.Stage)
			require.Equal(t, test.rank, classification.StageRank)
			require.Equal(t, test.exception, classification.IsException)
		})
	}
}

func TestClassifyYanwenTrackingProgressUsesOnlyKnownOfficialCheckpointStagesForExceptions(t *testing.T) {
	classification := ClassifyYanwenTrackingProgress("IC51", []string{"OR10", "LH20", "IC50", "IC51"})
	require.Equal(t, YanwenTrackingStageCustoms, classification.Stage)
	require.Equal(t, 3, classification.StageRank)
	require.True(t, classification.IsException)

	unknown := ClassifyYanwenTrackingProgress("LM90", []string{"LM90"})
	require.Equal(t, YanwenTrackingStageUnknown, unknown.Stage)
	require.Zero(t, unknown.StageRank)
	require.True(t, unknown.IsException)
}
