package main

import (
	"testing"

	"github.com/stretchr/testify/require"

	"commerce-platform/internal/pkg/config"
)

func TestRecoveryKeysIncludeEffectiveOrderIdentityAndEveryConfiguredEncryptionKey(t *testing.T) {
	for _, name := range []string{"PAYMENT_CONFIG_MASTER_KEY", "OPS_CONNECTOR_MASTER_KEY", "EMAIL_PROVIDER_MASTER_KEY", "SECURITY_MASTER_KEY", "FPX_API_MASTER_KEY", "YANWEN_API_MASTER_KEY"} {
		t.Setenv(name, "synthetic-original-"+name)
	}
	cfg := &config.Config{}
	cfg.JWT.Secret = "synthetic-jwt-key"
	keys := collectRecoverySigningAndEncryptionKeysFromApplicationConfiguration(cfg)
	require.Equal(t, cfg.JWT.Secret, keys["ORDER_NUMBER_SECRET"])
	cfg.OrderNumber.Secret = "synthetic-original-order-key"
	cfg.OrderNumber.PreviousSecret = "synthetic-previous-order-key"
	cfg.GoogleMerchant.TokenEncryptionKey = "synthetic-merchant-encryption-key"
	cfg.SocialOAuth.TokenEncryptionKey = "synthetic-social-encryption-key"
	keys = collectRecoverySigningAndEncryptionKeysFromApplicationConfiguration(cfg)
	require.Equal(t, cfg.OrderNumber.Secret, keys["ORDER_NUMBER_SECRET"])
	require.Equal(t, cfg.OrderNumber.PreviousSecret, keys["ORDER_NUMBER_PREVIOUS_SECRET"])
	require.Equal(t, cfg.GoogleMerchant.TokenEncryptionKey, keys["GOOGLE_MERCHANT_TOKEN_ENCRYPTION_KEY"])
	require.Equal(t, cfg.SocialOAuth.TokenEncryptionKey, keys["SOCIAL_OAUTH_TOKEN_ENCRYPTION_KEY"])
	for _, name := range []string{"PAYMENT_CONFIG_MASTER_KEY", "OPS_CONNECTOR_MASTER_KEY", "EMAIL_PROVIDER_MASTER_KEY", "SECURITY_MASTER_KEY", "FPX_API_MASTER_KEY", "YANWEN_API_MASTER_KEY"} {
		require.Equal(t, "synthetic-original-"+name, keys[name])
	}
}
