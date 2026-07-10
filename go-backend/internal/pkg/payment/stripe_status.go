package payment

import "github.com/stripe/stripe-go/v76"

// GetStripePaymentStatus 将Stripe状态映射为统一状态
func GetStripePaymentStatus(stripeStatus stripe.PaymentIntentStatus) string {
	switch stripeStatus {
	case stripe.PaymentIntentStatusSucceeded:
		return "succeeded"
	case stripe.PaymentIntentStatusProcessing:
		return "processing"
	case stripe.PaymentIntentStatusRequiresPaymentMethod:
		return "pending"
	case stripe.PaymentIntentStatusRequiresConfirmation:
		return "pending"
	case stripe.PaymentIntentStatusRequiresAction:
		return "pending"
	case stripe.PaymentIntentStatusRequiresCapture:
		return "authorized"
	case stripe.PaymentIntentStatusCanceled:
		return "canceled"
	default:
		return string(stripeStatus)
	}
}
