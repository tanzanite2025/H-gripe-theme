package service

import (
	"commerce-platform/internal/domain/setting"
	"errors"
	"fmt"
	"strconv"
)

// GetLoyaltySettings 获取公开的积分获取规则。
func (s *SettingService) GetLoyaltySettings(locale string) (*setting.LoyaltySettings, error) {
	if locale == "" {
		locale = "en"
	}

	settings, err := s.GetPublicByGroup("loyalty", locale)
	if err != nil {
		return nil, err
	}
	if len(settings) == 0 && locale != "en" {
		settings, err = s.GetPublicByGroup("loyalty", "en")
		if err != nil {
			return nil, err
		}
	}

	values := make(map[string]string, len(settings))
	for _, item := range settings {
		values[item.Key] = item.Value
	}

	result := setting.DefaultLoyaltySettings()
	fields := []struct {
		key    string
		target *int
	}{
		{"tz_loyalty_referral_referrer_points", &result.ReferralReferrerPoints},
		{"tz_loyalty_referral_referee_points", &result.ReferralRefereePoints},
		{"tz_loyalty_checkin_base_points", &result.CheckInBasePoints},
		{"tz_loyalty_checkin_streak_interval_days", &result.CheckInStreakIntervalDays},
		{"tz_loyalty_checkin_streak_bonus_points", &result.CheckInStreakBonusPoints},
		{"tz_loyalty_checkin_max_points", &result.CheckInMaxPoints},
	}

	for _, field := range fields {
		raw, ok := values[field.key]
		if !ok {
			return nil, fmt.Errorf("[CRITICAL] Loyalty setting '%s' is missing in settings table", field.key)
		}
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil || parsed < 0 {
			return nil, fmt.Errorf("[CRITICAL] Invalid format for %s", field.key)
		}
		*field.target = parsed
	}

	if result.CheckInStreakIntervalDays <= 0 {
		return nil, errors.New("[CRITICAL] Loyalty check-in streak interval must be greater than zero")
	}
	if result.CheckInMaxPoints < result.CheckInBasePoints {
		return nil, errors.New("[CRITICAL] Loyalty check-in max points cannot be lower than base points")
	}

	return &result, nil
}
