package payment

import (
	"os"
	"strings"
)

// LoadConfigFromEnv 从环境变量加载配置
func LoadConfigFromEnv(gatewayType GatewayType) *Config {
	prefix := strings.ToUpper(string(gatewayType))
	config := &Config{
		Type:          gatewayType,
		APIKey:        os.Getenv(prefix + "_API_KEY"),
		SecretKey:     os.Getenv(prefix + "_SECRET_KEY"),
		WebhookSecret: os.Getenv(prefix + "_WEBHOOK_SECRET"),
		Environment:   getEnv(prefix+"_ENVIRONMENT", "sandbox"),
	}

	switch gatewayType {
	case GatewayStripe:
		if config.SecretKey == "" {
			config.SecretKey = config.APIKey
		}
	case GatewayPayPal:
		if config.APIKey == "" {
			config.APIKey = os.Getenv("PAYPAL_CLIENT_ID")
		}
		if config.SecretKey == "" {
			config.SecretKey = os.Getenv("PAYPAL_SECRET")
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
