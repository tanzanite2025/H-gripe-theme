package payment

import (
	"fmt"
	"time"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

const stripeWebhookReplayTolerance = 5 * time.Minute

// VerifyWebhook 验证Stripe Webhook签名
func (g *stripeGatewayImpl) VerifyWebhook(payload []byte, signature string) (bool, error) {
	if g.config.WebhookSecret == "" {
		return false, fmt.Errorf("webhook secret is not configured")
	}

	_, err := webhook.ConstructEventWithTolerance(payload, signature, g.config.WebhookSecret, stripeWebhookReplayTolerance)
	if err != nil {
		return false, fmt.Errorf("webhook signature verification failed: %w", err)
	}

	return true, nil
}

// ParseWebhookEvent 解析Stripe Webhook事件（辅助方法）
func ParseStripeWebhookEvent(payload []byte, signature, webhookSecret string) (stripe.Event, error) {
	event, err := webhook.ConstructEventWithTolerance(payload, signature, webhookSecret, stripeWebhookReplayTolerance)
	if err != nil {
		return stripe.Event{}, fmt.Errorf("failed to parse stripe webhook: %w", err)
	}
	return event, nil
}
