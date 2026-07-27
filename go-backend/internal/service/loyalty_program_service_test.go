package service

import (
	"testing"

	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestLoyaltyProgramServiceCreatesImmutableVersions(t *testing.T) {
	db := openLoyaltyProgramTestDB(t)
	repo := repository.NewLoyaltyProgramRepository(db)
	service := NewLoyaltyProgramService(repo)

	first, err := service.Update(LoyaltyProgramConfigInput{
		Enabled:                   true,
		Currency:                  "usd",
		ExchangeRatePoints:        100,
		MinRedeemPoints:           1000,
		MaxValuePerDayCents:       50000,
		CardExpiryDays:            365,
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
		RedeemValuesCents:         []int64{1000, 5000},
	})
	require.NoError(t, err)
	require.Equal(t, "USD", first.Currency)
	require.Equal(t, 1, first.Version)

	second, err := service.Update(LoyaltyProgramConfigInput{
		Enabled:                   false,
		Currency:                  "USD",
		ExchangeRatePoints:        80,
		MinRedeemPoints:           800,
		MaxValuePerDayCents:       25000,
		CardExpiryDays:            90,
		ReferralReferrerPoints:    120,
		ReferralRefereePoints:     60,
		CheckInBasePoints:         8,
		CheckInStreakIntervalDays: 5,
		CheckInStreakBonusPoints:  3,
		CheckInMaxPoints:          30,
		RedeemValuesCents:         []int64{1000, 2500},
	})
	require.NoError(t, err)
	require.Equal(t, 2, second.Version)

	active, err := service.GetActive()
	require.NoError(t, err)
	require.Equal(t, second.ID, active.ID)
	require.Equal(t, 2, active.Version)

	var archived loyalty.ProgramConfig
	require.NoError(t, db.Where("id = ?", first.ID).First(&archived).Error)
	require.Equal(t, "archived", archived.Status)
	require.Equal(t, 100, archived.ExchangeRatePoints)
}

func TestLoyaltyProgramPublicConfigMarksUnavailableOptionsInactive(t *testing.T) {
	db := openLoyaltyProgramTestDB(t)
	repo := repository.NewLoyaltyProgramRepository(db)
	service := NewLoyaltyProgramService(repo)

	_, err := service.Update(LoyaltyProgramConfigInput{
		Enabled:                   true,
		Currency:                  "USD",
		ExchangeRatePoints:        100,
		MinRedeemPoints:           1000,
		MaxValuePerDayCents:       50000,
		CardExpiryDays:            365,
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
		RedeemValuesCents:         []int64{500, 1000},
	})
	require.NoError(t, err)

	response, err := service.GetPublicConfig()
	require.NoError(t, err)
	require.Len(t, response.RedeemOptions, 2)
	require.Equal(t, "inactive", response.RedeemOptions[0].Status)
	require.Equal(t, "active", response.RedeemOptions[1].Status)
}

func openLoyaltyProgramTestDB(t *testing.T) *gorm.DB {
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
		&loyalty.ProgramConfig{},
		&loyalty.ProgramRedeemOption{},
	))
	return db
}
