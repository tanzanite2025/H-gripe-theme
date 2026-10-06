package repository

import (
	"testing"

	"commerce-platform/internal/domain/shipping"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestYanwenPublishedChannelRepositorySeparatesSameProductByEnvironment(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.YanwenPublishedChannel{}))

	repo := NewYanwenPublishedChannelRepository(db)
	productionChannel := &shipping.YanwenPublishedChannel{
		Environment: shipping.YanwenPublishedChannelEnvironmentProduction,
		ProductCode: "481", DisplayName: "燕文生产普货", PackageType: "普货配件",
		VolumetricDivisor: 8000, Enabled: true,
	}
	fatChannel := &shipping.YanwenPublishedChannel{
		Environment: shipping.YanwenPublishedChannelEnvironmentFAT,
		ProductCode: "481", DisplayName: "燕文 FAT 普货", PackageType: "普货配件",
		VolumetricDivisor: 8000, Enabled: true,
	}
	require.NoError(t, repo.CreateYanwenPublishedChannel(productionChannel))
	require.NoError(t, repo.CreateYanwenPublishedChannel(fatChannel))

	productionChannels, err := repo.FindYanwenPublishedChannelsByEnvironment(shipping.YanwenPublishedChannelEnvironmentProduction, true)
	require.NoError(t, err)
	require.Len(t, productionChannels, 1)
	require.Equal(t, "燕文生产普货", productionChannels[0].DisplayName)

	fatChannels, err := repo.FindYanwenPublishedChannelsByEnvironment(shipping.YanwenPublishedChannelEnvironmentFAT, true)
	require.NoError(t, err)
	require.Len(t, fatChannels, 1)
	require.Equal(t, "燕文 FAT 普货", fatChannels[0].DisplayName)

	productionChannel.Enabled = false
	productionChannel.RequireReceiverTaxNumber = true
	productionChannel.RequireIOSS = true
	productionChannel.RequireEORI = true
	require.NoError(t, repo.UpdateYanwenPublishedChannel(productionChannel))
	productionChannels, err = repo.FindYanwenPublishedChannelsByEnvironment(shipping.YanwenPublishedChannelEnvironmentProduction, true)
	require.NoError(t, err)
	require.Empty(t, productionChannels)
	productionChannels, err = repo.FindYanwenPublishedChannelsByEnvironment(shipping.YanwenPublishedChannelEnvironmentProduction, false)
	require.NoError(t, err)
	require.Len(t, productionChannels, 1)
	require.True(t, productionChannels[0].RequireReceiverTaxNumber)
	require.True(t, productionChannels[0].RequireIOSS)
	require.True(t, productionChannels[0].RequireEORI)
	fatChannels, err = repo.FindYanwenPublishedChannelsByEnvironment(shipping.YanwenPublishedChannelEnvironmentFAT, true)
	require.NoError(t, err)
	require.Len(t, fatChannels, 1, "changing production approval must not affect the FAT channel")

	require.NoError(t, repo.DeleteYanwenPublishedChannel(productionChannel.ID))
	productionChannels, err = repo.FindYanwenPublishedChannelsByEnvironment(shipping.YanwenPublishedChannelEnvironmentProduction, false)
	require.NoError(t, err)
	require.Empty(t, productionChannels)
}
