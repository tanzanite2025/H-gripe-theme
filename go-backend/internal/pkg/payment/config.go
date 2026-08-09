package payment

import (
	"os"
	"strings"
)

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv(gatewayType GatewayType) *Config {
	prefix := strings.ToUpper(string(gatewayType))
	config := &Config{
		Type:           gatewayType,
		APIKey:         os.Getenv(prefix + "_API_KEY"),
		SecretKey:      os.Getenv(prefix + "_SECRET_KEY"),
		PublishableKey: os.Getenv(prefix + "_PUBLISHABLE_KEY"),
		WebhookSecret:  os.Getenv(prefix + "_WEBHOOK_SECRET"),
		ThreeDSecure:   getEnv(prefix+"_3DS_MODE", "automatic"),
		Environment:    getEnv(prefix+"_ENVIRONMENT", "sandbox"),
	}

	switch gatewayType {
	case GatewayStripe:
		if config.APIKey == "" {
			config.APIKey = config.SecretKey
		}
		if config.SecretKey == "" {
			config.SecretKey = config.APIKey
		}
		if config.PublishableKey == "" {
			config.PublishableKey = os.Getenv("STRIPE_PUBLIC_KEY")
		}
	case GatewayPayPal:
		if config.APIKey == "" {
			config.APIKey = os.Getenv("PAYPAL_CLIENT_ID")
		}
		if config.SecretKey == "" {
			config.SecretKey = os.Getenv("PAYPAL_SECRET")
		}
		if config.WebhookSecret == "" {
			config.WebhookSecret = os.Getenv("PAYPAL_WEBHOOK_ID")
		}
		if config.Environment == "" || config.Environment == "sandbox" {
			config.Environment = getEnv("PAYPAL_MODE", config.Environment)
		}
		if strings.EqualFold(config.Environment, "live") {
			config.Environment = "production"
		}
	case GatewayAlipay:
		if config.APIKey == "" {
			config.APIKey = os.Getenv("ALIPAY_APP_ID")
		}
		if config.SecretKey == "" {
			config.SecretKey = os.Getenv("ALIPAY_PRIVATE_KEY")
		}
		if config.WebhookSecret == "" {
			config.WebhookSecret = os.Getenv("ALIPAY_PUBLIC_KEY")
		}
	case GatewayWechat:
		if config.APIKey == "" {
			config.APIKey = os.Getenv("WECHAT_MCH_ID")
		}
		config.WechatAppID = os.Getenv("WECHAT_APP_ID")
		if config.SecretKey == "" {
			config.SecretKey = os.Getenv("WECHAT_PRIVATE_KEY_PATH")
		}
		if config.WebhookSecret == "" {
			config.WebhookSecret = os.Getenv("WECHAT_MERCHANT_SERIAL")
		}
		config.WechatAPIv3Key = os.Getenv("WECHAT_API_V3_KEY")
		config.WechatPayPlatformCertificate = os.Getenv("WECHAT_PAY_PLATFORM_CERTIFICATE")
		config.WechatPayPlatformPublicKey = os.Getenv("WECHAT_PAY_PLATFORM_PUBLIC_KEY")
		config.WechatPayPlatformPublicKeyID = os.Getenv("WECHAT_PAY_PLATFORM_PUBLIC_KEY_ID")
	}

	return config
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
