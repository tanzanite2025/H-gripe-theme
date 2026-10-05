package repository

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/shipping"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestDisableYanwenPublishedChannelsMissingOfficialProductsIsEnvironmentScoped(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenPublishedChannel{}))

	productionKept := shipping.YanwenPublishedChannel{
		Environment: shipping.YanwenPublishedChannelEnvironmentProduction,
		ProductCode: "481", DisplayName: "Production kept", VolumetricDivisor: 8000, Enabled: true,
	}
	productionDisabled := shipping.YanwenPublishedChannel{
		Environment: shipping.YanwenPublishedChannelEnvironmentProduction,
		ProductCode: "484", DisplayName: "Production removed", VolumetricDivisor: 8000, Enabled: true,
	}
	fatChannel := shipping.YanwenPublishedChannel{
		Environment: shipping.YanwenPublishedChannelEnvironmentFAT,
		ProductCode: "484", DisplayName: "FAT remains independent", VolumetricDivisor: 8000, Enabled: true,
	}
	require.NoError(t, db.Create(&productionKept).Error)
	require.NoError(t, db.Create(&productionDisabled).Error)
	require.NoError(t, db.Create(&fatChannel).Error)

	repo := NewYanwenPublishedChannelRepository(db)
	updatedAt := time.Date(2026, 10, 5, 8, 0, 0, 0, time.UTC)
	disabledCount, err := repo.DisableMissingOfficialProducts(
		shipping.YanwenPublishedChannelEnvironmentProduction,
		[]string{"481"},
		updatedAt,
	)
	require.NoError(t, err)
	require.Equal(t, int64(1), disabledCount)

	var loadedProductionKept shipping.YanwenPublishedChannel
	require.NoError(t, db.First(&loadedProductionKept, productionKept.ID).Error)
	require.True(t, loadedProductionKept.Enabled)
	var loadedProductionDisabled shipping.YanwenPublishedChannel
	require.NoError(t, db.First(&loadedProductionDisabled, productionDisabled.ID).Error)
	require.False(t, loadedProductionDisabled.Enabled)
	require.True(t, loadedProductionDisabled.UpdatedAt.Equal(updatedAt))
	var loadedFat shipping.YanwenPublishedChannel
	require.NoError(t, db.First(&loadedFat, fatChannel.ID).Error)
	require.True(t, loadedFat.Enabled)
}
