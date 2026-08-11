package admin

import (
	"testing"

	settingdomain "commerce-platform/internal/domain/setting"
	pgateway "commerce-platform/internal/pkg/payment"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRefundExecutionGatewayConfigUsesDomainManagedEncryptedConfig(t *testing.T) {
	masterKey := "admin-refund-payment-config-test-master-key"
	t.Setenv(pgateway.PaymentConfigMasterKeyEnv, masterKey)
	t.Setenv("PAYPAL_CLIENT_ID", "env-paypal-client")
	t.Setenv("PAYPAL_SECRET", "env-paypal-secret")
	t.Setenv("PAYPAL_WEBHOOK_ID", "env-paypal-webhook")

	adminSettings := newAdminPaymentSettingsServiceForConfigTest(t)
	encrypted, err := pgateway.EncryptSecureGatewayConfig(pgateway.SecureGatewayConfig{
		Provider:    pgateway.GatewayPayPal,
		Environment: "production",
		Credentials: map[string]string{
			"client_id":  "admin-paypal-client",
			"secret":     "admin-paypal-secret",
			"webhook_id": "admin-paypal-webhook",
		},
	}, masterKey)
	require.NoError(t, err)
	_, err = adminSettings.UpdateDomainManagedSetting(settingdomain.UpdateSettingRequest{
		Key:      pgateway.SecureGatewaySettingKey(pgateway.GatewayPayPal),
		Value:    encrypted,
		Type:     "encrypted_json",
		Group:    "payment_secret",
		Locale:   "global",
		IsPublic: false,
	})
	require.NoError(t, err)

	handler := NewPaymentRefundExecutionHandler(nil, adminSettings)
	config, err := handler.gatewayConfig(pgateway.GatewayPayPal)

	require.NoError(t, err)
	require.Equal(t, "admin-paypal-client", config.APIKey)
	require.Equal(t, "admin-paypal-secret", config.SecretKey)
	require.Equal(t, "admin-paypal-webhook", config.WebhookSecret)
	require.Equal(t, "production", config.Environment)
}

func newAdminPaymentSettingsServiceForConfigTest(t *testing.T) *service.AdminSettingsService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&settingdomain.Setting{}))
	settingService := service.NewSettingService(repository.NewSettingRepository(db), nil, 0)
	return service.NewAdminSettingsService(settingService)
}
