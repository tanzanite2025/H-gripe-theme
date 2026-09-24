package pricing

import (
	"encoding/json"
	"errors"
	"fmt"

	"commerce-platform/internal/domain/money"
)

const OrderSnapshotSchemaVersion = 1

var ErrInvalidOrderSnapshot = errors.New("invalid order pricing snapshot")

// OrderPricingSnapshotInput is the exact-money boundary used when an order is
// created. It mirrors the checkout quote without persisting float64 values.
type OrderPricingSnapshotInput struct {
	Currency       string
	Subtotal       money.Money
	Shipping       money.Money
	Tax            money.Money
	MemberDiscount money.Money
	CouponDiscount money.Money
	DiscountTotal  money.Money
	Total          money.Money
}

type OrderPricingSnapshotPayload struct {
	SchemaVersion       int    `json:"schema_version"`
	Currency            string `json:"currency"`
	SubtotalMinor       int64  `json:"subtotal_minor"`
	ShippingMinor       int64  `json:"shipping_minor"`
	TaxMinor            int64  `json:"tax_minor"`
	MemberDiscountMinor int64  `json:"member_discount_minor"`
	CouponDiscountMinor int64  `json:"coupon_discount_minor"`
	DiscountTotalMinor  int64  `json:"discount_total_minor"`
	TotalMinor          int64  `json:"total_minor"`
}

func MarshalOrderPricingSnapshot(input OrderPricingSnapshotInput) ([]byte, error) {
	payload, err := input.Payload()
	if err != nil {
		return nil, err
	}
	return json.Marshal(payload)
}

func (input OrderPricingSnapshotInput) Payload() (OrderPricingSnapshotPayload, error) {
	values := []money.Money{
		input.Subtotal, input.Shipping, input.Tax,
		input.MemberDiscount, input.CouponDiscount,
		input.DiscountTotal, input.Total,
	}
	for _, value := range values {
		if err := value.Validate(); err != nil {
			return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: invalid money: %v", ErrInvalidOrderSnapshot, err)
		}
		if value.AmountMinor() < 0 {
			return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: negative amount", ErrInvalidOrderSnapshot)
		}
		if value.Currency().String() != input.Currency {
			return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: currency mismatch", ErrInvalidOrderSnapshot)
		}
	}
	discountSum, err := input.MemberDiscount.Add(input.CouponDiscount)
	if err != nil {
		return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: discount total overflows", ErrInvalidOrderSnapshot)
	}
	if input.DiscountTotal.AmountMinor() != discountSum.AmountMinor() {
		return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: discount total is inconsistent", ErrInvalidOrderSnapshot)
	}
	gross, err := input.Subtotal.Add(input.Shipping)
	if err != nil {
		return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: gross total overflows", ErrInvalidOrderSnapshot)
	}
	gross, err = gross.Add(input.Tax)
	if err != nil {
		return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: gross total overflows", ErrInvalidOrderSnapshot)
	}
	expected, err := gross.Subtract(input.DiscountTotal)
	if err != nil {
		return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: total calculation failed", ErrInvalidOrderSnapshot)
	}
	if expected.AmountMinor() < 0 {
		expected, err = money.New(0, input.Currency)
		if err != nil {
			return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: invalid zero total", ErrInvalidOrderSnapshot)
		}
	}
	if input.Total.AmountMinor() != expected.AmountMinor() {
		return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: total is inconsistent", ErrInvalidOrderSnapshot)
	}
	return OrderPricingSnapshotPayload{
		SchemaVersion:       OrderSnapshotSchemaVersion,
		Currency:            input.Currency,
		SubtotalMinor:       input.Subtotal.AmountMinor(),
		ShippingMinor:       input.Shipping.AmountMinor(),
		TaxMinor:            input.Tax.AmountMinor(),
		MemberDiscountMinor: input.MemberDiscount.AmountMinor(),
		CouponDiscountMinor: input.CouponDiscount.AmountMinor(),
		DiscountTotalMinor:  input.DiscountTotal.AmountMinor(),
		TotalMinor:          input.Total.AmountMinor(),
	}, nil
}

func ParseOrderPricingSnapshot(raw []byte) (OrderPricingSnapshotPayload, error) {
	if len(raw) == 0 {
		return OrderPricingSnapshotPayload{}, ErrInvalidOrderSnapshot
	}
	var payload OrderPricingSnapshotPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return OrderPricingSnapshotPayload{}, fmt.Errorf("%w: decode JSON: %v", ErrInvalidOrderSnapshot, err)
	}
	if payload.SchemaVersion != OrderSnapshotSchemaVersion || payload.Currency == "" {
		return OrderPricingSnapshotPayload{}, ErrInvalidOrderSnapshot
	}
	input := OrderPricingSnapshotInput{Currency: payload.Currency}
	var err error
	input.Subtotal, err = money.New(payload.SubtotalMinor, payload.Currency)
	if err != nil {
		return OrderPricingSnapshotPayload{}, ErrInvalidOrderSnapshot
	}
	input.Shipping, err = money.New(payload.ShippingMinor, payload.Currency)
	if err != nil {
		return OrderPricingSnapshotPayload{}, ErrInvalidOrderSnapshot
	}
	input.Tax, err = money.New(payload.TaxMinor, payload.Currency)
	if err != nil {
		return OrderPricingSnapshotPayload{}, ErrInvalidOrderSnapshot
	}
	input.MemberDiscount, err = money.New(payload.MemberDiscountMinor, payload.Currency)
	if err != nil {
		return OrderPricingSnapshotPayload{}, ErrInvalidOrderSnapshot
	}
	input.CouponDiscount, err = money.New(payload.CouponDiscountMinor, payload.Currency)
	if err != nil {
		return OrderPricingSnapshotPayload{}, ErrInvalidOrderSnapshot
	}
	input.DiscountTotal, err = money.New(payload.DiscountTotalMinor, payload.Currency)
	if err != nil {
		return OrderPricingSnapshotPayload{}, ErrInvalidOrderSnapshot
	}
	input.Total, err = money.New(payload.TotalMinor, payload.Currency)
	if err != nil {
		return OrderPricingSnapshotPayload{}, ErrInvalidOrderSnapshot
	}
	if _, err := input.Payload(); err != nil {
		return OrderPricingSnapshotPayload{}, err
	}
	return payload, nil
}
