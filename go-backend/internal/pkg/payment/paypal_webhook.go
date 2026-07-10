package payment

import "fmt"

// VerifyWebhook 验证PayPal Webhook签名
func (g *paypalGatewayImpl) VerifyWebhook(payload []byte, signature string) (bool, error) {
	if g.config.WebhookSecret == "" {
		return false, fmt.Errorf("webhook secret is not configured")
	}

	// PayPal webhook验证使用HMAC SHA256
	// 注意：在生产环境中，建议使用PayPal的官方验证API
	// POST /v1/notifications/verify-webhook-signature

	// 基本的HMAC验证
	isValid := verifyHMACSHA256(payload, signature, g.config.WebhookSecret)
	if !isValid {
		return false, fmt.Errorf("webhook signature verification failed")
	}

	return true, nil
}
