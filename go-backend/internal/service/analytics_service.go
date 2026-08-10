package service

import (
	"errors"
	"fmt"

	"tanzanite/internal/domain/analytics"
)

var ErrInvalidAnalyticsSettings = errors.New("invalid analytics settings")

const analyticsIdentifierLimit = 128

type AnalyticsService struct {
	settings *SettingService
}

func NewAnalyticsService(settings *SettingService) *AnalyticsService {
	return &AnalyticsService{settings: settings}
}

func (s *AnalyticsService) Get(locale string) (*analytics.Settings, error) {
	if s == nil || s.settings == nil {
		return nil, errors.New("analytics service is not configured")
	}

	values, err := loadPublicManagedSettingValues(s.settings, analytics.Group, locale)
	if err != nil {
		return nil, err
	}

	return &analytics.Settings{
		GoogleAnalytics:  values["google_analytics"],
		GoogleTagManager: values["google_tag_manager"],
	}, nil
}

func (s *AnalyticsService) Update(request analytics.UpdateRequest) (*analytics.Settings, error) {
	if s == nil || s.settings == nil {
		return nil, errors.New("analytics service is not configured")
	}

	locale := normalizeManagedSettingLocale(request.Locale)
	current, err := s.Get(locale)
	if err != nil {
		return nil, err
	}

	next := *current
	if request.GoogleAnalytics != nil {
		next.GoogleAnalytics = *request.GoogleAnalytics
	}
	if request.GoogleTagManager != nil {
		next.GoogleTagManager = *request.GoogleTagManager
	}

	normalized, err := normalizeAnalyticsSettings(next)
	if err != nil {
		return nil, err
	}

	values := map[string]string{
		"google_analytics":   normalized.GoogleAnalytics,
		"google_tag_manager": normalized.GoogleTagManager,
	}
	descriptions := map[string]string{
		"google_analytics":   "Google Analytics measurement ID",
		"google_tag_manager": "Google Tag Manager container ID",
	}
	if err := s.settings.BatchSet(managedSettingRecords(analytics.Group, locale, values, descriptions)); err != nil {
		return nil, err
	}

	return &normalized, nil
}

func normalizeAnalyticsSettings(values analytics.Settings) (analytics.Settings, error) {
	values.GoogleAnalytics = managedSettingValue(values.GoogleAnalytics)
	values.GoogleTagManager = managedSettingValue(values.GoogleTagManager)

	if len([]rune(values.GoogleAnalytics)) > analyticsIdentifierLimit {
		return analytics.Settings{}, fmt.Errorf("%w: google_analytics exceeds %d characters", ErrInvalidAnalyticsSettings, analyticsIdentifierLimit)
	}
	if len([]rune(values.GoogleTagManager)) > analyticsIdentifierLimit {
		return analytics.Settings{}, fmt.Errorf("%w: google_tag_manager exceeds %d characters", ErrInvalidAnalyticsSettings, analyticsIdentifierLimit)
	}

	return values, nil
}
