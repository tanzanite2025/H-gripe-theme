package service

import (
	"testing"

	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/pricing"

	"github.com/stretchr/testify/require"
)

func TestBuildRefundLineItemSnapshotUsesPersistedPricingAllocation(t *testing.T) {
	line := persistedPricingLine(t, 3, 100, 100)
	item := order.OrderItem{
		ID:                  1,
		OrderID:             2,
		ProductID:           3,
		Quantity:            3,
		Currency:            "USD",
		PriceMinor:          100,
		SubtotalMinor:       300,
		DiscountMinor:       100,
		TaxAmountMinor:      30,
		TotalMinor:          230,
		Price:               1,
		Subtotal:            3,
		Discount:            1,
		TaxAmount:           0.30,
		Total:               2.30,
		PricingSnapshotData: line,
	}

	refunded, allocation, err := buildRefundLineItemSnapshot(item, 0, 1, false, "USD")
	require.NoError(t, err)
	require.Equal(t, 1.0, refunded.LineSubtotalAmount)
	require.Equal(t, 0.33, refunded.LineDiscountAmount)
	require.Equal(t, 0.10, refunded.LineTaxAmount)
	require.Equal(t, 0.77, refunded.LineTotalAmount)
	require.Equal(t, int64(100), allocation.LineSubtotalAmount.AmountMinor())
	require.Equal(t, int64(33), allocation.LineDiscountAmount.AmountMinor())
	require.Equal(t, int64(10), allocation.LineTaxAmount.AmountMinor())
	require.Equal(t, int64(77), allocation.LineTotalAmount.AmountMinor())
}

func TestBuildRefundLineItemSnapshotPreservesRemaindersAcrossPartialRefunds(t *testing.T) {
	item := order.OrderItem{
		ID:                  1,
		OrderID:             2,
		ProductID:           3,
		Quantity:            3,
		Currency:            "USD",
		PriceMinor:          100,
		SubtotalMinor:       300,
		DiscountMinor:       100,
		TaxAmountMinor:      1,
		TotalMinor:          201,
		Price:               1,
		Subtotal:            3,
		Discount:            1,
		TaxAmount:           0.01,
		Total:               2.01,
		PricingSnapshotData: persistedPricingLine(t, 3, 100, 100),
	}

	discountMinor := int64(0)
	taxMinor := int64(0)
	totalMinor := int64(0)
	for alreadyRefunded := 0; alreadyRefunded < item.Quantity; alreadyRefunded++ {
		_, allocation, err := buildRefundLineItemSnapshot(item, alreadyRefunded, 1, false, "USD")
		require.NoError(t, err)
		discountMinor += allocation.LineDiscountAmount.AmountMinor()
		taxMinor += allocation.LineTaxAmount.AmountMinor()
		totalMinor += allocation.LineTotalAmount.AmountMinor()
	}

	require.Equal(t, int64(100), discountMinor)
	require.Equal(t, int64(1), taxMinor)
	require.Equal(t, int64(201), totalMinor)
}

func TestBuildRefundLineItemSnapshotRejectsMissingPricingSnapshot(t *testing.T) {
	item := order.OrderItem{Quantity: 2, Price: 1.25, Subtotal: 2.5, Discount: 0.25, Total: 2.25}
	_, _, err := buildRefundLineItemSnapshot(item, 0, 1, false, "USD")
	require.ErrorIs(t, err, errInvalidOrderItemPricingSnapshot)
}

func TestBuildRefundLineItemSnapshotRejectsCorruptPersistedSnapshot(t *testing.T) {
	item := order.OrderItem{Quantity: 1, Subtotal: 1, Total: 1, PricingSnapshotData: []byte(`{"schema_version":1,"key":"broken"}`)}
	_, _, err := buildRefundLineItemSnapshot(item, 0, 1, false, "USD")
	require.ErrorIs(t, err, errInvalidOrderItemPricingSnapshot)
}

func TestBuildRefundLineItemSnapshotRejectsMismatchedPersistedIdentity(t *testing.T) {
	line := persistedPricingLine(t, 1, 100, 0)
	item := order.OrderItem{ProductID: 99, Quantity: 1, PricingSnapshotData: line}
	_, _, err := buildRefundLineItemSnapshot(item, 0, 1, false, "USD")
	require.ErrorIs(t, err, errInvalidOrderItemPricingSnapshot)
}

func TestRefundAmountsEqualInMinorUnitsUsesCurrencyPrecision(t *testing.T) {
	equal, err := refundAmountsEqualInMinorUnits(mustTestMoney(t, 1.00, "USD"), mustTestMoney(t, 1.004, "USD"))
	require.NoError(t, err)
	require.True(t, equal)

	equal, err = refundAmountsEqualInMinorUnits(mustTestMoney(t, 1.00, "USD"), mustTestMoney(t, 1.01, "USD"))
	require.NoError(t, err)
	require.False(t, equal)

	equal, err = refundAmountsEqualInMinorUnits(mustTestMoney(t, 100, "JPY"), mustTestMoney(t, 101, "JPY"))
	require.NoError(t, err)
	require.False(t, equal)
}

func TestRoundRefundMoneyUsesExplicitCurrencyPrecision(t *testing.T) {
	require.Equal(t, 1.0, roundRefundMoney(1.49, "JPY"))
	require.Equal(t, 1.49, roundRefundMoney(1.494, "USD"))
}

func TestReadOrderPricingRefundBaselinePrefersPersistedMinorUnits(t *testing.T) {
	raw, err := pricing.MarshalOrderPricingSnapshot(pricing.OrderPricingSnapshotInput{
		Currency:         "USD",
		Subtotal:         money.MustNew(12345, "USD"),
		Shipping:         money.MustNew(0, "USD"),
		Tax:              money.MustNew(0, "USD"),
		MemberDiscount:   money.MustNew(0, "USD"),
		CouponDiscount:   money.MustNew(2345, "USD"),
		PointsDiscount:   money.MustNew(0, "USD"),
		GiftCardDiscount: money.MustNew(0, "USD"),
		DiscountTotal:    money.MustNew(2345, "USD"),
		Total:            money.MustNew(10000, "USD"),
	})
	require.NoError(t, err)
	o := &order.Order{Currency: "USD", SubtotalAmount: 999, PricingSnapshotData: raw}
	subtotal, couponDiscount, present, err := readOrderPricingRefundBaseline(o)
	require.NoError(t, err)
	require.True(t, present)
	require.Equal(t, int64(12345), subtotal.AmountMinor())
	require.Equal(t, int64(2345), couponDiscount.AmountMinor())
}

func TestReadOrderPricingRefundBaselineRejectsCorruptOrMismatchedSnapshot(t *testing.T) {
	_, _, _, err := readOrderPricingRefundBaseline(&order.Order{Currency: "USD", PricingSnapshotData: []byte(`{"schema_version":1}`)})
	require.ErrorIs(t, err, errInvalidOrderPricingSnapshot)

	raw, err := pricing.MarshalOrderPricingSnapshot(pricing.OrderPricingSnapshotInput{
		Currency:         "EUR",
		Subtotal:         money.MustNew(100, "EUR"),
		Shipping:         money.MustNew(0, "EUR"),
		Tax:              money.MustNew(0, "EUR"),
		MemberDiscount:   money.MustNew(0, "EUR"),
		CouponDiscount:   money.MustNew(0, "EUR"),
		PointsDiscount:   money.MustNew(0, "EUR"),
		GiftCardDiscount: money.MustNew(0, "EUR"),
		DiscountTotal:    money.MustNew(0, "EUR"),
		Total:            money.MustNew(100, "EUR"),
	})
	require.NoError(t, err)
	_, _, _, err = readOrderPricingRefundBaseline(&order.Order{Currency: "USD", PricingSnapshotData: raw})
	require.ErrorIs(t, err, errInvalidOrderPricingSnapshot)
}

func TestApplyRefundPromotionClawbackUsesZeroDecimalCurrency(t *testing.T) {
	o := &order.Order{Currency: "JPY", SubtotalAmount: 1200}
	couponRecord := &coupon.Coupon{ID: 7, Code: "JPY200", Type: "fixed", Value: 200, MinAmount: 1000}
	usage := &coupon.CouponUsage{CouponID: couponRecord.ID, Discount: 200}

	adjustment, err := applyRefundPromotionClawback(
		o,
		couponRecord,
		usage,
		money.MustNew(300, "JPY"),
		money.MustNew(300, "JPY"),
		money.MustNew(1200, "JPY"),
		money.MustNew(200, "JPY"),
		money.MustNew(0, "JPY"),
		money.MustNew(0, "JPY"),
	)
	require.NoError(t, err)
	require.Equal(t, int64(300), adjustment.RequestedAmount.AmountMinor())
	require.Equal(t, int64(200), adjustment.DiscountClawbackAmount.AmountMinor())
	require.Equal(t, int64(100), adjustment.NetAmount.AmountMinor())
}

func TestProportionalRefundPointsUsesMinorUnitIntegerRatio(t *testing.T) {
	require.Equal(t, 33, proportionalRefundPoints(
		100,
		money.MustNew(100, "USD"),
		money.MustNew(300, "USD"),
	))
	require.Equal(t, 3, proportionalRefundPoints(
		10,
		money.MustNew(1, "JPY"),
		money.MustNew(3, "JPY"),
	))
	require.Equal(t, 100, proportionalRefundPoints(
		100,
		money.MustNew(300, "USD"),
		money.MustNew(300, "USD"),
	))
	require.Equal(t, 0, proportionalRefundPoints(
		100,
		money.MustNew(100, "USD"),
		money.MustNew(0, "USD"),
	))
	require.Equal(t, 0, proportionalRefundPoints(
		100,
		money.MustNew(100, "USD"),
		money.MustNew(100, "EUR"),
	))
}

func persistedPricingLine(t *testing.T, quantity int, unitPriceMinor, discountMinor int64) []byte {
	t.Helper()
	snapshot, err := pricing.NewSnapshot([]pricing.LineInput{{
		Key: "0:3:0", ProductID: 3, Quantity: quantity, UnitPrice: money.MustNew(unitPriceMinor, "USD"),
	}})
	require.NoError(t, err)
	snapshot, err = snapshot.AllocateDiscount(pricing.DiscountInput{
		Kind: pricing.DiscountKindCoupon, Amount: money.MustNew(discountMinor, "USD"),
	})
	require.NoError(t, err)
	raw, err := pricing.MarshalLineSnapshot(snapshot.Lines()[0])
	require.NoError(t, err)
	return raw
}

func refundPricingLineForTest(t *testing.T, productID uint, variantID *uint, quantity int, unitPrice, discount float64, currencyCode string) []byte {
	t.Helper()
	unitPriceMoney, err := money.FromMajorFloat(unitPrice, currencyCode)
	require.NoError(t, err)
	resolvedVariantID := uint(0)
	if variantID != nil {
		resolvedVariantID = *variantID
	}
	snapshot, err := pricing.NewSnapshot([]pricing.LineInput{{
		Key:       "test-refund-line",
		ProductID: productID,
		VariantID: resolvedVariantID,
		Quantity:  quantity,
		UnitPrice: unitPriceMoney,
	}})
	require.NoError(t, err)
	if discount > 0 {
		discountMoney, moneyErr := money.FromMajorFloat(discount, currencyCode)
		require.NoError(t, moneyErr)
		snapshot, err = snapshot.AllocateDiscount(pricing.DiscountInput{
			Kind:      pricing.DiscountKindCoupon,
			Reference: "test-coupon",
			Amount:    discountMoney,
		})
		require.NoError(t, err)
	}
	raw, err := pricing.MarshalLineSnapshot(snapshot.Lines()[0])
	require.NoError(t, err)
	return raw
}
