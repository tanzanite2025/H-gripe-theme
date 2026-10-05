package service

import (
	"testing"

	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreateTemplateWithCarrierServicesUsesPublishedFpxCountryScope(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&shippingdomain.Carrier{},
		&shippingdomain.ShippingTemplate{},
		&shippingdomain.ShippingRule{},
		&shippingdomain.CarrierService{},
		&shippingdomain.FpxChannel{},
		&shippingdomain.YanwenPublishedChannel{},
	))

	carrier := shippingdomain.Carrier{Name: "4PX", Code: "4PX", Enabled: true}
	require.NoError(t, db.Create(&carrier).Error)
	channel := shippingdomain.FpxChannel{
		ServiceCode: "EU-EXPRESS",
		DisplayName: "4PX Europe Express",
		Countries:   `["DE","FR"]`,
		Enabled:     true,
	}
	require.NoError(t, db.Create(&channel).Error)

	shippingService := NewShippingService(repository.NewShippingRepository(db))
	template := shippingdomain.ShippingTemplate{Name: "Collection-backed template", Type: "weight", Currency: "USD"}
	services := []shippingdomain.CarrierService{{
		CarrierID:         carrier.ID,
		FpxChannelID:      &channel.ID,
		ProviderCode:      "4PX",
		ServiceCode:       "CLIENT-SUPPLIED-CODE",
		ServiceName:       "Client-supplied name",
		Countries:         `["CN"]`,
		Currency:          "USD",
		BillingMode:       "actual_weight",
		Enabled:           true,
		VolumetricDivisor: 6000,
	}}

	require.NoError(t, shippingService.CreateTemplateWithCarrierServices(&template, services))
	var stored shippingdomain.CarrierService
	require.NoError(t, db.Where("template_id = ?", template.ID).First(&stored).Error)
	require.NotNil(t, stored.FpxChannelID)
	require.Equal(t, channel.ID, *stored.FpxChannelID)
	require.Equal(t, "EU-EXPRESS", stored.ServiceCode)
	require.Equal(t, "4PX Europe Express", stored.ServiceName)
	require.Equal(t, `["DE","FR"]`, stored.Countries)
}

func TestCreateTemplateRejectsUnpublishedFpxService(t *testing.T) {
	db, shippingService, carrier := newShippingServiceCollectionTestFixture(t, "4PX")
	template := shippingdomain.ShippingTemplate{Name: "Unknown 4PX service", Type: "weight", Currency: "USD"}
	services := []shippingdomain.CarrierService{{
		CarrierID: carrier.ID, ProviderCode: "4PX", ServiceCode: "NOT-PUBLISHED",
		ServiceName: "Forged service", Countries: `["US"]`, Currency: "USD", Enabled: true,
	}}

	err := shippingService.CreateTemplateWithCarrierServices(&template, services)
	require.ErrorContains(t, err, "select a 4PX service collection record")
	var templateCount int64
	require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).Count(&templateCount).Error)
	require.Zero(t, templateCount)
}

func TestCreateTemplateRejectsDisabledYanwenProduct(t *testing.T) {
	db, shippingService, carrier := newShippingServiceCollectionTestFixture(t, "YANWEN")
	channel := shippingdomain.YanwenPublishedChannel{
		ProductCode: "481", DisplayName: "Yanwen standard", Countries: `["US"]`, Enabled: false, VolumetricDivisor: 8000,
	}
	require.NoError(t, db.Create(&channel).Error)
	template := shippingdomain.ShippingTemplate{Name: "Disabled Yanwen service", Type: "weight", Currency: "USD"}
	services := []shippingdomain.CarrierService{{
		CarrierID: carrier.ID, ProviderCode: "YANWEN", ServiceCode: "YANWEN:481",
		YanwenPublishedChannelID: &channel.ID,
		ServiceName:              "Yanwen standard", Countries: `["US"]`, Currency: "USD", Enabled: true,
	}}

	err := shippingService.CreateTemplateWithCarrierServices(&template, services)
	require.ErrorContains(t, err, "is disabled in the service collection")
	var templateCount int64
	require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).Count(&templateCount).Error)
	require.Zero(t, templateCount)
}

func TestCreateTemplateRejectsProviderCodeThatDoesNotMatchCarrier(t *testing.T) {
	_, shippingService, carrier := newShippingServiceCollectionTestFixture(t, "4PX")
	template := shippingdomain.ShippingTemplate{Name: "Mismatched provider", Type: "weight", Currency: "USD"}
	services := []shippingdomain.CarrierService{{
		CarrierID: carrier.ID, ProviderCode: "YANWEN", ServiceCode: "YANWEN:481",
		ServiceName: "Yanwen standard", Currency: "USD", Enabled: true,
	}}

	err := shippingService.CreateTemplateWithCarrierServices(&template, services)
	require.ErrorContains(t, err, "does not match carrier")
}

func TestCreateTemplateWithCarrierServicesUsesPublishedYanwenCountryScope(t *testing.T) {
	db, shippingService, carrier := newShippingServiceCollectionTestFixture(t, "YANWEN")
	require.NoError(t, db.Create(&shippingdomain.YanwenPublishedChannel{
		Environment: shippingdomain.YanwenPublishedChannelEnvironmentFAT,
		ProductCode: "481", DisplayName: "Yanwen FAT standard", Countries: `["CA"]`, Enabled: true, VolumetricDivisor: 8000,
	}).Error)
	productionChannel := shippingdomain.YanwenPublishedChannel{
		Environment: shippingdomain.YanwenPublishedChannelEnvironmentProduction,
		ProductCode: "481", DisplayName: "Yanwen production standard", Countries: `["US","CA"]`, Enabled: true, VolumetricDivisor: 8000,
	}
	require.NoError(t, db.Create(&productionChannel).Error)
	template := shippingdomain.ShippingTemplate{Name: "Yanwen collection template", Type: "weight", Currency: "USD"}
	services := []shippingdomain.CarrierService{{
		CarrierID: carrier.ID, ProviderCode: "YANWEN", ServiceCode: "481",
		YanwenPublishedChannelID: &productionChannel.ID,
		ServiceName:              "Client-supplied name", Countries: `["CN"]`, Currency: "USD", Enabled: true,
	}}

	require.NoError(t, shippingService.CreateTemplateWithCarrierServices(&template, services))
	var stored shippingdomain.CarrierService
	require.NoError(t, db.Where("template_id = ?", template.ID).First(&stored).Error)
	require.NotNil(t, stored.YanwenPublishedChannelID)
	require.Equal(t, productionChannel.ID, *stored.YanwenPublishedChannelID)
	require.Equal(t, "YANWEN:481", stored.ServiceCode)
	require.Equal(t, "Yanwen production standard", stored.ServiceName)
	require.Equal(t, `["US","CA"]`, stored.Countries)
}

func TestCreateTemplateRejectsFatOnlyYanwenProduct(t *testing.T) {
	db, shippingService, carrier := newShippingServiceCollectionTestFixture(t, "YANWEN")
	fatChannel := shippingdomain.YanwenPublishedChannel{
		Environment: shippingdomain.YanwenPublishedChannelEnvironmentFAT,
		ProductCode: "481", DisplayName: "Yanwen FAT standard", Countries: `["US"]`, Enabled: true, VolumetricDivisor: 8000,
	}
	require.NoError(t, db.Create(&fatChannel).Error)
	template := shippingdomain.ShippingTemplate{Name: "FAT-only Yanwen service", Type: "weight", Currency: "USD"}
	services := []shippingdomain.CarrierService{{
		CarrierID: carrier.ID, ProviderCode: "YANWEN", ServiceCode: "YANWEN:481",
		YanwenPublishedChannelID: &fatChannel.ID,
		ServiceName:              "Yanwen FAT standard", Countries: `["US"]`, Currency: "USD", Enabled: true,
	}}

	err := shippingService.CreateTemplateWithCarrierServices(&template, services)
	require.ErrorContains(t, err, "missing or is not in production")
	var templateCount int64
	require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).Count(&templateCount).Error)
	require.Zero(t, templateCount)
}

func newShippingServiceCollectionTestFixture(t *testing.T, carrierCode string) (*gorm.DB, *ShippingService, shippingdomain.Carrier) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&shippingdomain.Carrier{},
		&shippingdomain.ShippingTemplate{},
		&shippingdomain.ShippingRule{},
		&shippingdomain.CarrierService{},
		&shippingdomain.FpxChannel{},
		&shippingdomain.YanwenPublishedChannel{},
	))
	carrier := shippingdomain.Carrier{Name: carrierCode, Code: carrierCode, Enabled: true}
	require.NoError(t, db.Create(&carrier).Error)
	shippingService := NewShippingService(repository.NewShippingRepository(db))
	shippingService.ConfigureYanwenPublishedCollectionService(
		NewYanwenPublishedCollectionService(repository.NewYanwenPublishedChannelRepository(db), nil),
	)
	return db, shippingService, carrier
}
