package admin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParsePaymentRiskMonitoringProvidersDefaultsToSupportedProviders(t *testing.T) {
	providers, err := parsePaymentRiskMonitoringProviders("")

	require.NoError(t, err)
	require.Equal(t, []string{"stripe", "paypal"}, providers)
}

func TestParsePaymentRiskMonitoringProvidersRejectsUnsupportedProvider(t *testing.T) {
	_, err := parsePaymentRiskMonitoringProviders("stripe,card_processor")

	require.Error(t, err)
}
