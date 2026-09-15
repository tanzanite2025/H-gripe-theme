package loyalty

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReferralRecordStateMachine(t *testing.T) {
	allowed := map[string][]string{
		ReferralStatusPending:  {ReferralStatusOrdered, ReferralStatusExpired, ReferralStatusRevoked},
		ReferralStatusOrdered:  {ReferralStatusVesting, ReferralStatusRevoked},
		ReferralStatusVesting:  {ReferralStatusSettled, ReferralStatusRevoked},
		ReferralStatusSettled:  {ReferralStatusReversed},
		ReferralStatusExpired:  {},
		ReferralStatusRevoked:  {},
		ReferralStatusReversed: {},
	}
	allStatuses := []string{
		ReferralStatusPending,
		ReferralStatusOrdered,
		ReferralStatusVesting,
		ReferralStatusSettled,
		ReferralStatusExpired,
		ReferralStatusRevoked,
		ReferralStatusReversed,
	}

	for current, expectedNext := range allowed {
		record := ReferralRecord{Status: current}
		for _, next := range allStatuses {
			assert.Equal(t, containsStatus(expectedNext, next), record.CanTransitionTo(next), "%s -> %s", current, next)
		}
	}
}

func TestReferralCodeGenerationAndValidation(t *testing.T) {
	code, err := GenerateReferralCode()
	require.NoError(t, err)
	require.Len(t, code, ReferralCodeLength)
	require.NoError(t, ValidateReferralCode(code))

	assert.Equal(t, "ABCD23", NormalizeReferralCode("  abcd23 "))
	assert.NoError(t, ValidateReferralCode("ABCD23"))
	assert.ErrorIs(t, ValidateReferralCode("SELF-I"), ErrInvalidReferralCode)
	assert.ErrorIs(t, ValidateReferralCode("SHORT"), ErrInvalidReferralCode)
}

func containsStatus(statuses []string, target string) bool {
	for _, status := range statuses {
		if status == target {
			return true
		}
	}
	return false
}
