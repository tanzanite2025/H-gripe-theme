package payment

import (
	"testing"

	settingdomain "commerce-platform/internal/domain/setting"

	"github.com/stretchr/testify/require"
)

func TestResolveStripePaymentMethodTypesUsesProviderInstallmentsSettings(t *testing.T) {
	adminSettings := newPaymentRuntimeAdminSettingsService(t)
	handler := NewHandler(nil, nil, adminSettings, nil, nil, nil, nil, nil, nil)

	value, err := settingdomain.PaymentProviderInstallmentsSettings{
		Provider:           "stripe",
		Enabled:            true,
		PaymentMethodTypes: []string{"card", "klarna"},
		Countries:          []string{"US"},
		Currencies:         []string{"USD"},
	}.Value()
	require.NoError(t, err)

	_, err = adminSettings.UpdateDomainManagedSetting(settingdomain.UpdateSettingRequest{
		Key:         settingdomain.PaymentInstallmentsStripeKey,
		Value:       value,
		Type:        "json",
		Group:       settingdomain.PaymentInstallmentsGroup,
		Locale:      "global",
		IsPublic:    false,
		Description: "stripe installments",
	})
	require.NoError(t, err)

	methodTypes, err := handler.resolveStripePaymentMethodTypes("US", "USD", 250, []string{"card"})
	require.NoError(t, err)
	require.Equal(t, []string{"card", "klarna"}, methodTypes)
}

func TestResolveStripePaymentMethodTypesFallsBackToCard(t *testing.T) {
	handler := NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil)

	methodTypes, err := handler.resolveStripePaymentMethodTypes("US", "USD", 250, nil)
	require.NoError(t, err)
	require.Equal(t, []string{"card"}, methodTypes)
}

func TestResolveStripePaymentMethodTypesKeepsCardWhenSettingsOmitIt(t *testing.T) {
	adminSettings := newPaymentRuntimeAdminSettingsService(t)
	handler := NewHandler(nil, nil, adminSettings, nil, nil, nil, nil, nil, nil)

	value, err := settingdomain.PaymentProviderInstallmentsSettings{
		Provider:           "stripe",
		Enabled:            true,
		PaymentMethodTypes: []string{"affirm"},
		Countries:          []string{"US"},
		Currencies:         []string{"USD"},
	}.Value()
	require.NoError(t, err)

	_, err = adminSettings.UpdateDomainManagedSetting(settingdomain.UpdateSettingRequest{
		Key:         settingdomain.PaymentInstallmentsStripeKey,
		Value:       value,
		Type:        "json",
		Group:       settingdomain.PaymentInstallmentsGroup,
		Locale:      "global",
		IsPublic:    false,
		Description: "stripe installments",
	})
	require.NoError(t, err)

	methodTypes, err := handler.resolveStripePaymentMethodTypes("US", "USD", 250, []string{"card"})
	require.NoError(t, err)
	require.Equal(t, []string{"card", "affirm"}, methodTypes)
}

func TestResolveStripePaymentMethodTypesFiltersUnsupportedCurrencyMethods(t *testing.T) {
	adminSettings := newPaymentRuntimeAdminSettingsService(t)
	handler := NewHandler(nil, nil, adminSettings, nil, nil, nil, nil, nil, nil)

	value, err := settingdomain.PaymentProviderInstallmentsSettings{
		Provider:           "stripe",
		Enabled:            true,
		PaymentMethodTypes: []string{"card", "affirm", "klarna", "afterpay_clearpay"},
		Currencies:         []string{"EUR"},
	}.Value()
	require.NoError(t, err)

	_, err = adminSettings.UpdateDomainManagedSetting(settingdomain.UpdateSettingRequest{
		Key:         settingdomain.PaymentInstallmentsStripeKey,
		Value:       value,
		Type:        "json",
		Group:       settingdomain.PaymentInstallmentsGroup,
		Locale:      "global",
		IsPublic:    false,
		Description: "stripe installments",
	})
	require.NoError(t, err)

	methodTypes, err := handler.resolveStripePaymentMethodTypes("DE", "EUR", 250, []string{"card"})
	require.NoError(t, err)
	require.Equal(t, []string{"card", "klarna"}, methodTypes)
}

func TestResolveStripePaymentMethodTypesFallsBackOutsideAmountThresholds(t *testing.T) {
	adminSettings := newPaymentRuntimeAdminSettingsService(t)
	handler := NewHandler(nil, nil, adminSettings, nil, nil, nil, nil, nil, nil)

	value, err := settingdomain.PaymentProviderInstallmentsSettings{
		Provider:           "stripe",
		Enabled:            true,
		PaymentMethodTypes: []string{"klarna"},
		Currencies:         []string{"USD"},
		MinAmount:          100,
		MaxAmount:          5000,
	}.Value()
	require.NoError(t, err)

	_, err = adminSettings.UpdateDomainManagedSetting(settingdomain.UpdateSettingRequest{
		Key:         settingdomain.PaymentInstallmentsStripeKey,
		Value:       value,
		Type:        "json",
		Group:       settingdomain.PaymentInstallmentsGroup,
		Locale:      "global",
		IsPublic:    false,
		Description: "stripe installments",
	})
	require.NoError(t, err)

	methodTypes, err := handler.resolveStripePaymentMethodTypes("US", "USD", 5, []string{"card"})
	require.NoError(t, err)
	require.Equal(t, []string{"card"}, methodTypes)

	methodTypes, err = handler.resolveStripePaymentMethodTypes("US", "USD", 250, []string{"card"})
	require.NoError(t, err)
	require.Equal(t, []string{"card", "klarna"}, methodTypes)

	methodTypes, err = handler.resolveStripePaymentMethodTypes("US", "USD", 6000, []string{"card"})
	require.NoError(t, err)
	require.Equal(t, []string{"card"}, methodTypes)
}
