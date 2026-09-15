package service

import (
	"testing"
	"time"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/loyalty"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	shippingdomain "commerce-platform/internal/domain/shipping"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCheckoutCalculatePointsDiscountConvertsUSDValueToOrderCurrency(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&loyalty.UserLoyalty{}))
	require.NoError(t, db.Create(&loyalty.UserLoyalty{UserID: 7, AvailablePoints: 100, TotalPoints: 100}).Error)

	checkoutService := &CheckoutService{}
	pointsToUse, discount, _, err := checkoutService.calculatePointsDiscountMoney(
		repository.NewLoyaltyRepository(db),
		7,
		100,
		domainmoney.MustNew(2000, "JPY"),
		&loyalty.ProgramConfig{ID: 3, Enabled: true, Currency: "USD", ExchangeRatePoints: 100},
		currency.OrderFXSnapshot{
			Version:         currency.OrderFXSnapshotVersion,
			BaseCurrency:    "USD",
			OrderCurrency:   "JPY",
			BaseToOrderRate: 150,
			Source:          "test",
			CapturedAt:      time.Now().UTC(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, 100, pointsToUse)
	assert.Equal(t, int64(150), discount.AmountMinor())
}

func TestCheckoutCalculatePointsDiscountCapsAfterFXConversion(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&loyalty.UserLoyalty{}))
	require.NoError(t, db.Create(&loyalty.UserLoyalty{UserID: 7, AvailablePoints: 100000, TotalPoints: 100000}).Error)

	_, discount, _, err := (&CheckoutService{}).calculatePointsDiscountMoney(
		repository.NewLoyaltyRepository(db),
		7,
		100000,
		domainmoney.MustNew(2000, "JPY"),
		&loyalty.ProgramConfig{ID: 3, Enabled: true, Currency: "USD", ExchangeRatePoints: 100},
		currency.OrderFXSnapshot{
			Version:         currency.OrderFXSnapshotVersion,
			BaseCurrency:    "USD",
			OrderCurrency:   "JPY",
			BaseToOrderRate: 150,
			Source:          "test",
			CapturedAt:      time.Now().UTC(),
		},
	)

	require.NoError(t, err)
	assert.LessOrEqual(t, discount.AmountMinor(), int64(1000))
	assert.Equal(t, int64(999), discount.AmountMinor())
}

func TestCheckoutCalculatePointsDiscountUsesConfiguredPointsCurrency(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&loyalty.UserLoyalty{}))
	require.NoError(t, db.Create(&loyalty.UserLoyalty{UserID: 7, AvailablePoints: 100, TotalPoints: 100}).Error)

	_, discount, _, err := (&CheckoutService{}).calculatePointsDiscountMoney(
		repository.NewLoyaltyRepository(db),
		7,
		100,
		domainmoney.MustNew(10000, "USD"),
		&loyalty.ProgramConfig{ID: 3, Enabled: true, Currency: "JPY", ExchangeRatePoints: 100},
		currency.OrderFXSnapshot{
			Version:         currency.OrderFXSnapshotVersion,
			BaseCurrency:    "JPY",
			OrderCurrency:   "USD",
			BaseToOrderRate: 0.0067,
			Source:          "test",
			CapturedAt:      time.Now().UTC(),
		},
	)

	require.NoError(t, err)
	assert.Equal(t, int64(1), discount.AmountMinor())
}

func TestCheckoutCalculatePointsDiscountMoneyRoundsOnlyAtOrderCurrencyBoundary(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&loyalty.UserLoyalty{}))
	require.NoError(t, db.Create(&loyalty.UserLoyalty{UserID: 7, AvailablePoints: 100, TotalPoints: 100}).Error)

	subtotal := domainmoney.MustNew(10000, "USD")
	pointsToUse, discount, _, err := (&CheckoutService{}).calculatePointsDiscountMoney(
		repository.NewLoyaltyRepository(db),
		7,
		100,
		subtotal,
		&loyalty.ProgramConfig{ID: 3, Enabled: true, Currency: "JPY", ExchangeRatePoints: 100},
		currency.OrderFXSnapshot{
			Version:         currency.OrderFXSnapshotVersion,
			BaseCurrency:    "JPY",
			OrderCurrency:   "USD",
			BaseToOrderRate: 0.0067,
			Source:          "test",
			CapturedAt:      time.Now().UTC(),
		},
	)
	require.NoError(t, err)
	require.Equal(t, 100, pointsToUse)
	require.Equal(t, int64(1), discount.AmountMinor())
}

func TestCheckoutMemberDiscountUsesExactMinorUnitRounding(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&loyalty.MemberLevel{}, &loyalty.UserLoyalty{}))
	require.NoError(t, db.Create(&loyalty.MemberLevel{
		Name:         "Precision",
		MinPoints:    0,
		MaxPoints:    999999,
		DiscountRate: 5.5,
	}).Error)
	require.NoError(t, db.Create(&loyalty.UserLoyalty{
		UserID:          7,
		TotalPoints:     100,
		AvailablePoints: 100,
	}).Error)

	discount, err := (&CheckoutService{}).calculateMemberDiscountMoney(
		repository.NewLoyaltyRepository(db),
		7,
		domainmoney.MustNew(333, "USD"),
	)
	require.NoError(t, err)
	assert.Equal(t, int64(18), discount.AmountMinor())
}

func TestCheckoutQuoteAppliesMerchandiseDiscountsBeforeShippingAndTax(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 1000, 5)
	require.NotNil(t, productRecord.ShippingTemplateID)
	require.NoError(t, db.Model(&shippingdomain.ShippingTemplate{}).
		Where("id = ?", *productRecord.ShippingTemplateID).
		Update("free_shipping", false).Error)
	seedUserLoyalty(t, db, 42, 100000)
	seedCoupon(t, db, "WATERFALL-900", "fixed", 900, 0)
	require.NoError(t, db.Create(&paymentdomain.TaxRate{
		Name:       "California sales tax",
		Country:    "US",
		State:      "CA",
		PostalCode: "90001",
		Rate:       10,
		Enabled:    true,
	}).Error)

	quote, err := orderService.checkout.Quote(CheckoutQuoteInput{
		UserID:          42,
		Items:           []order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		ShippingAddress: testAddress(),
		CouponCode:      "WATERFALL-900",
		PointsToUse:     100000,
	})
	require.NoError(t, err)

	assert.Equal(t, 50.0, quote.MemberDiscount)
	assert.Equal(t, 900.0, quote.CouponDiscount)
	assert.Equal(t, 25.0, quote.PointsDiscount)
	assert.Equal(t, 2500, quote.PointsToUse)
	assert.Equal(t, 10.0, quote.ShippingFee)
	assert.Equal(t, 2.5, quote.TaxAmount)
	assert.Equal(t, 37.5, quote.TotalAmount)
	assert.LessOrEqual(t, quote.MemberDiscount+quote.CouponDiscount+quote.PointsDiscount, quote.SubtotalAmount)
	assert.Equal(t, int64(100000), quote.PricingSnapshot.BaseTotal().AmountMinor())
	assert.Equal(t, int64(97500), quote.PricingSnapshot.DiscountTotal().AmountMinor())
	assert.Equal(t, int64(2500), quote.PricingSnapshot.NetTotal().AmountMinor())
	lineTaxes := quote.PricingSnapshot.Lines()
	require.Len(t, lineTaxes, 1)
	assert.Equal(t, int64(250), lineTaxes[0].Tax().AmountMinor())
	assert.Equal(t, int64(100000), quote.SubtotalMinor)
	assert.Equal(t, int64(1000), quote.ShippingFeeMinor)
	assert.Equal(t, int64(250), quote.TaxMinor)
	assert.Equal(t, int64(5000), quote.MemberDiscountMinor)
	assert.Equal(t, int64(90000), quote.CouponDiscountMinor)
	assert.Equal(t, int64(2500), quote.PointsDiscountMinor)
	assert.Equal(t, int64(97500), quote.DiscountMinor)
	assert.Equal(t, int64(3750), quote.TotalMinor)
	assert.Equal(t, int64(3750), quote.PaymentAmountMinor)
}

func TestCheckoutCalculateTaxReturnsZeroWhenLocationHasNoTaxRule(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&paymentdomain.TaxRate{}))

	checkoutService := &CheckoutService{}
	taxMoney, err := checkoutService.calculateTaxMoney(
		repository.NewPaymentRepository(db),
		domainmoney.MustNew(10000, "USD"),
		"US",
		"CA",
		"90210",
		"USD",
	)

	require.NoError(t, err)
	assert.Zero(t, taxMoney.AmountMinor())
}

func TestCheckoutCalculateTaxPropagatesTaxRateLookupFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	require.NoError(t, sqlDB.Close())

	checkoutService := &CheckoutService{}
	taxMoney, err := checkoutService.calculateTaxMoney(
		repository.NewPaymentRepository(db),
		domainmoney.MustNew(10000, "USD"),
		"US",
		"CA",
		"90210",
		"USD",
	)

	assert.Zero(t, taxMoney.AmountMinor())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load tax rate")
}

func TestCheckoutCalculateTaxRoundsUsingCurrencyMinorUnits(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&paymentdomain.TaxRate{}))
	require.NoError(t, db.Create(&paymentdomain.TaxRate{
		Name:       "California VAT",
		Country:    "US",
		State:      "CA",
		PostalCode: "90210",
		Rate:       20,
		Enabled:    true,
	}).Error)

	checkoutService := &CheckoutService{}
	usdTax, err := checkoutService.calculateTaxMoney(
		repository.NewPaymentRepository(db),
		domainmoney.MustNew(142405, "USD"),
		"US",
		"CA",
		"90210",
		"USD",
	)
	require.NoError(t, err)
	assert.Equal(t, int64(28481), usdTax.AmountMinor())

	assert.Equal(t, 123.46, roundMoney(123.456, "USD"))
	assert.Equal(t, 123.0, roundMoney(123.456, "JPY"))
}
