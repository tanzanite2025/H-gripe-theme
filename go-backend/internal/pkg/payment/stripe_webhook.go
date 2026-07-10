package payment

import (
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

// VerifyWebhook 验证Stripe Webhook签名
func (g *stripeGatewayImpl) VerifyWebhook(payload []byte, signature string) (bool, error) {
	if g.config.WebhookSecret == "" {
		return false, fmt.Errorf("webhook secret is not configured")
	}

	// 使用Stripe SDK验证webhook签名
	_, err := webhook.ConstructEvent(payload, signature, g.config.WebhookSecret)
	if err != nil {
		return false, fmt.Errorf("webhook signature verification failed: %w", err)
	}

	return true, nil
}

// ParseWebhookEvent 解析Stripe Webhook事件（辅助方法）
func ParseStripeWebhookEvent(payload []byte, signature, webhookSecret string) (stripe.Event, error) {
	event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	if err != nil {
		return stripe.Event{}, fmt.Errorf("failed to parse stripe webhook: %w", err)
	}
	return event, nil
}
