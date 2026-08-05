package service

import (
	"encoding/json"
	"errors"
	"math"

	"tanzanite/internal/domain/coupon"
	"tanzanite/internal/domain/order"
	"tanzanite/internal/repository"
)

type refundPromotionAdjustment struct {
	RequestedAmount        float64
	NetAmount              float64
	DiscountClawbackAmount float64
	CalculationSnapshot    string
}

type refundPromotionAdjustmentSnapshot struct {
	Version                          int     `json:"version"`
	Policy                           string  `json:"policy"`
	Reason                           string  `json:"reason,omitempty"`
	RequestedAmount                  float64 `json:"requested_amount"`
	RequestedSubtotalAmount          float64 `json:"requested_subtotal_amount"`
	NetRefundAmount                  float64 `json:"net_refund_amount"`
	DiscountClawbackAmount           float64 `json:"discount_clawback_amount"`
	OrderSubtotalAmount              float64 `json:"order_subtotal_amount"`
	PreviousRefundedSubtotalAmount   float64 `json:"previous_refunded_subtotal_amount"`
	PreviousDiscountClawbackAmount   float64 `json:"previous_discount_clawback_amount"`
	RemainingSubtotalBeforeRefund    float64 `json:"remaining_subtotal_before_refund"`
	RemainingSubtotalAfterRefund     float64 `json:"remaining_subtotal_after_refund"`
	OriginalCouponDiscountAmount     float64 `json:"original_coupon_discount_amount"`
	RecalculatedCouponDiscountAmount float64 `json:"recalculated_coupon_discount_amount"`
	CouponID                         uint    `json:"coupon_id,omitempty"`
	CouponCode                       string  `json:"coupon_code,omitempty"`
	CouponType                       string  `json:"coupon_type,omitempty"`
	CouponValue                      float64 `json:"coupon_value,omitempty"`
	CouponMinAmount                  float64 `json:"coupon_min_amount,omitempty"`
	CouponMaxDiscount                float64 `json:"coupon_max_discount,omitempty"`
}

func calculateRefundPromotionAdjustment(repos repository.TxRepositories, o *order.Order, requestedAmount float64, requestedSubtotalAmount float64) (refundPromotionAdjustment, error) {
	requestedAmount = roundRefundMoney(requestedAmount)
	requestedSubtotalAmount = roundRefundMoney(requestedSubtotalAmount)
	if requestedAmount <= 0 {
		return refundPromotionAdjustment{}, errors.New("amount must be greater than zero")
	}
	if requestedSubtotalAmount <= 0 {
		requestedSubtotalAmount = requestedAmount
	}

	usage, err := repos.Coupon.FindCouponUsageByOrderID(o.ID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return newRefundPromotionAdjustmentSnapshot(o, nil, nil, requestedAmount, requestedSubtotalAmount, requestedAmount, 0, 0, 0, 0, 0, "no_coupon_usage"), nil
		}
		return refundPromotionAdjustment{}, err
	}
	if usage.Discount <= 0 {
		return newRefundPromotionAdjustmentSnapshot(o, nil, usage, requestedAmount, requestedSubtotalAmount, requestedAmount, 0, 0, 0, 0, 0, "coupon_discount_zero"), nil
	}

	couponRecord, err := repos.Coupon.FindCouponByIDIncludingDeleted(usage.CouponID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return refundPromotionAdjustment{}, errors.New("coupon used by order is missing; cannot calculate promotional refund adjustment")
		}
		return refundPromotionAdjustment{}, err
	}

	previousRefundedSubtotal, err := repos.Payment.SumRefundedSubtotalAmountByOrderID(o.ID, "pending", "completed")
	if err != nil {
		return refundPromotionAdjustment{}, err
	}
	previousDiscountClawback, err := repos.Payment.SumRefundDiscountClawbackByOrderID(o.ID, "pending", "completed")
	if err != nil {
		return refundPromotionAdjustment{}, err
	}

	return applyRefundPromotionClawback(o, couponRecord, usage, requestedAmount, requestedSubtotalAmount, previousRefundedSubtotal, previousDiscountClawback)
}

func applyRefundPromotionClawback(
	o *order.Order,
	couponRecord *coupon.Coupon,
	usage *coupon.CouponUsage,
	requestedAmount float64,
	requestedSubtotalAmount float64,
	previousRefundedSubtotal float64,
	previousDiscountClawback float64,
) (refundPromotionAdjustment, error) {
	requestedAmount = roundRefundMoney(requestedAmount)
	requestedSubtotalAmount = roundRefundMoney(requestedSubtotalAmount)
	if requestedSubtotalAmount <= 0 {
		requestedSubtotalAmount = requestedAmount
	}
	orderSubtotal := roundRefundMoney(o.SubtotalAmount)
	if orderSubtotal <= 0 {
		return newRefundPromotionAdjustmentSnapshot(o, couponRecord, usage, requestedAmount, requestedSubtotalAmount, requestedAmount, 0, previousRefundedSubtotal, previousDiscountClawback, 0, 0, "order_subtotal_missing"), nil
	}

	previousRefundedSubtotal = clampRefundMoney(previousRefundedSubtotal, 0, orderSubtotal)
	previousDiscountClawback = clampRefundMoney(previousDiscountClawback, 0, usage.Discount)
	remainingBefore := roundRefundMoney(orderSubtotal - previousRefundedSubtotal)
	if remainingBefore < 0 {
		remainingBefore = 0
	}
	remainingAfter := roundRefundMoney(remainingBefore - requestedSubtotalAmount)
	if remainingAfter < 0 {
		remainingAfter = 0
	}

	originalCouponDiscount := roundRefundMoney(usage.Discount)
	recalculatedCouponDiscount := roundRefundMoney(couponRecord.CalculateDiscount(remainingAfter))
	if recalculatedCouponDiscount > originalCouponDiscount {
		recalculatedCouponDiscount = originalCouponDiscount
	}

	requiredTotalClawback := roundRefundMoney(originalCouponDiscount - recalculatedCouponDiscount)
	if requiredTotalClawback < 0 {
		requiredTotalClawback = 0
	}
	discountClawback := roundRefundMoney(requiredTotalClawback - previousDiscountClawback)
	if discountClawback < 0 {
		discountClawback = 0
	}
	if discountClawback > requestedAmount {
		discountClawback = requestedAmount
	}

	netAmount := roundRefundMoney(requestedAmount - discountClawback)
	if netAmount <= 0 {
		return refundPromotionAdjustment{}, errors.New("discount clawback consumes the requested refund; no gateway refund amount remains")
	}

	return newRefundPromotionAdjustmentSnapshot(
		o,
		couponRecord,
		usage,
		requestedAmount,
		requestedSubtotalAmount,
		netAmount,
		discountClawback,
		previousRefundedSubtotal,
		previousDiscountClawback,
		remainingBefore,
		remainingAfter,
		"coupon_recalculation",
	), nil
}

func newRefundPromotionAdjustmentSnapshot(
	o *order.Order,
	couponRecord *coupon.Coupon,
	usage *coupon.CouponUsage,
	requestedAmount float64,
	requestedSubtotalAmount float64,
	netAmount float64,
	discountClawback float64,
	previousRefundedSubtotal float64,
	previousDiscountClawback float64,
	remainingBefore float64,
	remainingAfter float64,
	reason string,
) refundPromotionAdjustment {
	snapshot := refundPromotionAdjustmentSnapshot{
		Version:                        1,
		Policy:                         "remaining-subtotal-coupon-recalculation",
		Reason:                         reason,
		RequestedAmount:                roundRefundMoney(requestedAmount),
		RequestedSubtotalAmount:        roundRefundMoney(requestedSubtotalAmount),
		NetRefundAmount:                roundRefundMoney(netAmount),
		DiscountClawbackAmount:         roundRefundMoney(discountClawback),
		OrderSubtotalAmount:            roundRefundMoney(o.SubtotalAmount),
		PreviousRefundedSubtotalAmount: roundRefundMoney(previousRefundedSubtotal),
		PreviousDiscountClawbackAmount: roundRefundMoney(previousDiscountClawback),
		RemainingSubtotalBeforeRefund:  roundRefundMoney(remainingBefore),
		RemainingSubtotalAfterRefund:   roundRefundMoney(remainingAfter),
	}
	if usage != nil {
		snapshot.OriginalCouponDiscountAmount = roundRefundMoney(usage.Discount)
	}
	if couponRecord != nil {
		snapshot.CouponID = couponRecord.ID
		snapshot.CouponCode = couponRecord.Code
		snapshot.CouponType = couponRecord.Type
		snapshot.CouponValue = couponRecord.Value
		snapshot.CouponMinAmount = couponRecord.MinAmount
		snapshot.CouponMaxDiscount = couponRecord.MaxDiscount
		snapshot.RecalculatedCouponDiscountAmount = roundRefundMoney(couponRecord.CalculateDiscount(snapshot.RemainingSubtotalAfterRefund))
		if snapshot.RecalculatedCouponDiscountAmount > snapshot.OriginalCouponDiscountAmount && snapshot.OriginalCouponDiscountAmount > 0 {
			snapshot.RecalculatedCouponDiscountAmount = snapshot.OriginalCouponDiscountAmount
		}
	}

	snapshotJSON, _ := json.Marshal(snapshot)
	return refundPromotionAdjustment{
		RequestedAmount:        snapshot.RequestedAmount,
		NetAmount:              snapshot.NetRefundAmount,
		DiscountClawbackAmount: snapshot.DiscountClawbackAmount,
		CalculationSnapshot:    string(snapshotJSON),
	}
}

func roundRefundMoney(value float64) float64 {
	return math.Round(value*100) / 100
}

func clampRefundMoney(value, min, max float64) float64 {
	value = roundRefundMoney(value)
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
