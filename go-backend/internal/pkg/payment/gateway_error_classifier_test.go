package payment

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsTransientGatewayNetworkOrServerErrorRecognizesProviderIncidents(t *testing.T) {
	require.True(t, IsTransientGatewayNetworkOrServerError(errors.New("provider returned status code: 503")))
	require.True(t, IsTransientGatewayNetworkOrServerError(context.DeadlineExceeded))
	require.True(t, IsTransientGatewayNetworkOrServerError(errors.New("dial tcp: connection refused")))
}

func TestIsTransientGatewayNetworkOrServerErrorIgnoresBusinessPaymentFailures(t *testing.T) {
	require.False(t, IsTransientGatewayNetworkOrServerError(errors.New("card declined")))
	require.False(t, IsTransientGatewayNetworkOrServerError(errors.New("invalid payment request: invalid currency")))
	require.False(t, IsTransientGatewayNetworkOrServerError(errors.New("PayPal order mismatch")))
	require.False(t, IsTransientGatewayNetworkOrServerError(context.Canceled))
}
