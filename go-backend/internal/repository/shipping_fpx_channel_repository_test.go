package repository

import (
	"testing"

	"commerce-platform/internal/domain/shipping"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestUpsertFpxChannelsPreservesApprovalAndDisablesNewChannels(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shipping.FpxChannel{}))

	repo := NewShippingRepository(db)
	require.NoError(t, repo.CreateFpxChannel(&shipping.FpxChannel{
		ServiceCode: "APPROVED-1",
		DisplayName: "旧名称",
		Enabled:     true,
	}))

	stats, err := repo.UpsertFpxChannels([]shipping.FpxChannel{
		{ServiceCode: "APPROVED-1", DisplayName: "官方新名称", Enabled: false},
		{ServiceCode: "NEW-1", DisplayName: "新官方服务", Enabled: true},
	})
	require.NoError(t, err)
	require.Equal(t, 2, stats.Scanned)
	require.Equal(t, 1, stats.Added)
	require.Equal(t, 1, stats.Updated)
	require.Equal(t, 1, stats.PreservedEnabled)

	channels, err := repo.FindAllFpxChannels(false)
	require.NoError(t, err)
	require.Len(t, channels, 2)
	byCode := make(map[string]shipping.FpxChannel, len(channels))
	for _, channel := range channels {
		byCode[channel.ServiceCode] = channel
	}
	require.Equal(t, "官方新名称", byCode["APPROVED-1"].DisplayName)
	require.True(t, byCode["APPROVED-1"].Enabled, "sync must preserve the operator's approval")
	require.False(t, byCode["NEW-1"].Enabled, "new official services must require explicit approval")
}
