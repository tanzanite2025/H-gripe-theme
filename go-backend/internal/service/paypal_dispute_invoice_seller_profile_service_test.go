package service

import (
	"testing"

	settingdomain "commerce-platform/internal/domain/setting"

	"github.com/stretchr/testify/require"
)

func TestPayPalDisputeInvoiceSellerProfileUsesEnvironmentFallback(t *testing.T) {
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_NAME", "Env Seller")
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_ADDRESS", "Env Address")
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_EMAIL", "env@example.test")
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_PHONE", "+1 555 0100")
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_WEBSITE", "https://env.example.test")
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_TAX_ID", "ENV-TAX-ID")

	_, settingService := newTestSettingService(t)
	service := NewPayPalDisputeInvoiceSellerProfileService(settingService)

	profile, err := service.Get()
	require.NoError(t, err)
	require.Equal(t, "Env Seller", profile.Name)
	require.Equal(t, "Env Address", profile.Address)
	require.Equal(t, "env@example.test", profile.Email)
	require.Equal(t, "https://env.example.test", profile.Website)
	require.Equal(t, "ENV-TAX-ID", profile.TaxID)
}

func TestPayPalDisputeInvoiceSellerProfileUpdateOverridesEnvironment(t *testing.T) {
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_NAME", "Env Seller")
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_ADDRESS", "Env Address")
	t.Setenv("PAYPAL_DISPUTE_INVOICE_SELLER_EMAIL", "env@example.test")

	_, settingService := newTestSettingService(t)
	profileService := NewPayPalDisputeInvoiceSellerProfileService(settingService)
	profile, err := profileService.Update(settingdomain.PayPalDisputeInvoiceSellerProfileUpdateRequest{
		Name:    "Database Seller",
		Address: "Database Address",
		Email:   "database@example.test",
		Website: "https://database.example.test",
		TaxID:   "DB-TAX-ID",
	})
	require.NoError(t, err)
	require.Equal(t, "Database Seller", profile.Name)
	require.Equal(t, "Database Address", profile.Address)
	require.Equal(t, "database@example.test", profile.Email)
	require.Equal(t, "https://database.example.test", profile.Website)
	require.Equal(t, "DB-TAX-ID", profile.TaxID)

	record, err := settingService.Get(settingdomain.PayPalDisputeInvoiceSellerProfileKeyWebsite, "global")
	require.NoError(t, err)
	require.Equal(t, "https://database.example.test", record.Value)
	require.False(t, record.IsPublic)
	require.Equal(t, settingdomain.PayPalDisputeInvoiceSellerProfileGroup, record.Group)
}

func TestPayPalDisputeInvoiceSellerProfileRequiresNameAndAddress(t *testing.T) {
	_, settingService := newTestSettingService(t)
	profileService := NewPayPalDisputeInvoiceSellerProfileService(settingService)

	_, err := profileService.Update(settingdomain.PayPalDisputeInvoiceSellerProfileUpdateRequest{
		Email: "invoice@example.test",
	})
	require.ErrorIs(t, err, ErrPayPalDisputeInvoiceSellerProfileIncomplete)
}
