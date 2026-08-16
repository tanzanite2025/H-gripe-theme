package payment

import (
	"net/url"
	"os"
	"strings"
)

type RuntimeReadiness struct {
	RuntimeSource         string                 `json:"runtime_source"`
	SecretStoreConfigured bool                   `json:"secret_store_configured"`
	SecretStoreEnv        string                 `json:"secret_store_env"`
	Gateways              []GatewayRuntimeStatus `json:"gateways"`
}

type GatewayRuntimeStatus struct {
	Provider              GatewayType `json:"provider"`
	Label                 string      `json:"label"`
	Environment           string      `json:"environment"`
	ThreeDSMode           string      `json:"three_ds_mode,omitempty"`
	CallbackURL           string      `json:"callback_url"`
	Configured            bool        `json:"configured"`
	WebhookConfigured     bool        `json:"webhook_configured"`
	WebhookSupported      bool        `json:"webhook_supported"`
	ProductionReady       bool        `json:"production_ready"`
	RuntimeSource         string      `json:"runtime_source"`
	AdminConfigConfigured bool        `json:"admin_config_configured"`
	AdminConfigReadable   bool        `json:"admin_config_readable"`
	SecretStoreConfigured bool        `json:"secret_store_configured"`
	Missing               []string    `json:"missing"`
	Blockers              []string    `json:"blockers"`
	Warnings              []string    `json:"warnings"`
	ConfiguredFields      []string    `json:"configured_fields"`
	RequiredFields        []string    `json:"required_fields"`
	DocumentationLabel    string      `json:"documentation_label"`
	DocumentationURL      string      `json:"documentation_url"`
}

func BuildRuntimeReadiness(baseURL string) RuntimeReadiness {
	gateways := []GatewayType{GatewayStripe, GatewayPayPal, GatewayAlipay, GatewayWechat}
	statuses := make([]GatewayRuntimeStatus, 0, len(gateways))
	for _, gatewayType := range gateways {
		statuses = append(statuses, buildGatewayRuntimeStatus(gatewayType, baseURL))
	}

	return RuntimeReadiness{
		RuntimeSource:         "environment",
		SecretStoreConfigured: PaymentConfigMasterKeyConfigured(),
		SecretStoreEnv:        PaymentConfigMasterKeyEnv,
		Gateways:              statuses,
	}
}

func ApplySecureGatewayStatuses(readiness *RuntimeReadiness, secureStatuses []SecureGatewayConfigStatus) {
	if readiness == nil {
		return
	}
	hasAdminRuntime := false
	for i := range readiness.Gateways {
		for _, secureStatus := range secureStatuses {
			if readiness.Gateways[i].Provider == secureStatus.Provider {
				applySecureGatewayStatus(&readiness.Gateways[i], secureStatus)
				if readiness.Gateways[i].RuntimeSource == "admin-encrypted" {
					hasAdminRuntime = true
				}
				break
			}
		}
	}
	if hasAdminRuntime {
		readiness.RuntimeSource = "mixed"
	}
}

func buildGatewayRuntimeStatus(gatewayType GatewayType, baseURL string) GatewayRuntimeStatus {
	config := LoadConfigFromEnv(gatewayType)
	status := GatewayRuntimeStatus{
		Provider:              gatewayType,
		Label:                 gatewayLabel(gatewayType),
		Environment:           config.Environment,
		CallbackURL:           paymentWebhookURL(baseURL, gatewayType),
		RuntimeSource:         "environment",
		WebhookSupported:      true,
		SecretStoreConfigured: PaymentConfigMasterKeyConfigured(),
	}

	switch gatewayType {
	case GatewayStripe:
		status.RequiredFields = []string{"STRIPE_API_KEY or STRIPE_SECRET_KEY", "STRIPE_WEBHOOK_SECRET"}
		status.ConfiguredFields = configuredEnvFields([]string{"STRIPE_API_KEY", "STRIPE_SECRET_KEY", "STRIPE_PUBLISHABLE_KEY", "STRIPE_WEBHOOK_SECRET"})
		if !envAnySet("STRIPE_API_KEY", "STRIPE_SECRET_KEY") {
			status.Missing = append(status.Missing, "STRIPE_API_KEY or STRIPE_SECRET_KEY")
		}
		if !envSet("STRIPE_WEBHOOK_SECRET") {
			status.Missing = append(status.Missing, "STRIPE_WEBHOOK_SECRET")
		}
		status.Configured = envAnySet("STRIPE_API_KEY", "STRIPE_SECRET_KEY")
		status.WebhookConfigured = envSet("STRIPE_WEBHOOK_SECRET")
		status.ProductionReady = status.Configured && status.WebhookConfigured
		status.ThreeDSMode = NormalizeThreeDSecureMode(config.ThreeDSecure)
		status.DocumentationLabel = "Stripe webhook signatures"
		status.DocumentationURL = "https://docs.stripe.com/webhooks/signature"
	case GatewayPayPal:
		status.RequiredFields = []string{"PAYPAL_CLIENT_ID or PAYPAL_API_KEY", "PAYPAL_SECRET or PAYPAL_SECRET_KEY", "PAYPAL_WEBHOOK_ID"}
		status.ConfiguredFields = configuredEnvFields([]string{"PAYPAL_CLIENT_ID", "PAYPAL_API_KEY", "PAYPAL_SECRET", "PAYPAL_SECRET_KEY", "PAYPAL_WEBHOOK_ID"})
		if !envAnySet("PAYPAL_CLIENT_ID", "PAYPAL_API_KEY") {
			status.Missing = append(status.Missing, "PAYPAL_CLIENT_ID or PAYPAL_API_KEY")
		}
		if !envAnySet("PAYPAL_SECRET", "PAYPAL_SECRET_KEY") {
			status.Missing = append(status.Missing, "PAYPAL_SECRET or PAYPAL_SECRET_KEY")
		}
		if !envSet("PAYPAL_WEBHOOK_ID") {
			status.Missing = append(status.Missing, "PAYPAL_WEBHOOK_ID")
		}
		status.Configured = envAnySet("PAYPAL_CLIENT_ID", "PAYPAL_API_KEY") && envAnySet("PAYPAL_SECRET", "PAYPAL_SECRET_KEY")
		status.WebhookConfigured = envSet("PAYPAL_WEBHOOK_ID")
		status.ProductionReady = status.Configured && status.WebhookConfigured
		status.DocumentationLabel = "PayPal verify webhook signature"
		status.DocumentationURL = "https://developer.paypal.com/docs/api/webhooks/v1/#verify-webhook-signature_post"
	case GatewayAlipay:
		status.RequiredFields = []string{"ALIPAY_APP_ID", "ALIPAY_PRIVATE_KEY", "ALIPAY_PUBLIC_KEY"}
		status.ConfiguredFields = configuredEnvFields(status.RequiredFields)
		if !envSet("ALIPAY_APP_ID") {
			status.Missing = append(status.Missing, "ALIPAY_APP_ID")
		}
		if !envSet("ALIPAY_PRIVATE_KEY") {
			status.Missing = append(status.Missing, "ALIPAY_PRIVATE_KEY")
		}
		if !envSet("ALIPAY_PUBLIC_KEY") {
			status.Missing = append(status.Missing, "ALIPAY_PUBLIC_KEY")
		}
		status.Configured = envSet("ALIPAY_APP_ID") && envSet("ALIPAY_PRIVATE_KEY")
		status.WebhookConfigured = envSet("ALIPAY_PUBLIC_KEY")
		status.ProductionReady = status.Configured && status.WebhookConfigured
		status.DocumentationLabel = "Alipay+ notifyPayment"
		status.DocumentationURL = "https://docs.alipayplus.com/alipayplus/alipayplus/api_acq/notify_payment"
	case GatewayWechat:
		status.RequiredFields = []string{
			"WECHAT_MCH_ID",
			"WECHAT_APP_ID",
			"WECHAT_PRIVATE_KEY_PATH",
			"WECHAT_MERCHANT_SERIAL",
			"WECHAT_API_V3_KEY",
			"WECHAT_PAY_PLATFORM_CERTIFICATE or WECHAT_PAY_PLATFORM_PUBLIC_KEY + WECHAT_PAY_PLATFORM_PUBLIC_KEY_ID",
		}
		status.ConfiguredFields = configuredEnvFields([]string{
			"WECHAT_MCH_ID",
			"WECHAT_APP_ID",
			"WECHAT_PRIVATE_KEY_PATH",
			"WECHAT_MERCHANT_SERIAL",
			"WECHAT_API_V3_KEY",
			"WECHAT_PAY_PLATFORM_CERTIFICATE",
			"WECHAT_PAY_PLATFORM_PUBLIC_KEY",
			"WECHAT_PAY_PLATFORM_PUBLIC_KEY_ID",
		})
		if !envSet("WECHAT_MCH_ID") {
			status.Missing = append(status.Missing, "WECHAT_MCH_ID")
		}
		if !envSet("WECHAT_APP_ID") {
			status.Missing = append(status.Missing, "WECHAT_APP_ID")
		}
		if !envSet("WECHAT_PRIVATE_KEY_PATH") {
			status.Missing = append(status.Missing, "WECHAT_PRIVATE_KEY_PATH")
		}
		if !envSet("WECHAT_MERCHANT_SERIAL") {
			status.Missing = append(status.Missing, "WECHAT_MERCHANT_SERIAL")
		}
		if !envSet("WECHAT_API_V3_KEY") {
			status.Missing = append(status.Missing, "WECHAT_API_V3_KEY")
		}
		wechatPlatformVerifierConfigured := envSet("WECHAT_PAY_PLATFORM_CERTIFICATE") || (envSet("WECHAT_PAY_PLATFORM_PUBLIC_KEY") && envSet("WECHAT_PAY_PLATFORM_PUBLIC_KEY_ID"))
		if !wechatPlatformVerifierConfigured {
			status.Missing = append(status.Missing, "WECHAT_PAY_PLATFORM_CERTIFICATE or WECHAT_PAY_PLATFORM_PUBLIC_KEY + WECHAT_PAY_PLATFORM_PUBLIC_KEY_ID")
		}
		status.Configured = envSet("WECHAT_MCH_ID") && envSet("WECHAT_APP_ID") && envSet("WECHAT_PRIVATE_KEY_PATH") && envSet("WECHAT_MERCHANT_SERIAL")
		status.WebhookConfigured = envSet("WECHAT_API_V3_KEY") && wechatPlatformVerifierConfigured
		status.ProductionReady = status.Configured && status.WebhookConfigured
		status.DocumentationLabel = "WeChat Pay APIv3 signatures"
		status.DocumentationURL = "https://pay.wechatpay.cn/doc/v3/merchant/4012365342"
	}

	if len(status.Missing) > 0 {
		status.Warnings = append(status.Warnings, "Missing runtime environment fields: "+strings.Join(status.Missing, ", "))
	}

	return status
}

func paymentWebhookURL(baseURL string, gatewayType GatewayType) string {
	baseURL = NormalizePublicBaseURL(baseURL)
	if baseURL == "" {
		return "/api/v1/payment/webhook/" + string(gatewayType)
	}
	return baseURL + "/api/v1/payment/webhook/" + string(gatewayType)
}

func GatewayWebhookURL(baseURL string, gatewayType GatewayType) string {
	return paymentWebhookURL(baseURL, gatewayType)
}

func NormalizePublicBaseURL(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 2048 {
		return ""
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.User != nil {
		return ""
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}
	if strings.TrimSpace(parsed.Host) == "" {
		return ""
	}
	return strings.TrimRight(parsed.String(), "/")
}

func applySecureGatewayStatus(status *GatewayRuntimeStatus, secureStatus SecureGatewayConfigStatus) {
	status.AdminConfigConfigured = secureStatus.Configured
	status.AdminConfigReadable = secureStatus.Readable
	status.SecretStoreConfigured = PaymentConfigMasterKeyConfigured()
	if !secureStatus.Configured {
		return
	}
	if !secureStatus.Readable {
		if secureStatus.Error != "" {
			status.Blockers = append(status.Blockers, secureStatus.Error)
		}
		status.Warnings = append(status.Warnings, "Encrypted admin payment config exists but cannot be read.")
		return
	}

	status.RuntimeSource = "admin-encrypted"
	if secureStatus.Environment != "" {
		status.Environment = secureStatus.Environment
	}
	if status.Provider == GatewayStripe {
		status.ThreeDSMode = NormalizeThreeDSecureMode(secureStatus.ThreeDSMode)
	}
	status.RequiredFields = SecureGatewayRequiredCredentialFields(status.Provider)
	status.ConfiguredFields = secureStatus.ConfiguredFields
	status.Missing = missingSecureCredentialFields(status.Provider, secureStatus.ConfiguredFields)
	status.Warnings = nil
	if len(status.Missing) > 0 {
		status.Warnings = append(status.Warnings, "Missing encrypted admin fields: "+strings.Join(status.Missing, ", "))
	}

	switch status.Provider {
	case GatewayStripe:
		status.Configured = containsString(secureStatus.ConfiguredFields, "api_key")
		status.WebhookConfigured = containsString(secureStatus.ConfiguredFields, "webhook_secret")
		status.ProductionReady = status.Configured && status.WebhookConfigured
		status.Blockers = nil
	case GatewayPayPal:
		status.Configured = containsString(secureStatus.ConfiguredFields, "client_id") && containsString(secureStatus.ConfiguredFields, "secret")
		status.WebhookConfigured = containsString(secureStatus.ConfiguredFields, "webhook_id")
		status.ProductionReady = status.Configured && status.WebhookConfigured
	case GatewayAlipay:
		status.Configured = containsString(secureStatus.ConfiguredFields, "app_id") && containsString(secureStatus.ConfiguredFields, "private_key")
		status.WebhookConfigured = containsString(secureStatus.ConfiguredFields, "public_key")
		status.ProductionReady = status.Configured && status.WebhookConfigured
	case GatewayWechat:
		wechatPlatformVerifierConfigured := containsString(secureStatus.ConfiguredFields, "platform_certificate") || (containsString(secureStatus.ConfiguredFields, "platform_public_key") && containsString(secureStatus.ConfiguredFields, "platform_public_key_id"))
		status.Configured = containsString(secureStatus.ConfiguredFields, "mch_id") && containsString(secureStatus.ConfiguredFields, "app_id") && containsString(secureStatus.ConfiguredFields, "private_key_path") && containsString(secureStatus.ConfiguredFields, "merchant_serial")
		status.WebhookConfigured = containsString(secureStatus.ConfiguredFields, "api_v3_key") && wechatPlatformVerifierConfigured
		status.ProductionReady = status.Configured && status.WebhookConfigured
	}
}

func missingSecureCredentialFields(provider GatewayType, configured []string) []string {
	if provider == GatewayWechat {
		return missingWechatSecureCredentialFields(configured)
	}
	missing := []string{}
	for _, field := range SecureGatewayRequiredCredentialFields(provider) {
		if !containsString(configured, field) {
			missing = append(missing, field)
		}
	}
	return missing
}

func missingWechatSecureCredentialFields(configured []string) []string {
	missing := []string{}
	for _, field := range []string{"mch_id", "app_id", "private_key_path", "merchant_serial", "api_v3_key"} {
		if !containsString(configured, field) {
			missing = append(missing, field)
		}
	}
	if containsString(configured, "platform_certificate") {
		return missing
	}
	if containsString(configured, "platform_public_key") && containsString(configured, "platform_public_key_id") {
		return missing
	}
	return append(missing, "platform_certificate or platform_public_key + platform_public_key_id")
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func gatewayLabel(gatewayType GatewayType) string {
	switch gatewayType {
	case GatewayStripe:
		return "Stripe"
	case GatewayPayPal:
		return "PayPal"
	case GatewayAlipay:
		return "Alipay"
	case GatewayWechat:
		return "WeChat Pay"
	default:
		return string(gatewayType)
	}
}

func configuredEnvFields(keys []string) []string {
	fields := make([]string, 0, len(keys))
	for _, key := range keys {
		if envSet(key) {
			fields = append(fields, key)
		}
	}
	return fields
}

func envSet(key string) bool {
	return strings.TrimSpace(os.Getenv(key)) != ""
}

func envAnySet(keys ...string) bool {
	for _, key := range keys {
		if envSet(key) {
			return true
		}
	}
	return false
}
