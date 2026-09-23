package payment

// RequiredWebhookEvents returns the concrete provider event names that the
// application consumes for payment, refund, fraud, and dispute state
// transitions. These lists are deliberately kept beside runtime readiness so
// the admin console cannot present credentials as production-ready while the
// provider endpoint is subscribed to an incomplete event set.
func RequiredWebhookEvents(provider GatewayType) []string {
	var events []string
	switch provider {
	case GatewayStripe:
		events = []string{
			"payment_intent.succeeded",
			"payment_intent.payment_failed",
			"payment_intent.requires_action",
			"payment_intent.processing",
			"charge.refunded",
			"refund.created",
			"refund.updated",
			"charge.dispute.created",
			"charge.dispute.updated",
			"charge.dispute.funds_withdrawn",
			"charge.dispute.funds_reinstated",
			"charge.dispute.closed",
			"radar.early_fraud_warning.created",
			"review.opened",
			"review.closed",
		}
	case GatewayPayPal:
		events = []string{
			"CHECKOUT.ORDER.APPROVED",
			"PAYMENT.CAPTURE.COMPLETED",
			"PAYMENT.CAPTURE.DENIED",
			"PAYMENT.CAPTURE.REFUNDED",
			"CUSTOMER.DISPUTE.CREATED",
			"CUSTOMER.DISPUTE.RESOLVED",
			"CUSTOMER.DISPUTE.UPDATED",
			"RISK.DISPUTE.CREATED",
		}
	}
	return append([]string(nil), events...)
}

// RequiredWebhookEventChecklist returns the operator-facing checklist. Stripe
// documents review.opened and review.closed as one review lifecycle item, but
// both concrete event names remain in RequiredWebhookEvents so neither can be
// omitted from the provider subscription.
func RequiredWebhookEventChecklist(provider GatewayType) []string {
	if provider != GatewayStripe {
		return RequiredWebhookEvents(provider)
	}
	return []string{
		"payment_intent.succeeded",
		"payment_intent.payment_failed",
		"payment_intent.requires_action",
		"payment_intent.processing",
		"charge.refunded",
		"refund.created",
		"refund.updated",
		"charge.dispute.created",
		"charge.dispute.updated",
		"charge.dispute.funds_withdrawn",
		"charge.dispute.funds_reinstated",
		"charge.dispute.closed",
		"radar.early_fraud_warning.created",
		"review.opened / review.closed",
	}
}

func WebhookEventSetupURL(provider GatewayType) string {
	switch provider {
	case GatewayStripe:
		return "https://dashboard.stripe.com/webhooks"
	case GatewayPayPal:
		return "https://developer.paypal.com/dashboard/webhooks"
	default:
		return ""
	}
}
