package service

import (
	"errors"
	"testing"

	seodomain "commerce-platform/internal/domain/seo"
	settingdomain "commerce-platform/internal/domain/setting"

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

	_, err = adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:   "anything",
		Group: "seo",
	})
	require.True(t, errors.Is(err, ErrSettingManagedByDomainService))

	_, err = adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:   "anything",
		Group: "analytics",
	})
	require.True(t, errors.Is(err, ErrSettingManagedByDomainService))

	_, err = adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:   seodomain.HomeKeys.MetaTitle,
		Group: "site",
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
		{Key: "site_name", Value: "Commerce Platform", Locale: "en", Group: "site", IsPublic: true},
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

func TestAdminSettingsSocialGroupOnlyExposesSupportedPlatforms(t *testing.T) {
	_, settingService := newTestSettingService(t)
	adminSettings := NewAdminSettingsService(settingService)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{Key: "facebook", Value: "https://facebook.example", Locale: "en", Group: "social", IsPublic: true},
		{Key: "x", Value: "https://x.example", Locale: "en", Group: "social", IsPublic: true},
		{Key: "twitter", Value: "https://old.example", Locale: "en", Group: "social", IsPublic: true},
		{Key: "linkedin", Value: "https://linkedin.example", Locale: "en", Group: "social", IsPublic: true},
		{Key: "wechat", Value: "wechat-id", Locale: "en", Group: "social", IsPublic: true},
	}))

	settings, err := adminSettings.GetByGroup("social", "en")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"facebook", "x"}, settingKeys(settings))

	_, err = adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:    "twitter",
		Value:  "https://old.example",
		Group:  "social",
		Locale: "en",
	})
	require.ErrorIs(t, err, ErrSettingInvalid)
}

func TestAdminSettingsMasksSecretsAndPreservesMaskedUpdates(t *testing.T) {
	_, settingService := newTestSettingService(t)
	adminSettings := NewAdminSettingsService(settingService)

	require.NoError(t, settingService.BatchSet([]settingdomain.Setting{
		{
			Key:      "customs_lookup_us_hts_api_key",
			Value:    "real-secret",
			Type:     "string",
			Group:    "api",
			Locale:   "en",
			IsPublic: false,
		},
	}))

	settings, err := adminSettings.GetByGroup("api", "en")
	require.NoError(t, err)
	require.Len(t, settings, 1)
	require.Equal(t, maskedSettingValue, settings[0].Value)

	updated, err := adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:      "customs_lookup_us_hts_api_key",
		Value:    maskedSettingValue,
		Type:     "string",
		Group:    "api",
		Locale:   "en",
		IsPublic: true,
	})
	require.NoError(t, err)
	require.Equal(t, maskedSettingValue, updated.Value)

	stored, err := settingService.Get("customs_lookup_us_hts_api_key", "en")
	require.NoError(t, err)
	require.Equal(t, "real-secret", stored.Value)
	require.False(t, stored.IsPublic)
}

func settingKeys(items []settingdomain.Setting) []string {
	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.Key)
	}
	return keys
}

func TestAdminSettingsRejectsUnsafeCustomsEndpoints(t *testing.T) {
	_, settingService := newTestSettingService(t)
	adminSettings := NewAdminSettingsService(settingService)

	_, err := adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:    "customs_lookup_us_hts_endpoint",
		Value:  "http://127.0.0.1:8080/search",
		Type:   "string",
		Group:  "api",
		Locale: "en",
	})
	require.ErrorIs(t, err, ErrSettingInvalid)

	_, err = adminSettings.UpdateSetting(settingdomain.UpdateSettingRequest{
		Key:    "customs_lookup_us_hts_endpoint",
		Value:  "https://1.1.1.1/reststop/search",
		Type:   "string",
		Group:  "api",
		Locale: "en",
	})
	require.NoError(t, err)
}
