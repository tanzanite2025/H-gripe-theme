package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarketingServiceValidateCouponRejectsPerUserUsageLimit(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:              "WELCOME20",
		Type:              "fixed",
		Value:             20,
		UsageLimitPerUser: 1,
		StartDate:         now.Add(-time.Hour),
		EndDate:           now.Add(time.Hour),
		Enabled:           true,
	}
	require.NoError(t, db.Create(&cp).Error)
	require.NoError(t, db.Create(&coupon.CouponUsage{
		CouponID: cp.ID,
		UserID:   42,
		OrderID:  1001,
		Discount: 20,
	}).Error)

	_, discount, err := marketingService.ValidateCoupon("WELCOME20", 42, 100)

	require.ErrorIs(t, err, ErrCouponPerUserUsageLimitReached)
	assert.Zero(t, discount)

	validCoupon, discount, err := marketingService.ValidateCoupon("WELCOME20", 7, 100)
	require.NoError(t, err)
	require.NotNil(t, validCoupon)
	assert.Equal(t, cp.ID, validCoupon.ID)
	assert.InDelta(t, 20, discount, 0.001)
}

func TestMarketingServiceUseCouponRejectsPerUserUsageLimit(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:              "WELCOME10",
		Type:              "fixed",
		Value:             10,
		UsageLimitPerUser: 1,
		StartDate:         now.Add(-time.Hour),
		EndDate:           now.Add(time.Hour),
		Enabled:           true,
	}
	require.NoError(t, db.Create(&cp).Error)
	require.NoError(t, db.Create(&coupon.CouponUsage{
		CouponID: cp.ID,
		UserID:   42,
		OrderID:  1001,
		Discount: 10,
	}).Error)

	err := marketingService.UseCoupon(cp.ID, 42, 1002, 10)

	require.ErrorIs(t, err, ErrCouponPerUserUsageLimitReached)

	var savedCoupon coupon.Coupon
	require.NoError(t, db.First(&savedCoupon, cp.ID).Error)
	assert.Equal(t, 0, savedCoupon.UsedCount)

	var usageCount int64
	require.NoError(t, db.Model(&coupon.CouponUsage{}).Where("coupon_id = ? AND user_id = ?", cp.ID, uint(42)).Count(&usageCount).Error)
	assert.Equal(t, int64(1), usageCount)
}

func TestMarketingServiceUseCouponRejectsGlobalUsageLimitInsideTransaction(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:       "FLASH100",
		Type:       "fixed",
		Value:      100,
		UsageLimit: 1,
		UsedCount:  1,
		StartDate:  now.Add(-time.Hour),
		EndDate:    now.Add(time.Hour),
		Enabled:    true,
	}
	require.NoError(t, db.Create(&cp).Error)

	err := marketingService.UseCoupon(cp.ID, 42, 1001, 100)

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
		Code:      "BIG50",
		Type:      "fixed",
		Value:     50,
		MinAmount: 25,
		StartDate: now.Add(-time.Hour),
		EndDate:   now.Add(time.Hour),
		Enabled:   true,
	}
	require.NoError(t, db.Create(&cp).Error)

	validCoupon, discount, err := marketingService.ValidateCoupon("BIG50", 42, 30)

	require.NoError(t, err)
	require.NotNil(t, validCoupon)
	assert.Equal(t, cp.ID, validCoupon.ID)
	assert.InDelta(t, 30, discount, 0.001)
}

func TestMarketingServiceAnalyzePromotionStackingRiskFlagsZeroTotalStacks(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	now := time.Now()
	cp := coupon.Coupon{
		Code:      "STACK80",
		Type:      "fixed",
		Value:     80,
		MinAmount: 100,
		StartDate: now.Add(-time.Hour),
		EndDate:   now.Add(time.Hour),
		Enabled:   true,
	}
	require.NoError(t, db.Create(&cp).Error)
	require.NoError(t, db.Create(&loyalty.MemberLevel{
		Name:         "Gold",
		MinPoints:    0,
		MaxPoints:    999999,
		DiscountRate: 20,
	}).Error)

	programService := NewLoyaltyProgramService(repository.NewLoyaltyProgramRepository(db))
	_, err := programService.Update(LoyaltyProgramConfigInput{
		Enabled:                   true,
		Currency:                  "USD",
		PurchaseEarnPointsPerUnit: 1,
		ExchangeRatePoints:        100,
		MinRedeemPoints:           100,
		MaxValuePerDayCents:       10000,
		CardExpiryDays:            365,
		CheckInStreakIntervalDays: 1,
		RedeemOptions: []LoyaltyProgramOptionInput{{
			ValueCents:    1000,
			Currency:      "USD",
			StockQuantity: 2,
		}},
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
	assert.Equal(t, 10.0, analysis.Summary.MaxRedeemGiftCardValue)
	require.NotEmpty(t, analysis.Items)

	item := analysis.Items[0]
	assert.Equal(t, "zero_total", item.Kind)
	assert.Equal(t, "STACK80", item.CouponCode)
	assert.Equal(t, "active", item.CouponStatus)
	assert.InDelta(t, 100, item.EstimatedSubtotal, 0.001)
	assert.InDelta(t, 0, item.EstimatedPayableAmount, 0.001)
	assert.Contains(t, item.Factors, "fixed_coupon")
	assert.Contains(t, item.Factors, "member_level_discount")
	assert.Contains(t, item.Factors, "direct_points_discount")
}
