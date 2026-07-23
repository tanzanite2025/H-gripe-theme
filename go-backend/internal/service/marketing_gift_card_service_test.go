package service

import (
	"errors"
	"testing"

	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/loyalty"
	"tanzanite/internal/domain/order"
	"tanzanite/internal/domain/payment"
	"tanzanite/internal/domain/product"
	"tanzanite/internal/repository"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMarketingServiceRejectsUnsafeGiftCardStatusTransitions(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	card := coupon.GiftCard{
		Code:         "GC-ACTIVE",
		InitialValue: 50,
		Balance:      50,
		Currency:     "USD",
		Status:       "active",
	}
	require.NoError(t, db.Create(&card).Error)

	_, err := marketingService.UpdateGiftCardStatus(card.ID, "used")

	require.ErrorIs(t, err, ErrInvalidGiftCardStatusTransition)
	var saved coupon.GiftCard
	require.NoError(t, db.First(&saved, card.ID).Error)
	assert.Equal(t, "active", saved.Status)
}

func TestMarketingServiceAllowsTerminalGiftCardTransitionFromActive(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	card := coupon.GiftCard{
		Code:         "GC-CANCEL",
		InitialValue: 50,
		Balance:      50,
		Currency:     "USD",
		Status:       "active",
	}
	require.NoError(t, db.Create(&card).Error)

	updated, err := marketingService.UpdateGiftCardStatus(card.ID, "cancelled")

	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "cancelled", updated.Status)
}

func TestMarketingServiceRejectsOverlappingMemberLevels(t *testing.T) {
	db, marketingService := newTestMarketingService(t)
	require.NoError(t, db.Create(&loyalty.MemberLevel{
		Name:             "Silver",
		MinPoints:        1000,
		MaxPoints:        4999,
		DiscountRate:     5,
		PointsMultiplier: 1,
	}).Error)

	_, err := marketingService.CreateMemberLevelAdmin(MemberLevelCreateInput{
		Name:             "Overlap",
		MinPoints:        4000,
		MaxPoints:        8000,
		DiscountRate:     10,
		PointsMultiplier: 1,
	})

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidMemberLevel))
}

func newTestMarketingService(t *testing.T) (*gorm.DB, *MarketingService) {
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
		&coupon.Coupon{},
		&coupon.CouponUsage{},
		&coupon.GiftCard{},
		&coupon.GiftCardTransaction{},
		&loyalty.LoyaltyTransaction{},
		&loyalty.MemberLevel{},
		&loyalty.UserLoyalty{},
		&product.Product{},
		&product.ProductVariant{},
		&order.Order{},
		&order.OrderItem{},
		&payment.PaymentMethod{},
		&payment.TaxRate{},
	))

	orderRepo := repository.NewOrderRepository(db)
	productRepo := repository.NewProductRepository(db)
	couponRepo := repository.NewCouponRepository(db)
	loyaltyRepo := repository.NewLoyaltyRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)
	txManager := repository.NewTxManager(db, orderRepo, productRepo, couponRepo, loyaltyRepo, paymentRepo)

	return db, NewMarketingService(txManager, couponRepo, loyaltyRepo)
}
