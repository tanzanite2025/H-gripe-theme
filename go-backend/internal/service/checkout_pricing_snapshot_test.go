package service

import (
	"testing"

	"commerce-platform/internal/domain/coupon"
	"commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/pricing"

	"github.com/stretchr/testify/require"
)

func TestBuildCheckoutPricingSnapshotAllocatesDiscountsByEligibleLine(t *testing.T) {
	variantA := uint(11)
	variantB := uint(22)
	snapshot, err := buildCheckoutPricingSnapshot(checkoutPricingInput{
		Items: []checkoutPricingLineInput{
			{ProductID: 1, VariantID: variantA, Quantity: 2, UnitPrice: money.MustNew(1000, "USD"), Subtotal: money.MustNew(2000, "USD")},
			{ProductID: 2, VariantID: variantB, Quantity: 1, UnitPrice: money.MustNew(1000, "USD"), Subtotal: money.MustNew(1000, "USD")},
		},
		Currency:       "USD",
		Subtotal:       money.MustNew(3000, "USD"),
		MemberDiscount: money.MustNew(300, "USD"),
		CouponDiscount: money.MustNew(400, "USD"),
		Coupon: &coupon.Coupon{
			Code:               "ONLY-ONE",
			ApplicableProducts: "[1]",
			ExcludedProducts:   "[2]",
		},
		PointsDiscount:      money.MustNew(200, "USD"),
		PointsToUse:         200,
		MerchandiseNetTotal: money.MustNew(2100, "USD"),
	})
	require.NoError(t, err)
	require.Equal(t, int64(3000), snapshot.BaseTotal().AmountMinor())
	require.Equal(t, int64(900), snapshot.DiscountTotal().AmountMinor())
	require.Equal(t, int64(2100), snapshot.NetTotal().AmountMinor())
	require.Equal(t, []int64{722, 178}, checkoutLineDiscountMinor(snapshot))
	require.Equal(t,
		[]pricing.DiscountKind{pricing.DiscountKindMember, pricing.DiscountKindCoupon, pricing.DiscountKindPoints},
		checkoutAllocationKinds(snapshot.Lines()[0].DiscountAllocations()),
	)
	require.Equal(t,
		[]pricing.DiscountKind{pricing.DiscountKindMember, pricing.DiscountKindPoints},
		checkoutAllocationKinds(snapshot.Lines()[1].DiscountAllocations()),
	)
}

func TestBuildCheckoutPricingSnapshotRejectsLegacyLineRoundingDrift(t *testing.T) {
	_, err := buildCheckoutPricingSnapshot(checkoutPricingInput{
		Items: []checkoutPricingLineInput{
			{ProductID: 1, Quantity: 3, UnitPrice: money.MustNew(333, "USD"), Subtotal: money.MustNew(1000, "USD")},
		},
		Currency:            "USD",
		Subtotal:            money.MustNew(1000, "USD"),
		MemberDiscount:      money.MustNew(0, "USD"),
		CouponDiscount:      money.MustNew(0, "USD"),
		PointsDiscount:      money.MustNew(0, "USD"),
		MerchandiseNetTotal: money.MustNew(1000, "USD"),
	})
	require.ErrorIs(t, err, ErrCheckoutPricingSnapshotInvalid)
	require.ErrorContains(t, err, "exact subtotal is 999 minor units")
}

func TestBuildCheckoutPricingSnapshotAcceptsExactFractionalUnitSubtotal(t *testing.T) {
	snapshot, err := buildCheckoutPricingSnapshot(checkoutPricingInput{
		Items: []checkoutPricingLineInput{
			{ProductID: 1, Quantity: 3, UnitPrice: money.MustNew(333, "USD"), Subtotal: money.MustNew(999, "USD")},
		},
		Currency:            "USD",
		Subtotal:            money.MustNew(999, "USD"),
		MemberDiscount:      money.MustNew(0, "USD"),
		CouponDiscount:      money.MustNew(0, "USD"),
		PointsDiscount:      money.MustNew(0, "USD"),
		MerchandiseNetTotal: money.MustNew(999, "USD"),
	})
	require.NoError(t, err)
	require.Equal(t, int64(999), snapshot.BaseTotal().AmountMinor())
	require.Equal(t, int64(999), snapshot.Lines()[0].BaseSubtotal().AmountMinor())
}

func TestBuildCheckoutPricingSnapshotUsesCurrencyMinorUnits(t *testing.T) {
	snapshot, err := buildCheckoutPricingSnapshot(checkoutPricingInput{
		Items: []checkoutPricingLineInput{
			{ProductID: 1, Quantity: 2, UnitPrice: money.MustNew(101, "JPY"), Subtotal: money.MustNew(202, "JPY")},
		},
		Currency:            "JPY",
		Subtotal:            money.MustNew(202, "JPY"),
		MemberDiscount:      money.MustNew(1, "JPY"),
		CouponDiscount:      money.MustNew(0, "JPY"),
		PointsDiscount:      money.MustNew(0, "JPY"),
		MerchandiseNetTotal: money.MustNew(201, "JPY"),
	})
	require.NoError(t, err)
	require.Equal(t, int64(201), snapshot.NetTotal().AmountMinor())
}

func TestBuildCheckoutPricingSnapshotRejectsMixedMoneyCurrency(t *testing.T) {
	_, err := buildCheckoutPricingSnapshot(checkoutPricingInput{
		Items: []checkoutPricingLineInput{
			{
				ProductID: 1,
				Quantity:  1,
				UnitPrice: money.MustNew(1000, "EUR"),
				Subtotal:  money.MustNew(1000, "EUR"),
			},
		},
		Currency:            "USD",
		Subtotal:            money.MustNew(1000, "USD"),
		MemberDiscount:      money.MustNew(0, "USD"),
		CouponDiscount:      money.MustNew(0, "USD"),
		PointsDiscount:      money.MustNew(0, "USD"),
		MerchandiseNetTotal: money.MustNew(1000, "USD"),
	})
	require.ErrorIs(t, err, ErrCheckoutPricingSnapshotInvalid)
}

func TestCouponQualifiedSubtotalMoneyPreservesLineMinorUnits(t *testing.T) {
	qualified, total, err := couponQualifiedSubtotalMoneyWithCategories(
		[]checkoutPricingLineInput{
			{ProductID: 1, Quantity: 3, UnitPrice: money.MustNew(333, "USD"), Subtotal: money.MustNew(999, "USD")},
			{ProductID: 2, Quantity: 1, UnitPrice: money.MustNew(101, "USD"), Subtotal: money.MustNew(101, "USD")},
		},
		map[uint]struct{}{1: {}},
		map[uint]struct{}{},
		couponCategoryValues{},
		"USD",
	)
	require.NoError(t, err)
	require.Equal(t, int64(999), qualified.AmountMinor())
	require.Equal(t, int64(1100), total.AmountMinor())
}

func TestCouponCategoryRestrictionMatchesIDOrSlug(t *testing.T) {
	categoryID := uint(7)
	values, err := parseCouponCategoryValues(`[7,"rings"]`)
	require.NoError(t, err)
	qualified, _, err := couponQualifiedSubtotalMoneyWithCategories([]checkoutPricingLineInput{
		{ProductID: 1, Quantity: 1, ProductCategoryID: &categoryID, UnitPrice: money.MustNew(1000, "USD"), Subtotal: money.MustNew(1000, "USD")},
		{ProductID: 2, Quantity: 1, ProductCategorySlug: "rings", UnitPrice: money.MustNew(500, "USD"), Subtotal: money.MustNew(500, "USD")},
		{ProductID: 3, Quantity: 1, ProductCategorySlug: "bikes", UnitPrice: money.MustNew(2000, "USD"), Subtotal: money.MustNew(2000, "USD")},
	}, nil, nil, values, "USD")
	require.NoError(t, err)
	major, err := qualified.MajorFloat()
	require.NoError(t, err)
	require.Equal(t, 15.0, major)
}

func checkoutLineDiscountMinor(snapshot pricing.Snapshot) []int64 {
	lines := snapshot.Lines()
	result := make([]int64, len(lines))
	for i, line := range lines {
		result[i] = line.DiscountTotal().AmountMinor()
	}
	return result
}

func checkoutAllocationKinds(allocations []pricing.DiscountAllocation) []pricing.DiscountKind {
	result := make([]pricing.DiscountKind, len(allocations))
	for i, allocation := range allocations {
		result[i] = allocation.Kind()
	}
	return result
}
