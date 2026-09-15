package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLegacyReferralWritesAreDisabled(t *testing.T) {
	marketing := &MarketingService{}

	assert.ErrorIs(t, marketing.CreateReferral(1, 2), ErrLegacyReferralWriteDisabled)
	assert.ErrorIs(t, marketing.CompleteReferral(2, 3), ErrLegacyReferralWriteDisabled)
	_, err := marketing.UpdateReferralStatus(1, "completed")
	assert.ErrorIs(t, err, ErrLegacyReferralWriteDisabled)
}
