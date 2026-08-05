package payment

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	PaymentConfigMasterKeyEnv = "PAYMENT_CONFIG_MASTER_KEY"
	secureConfigPrefix        = "v1:"
)

type SecureGatewayConfig struct {
	Provider    GatewayType       `json:"provider"`
	Environment string            `json:"environment"`
	Credentials map[string]string `json:"credentials"`
}

type SecureGatewayConfigStatus struct {
	Provider         GatewayType `json:"provider"`
	Environment      string      `json:"environment"`
	Configured       bool        `json:"configured"`
	Readable         bool        `json:"readable"`
	RuntimeSource    string      `json:"runtime_source"`
	ConfiguredFields []string    `json:"configured_fields"`
	Error            string      `json:"error,omitempty"`
}

func SecureGatewaySettingKey(provider GatewayType) string {
	return "payment_gateway_" + string(provider)
}

func PaymentConfigMasterKey() string {
	return strings.TrimSpace(os.Getenv(PaymentConfigMasterKeyEnv))
}

func PaymentConfigMasterKeyConfigured() bool {
	return PaymentConfigMasterKey() != ""
}

func EncryptSecureGatewayConfig(config SecureGatewayConfig, masterKey string) (string, error) {
	if strings.TrimSpace(masterKey) == "" {
		return "", fmt.Errorf("%s is required", PaymentConfigMasterKeyEnv)
	}
	if config.Provider == "" {
		return "", fmt.Errorf("provider is required")
	}
	if config.Environment == "" {
		config.Environment = "sandbox"
	}
	if config.Credentials == nil {
		config.Credentials = map[string]string{}
	}

	plaintext, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(derivePaymentConfigKey(masterKey))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, plaintext, []byte(config.Provider))
	payload := make([]byte, 0, len(nonce)+len(ciphertext))
	payload = append(payload, nonce...)
	payload = append(payload, ciphertext...)
	return secureConfigPrefix + base64.StdEncoding.EncodeToString(payload), nil
}

func DecryptSecureGatewayConfig(value, masterKey string) (SecureGatewayConfig, error) {
	if strings.TrimSpace(masterKey) == "" {
		return SecureGatewayConfig{}, fmt.Errorf("%s is required", PaymentConfigMasterKeyEnv)
	}
	value = strings.TrimSpace(value)
	if !strings.HasPrefix(value, secureConfigPrefix) {
		return SecureGatewayConfig{}, fmt.Errorf("unsupported payment config format")
	}

	payload, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(value, secureConfigPrefix))
	if err != nil {
		return SecureGatewayConfig{}, err
	}

	block, err := aes.NewCipher(derivePaymentConfigKey(masterKey))
	if err != nil {
		return SecureGatewayConfig{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return SecureGatewayConfig{}, err
	}
	if len(payload) <= gcm.NonceSize() {
		return SecureGatewayConfig{}, fmt.Errorf("invalid payment config payload")
	}

	nonce := payload[:gcm.NonceSize()]
	ciphertext := payload[gcm.NonceSize():]

	// Provider is validated after decrypting by trying every known provider as AAD.
	for _, provider := range []GatewayType{GatewayStripe, GatewayPayPal, GatewayAlipay, GatewayWechat} {
		plaintext, err := gcm.Open(nil, nonce, ciphertext, []byte(provider))
		if err != nil {
			continue
		}

		var config SecureGatewayConfig
		if err := json.Unmarshal(plaintext, &config); err != nil {
			return SecureGatewayConfig{}, err
		}
		if config.Provider != provider {
			return SecureGatewayConfig{}, fmt.Errorf("payment config provider mismatch")
		}
		if config.Credentials == nil {
			config.Credentials = map[string]string{}
		}
		return config, nil
	}

	return SecureGatewayConfig{}, fmt.Errorf("payment config decrypt failed")
}

func DecodeStoredSecureGatewayConfig(value string, provider GatewayType) (SecureGatewayConfig, error) {
	if !PaymentConfigMasterKeyConfigured() {
		return SecureGatewayConfig{}, fmt.Errorf("%s is required to read encrypted payment config", PaymentConfigMasterKeyEnv)
	}
	config, err := DecryptSecureGatewayConfig(value, PaymentConfigMasterKey())
	if err != nil {
		return SecureGatewayConfig{}, err
	}
	if config.Provider != provider {
		return SecureGatewayConfig{}, fmt.Errorf("payment config provider mismatch")
	}
	return config, nil
}

func GatewayConfigFromSecureConfig(config SecureGatewayConfig) *Config {
	credentials := config.Credentials
	gatewayConfig := &Config{
		Type:        config.Provider,
		Environment: config.Environment,
	}
	if gatewayConfig.Environment == "" {
		gatewayConfig.Environment = "sandbox"
	}

	switch config.Provider {
	case GatewayStripe:
		gatewayConfig.APIKey = credentials["api_key"]
		gatewayConfig.SecretKey = credentials["api_key"]
		gatewayConfig.PublishableKey = credentials["publishable_key"]
		gatewayConfig.WebhookSecret = credentials["webhook_secret"]
		gatewayConfig.ThreeDSecure = NormalizeThreeDSecureMode(credentials["three_ds_mode"])
	case GatewayPayPal:
		gatewayConfig.APIKey = credentials["client_id"]
		gatewayConfig.SecretKey = credentials["secret"]
		gatewayConfig.WebhookSecret = credentials["webhook_id"]
	case GatewayAlipay:
		gatewayConfig.APIKey = credentials["app_id"]
		gatewayConfig.SecretKey = credentials["private_key"]
		gatewayConfig.WebhookSecret = credentials["public_key"]
	case GatewayWechat:
		gatewayConfig.APIKey = credentials["mch_id"]
		gatewayConfig.WechatAppID = credentials["app_id"]
		gatewayConfig.SecretKey = credentials["private_key_path"]
		gatewayConfig.WebhookSecret = credentials["merchant_serial"]
		gatewayConfig.WechatAPIv3Key = credentials["api_v3_key"]
		gatewayConfig.WechatPayPlatformCertificate = credentials["platform_certificate"]
		gatewayConfig.WechatPayPlatformPublicKey = credentials["platform_public_key"]
		gatewayConfig.WechatPayPlatformPublicKeyID = credentials["platform_public_key_id"]
	}

	return gatewayConfig
}

func SecureGatewayConfiguredFields(config SecureGatewayConfig) []string {
	fields := SecureGatewayCredentialFields(config.Provider)
	configured := make([]string, 0, len(fields))
	for _, field := range fields {
		if strings.TrimSpace(config.Credentials[field]) != "" {
			configured = append(configured, field)
		}
	}
	return configured
}

func SecureGatewayCredentialFields(provider GatewayType) []string {
	switch provider {
	case GatewayStripe:
		return []string{"api_key", "publishable_key", "webhook_secret", "three_ds_mode"}
	case GatewayPayPal:
		return []string{"client_id", "secret", "webhook_id"}
	case GatewayAlipay:
		return []string{"app_id", "private_key", "public_key"}
	case GatewayWechat:
		return []string{"mch_id", "app_id", "private_key_path", "merchant_serial", "api_v3_key", "platform_certificate", "platform_public_key", "platform_public_key_id"}
	default:
		return nil
	}
}

func NormalizeThreeDSecureMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "any":
		return "any"
	case "challenge":
		return "challenge"
	default:
		return "automatic"
	}
}

func NormalizeGatewayEnvironment(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "production", "live":
		return "production"
	default:
		return "sandbox"
	}
}

func ParseGatewayType(value string) (GatewayType, error) {
	switch GatewayType(strings.ToLower(strings.TrimSpace(value))) {
	case GatewayStripe:
		return GatewayStripe, nil
	case GatewayPayPal:
		return GatewayPayPal, nil
	case GatewayAlipay:
		return GatewayAlipay, nil
	case GatewayWechat:
		return GatewayWechat, nil
	default:
		return "", fmt.Errorf("unsupported gateway type: %s", value)
	}
}

func derivePaymentConfigKey(masterKey string) []byte {
	hash := sha256.Sum256([]byte(masterKey))
	return hash[:]
}
