package service

import (
	"errors"
	"testing"

	settingdomain "tanzanite/internal/domain/setting"

	"github.com/stretchr/testify/require"
)

func TestAdminSettingsRejectsDomainManagedGroupsAndKeys(t *testing.T) {
	_, settingService := newTestSettingService(t)
	adminSettings := NewAdminSettingsService(settingService)

	_, err := adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:    "anything",
		Group:  "loyalty",
		Locale: "en",
	})
	require.True(t, errors.Is(err, ErrSettingManagedByDomainService))

	_, err = adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:    "tz_redeem_exchange_rate",
		Group:  "site",
		Locale: "en",
	})
	require.True(t, errors.Is(err, ErrSettingManagedByDomainService))
}

func TestAdminSettingsFiltersDomainManagedSettingsFromGenericLists(t *testing.T) {
	_, settingService := newTestSettingService(t)
	adminSettings := NewAdminSettingsService(settingService)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "Tanzanite", Locale: "en", Group: "site", IsPublic: true},
		{Key: "tz_loyalty_checkin_base_points", Value: "10", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_redeem_exchange_rate", Value: "100", Locale: "en", Group: "redeem", IsPublic: true},
	}))

	settings, err := adminSettings.ListSettings("en", "")
	require.NoError(t, err)
	require.Len(t, settings, 1)
	require.Equal(t, "site_name", settings[0].Key)

	groups, err := adminSettings.GetGroups()
	require.NoError(t, err)
	require.Equal(t, []string{"site"}, groups)
}
