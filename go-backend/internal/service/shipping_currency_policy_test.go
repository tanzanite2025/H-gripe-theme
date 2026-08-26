package service

import (
	"testing"

	"commerce-platform/internal/domain/currency"
	settingdomain "commerce-platform/internal/domain/setting"
	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestShippingServiceDefaultsNewSourceCurrenciesToBackendEntryCurrency(t *testing.T) {
	db, shippingService, policyService := newCurrencyAwareShippingService(t)
	_, err := policyService.UpdatePolicy(currency.Policy{PrimaryCurrency: "CNY"})
	require.NoError(t, err)

	template := shippingdomain.ShippingTemplate{
		Name:       "Policy default shipping",
		Type:       "weight",
		DefaultFee: 10,
		Enabled:    true,
		Rules: []shippingdomain.ShippingRule{
			{Region: "US", Fee: 5},
		},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))
	require.Equal(t, "CNY", template.Currency)
	require.Len(t, template.Rules, 1)
	require.Equal(t, "CNY", template.Rules[0].Currency)

	var storedTemplate shippingdomain.ShippingTemplate
	require.NoError(t, db.Preload("Rules").First(&storedTemplate, template.ID).Error)
	require.Equal(t, "CNY", storedTemplate.Currency)
	require.Equal(t, "CNY", storedTemplate.Rules[0].Currency)

	carrier := shippingdomain.Carrier{Name: "Policy carrier", Code: "policy-carrier", Enabled: true}
	require.NoError(t, shippingService.CreateCarrier(&carrier))
	carrierService := shippingdomain.CarrierService{
		CarrierID:   carrier.ID,
		ServiceCode: "policy-service",
		ServiceName: "Policy service",
		Enabled:     true,
	}
	require.NoError(t, shippingService.CreateCarrierService(&carrierService))
	require.Equal(t, "CNY", carrierService.Currency)
}

func TestShippingServicePreservesExistingCurrencyWhenUpdateOmitsIt(t *testing.T) {
	_, shippingService, policyService := newCurrencyAwareShippingService(t)
	_, err := policyService.UpdatePolicy(currency.Policy{PrimaryCurrency: "CNY"})
	require.NoError(t, err)

	template := shippingdomain.ShippingTemplate{
		Name:       "Historical USD shipping",
		Type:       "weight",
		Currency:   "USD",
		DefaultFee: 10,
		Enabled:    true,
		Rules: []shippingdomain.ShippingRule{
			{Region: "US", Currency: "USD", Fee: 5},
		},
	}
	require.NoError(t, shippingService.CreateTemplate(&template))

	updated := shippingdomain.ShippingTemplate{
		ID:         template.ID,
		Name:       "Historical USD shipping updated",
		Type:       "weight",
		DefaultFee: 12,
		Enabled:    true,
		Rules: []shippingdomain.ShippingRule{
			{Region: "US", Fee: 6},
		},
	}
	require.NoError(t, shippingService.UpdateTemplate(&updated))

	found, err := shippingService.GetTemplate(template.ID)
	require.NoError(t, err)
	require.Equal(t, "USD", found.Currency)
	require.Len(t, found.Rules, 1)
	require.Equal(t, "USD", found.Rules[0].Currency)
}

func newCurrencyAwareShippingService(t *testing.T) (*gorm.DB, *ShippingService, *CurrencyPolicyService) {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	require.NoError(t, db.AutoMigrate(
		&settingdomain.Setting{},
		&shippingdomain.ShippingTemplate{},
		&shippingdomain.ShippingRule{},
		&shippingdomain.Carrier{},
		&shippingdomain.CarrierService{},
	))

	settingRepo := repository.NewSettingRepository(db)
	policyService := NewCurrencyPolicyService(settingRepo)
	shippingService := NewShippingService(repository.NewShippingRepository(db))
	shippingService.ConfigureCurrencyPolicy(policyService)
	return db, shippingService, policyService
}
