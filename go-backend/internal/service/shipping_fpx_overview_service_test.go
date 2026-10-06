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

func TestFpxOverviewCountsServiceDirectory(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&shippingdomain.FpxChannel{}))

	repo := repository.NewShippingRepository(db)
	enabledChannel := &shippingdomain.FpxChannel{ServiceCode: "FPX-OVERVIEW-ENABLED", DisplayName: "Enabled", Enabled: true}
	disabledChannel := &shippingdomain.FpxChannel{ServiceCode: "FPX-OVERVIEW-DISABLED", DisplayName: "Disabled", Enabled: false}
	testChannel := &shippingdomain.FpxChannel{Environment: "test", ServiceCode: "FPX-OVERVIEW-TEST", DisplayName: "Test", Enabled: true}
	require.NoError(t, repo.CreateFpxChannel(enabledChannel))
	require.NoError(t, repo.CreateFpxChannel(disabledChannel))
	require.NoError(t, repo.CreateFpxChannel(testChannel))

	overview, err := NewShippingService(repo).GetFpxOverview()
	require.NoError(t, err)
	require.False(t, overview.GeneratedAt.IsZero())
	require.Equal(t, int64(2), overview.Channels.Total)
	require.Equal(t, int64(1), overview.Channels.Enabled)
}
