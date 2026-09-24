package payment

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequiredWebhookEventListsKeepConcreteAndOperatorViewsAligned(t *testing.T) {
	stripeEvents := RequiredWebhookEvents(GatewayStripe)
	stripeChecklist := RequiredWebhookEventChecklist(GatewayStripe)

	require.Len(t, stripeEvents, 15)
	require.Len(t, stripeChecklist, 14)
	require.Contains(t, stripeEvents, "review.opened")
	require.Contains(t, stripeEvents, "review.closed")
	require.Contains(t, stripeChecklist, "review.opened / review.closed")
	require.Len(t, RequiredWebhookEvents(GatewayPayPal), 8)
	require.Len(t, RequiredWebhookEventChecklist(GatewayPayPal), 8)

	stripeEvents[0] = "mutated"
	require.Equal(t, "payment_intent.succeeded", RequiredWebhookEvents(GatewayStripe)[0])
}
