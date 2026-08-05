package payment

import "strings"

// ProviderForPaymentMethod maps storefront payment method codes to their
// backing gateway provider. Unknown methods return an empty provider so global,
// country, and payment-method scoped controls can still be evaluated.
func ProviderForPaymentMethod(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "card", "credit_card", "debit_card", string(GatewayStripe):
		return string(GatewayStripe)
	case string(GatewayPayPal):
		return string(GatewayPayPal)
	case string(GatewayAlipay):
		return string(GatewayAlipay)
	case "wechatpay", "wechat_pay", string(GatewayWechat):
		return string(GatewayWechat)
	default:
		return ""
	}
}
