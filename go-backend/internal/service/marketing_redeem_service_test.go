package service

import (
	"testing"

	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/domain/setting"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRedeemPointsForGiftCardIsIdempotentAndLossless(t *testing.T) {
	db := openRedemptionTestDB(t)
	couponRepo := repository.NewCouponRepository(db)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	redemptionRepo := repository.NewGiftCardRedemptionRepository(db)
	programRepo := repository.NewLoyaltyProgramRepository(db)
	programService := NewLoyaltyProgramService(programRepo)
	currencyPolicyService := seedTestCurrencyPolicy(t, db)
	programService.ConfigureCurrencyPolicy(currencyPolicyService)

	txManager := repository.NewTxManager(
		db,
		repository.NewOrderRepository(db),
		repository.NewProductRepository(db),
		couponRepo,
		loyaltyRepo,
		repository.NewPaymentRepository(db),
	)
	txManager.ConfigureGiftCardRedemptionRepository(redemptionRepo)
	txManager.ConfigureLoyaltyProgramRepository(programRepo)
	marketingService := NewMarketingService(txManager, couponRepo, loyaltyRepo)
	marketingService.ConfigureLoyaltyProgram(programService)
	marketingService.ConfigureGiftCardRedemptions(redemptionRepo)
	marketingService.ConfigureCurrencyPolicy(currencyPolicyService)

	config, err := programService.Update(LoyaltyProgramConfigInput{
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
		RedeemOptions: []LoyaltyProgramOptionInput{{
			ValueCents:    1000,
			Currency:      "USD",
			StockQuantity: 1,
		}},
	})
	require.NoError(t, err)

	require.NoError(t, db.Create(&loyalty.UserLoyalty{
		UserID:          42,
		TotalPoints:     1500,
		AvailablePoints: 1500,
	}).Error)

	req := RedeemGiftCardRequest{
		OptionID:       config.RedeemOptions[0].ID,
		IdempotencyKey: "test-redemption-42",
	}
	first, err := marketingService.RedeemPointsForGiftCard(42, req, config)
	require.NoError(t, err)
	require.Equal(t, 1000, first.PointsSpent)
	require.Equal(t, 500, first.PointsRemaining)
	require.Equal(t, int64(1000), first.GiftCardValueCents)

	second, err := marketingService.RedeemPointsForGiftCard(42, req, config)
	require.NoError(t, err)
	require.Equal(t, first.RedemptionID, second.RedemptionID)
	require.Equal(t, first.GiftCardID, second.GiftCardID)

	var giftCards []coupon.GiftCard
	require.NoError(t, db.Find(&giftCards).Error)
	assert.Len(t, giftCards, 1)
	assert.Equal(t, uint(42), *giftCards[0].OwnerUserID)
	assert.Equal(t, "loyalty_redemption", giftCards[0].Origin)

	var redemptions []coupon.GiftCardRedemption
	require.NoError(t, db.Find(&redemptions).Error)
	assert.Len(t, redemptions, 1)
	assert.NotNil(t, redemptions[0].LoyaltyTransactionID)

	var loyaltyTransactions []loyalty.LoyaltyTransaction
	require.NoError(t, db.Find(&loyaltyTransactions).Error)
	assert.Len(t, loyaltyTransactions, 1)
	assert.Equal(t, "gift_card_redemption", loyaltyTransactions[0].Source)
	assert.Equal(t, config.ID, *loyaltyTransactions[0].ProgramConfigID)

	var giftCardTransactions []coupon.GiftCardTransaction
	require.NoError(t, db.Find(&giftCardTransactions).Error)
	assert.Len(t, giftCardTransactions, 1)
	assert.Equal(t, "issue", giftCardTransactions[0].Type)
}

func TestRedeemPointsForGiftCardRollsBackOnInsufficientPoints(t *testing.T) {
	db := openRedemptionTestDB(t)
	couponRepo := repository.NewCouponRepository(db)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	redemptionRepo := repository.NewGiftCardRedemptionRepository(db)
	programRepo := repository.NewLoyaltyProgramRepository(db)
	programService := NewLoyaltyProgramService(programRepo)
	currencyPolicyService := seedTestCurrencyPolicy(t, db)
	programService.ConfigureCurrencyPolicy(currencyPolicyService)

	txManager := repository.NewTxManager(
		db,
		repository.NewOrderRepository(db),
		repository.NewProductRepository(db),
		couponRepo,
		loyaltyRepo,
		repository.NewPaymentRepository(db),
	)
	txManager.ConfigureGiftCardRedemptionRepository(redemptionRepo)
	txManager.ConfigureLoyaltyProgramRepository(programRepo)
	marketingService := NewMarketingService(txManager, couponRepo, loyaltyRepo)
	marketingService.ConfigureLoyaltyProgram(programService)
	marketingService.ConfigureCurrencyPolicy(currencyPolicyService)

	config, err := programService.Update(LoyaltyProgramConfigInput{
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
		RedeemOptions: []LoyaltyProgramOptionInput{{
			ValueCents:    1000,
			Currency:      "USD",
			StockQuantity: 1,
		}},
	})
	require.NoError(t, err)
	require.NoError(t, db.Create(&loyalty.UserLoyalty{UserID: 7, AvailablePoints: 10}).Error)

	_, err = marketingService.RedeemPointsForGiftCard(7, RedeemGiftCardRequest{
		OptionID:       config.RedeemOptions[0].ID,
		IdempotencyKey: "rollback-7",
	}, config)
	require.Error(t, err)

	var giftCardCount int64
	require.NoError(t, db.Model(&coupon.GiftCard{}).Count(&giftCardCount).Error)
	assert.Equal(t, int64(0), giftCardCount)
}

func openRedemptionTestDB(t *testing.T) *gorm.DB {
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
		&coupon.GiftCard{},
		&coupon.GiftCardTransaction{},
		&coupon.GiftCardRedemption{},
		&loyalty.LoyaltyTransaction{},
		&loyalty.UserLoyalty{},
		&loyalty.ProgramConfig{},
		&loyalty.ProgramRedeemOption{},
		&setting.Setting{},
	))
	return db
}
