package service

import (
	"errors"
	"fmt"

	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	domainpricing "commerce-platform/internal/domain/pricing"

	"gorm.io/datatypes"
)

var ErrOrderPricingSnapshotMismatch = errors.New("order pricing snapshot does not match order lines")

func attachPricingSnapshotsToOrderItems(items []order.OrderItem, snapshot domainpricing.Snapshot) ([]order.OrderItem, error) {
	lines := snapshot.Lines()
	if len(items) != len(lines) {
		return nil, fmt.Errorf("%w: item count %d, snapshot line count %d", ErrOrderPricingSnapshotMismatch, len(items), len(lines))
	}
	result := make([]order.OrderItem, len(items))
	copy(result, items)
	for i := range result {
		if result[i].ProductID != lines[i].ProductID() || orderItemVariantID(result[i]) != lines[i].VariantID() || result[i].Quantity != lines[i].Quantity() {
			return nil, fmt.Errorf("%w: line %d identity differs", ErrOrderPricingSnapshotMismatch, i)
		}
		encoded, err := domainpricing.MarshalLineSnapshot(lines[i])
		if err != nil {
			return nil, fmt.Errorf("encode line %d pricing snapshot: %w", i, err)
		}
		result[i].PricingSnapshotData = datatypes.JSON(encoded)
		result[i].Currency = lines[i].UnitPrice().Currency().String()
		result[i].DiscountMinor = lines[i].DiscountTotal().AmountMinor()
		result[i].TaxAmountMinor = lines[i].Tax().AmountMinor()
		netWithTax, err := lines[i].NetSubtotal().Add(lines[i].Tax())
		if err != nil {
			return nil, fmt.Errorf("calculate line %d total: %w", i, err)
		}
		result[i].TotalMinor = netWithTax.AmountMinor()
		if len(result[i].ConfigurationData) > 0 && string(result[i].ConfigurationData) != "{}" {
			result[i].ConfigurationSnapshotData = append([]byte(nil), result[i].ConfigurationData...)
		}
	}
	return result, nil
}

func marshalOrderPricingSnapshot(quote *CheckoutQuote) (datatypes.JSON, error) {
	if quote == nil {
		return nil, errors.New("checkout quote is required")
	}
	fromMinor := func(value int64) (domainmoney.Money, error) {
		return domainmoney.New(value, quote.Currency)
	}
	subtotal, err := fromMinor(quote.SubtotalMinor)
	if err != nil {
		return nil, err
	}
	shipping, err := fromMinor(quote.ShippingFeeMinor)
	if err != nil {
		return nil, err
	}
	tax, err := fromMinor(quote.TaxMinor)
	if err != nil {
		return nil, err
	}
	member, err := fromMinor(quote.MemberDiscountMinor)
	if err != nil {
		return nil, err
	}
	coupon, err := fromMinor(quote.CouponDiscountMinor)
	if err != nil {
		return nil, err
	}
	discount, err := fromMinor(quote.DiscountMinor)
	if err != nil {
		return nil, err
	}
	total, err := fromMinor(quote.TotalMinor)
	if err != nil {
		return nil, err
	}
	raw, err := domainpricing.MarshalOrderPricingSnapshot(domainpricing.OrderPricingSnapshotInput{
		Currency:       quote.Currency,
		Subtotal:       subtotal,
		Shipping:       shipping,
		Tax:            tax,
		MemberDiscount: member,
		CouponDiscount: coupon,
		DiscountTotal:  discount,
		Total:          total,
	})
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(raw), nil
}
