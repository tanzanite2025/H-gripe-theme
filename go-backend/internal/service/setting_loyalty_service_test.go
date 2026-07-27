package service

import (
	settingdomain "tanzanite/internal/domain/setting"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLoyaltySettingsParsesConfiguredRules(t *testing.T) {
	_, settingService := newTestSettingService(t)

	settings := []settingdomain.Setting{
		{Key: "tz_loyalty_referral_referrer_points", Value: "120", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_loyalty_referral_referee_points", Value: "60", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_loyalty_checkin_base_points", Value: "8", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_loyalty_checkin_streak_interval_days", Value: "5", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_loyalty_checkin_streak_bonus_points", Value: "3", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_loyalty_checkin_max_points", Value: "30", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
	}
	require.NoError(t, settingService.BatchSet(settings))

	actual, err := settingService.GetLoyaltySettings("en")
	require.NoError(t, err)
	assert.Equal(t, 120, actual.ReferralReferrerPoints)
	assert.Equal(t, 60, actual.ReferralRefereePoints)
	assert.Equal(t, 8, actual.CheckInBasePoints)
	assert.Equal(t, 5, actual.CheckInStreakIntervalDays)
	assert.Equal(t, 3, actual.CheckInStreakBonusPoints)
	assert.Equal(t, 30, actual.CheckInMaxPoints)
}

func TestGetLoyaltySettingsFallsBackToEnglishLocale(t *testing.T) {
	_, settingService := newTestSettingService(t)
	defaults := settingdomain.DefaultLoyaltySettings()

	settings := []settingdomain.Setting{
		{Key: "tz_loyalty_referral_referrer_points", Value: "100", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_loyalty_referral_referee_points", Value: "50", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_loyalty_checkin_base_points", Value: "10", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_loyalty_checkin_streak_interval_days", Value: "7", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_loyalty_checkin_streak_bonus_points", Value: "5", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_loyalty_checkin_max_points", Value: "50", Type: "number", Locale: "en", Group: "loyalty", IsPublic: true},
	}
	require.NoError(t, settingService.BatchSet(settings))

	actual, err := settingService.GetLoyaltySettings("zh_cn")
	require.NoError(t, err)
	assert.Equal(t, defaults, *actual)
}
