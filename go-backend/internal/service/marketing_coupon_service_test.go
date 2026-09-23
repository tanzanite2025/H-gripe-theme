package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMarketingServiceValidateCouponRejectsPerUserUsageLimit(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:              "WELCOME20",
		Type:              "fixed",
		ValueMinor:        2000,
		UsageLimitPerUser: 1,
		StartDate:         now.Add(-time.Hour),
		EndDate:           now.Add(time.Hour),
		Enabled:           true,
	}
	require.NoError(t, db.Create(&cp).Error)
	require.NoError(t, db.Create(&coupon.CouponUsage{
		CouponID:      cp.ID,
		UserID:        42,
		OrderID:       1001,
		DiscountMinor: 2000,
	}).Error)

	_, discount, err := marketingService.ValidateCoupon("WELCOME20", 42, 10000, "")

	require.ErrorIs(t, err, ErrCouponPerUserUsageLimitReached)
	assert.Zero(t, discount)

	validCoupon, discount, err := marketingService.ValidateCoupon("WELCOME20", 7, 10000, "")
	require.NoError(t, err)
	require.NotNil(t, validCoupon)
	assert.Equal(t, cp.ID, validCoupon.ID)
	assert.Equal(t, int64(2000), discount)
}

func TestMarketingServiceUseCouponRejectsPerUserUsageLimit(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:              "WELCOME10",
		Type:              "fixed",
		ValueMinor:        1000,
		UsageLimitPerUser: 1,
		StartDate:         now.Add(-time.Hour),
		EndDate:           now.Add(time.Hour),
		Enabled:           true,
	}
	require.NoError(t, db.Create(&cp).Error)
	require.NoError(t, db.Create(&coupon.CouponUsage{
		CouponID:      cp.ID,
		UserID:        42,
		OrderID:       1001,
		DiscountMinor: 1000,
	}).Error)

	err := marketingService.UseCoupon(cp.ID, 42, 1002, 1000, "")

	require.ErrorIs(t, err, ErrCouponPerUserUsageLimitReached)

	var savedCoupon coupon.Coupon
	require.NoError(t, db.First(&savedCoupon, cp.ID).Error)
	assert.Equal(t, 0, savedCoupon.UsedCount)

	var usageCount int64
	require.NoError(t, db.Model(&coupon.CouponUsage{}).Where("coupon_id = ? AND user_id = ?", cp.ID, uint(42)).Count(&usageCount).Error)
	assert.Equal(t, int64(1), usageCount)
}

func TestMarketingServiceValidateCouponEnforcesGuestEmailUsageLimit(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:              "GUESTWELCOME",
		Type:              "fixed",
		ValueMinor:        5000,
		UsageLimitPerUser: 1,
		StartDate:         now.Add(-time.Hour),
		EndDate:           now.Add(time.Hour),
		Enabled:           true,
	}
	require.NoError(t, db.Create(&cp).Error)
	require.NoError(t, db.Create(&coupon.CouponUsage{
		CouponID:      cp.ID,
		UserID:        0,
		Email:         "buyer@example.com",
		OrderID:       1001,
		DiscountMinor: 5000,
	}).Error)

	_, discount, err := marketingService.ValidateCoupon("GUESTWELCOME", 0, 10000, " Buyer@Example.com ")

	require.ErrorIs(t, err, ErrCouponPerUserUsageLimitReached)
	assert.Zero(t, discount)

	validCoupon, discount, err := marketingService.ValidateCoupon("GUESTWELCOME", 0, 10000, "other@example.com")
	require.NoError(t, err)
	require.NotNil(t, validCoupon)
	assert.Equal(t, int64(5000), discount)
}

func TestMarketingServiceValidateCouponRequiresGuestEmailForPerUserLimit(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:              "GUESTIDENTITY",
		Type:              "fixed",
		ValueMinor:        500,
		UsageLimitPerUser: 1,
		StartDate:         now.Add(-time.Hour),
		EndDate:           now.Add(time.Hour),
		Enabled:           true,
	}
	require.NoError(t, db.Create(&cp).Error)

	_, discount, err := marketingService.ValidateCoupon("GUESTIDENTITY", 0, 10000, "")

	require.ErrorIs(t, err, ErrCouponUsageIdentityRequired)
	assert.Zero(t, discount)
}

func TestMarketingServiceUseCouponStoresNormalizedGuestEmail(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:              "GUESTSAVE",
		Type:              "fixed",
		ValueMinor:        1000,
		UsageLimitPerUser: 1,
		StartDate:         now.Add(-time.Hour),
		EndDate:           now.Add(time.Hour),
		Enabled:           true,
	}
	require.NoError(t, db.Create(&cp).Error)

	require.NoError(t, marketingService.UseCoupon(cp.ID, 0, 1001, 1000, " Buyer@Example.com "))

	var usage coupon.CouponUsage
	require.NoError(t, db.Where("coupon_id = ? AND order_id = ?", cp.ID, 1001).First(&usage).Error)
	assert.Equal(t, "buyer@example.com", usage.Email)
}

func TestMarketingServiceUseCouponRejectsGlobalUsageLimitInsideTransaction(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:       "FLASH100",
		Type:       "fixed",
		ValueMinor: 10000,
		UsageLimit: 1,
		UsedCount:  1,
		StartDate:  now.Add(-time.Hour),
		EndDate:    now.Add(time.Hour),
		Enabled:    true,
	}
	require.NoError(t, db.Create(&cp).Error)

	err := marketingService.UseCoupon(cp.ID, 42, 1001, 10000, "")

	require.ErrorIs(t, err, repository.ErrCouponUsageLimitReached)

	var savedCoupon coupon.Coupon
	require.NoError(t, db.First(&savedCoupon, cp.ID).Error)
	assert.Equal(t, 1, savedCoupon.UsedCount)

	var usageCount int64
	require.NoError(t, db.Model(&coupon.CouponUsage{}).Where("coupon_id = ?", cp.ID).Count(&usageCount).Error)
	assert.Zero(t, usageCount)
}

func TestMarketingServiceValidateCouponCapsFixedDiscountAtAmount(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:           "BIG50",
		Type:           "fixed",
		ValueMinor:     5000,
		MinAmountMinor: 2500,
		StartDate:      now.Add(-time.Hour),
		EndDate:        now.Add(time.Hour),
		Enabled:        true,
	}
	require.NoError(t, db.Create(&cp).Error)

	validCoupon, discount, err := marketingService.ValidateCoupon("BIG50", 42, 3000, "")

	require.NoError(t, err)
	require.NotNil(t, validCoupon)
	assert.Equal(t, cp.ID, validCoupon.ID)
	assert.Equal(t, int64(3000), discount)
}

func TestMarketingServiceAnalyzePromotionStackingRiskFlagsZeroTotalStacks(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:           "STACK80",
		Type:           "fixed",
		ValueMinor:     8000,
		MinAmountMinor: 10000,
		StartDate:      now.Add(-time.Hour),
		EndDate:        now.Add(time.Hour),
		Enabled:        true,
	}
	require.NoError(t, db.Create(&cp).Error)
	require.NoError(t, db.Create(&loyalty.MemberLevel{
		Name:                "Gold",
		MinPoints:           0,
		MaxPoints:           999999,
		DiscountRateDecimal: "20",
	}).Error)

	programService := NewLoyaltyProgramService(repository.NewLoyaltyProgramRepository(db))
	_, err := programService.Update(LoyaltyProgramConfigInput{
		Enabled:                   true,
		Currency:                  "USD",
		PurchaseEarnPointsPerUnit: 1,
		ExchangeRatePoints:        100,
		ReferralReferrerPoints:    0,
		ReferralRefereePoints:     0,
		CheckInBasePoints:         0,
		CheckInStreakIntervalDays: 1,
		CheckInStreakBonusPoints:  0,
		CheckInMaxPoints:          0,
	})
	require.NoError(t, err)
	marketingService.ConfigureLoyaltyProgram(programService)

	analysis, err := marketingService.AnalyzePromotionStackingRisk()

	require.NoError(t, err)
	require.NotNil(t, analysis)
	assert.Equal(t, "critical", analysis.Summary.Severity)
	assert.Equal(t, 1, analysis.Summary.CandidateCouponCount)
	assert.Equal(t, 20.0, analysis.Summary.MaxMemberDiscountRate)
	assert.Equal(t, 50.0, analysis.Summary.DirectPointsDiscountCapRate)
	assert.Equal(t, 50.0, analysis.Summary.DirectPointsDiscountCapRate)
	require.NotEmpty(t, analysis.Items)

	item := analysis.Items[0]
	assert.Equal(t, "zero_total", item.Kind)
	assert.Equal(t, "STACK80", item.CouponCode)
	assert.Equal(t, "active", item.CouponStatus)
	assert.Equal(t, "100.00", item.EstimatedSubtotal)
	assert.Equal(t, "0.00", item.EstimatedPayableAmount)
	assert.Contains(t, item.Factors, "fixed_coupon")
	assert.Contains(t, item.Factors, "member_level_discount")
	assert.Contains(t, item.Factors, "direct_points_discount")
}

// newTestMarketingService builds the minimal transactional graph required by
// coupon validation/usage and promotion-risk analysis.  The production
// service requires every repository in TxManager because WithinTx materializes
// all transactional repository handles, even when a particular operation only
// touches coupons.
func newTestMarketingService(t *testing.T) (*gorm.DB, *MarketingService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(
		&coupon.Coupon{},
		&coupon.CouponUsage{},
		&loyalty.MemberLevel{},
		&loyalty.ProgramConfig{},
	))

	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	txManager := repository.NewTxManager(db, orderRepo, productRepo, couponRepo, loyaltyRepo, paymentRepo)
	service := NewMarketingService(txManager, couponRepo, loyaltyRepo)
	return db, service
}

func TestPromotionRiskMemberRateKeepsExactDecimalFraction(t *testing.T) {
	rate := memberLevelDiscountRateFraction(&loyalty.MemberLevel{
		DiscountRateDecimal: "33.333333333333333",
	})

	require.Equal(t, "33333333333333333/100000000000000000", rate.RatString())
}
