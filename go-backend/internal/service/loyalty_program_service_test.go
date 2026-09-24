package service

import (
	"testing"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/domain/setting"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestLoyaltyProgramServiceCreatesImmutableVersions(t *testing.T) {
	db := openLoyaltyProgramTestDB(t)
	service := newTestLoyaltyProgramService(t, db)

	first, err := service.Update(LoyaltyProgramConfigInput{
		Enabled:                   true,
		Currency:                  "usd",
		PurchaseEarnPointsPerUnit: 1,
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
	})
	require.NoError(t, err)
	require.Equal(t, "USD", first.Currency)
	require.Equal(t, 1, first.Version)

	second, err := service.Update(LoyaltyProgramConfigInput{
		Enabled:                   false,
		Currency:                  "USD",
		PurchaseEarnPointsPerUnit: 2,
		ReferralReferrerPoints:    120,
		ReferralRefereePoints:     60,
		CheckInBasePoints:         8,
		CheckInStreakIntervalDays: 5,
		CheckInStreakBonusPoints:  3,
		CheckInMaxPoints:          30,
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
}

func TestLoyaltyProgramPublicConfigContainsOnlyCurrentProgramRules(t *testing.T) {
	db := openLoyaltyProgramTestDB(t)
	service := newTestLoyaltyProgramService(t, db)

	_, err := service.Update(LoyaltyProgramConfigInput{
		Enabled:                   true,
		Currency:                  "USD",
		PurchaseEarnPointsPerUnit: 1,
		ReferralReferrerPoints:    100,
		ReferralRefereePoints:     50,
		CheckInBasePoints:         10,
		CheckInStreakIntervalDays: 7,
		CheckInStreakBonusPoints:  5,
		CheckInMaxPoints:          50,
	})
	require.NoError(t, err)

	response, err := service.GetPublicConfig()
	require.NoError(t, err)
	require.Equal(t, LoyaltyPointsBaseCurrency, response.PointsBaseCurrency)
	require.Equal(t, "USD", response.Currency)
	require.Equal(t, 1, response.PurchaseEarnPointsPerUnit)
	require.NotNil(t, response.AvailableCurrencies)
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
		&setting.Setting{},
		&loyalty.ProgramConfig{},
	))
	return db
}

func newTestLoyaltyProgramService(t *testing.T, db *gorm.DB) *LoyaltyProgramService {
	t.Helper()
	program := NewLoyaltyProgramService(repository.NewLoyaltyProgramRepository(db))
	program.ConfigureCurrencyPolicy(seedTestCurrencyPolicy(t, db))
	return program
}

func seedTestCurrencyPolicy(t *testing.T, db *gorm.DB) *CurrencyPolicyService {
	t.Helper()
	policy := NewCurrencyPolicyService(repository.NewSettingRepository(db))
	_, err := policy.UpdatePolicy(currency.Policy{
		DisplayCurrencies: []string{"USD"},
	})
	require.NoError(t, err)
	return policy
}
