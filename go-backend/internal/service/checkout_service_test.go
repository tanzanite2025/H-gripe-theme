package service

import (
	"context"
	"testing"

	"commerce-platform/internal/domain/loyalty"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	shippingdomain "commerce-platform/internal/domain/shipping"

	taxratedomain "commerce-platform/internal/domain/taxrate"
	"commerce-platform/internal/repository"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCheckoutMemberDiscountUsesExactMinorUnitRounding(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&loyalty.MemberLevel{}, &loyalty.UserLoyalty{}))
	require.NoError(t, db.Create(&loyalty.MemberLevel{
		Name:                "Precision",
		MinPoints:           0,
		MaxPoints:           999999,
		DiscountRateDecimal: "5.5",
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
	require.NoError(t, db.Create(&taxratedomain.TaxRate{
		Name:        "California sales tax",
		Country:     "US",
		State:       "CA",
		PostalCode:  "90001",
		RateDecimal: "10",
		Enabled:     true,
	}).Error)

	quote, err := orderService.checkout.Quote(CheckoutQuoteInput{
		UserID:          42,
		Items:           []order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		ShippingAddress: testAddress(),
		CouponCode:      "WATERFALL-900",
	})
	require.NoError(t, err)

	assert.Equal(t, int64(5000), quote.MemberDiscountMinor)
	assert.Equal(t, int64(90000), quote.CouponDiscountMinor)
	assert.Equal(t, int64(1000), quote.ShippingFeeMinor)
	assert.Equal(t, int64(500), quote.TaxMinor)
	assert.Equal(t, int64(6500), quote.TotalMinor)
	assert.LessOrEqual(t, quote.MemberDiscountMinor+quote.CouponDiscountMinor, quote.SubtotalMinor)
	assert.Equal(t, int64(100000), quote.PricingSnapshot.BaseTotal().AmountMinor())
	assert.Equal(t, int64(95000), quote.PricingSnapshot.DiscountTotal().AmountMinor())
	assert.Equal(t, int64(5000), quote.PricingSnapshot.NetTotal().AmountMinor())
	lineTaxes := quote.PricingSnapshot.Lines()
	require.Len(t, lineTaxes, 1)
	assert.Equal(t, int64(500), lineTaxes[0].Tax().AmountMinor())
	assert.Equal(t, int64(100000), quote.SubtotalMinor)
	assert.Equal(t, int64(1000), quote.ShippingFeeMinor)
	assert.Equal(t, int64(500), quote.TaxMinor)
	assert.Equal(t, int64(5000), quote.MemberDiscountMinor)
	assert.Equal(t, int64(90000), quote.CouponDiscountMinor)
	assert.Equal(t, int64(95000), quote.DiscountMinor)
	assert.Equal(t, int64(6500), quote.TotalMinor)
	assert.Equal(t, int64(6500), quote.PaymentAmountMinor)

	var lineBase, lineDiscount, lineNet, lineTax int64
	for _, line := range quote.PricingSnapshot.Lines() {
		lineBase += line.BaseSubtotal().AmountMinor()
		lineDiscount += line.DiscountTotal().AmountMinor()
		lineNet += line.NetSubtotal().AmountMinor()
		lineTax += line.Tax().AmountMinor()
		assert.Equal(t, line.BaseSubtotal().AmountMinor()-line.DiscountTotal().AmountMinor(), line.NetSubtotal().AmountMinor())
	}
	assert.Equal(t, quote.PricingSnapshot.BaseTotal().AmountMinor(), lineBase)
	assert.Equal(t, quote.PricingSnapshot.DiscountTotal().AmountMinor(), lineDiscount)
	assert.Equal(t, quote.PricingSnapshot.NetTotal().AmountMinor(), lineNet)
	assert.Equal(t, quote.PricingSnapshot.TaxTotal().AmountMinor(), lineTax)
	assert.Equal(t, quote.PricingSnapshot.NetTotal().AmountMinor()+quote.PricingSnapshot.TaxTotal().AmountMinor(), quote.TotalMinor-quote.ShippingFeeMinor)
}

func TestCheckoutAndOrderCreationRejectAddressWithoutConfiguredTaxRule(t *testing.T) {
	db, orderService := newTestOrderService(t)
	productRecord := seedProduct(t, db, 1000, 5)
	unsupportedTaxAddress := testAddress()
	unsupportedTaxAddress.State = "NY"
	unsupportedTaxAddress.PostalCode = "10001"

	_, err := orderService.checkout.Quote(CheckoutQuoteInput{
		UserID:          42,
		Items:           []order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		ShippingAddress: unsupportedTaxAddress,
	})
	require.ErrorIs(t, err, ErrTaxRateUnavailable)

	_, err = orderService.CreateOrder(
		context.Background(),
		42,
		[]order.OrderItem{{ProductID: productRecord.ID, Quantity: 1}},
		unsupportedTaxAddress,
		unsupportedTaxAddress,
		"card",
		"standard",
		"",
	)
	require.ErrorIs(t, err, ErrTaxRateUnavailable)

	var createdOrderCount int64
	require.NoError(t, db.Model(&order.Order{}).Count(&createdOrderCount).Error)
	assert.Zero(t, createdOrderCount)
}

func TestCheckoutCalculateTaxRejectsLocationWithoutTaxRule(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	require.NoError(t, db.AutoMigrate(&taxratedomain.TaxRate{}))

	taxRateRepository := repository.NewTaxRateRepository(db)
	checkoutService := &CheckoutService{taxRateService: NewTaxRateService(taxRateRepository)}
	taxMoney, err := checkoutService.calculateTaxMoney(
		taxRateRepository,
		domainmoney.MustNew(10000, "USD"),
		"US",
		"CA",
		"90210",
		"USD",
	)

	require.ErrorIs(t, err, ErrTaxRateUnavailable)
	assert.Zero(t, taxMoney.AmountMinor())
}

func TestCheckoutCalculateTaxPropagatesTaxRateLookupFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	sqlDB, err := db.DB()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })
	require.NoError(t, db.AutoMigrate(&taxratedomain.TaxRate{}))
	require.NoError(t, db.Migrator().DropTable(&taxratedomain.TaxRate{}))

	taxRateRepository := repository.NewTaxRateRepository(db)
	checkoutService := &CheckoutService{taxRateService: NewTaxRateService(taxRateRepository)}
	taxMoney, err := checkoutService.calculateTaxMoney(
		taxRateRepository,
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
	require.NoError(t, db.AutoMigrate(&taxratedomain.TaxRate{}))
	require.NoError(t, db.Create(&taxratedomain.TaxRate{
		Name:        "California VAT",
		Country:     "US",
		State:       "CA",
		PostalCode:  "90210",
		RateDecimal: "20",
		Enabled:     true,
	}).Error)

	taxRateRepository := repository.NewTaxRateRepository(db)
	checkoutService := &CheckoutService{taxRateService: NewTaxRateService(taxRateRepository)}
	usdTax, err := checkoutService.calculateTaxMoney(
		taxRateRepository,
		domainmoney.MustNew(142405, "USD"),
		"US",
		"CA",
		"90210",
		"USD",
	)
	require.NoError(t, err)
	assert.Equal(t, int64(28481), usdTax.AmountMinor())

	usdDisplay, err := domainmoney.FromMajorFloat(123.456, "USD")
	require.NoError(t, err)
	assert.Equal(t, "123.46", mustFormatTestMoney(usdDisplay))
	jpyDisplay, err := domainmoney.FromMajorFloat(123.456, "JPY")
	require.NoError(t, err)
	assert.Equal(t, "123", mustFormatTestMoney(jpyDisplay))
}

func mustFormatTestMoney(value domainmoney.Money) string {
	formatted, err := value.FormatMajor()
	if err != nil {
		panic(err)
	}
	return formatted
}
