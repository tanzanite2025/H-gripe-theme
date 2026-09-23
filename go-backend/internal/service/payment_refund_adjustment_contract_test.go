package service

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRefundPromotionAdjustmentSnapshotUsesMinorUnits(t *testing.T) {
	snapshot := refundPromotionAdjustmentSnapshot{
		Version:                           2,
		Policy:                            "minor-unit-coupon-recalculation",
		RequestedAmountMinor:              12345,
		NetRefundAmountMinor:              12000,
		OriginalCouponDiscountAmountMinor: 345,
		CouponValueRateDecimal:            "12.500000000000000",
		CouponValueMinor:                  500,
		CouponMinAmountMinor:              1000,
		CouponMaxDiscountMinor:            2500,
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	serialized := string(payload)
	for _, legacy := range []string{
		`"requested_amount"`,
		`"net_refund_amount"`,
		`"coupon_value"`,
		`"coupon_min_amount"`,
		`"coupon_max_discount"`,
	} {
		if strings.Contains(serialized, legacy) {
			t.Fatalf("snapshot contains legacy major-unit field %s: %s", legacy, serialized)
		}
	}
	for _, canonical := range []string{
		`"requested_amount_minor":12345`,
		`"net_refund_amount_minor":12000`,
		`"coupon_value_rate_decimal":"12.500000000000000"`,
		`"coupon_value_minor":500`,
	} {
		if !strings.Contains(serialized, canonical) {
			t.Fatalf("snapshot is missing canonical field %s: %s", canonical, serialized)
		}
	}
}
