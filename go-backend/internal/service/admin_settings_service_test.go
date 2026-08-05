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

	_, err = adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:    "payment_gateway_stripe",
		Group:  "site",
		Locale: "global",
	})
	require.True(t, errors.Is(err, ErrSettingManagedByDomainService))

	_, err = adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:    "anything",
		Group:  "payment_secret",
		Locale: "global",
	})
	require.True(t, errors.Is(err, ErrSettingManagedByDomainService))

	_, err = adminSettings.GetSetting("payment_gateway_paypal", "global")
	require.True(t, errors.Is(err, ErrSettingManagedByDomainService))

	err = adminSettings.DeleteSetting("payment_gateway_wechat", "global")
	require.True(t, errors.Is(err, ErrSettingManagedByDomainService))
}

func TestAdminSettingsDomainManagedMethodsAllowSpecializedHandlers(t *testing.T) {
	_, settingService := newTestSettingService(t)
	adminSettings := NewAdminSettingsService(settingService)

	stored, err := adminSettings.UpdateDomainManagedSetting(settingdomain.UpdateSettingRequest{
		Key:      "payment_gateway_stripe",
		Value:    "encrypted-payload",
		Type:     "encrypted_json",
		Group:    "payment_secret",
		Locale:   "global",
		IsPublic: false,
	})
	require.NoError(t, err)
	require.Equal(t, "payment_gateway_stripe", stored.Key)
	require.Equal(t, "payment_secret", stored.Group)

	loaded, err := adminSettings.GetDomainManagedSetting("payment_gateway_stripe", "global")
	require.NoError(t, err)
	require.Equal(t, "encrypted-payload", loaded.Value)

	require.NoError(t, adminSettings.DeleteDomainManagedSetting("payment_gateway_stripe", "global"))
}

func TestAdminSettingsFiltersDomainManagedSettingsFromGenericLists(t *testing.T) {
	_, settingService := newTestSettingService(t)
	adminSettings := NewAdminSettingsService(settingService)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "site_name", Value: "Tanzanite", Locale: "en", Group: "site", IsPublic: true},
		{Key: "tz_loyalty_checkin_base_points", Value: "10", Locale: "en", Group: "loyalty", IsPublic: true},
		{Key: "tz_redeem_exchange_rate", Value: "100", Locale: "en", Group: "redeem", IsPublic: true},
		{Key: "payment_gateway_stripe", Value: "encrypted", Locale: "global", Group: "payment_secret", IsPublic: false},
	}))

	settings, err := adminSettings.ListSettings("en", "")
	require.NoError(t, err)
	require.Len(t, settings, 1)
	require.Equal(t, "site_name", settings[0].Key)

	groups, err := adminSettings.GetGroups()
	require.NoError(t, err)
	require.Equal(t, []string{"site"}, groups)
}
