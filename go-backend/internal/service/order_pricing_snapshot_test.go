package service

import (
	"testing"

	"commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/pricing"

	"github.com/stretchr/testify/require"
)

func TestAttachPricingSnapshotsToOrderItemsCopiesValidatedLineJSON(t *testing.T) {
	variantID := uint(20)
	snapshot, err := pricing.NewSnapshot([]pricing.LineInput{{
		Key: "0:10:20", ProductID: 10, VariantID: variantID, Quantity: 2, UnitPrice: money.MustNew(125, "USD"),
	}})
	require.NoError(t, err)
	snapshot, err = snapshot.AllocateDiscount(pricing.DiscountInput{Kind: pricing.DiscountKindCoupon, Amount: money.MustNew(25, "USD")})
	require.NoError(t, err)
	snapshot, err = snapshot.AllocateTax(money.MustNew(10, "USD"))
	require.NoError(t, err)
	items, err := attachPricingSnapshotsToOrderItems([]order.OrderItem{{
		ProductID: 10, VariantID: &variantID, Quantity: 2,
	}}, snapshot)
	require.NoError(t, err)
	require.NotEqual(t, "{}", string(items[0].PricingSnapshotData))
	parsed, err := pricing.ParseLineSnapshot(items[0].PricingSnapshotData)
	require.NoError(t, err)
	require.Equal(t, int64(250), parsed.BaseSubtotal().AmountMinor())
	require.Equal(t, int64(25), items[0].DiscountMinor)
	require.Equal(t, int64(10), items[0].TaxAmountMinor)
	require.Equal(t, int64(235), items[0].TotalMinor)
}

func TestAttachPricingSnapshotsToOrderItemsRejectsIdentityMismatch(t *testing.T) {
	snapshot, err := pricing.NewSnapshot([]pricing.LineInput{{
		Key: "0:10:20", ProductID: 10, VariantID: 20, Quantity: 2, UnitPrice: money.MustNew(125, "USD"),
	}})
	require.NoError(t, err)
	_, err = attachPricingSnapshotsToOrderItems([]order.OrderItem{{ProductID: 99, Quantity: 2}}, snapshot)
	require.ErrorIs(t, err, ErrOrderPricingSnapshotMismatch)
}
