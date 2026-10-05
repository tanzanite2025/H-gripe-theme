package service

import (
	"testing"
	"time"

	shippingdomain "commerce-platform/internal/domain/shipping"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreateYanwenPublishedChannelRequiresOfficialProductCatalogEntry(t *testing.T) {
	db := openYanwenPublishedChannelCatalogValidationTestDatabase(t)
	productCatalogRepository := repository.NewYanwenProductCatalogRepository(db)
	collectionService := NewYanwenPublishedCollectionService(
		repository.NewYanwenPublishedChannelRepository(db),
		productCatalogRepository,
	)

	require.NoError(t, db.Create(&shippingdomain.YanwenProductCatalogEntry{
		Environment:  shippingdomain.YanwenPublishedChannelEnvironmentProduction,
		ProductID:    "481",
		NameChinese:  "燕文专线追踪",
		LastSyncedAt: time.Now().UTC(),
	}).Error)

	acceptedChannel := shippingdomain.YanwenPublishedChannel{
		Environment:       shippingdomain.YanwenPublishedChannelEnvironmentProduction,
		ProductCode:       "481",
		DisplayName:       "Yanwen tracked",
		VolumetricDivisor: 8000,
	}
	require.NoError(t, collectionService.CreateYanwenPublishedChannel(&acceptedChannel))

	rejectedChannel := shippingdomain.YanwenPublishedChannel{
		Environment:       shippingdomain.YanwenPublishedChannelEnvironmentProduction,
		ProductCode:       "999",
		DisplayName:       "Unknown Yanwen product",
		VolumetricDivisor: 8000,
	}
	err := collectionService.CreateYanwenPublishedChannel(&rejectedChannel)
	require.ErrorContains(t, err, `Yanwen product "999" is not present in the official production catalog`)
}

func TestUpdateYanwenPublishedChannelRequiresCurrentEnvironmentProductCatalogEntry(t *testing.T) {
	db := openYanwenPublishedChannelCatalogValidationTestDatabase(t)
	productCatalogRepository := repository.NewYanwenProductCatalogRepository(db)
	collectionService := NewYanwenPublishedCollectionService(
		repository.NewYanwenPublishedChannelRepository(db),
		productCatalogRepository,
	)

	channel := shippingdomain.YanwenPublishedChannel{
		Environment:       shippingdomain.YanwenPublishedChannelEnvironmentProduction,
		ProductCode:       "481",
		DisplayName:       "Yanwen tracked",
		VolumetricDivisor: 8000,
	}
	require.NoError(t, db.Create(&channel).Error)
	require.NoError(t, db.Create(&shippingdomain.YanwenProductCatalogEntry{
		Environment:  shippingdomain.YanwenPublishedChannelEnvironmentProduction,
		ProductID:    "481",
		NameChinese:  "燕文专线追踪",
		LastSyncedAt: time.Now().UTC(),
	}).Error)

	channel.ProductCode = "999"
	err := collectionService.UpdateYanwenPublishedChannel(&channel)
	require.ErrorContains(t, err, `Yanwen product "999" is not present in the official production catalog`)
}

func TestListYanwenPublishedChannelsDefaultsToProductionWithoutCrossEnvironmentRecords(t *testing.T) {
	db := openYanwenPublishedChannelCatalogValidationTestDatabase(t)
	collectionService := NewYanwenPublishedCollectionService(
		repository.NewYanwenPublishedChannelRepository(db),
		repository.NewYanwenProductCatalogRepository(db),
	)

	productionChannel := shippingdomain.YanwenPublishedChannel{
		Environment:       shippingdomain.YanwenPublishedChannelEnvironmentProduction,
		ProductCode:       "481",
		DisplayName:       "Production channel",
		VolumetricDivisor: 8000,
	}
	fatChannel := shippingdomain.YanwenPublishedChannel{
		Environment:       shippingdomain.YanwenPublishedChannelEnvironmentFAT,
		ProductCode:       "481-FAT",
		DisplayName:       "FAT channel",
		VolumetricDivisor: 8000,
	}
	require.NoError(t, db.Create(&productionChannel).Error)
	require.NoError(t, db.Create(&fatChannel).Error)

	channels, err := collectionService.ListYanwenPublishedChannels("", false)
	require.NoError(t, err)
	require.Len(t, channels, 1)
	require.Equal(t, shippingdomain.YanwenPublishedChannelEnvironmentProduction, channels[0].Environment)

	_, err = collectionService.ListYanwenPublishedChannels("sandbox", false)
	require.ErrorContains(t, err, "Yanwen channel environment must be fat or production")
}

func openYanwenPublishedChannelCatalogValidationTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&shippingdomain.YanwenPublishedChannel{},
		&shippingdomain.YanwenProductCatalogEntry{},
	))
	return db
}
