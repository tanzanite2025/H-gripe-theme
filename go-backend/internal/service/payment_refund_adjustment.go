package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/coupon"
	currencydomain "commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	domainpricing "commerce-platform/internal/domain/pricing"
	"commerce-platform/internal/repository"
)

type refundPromotionAdjustment struct {
	RequestedAmount        domainmoney.Money
	NetAmount              domainmoney.Money
	DiscountClawbackAmount domainmoney.Money
	CalculationSnapshot    string
}

type refundPromotionAdjustmentSnapshot struct {
	Version int    `json:"version"`
	Policy  string `json:"policy"`
	Reason  string `json:"reason,omitempty"`

	// Transactional evidence is persisted in the currency's smallest unit.
	// Percentage coupon values remain exact decimal strings because they are
	// rates rather than monetary amounts.
	RequestedAmountMinor                  int64  `json:"requested_amount_minor"`
	RequestedSubtotalAmountMinor          int64  `json:"requested_subtotal_amount_minor"`
	NetRefundAmountMinor                  int64  `json:"net_refund_amount_minor"`
	DiscountClawbackAmountMinor           int64  `json:"discount_clawback_amount_minor"`
	OrderSubtotalAmountMinor              int64  `json:"order_subtotal_amount_minor"`
	PreviousRefundedSubtotalAmountMinor   int64  `json:"previous_refunded_subtotal_amount_minor"`
	PreviousDiscountClawbackAmountMinor   int64  `json:"previous_discount_clawback_amount_minor"`
	RemainingSubtotalBeforeRefundMinor    int64  `json:"remaining_subtotal_before_refund_minor"`
	RemainingSubtotalAfterRefundMinor     int64  `json:"remaining_subtotal_after_refund_minor"`
	OriginalCouponDiscountAmountMinor     int64  `json:"original_coupon_discount_amount_minor"`
	RecalculatedCouponDiscountAmountMinor int64  `json:"recalculated_coupon_discount_amount_minor"`
	CouponID                              uint   `json:"coupon_id,omitempty"`
	CouponCode                            string `json:"coupon_code,omitempty"`
	CouponType                            string `json:"coupon_type,omitempty"`
	CouponValueRateDecimal                string `json:"coupon_value_rate_decimal,omitempty"`
	CouponValueMinor                      int64  `json:"coupon_value_minor,omitempty"`
	CouponMinAmountMinor                  int64  `json:"coupon_min_amount_minor,omitempty"`
	CouponMaxDiscountMinor                int64  `json:"coupon_max_discount_minor,omitempty"`
}

var errInvalidOrderPricingSnapshot = errors.New("invalid order pricing snapshot")

func readOrderPricingRefundBaseline(o *order.Order) (subtotal domainmoney.Money, couponDiscount domainmoney.Money, present bool, err error) {
	if o == nil || len(o.PricingSnapshotData) == 0 || string(o.PricingSnapshotData) == "{}" {
		return domainmoney.Money{}, domainmoney.Money{}, false, nil
	}
	payload, parseErr := domainpricing.ParseOrderPricingSnapshot(o.PricingSnapshotData)
	if parseErr != nil {
		return domainmoney.Money{}, domainmoney.Money{}, true, fmt.Errorf("%w: %v", errInvalidOrderPricingSnapshot, parseErr)
	}
	if o.Currency != "" && !strings.EqualFold(payload.Currency, o.Currency) {
		return domainmoney.Money{}, domainmoney.Money{}, true, fmt.Errorf("%w: snapshot currency %s does not match order currency %s", errInvalidOrderPricingSnapshot, payload.Currency, o.Currency)
	}
	subtotal, parseErr = domainmoney.New(payload.SubtotalMinor, payload.Currency)
	if parseErr != nil {
		return domainmoney.Money{}, domainmoney.Money{}, true, fmt.Errorf("%w: subtotal: %v", errInvalidOrderPricingSnapshot, parseErr)
	}
	couponDiscount, parseErr = domainmoney.New(payload.CouponDiscountMinor, payload.Currency)
	if parseErr != nil {
		return domainmoney.Money{}, domainmoney.Money{}, true, fmt.Errorf("%w: coupon discount: %v", errInvalidOrderPricingSnapshot, parseErr)
	}
	return subtotal, couponDiscount, true, nil
}

func calculateRefundPromotionAdjustment(
	repos repository.TxRepositories,
	o *order.Order,
	requestedAmount domainmoney.Money,
	requestedSubtotalAmount domainmoney.Money,
) (refundPromotionAdjustment, error) {
	if o == nil {
		return refundPromotionAdjustment{}, errors.New("order is required")
	}
	currencyCode := currencydomain.NormalizeCode(o.Currency)
	if currencyCode == "" {
		currencyCode = currencydomain.DefaultPrimaryCurrency
	}
	if err := validateRefundAdjustmentMoney(requestedAmount, currencyCode, "requested refund amount"); err != nil {
		return refundPromotionAdjustment{}, err
	}
	if requestedAmount.AmountMinor() <= 0 {
		return refundPromotionAdjustment{}, errors.New("amount must be greater than zero")
	}
	if requestedSubtotalAmount.Currency().String() == "" {
		requestedSubtotalAmount = requestedAmount
	} else if err := validateRefundAdjustmentMoney(requestedSubtotalAmount, currencyCode, "requested subtotal amount"); err != nil {
		return refundPromotionAdjustment{}, err
	} else if requestedSubtotalAmount.AmountMinor() <= 0 {
		requestedSubtotalAmount = requestedAmount
	}
	snapshotSubtotal, snapshotCouponDiscount, snapshotPresent, err := readOrderPricingRefundBaseline(o)
	if err != nil {
		return refundPromotionAdjustment{}, err
	}
	orderSubtotal, err := o.SubtotalMoney()
	if err != nil {
		return refundPromotionAdjustment{}, fmt.Errorf("order subtotal: %w", err)
	}
	if snapshotPresent {
		orderSubtotal = snapshotSubtotal
	}
	zero, err := domainmoney.New(0, currencyCode)
	if err != nil {
		return refundPromotionAdjustment{}, err
	}

	usage, err := repos.Coupon.FindCouponUsageByOrderID(o.ID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return newRefundPromotionAdjustmentSnapshot(
				nil, nil, requestedAmount, requestedSubtotalAmount, requestedAmount,
				zero, orderSubtotal, zero, zero, zero, zero, zero,
				"no_coupon_usage", nil,
			)
		}
		return refundPromotionAdjustment{}, err
	}
	persistedCouponDiscount, err := usage.DiscountMoney(currencyCode)
	if err != nil {
		return refundPromotionAdjustment{}, fmt.Errorf("original coupon discount: %w", err)
	}
	originalCouponDiscount := persistedCouponDiscount
	if snapshotPresent {
		if snapshotCouponDiscount.AmountMinor() <= 0 && persistedCouponDiscount.AmountMinor() > 0 {
			return refundPromotionAdjustment{}, fmt.Errorf("%w: coupon usage exists but order snapshot has no coupon discount", errInvalidOrderPricingSnapshot)
		}
		originalCouponDiscount = snapshotCouponDiscount
	}
	if originalCouponDiscount.AmountMinor() <= 0 {
		return newRefundPromotionAdjustmentSnapshot(
			nil, usage, requestedAmount, requestedSubtotalAmount, requestedAmount,
			zero, orderSubtotal, zero, zero, zero, zero, originalCouponDiscount,
			"coupon_discount_zero", nil,
		)
	}

	couponRecord, err := repos.Coupon.FindCouponByIDIncludingDeleted(usage.CouponID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return refundPromotionAdjustment{}, errors.New("coupon used by order is missing; cannot calculate promotional refund adjustment")
		}
		return refundPromotionAdjustment{}, err
	}

	previousRefundedSubtotalMinor, err := repos.Payment.SumRefundedSubtotalMinorAmountByOrderID(o.ID, "pending", "completed")
	if err != nil {
		return refundPromotionAdjustment{}, err
	}
	previousRefundedSubtotal, err := domainmoney.New(previousRefundedSubtotalMinor, currencyCode)
	if err != nil {
		return refundPromotionAdjustment{}, fmt.Errorf("previous refunded subtotal: %w", err)
	}
	previousDiscountClawbackMinor, err := repos.Payment.SumRefundDiscountClawbackMinorByOrderID(o.ID, "pending", "completed")
	if err != nil {
		return refundPromotionAdjustment{}, err
	}
	previousDiscountClawback, err := domainmoney.New(previousDiscountClawbackMinor, currencyCode)
	if err != nil {
		return refundPromotionAdjustment{}, fmt.Errorf("previous discount clawback: %w", err)
	}

	return applyRefundPromotionClawback(
		o,
		couponRecord,
		usage,
		requestedAmount,
		requestedSubtotalAmount,
		orderSubtotal,
		originalCouponDiscount,
		previousRefundedSubtotal,
		previousDiscountClawback,
	)
}

// calculateRefundPromotionAdjustmentFromPersistedPricing is the refund path
// for orders whose line snapshot already contains the coupon allocation. The
// checkout pipeline has permanently decided the net line amount, so a refund
// must not reverse-engineer an order-level coupon from the remaining subtotal
// (which can produce a different result after a partial return).
func calculateRefundPromotionAdjustmentFromPersistedPricing(
	o *order.Order,
	requestedAmount domainmoney.Money,
	requestedSubtotalAmount domainmoney.Money,
	originalCouponDiscount domainmoney.Money,
) (refundPromotionAdjustment, error) {
	if o == nil {
		return refundPromotionAdjustment{}, errors.New("order is required")
	}
	currencyCode := currencydomain.NormalizeCode(o.Currency)
	if currencyCode == "" {
		currencyCode = currencydomain.DefaultPrimaryCurrency
	}
	for label, value := range map[string]domainmoney.Money{
		"requested refund amount":   requestedAmount,
		"requested subtotal amount": requestedSubtotalAmount,
		"original coupon discount":  originalCouponDiscount,
	} {
		if err := validateRefundAdjustmentMoney(value, currencyCode, label); err != nil {
			return refundPromotionAdjustment{}, err
		}
	}
	orderSubtotal, err := o.SubtotalMoney()
	if err != nil {
		return refundPromotionAdjustment{}, fmt.Errorf("order subtotal: %w", err)
	}
	zero := zeroRefundMoney(currencyCode)
	return newRefundPromotionAdjustmentSnapshot(
		nil,
		nil,
		requestedAmount,
		requestedSubtotalAmount,
		requestedAmount,
		zero,
		orderSubtotal,
		zero,
		zero,
		zero,
		zero,
		originalCouponDiscount,
		"persisted_line_pricing",
		nil,
	)
}

func applyRefundPromotionClawback(
	o *order.Order,
	couponRecord *coupon.Coupon,
	usage *coupon.CouponUsage,
	requestedMoney domainmoney.Money,
	requestedSubtotalMoney domainmoney.Money,
	orderSubtotalMoney domainmoney.Money,
	originalCouponMoney domainmoney.Money,
	previousRefundedMoney domainmoney.Money,
	previousClawbackMoney domainmoney.Money,
) (refundPromotionAdjustment, error) {
	if o == nil {
		return refundPromotionAdjustment{}, errors.New("order is required")
	}
	currencyCode := currencydomain.NormalizeCode(o.Currency)
	if currencyCode == "" {
		currencyCode = currencydomain.DefaultPrimaryCurrency
	}
	values := []struct {
		name  string
		value domainmoney.Money
	}{
		{name: "requested refund amount", value: requestedMoney},
		{name: "requested subtotal amount", value: requestedSubtotalMoney},
		{name: "order subtotal", value: orderSubtotalMoney},
		{name: "original coupon discount", value: originalCouponMoney},
		{name: "previous refunded subtotal", value: previousRefundedMoney},
		{name: "previous discount clawback", value: previousClawbackMoney},
	}
	for _, item := range values {
		if err := validateRefundAdjustmentMoney(item.value, currencyCode, item.name); err != nil {
			return refundPromotionAdjustment{}, err
		}
	}
	if orderSubtotalMoney.AmountMinor() <= 0 {
		zero := zeroRefundMoney(currencyCode)
		return newRefundPromotionAdjustmentSnapshot(
			couponRecord, usage, requestedMoney, requestedSubtotalMoney, requestedMoney,
			zero, orderSubtotalMoney, previousRefundedMoney, previousClawbackMoney,
			zero, zero, originalCouponMoney, "order_subtotal_missing", nil,
		)
	}
	previousRefundedMoney = clampRefundMoneyValue(previousRefundedMoney, zeroRefundMoney(currencyCode), orderSubtotalMoney)
	previousClawbackMoney = clampRefundMoneyValue(previousClawbackMoney, zeroRefundMoney(currencyCode), originalCouponMoney)
	remainingBeforeMoney, err := orderSubtotalMoney.Subtract(previousRefundedMoney)
	if err != nil {
		return refundPromotionAdjustment{}, err
	}
	if remainingBeforeMoney.AmountMinor() < 0 {
		remainingBeforeMoney = zeroRefundMoney(currencyCode)
	}
	remainingAfterMoney, err := remainingBeforeMoney.Subtract(requestedSubtotalMoney)
	if err != nil {
		return refundPromotionAdjustment{}, err
	}
	if remainingAfterMoney.AmountMinor() < 0 {
		remainingAfterMoney = zeroRefundMoney(currencyCode)
	}
	recalculatedCouponMoney, err := couponRecord.CalculateDiscountMoney(remainingAfterMoney)
	if err != nil {
		return refundPromotionAdjustment{}, fmt.Errorf("recalculated coupon discount: %w", err)
	}
	if recalculatedCouponMoney.AmountMinor() > originalCouponMoney.AmountMinor() {
		recalculatedCouponMoney = originalCouponMoney
	}
	requiredTotalClawbackMoney, err := originalCouponMoney.Subtract(recalculatedCouponMoney)
	if err != nil {
		return refundPromotionAdjustment{}, err
	}
	if requiredTotalClawbackMoney.AmountMinor() < 0 {
		requiredTotalClawbackMoney = zeroRefundMoney(currencyCode)
	}
	discountClawbackMoney, err := requiredTotalClawbackMoney.Subtract(previousClawbackMoney)
	if err != nil {
		return refundPromotionAdjustment{}, err
	}
	if discountClawbackMoney.AmountMinor() < 0 {
		discountClawbackMoney = zeroRefundMoney(currencyCode)
	}
	discountClawbackMoney = clampRefundMoneyValue(discountClawbackMoney, zeroRefundMoney(currencyCode), requestedMoney)
	netMoney, err := requestedMoney.Subtract(discountClawbackMoney)
	if err != nil {
		return refundPromotionAdjustment{}, err
	}
	if netMoney.AmountMinor() < 0 {
		return refundPromotionAdjustment{}, errors.New("discount clawback consumes the requested refund; no gateway refund amount remains")
	}
	return newRefundPromotionAdjustmentSnapshot(
		couponRecord,
		usage,
		requestedMoney,
		requestedSubtotalMoney,
		netMoney,
		discountClawbackMoney,
		orderSubtotalMoney,
		previousRefundedMoney,
		previousClawbackMoney,
		remainingBeforeMoney,
		remainingAfterMoney,
		originalCouponMoney,
		"coupon_recalculation",
		&recalculatedCouponMoney,
	)
}

func zeroRefundMoney(currency string) domainmoney.Money {
	return domainmoney.MustNew(0, currency)
}

func clampRefundMoneyValue(value, min, max domainmoney.Money) domainmoney.Money {
	if value.AmountMinor() < min.AmountMinor() {
		return min
	}
	if value.AmountMinor() > max.AmountMinor() {
		return max
	}
	return value
}

func newRefundPromotionAdjustmentSnapshot(
	couponRecord *coupon.Coupon,
	usage *coupon.CouponUsage,
	requestedAmount domainmoney.Money,
	requestedSubtotalAmount domainmoney.Money,
	netAmount domainmoney.Money,
	discountClawback domainmoney.Money,
	orderSubtotal domainmoney.Money,
	previousRefundedSubtotal domainmoney.Money,
	previousDiscountClawback domainmoney.Money,
	remainingBefore domainmoney.Money,
	remainingAfter domainmoney.Money,
	originalCouponDiscount domainmoney.Money,
	reason string,
	recalculatedCouponMoney *domainmoney.Money,
) (refundPromotionAdjustment, error) {
	amounts := []domainmoney.Money{
		requestedAmount,
		requestedSubtotalAmount,
		netAmount,
		discountClawback,
		orderSubtotal,
		previousRefundedSubtotal,
		previousDiscountClawback,
		remainingBefore,
		remainingAfter,
		originalCouponDiscount,
	}
	minor := make([]int64, len(amounts))
	for index, amount := range amounts {
		if err := amount.Validate(); err != nil {
			return refundPromotionAdjustment{}, fmt.Errorf("validate refund promotion snapshot amount: %w", err)
		}
		minor[index] = amount.AmountMinor()
	}
	snapshot := refundPromotionAdjustmentSnapshot{
		Version:                             2,
		Policy:                              "minor-unit-coupon-recalculation",
		Reason:                              reason,
		RequestedAmountMinor:                minor[0],
		RequestedSubtotalAmountMinor:        minor[1],
		NetRefundAmountMinor:                minor[2],
		DiscountClawbackAmountMinor:         minor[3],
		OrderSubtotalAmountMinor:            minor[4],
		PreviousRefundedSubtotalAmountMinor: minor[5],
		PreviousDiscountClawbackAmountMinor: minor[6],
		RemainingSubtotalBeforeRefundMinor:  minor[7],
		RemainingSubtotalAfterRefundMinor:   minor[8],
	}
	if usage != nil {
		snapshot.OriginalCouponDiscountAmountMinor = minor[9]
	}
	if couponRecord != nil {
		snapshot.CouponID = couponRecord.ID
		snapshot.CouponCode = couponRecord.Code
		snapshot.CouponType = couponRecord.Type
		if couponRecord.Type == "percentage" {
			snapshot.CouponValueRateDecimal = strings.TrimSpace(couponRecord.ValueRateDecimal)
		}
		snapshot.CouponValueMinor = couponRecord.ValueMinor
		snapshot.CouponMinAmountMinor = couponRecord.MinAmountMinor
		snapshot.CouponMaxDiscountMinor = couponRecord.MaxDiscountMinor
		if recalculatedCouponMoney != nil {
			recalculated := *recalculatedCouponMoney
			if recalculated.AmountMinor() > originalCouponDiscount.AmountMinor() {
				recalculated = originalCouponDiscount
			}
			snapshot.RecalculatedCouponDiscountAmountMinor = recalculated.AmountMinor()
		}
	}

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		return refundPromotionAdjustment{}, fmt.Errorf("marshal refund promotion snapshot: %w", err)
	}
	return refundPromotionAdjustment{
		RequestedAmount:        requestedAmount,
		NetAmount:              netAmount,
		DiscountClawbackAmount: discountClawback,
		CalculationSnapshot:    string(snapshotJSON),
	}, nil
}

func validateRefundAdjustmentMoney(value domainmoney.Money, currencyCode string, label string) error {
	if err := value.Validate(); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if value.Currency().String() != currencyCode {
		return fmt.Errorf("%s: %w", label, domainmoney.ErrCurrencyMismatch)
	}
	return nil
}
