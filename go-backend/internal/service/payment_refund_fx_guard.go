package service

import (
	"errors"
	"fmt"
	"strings"

	currencydomain "tanzanite/internal/domain/currency"
	orderdomain "tanzanite/internal/domain/order"
	paymentdomain "tanzanite/internal/domain/payment"
)

var ErrHistoricalRefundFXSnapshotMissing = errors.New("historical refund FX snapshot is missing")

func ensureRefundFXSnapshot(
	refund *paymentdomain.Refund,
	orderRecord *orderdomain.Order,
	transactionCurrency string,
) (currencydomain.OrderFXSnapshot, bool, error) {
	if refund == nil || orderRecord == nil {
		return currencydomain.OrderFXSnapshot{}, false, errors.New("refund and order are required for FX validation")
	}

	if snapshot, err := currencydomain.ParseOrderFXSnapshot(refund.FXSnapshotData); err == nil {
		if err := snapshot.Validate(transactionCurrency); err != nil {
			return currencydomain.OrderFXSnapshot{}, false, err
		}
		return snapshot, false, nil
	}

	if snapshot, err := currencydomain.ParseOrderFXSnapshot(orderRecord.FXSnapshotData); err == nil {
		if err := snapshot.Validate(transactionCurrency); err != nil {
			return currencydomain.OrderFXSnapshot{}, false, err
		}
		return snapshot, len(refund.FXSnapshotData) == 0 || string(refund.FXSnapshotData) == "{}", nil
	}

	// Orders created before migration 106 can still be refunded safely when
	// their transaction is in the historical default currency. A non-default
	// currency without a captured snapshot is blocked instead of guessing a
	// current exchange rate.
	if strings.EqualFold(strings.TrimSpace(transactionCurrency), currencydomain.DefaultPrimaryCurrency) {
		snapshot := currencydomain.OrderFXSnapshot{
			Version:         currencydomain.OrderFXSnapshotVersion,
			BaseCurrency:    currencydomain.DefaultPrimaryCurrency,
			OrderCurrency:   currencydomain.DefaultPrimaryCurrency,
			BaseToOrderRate: 1,
			Source:          "legacy_same_currency",
			CapturedAt:      orderRecord.CreatedAt.UTC(),
		}
		if snapshot.CapturedAt.IsZero() {
			snapshot.CapturedAt = orderRecord.UpdatedAt.UTC()
		}
		if snapshot.CapturedAt.IsZero() {
			return currencydomain.OrderFXSnapshot{}, false, ErrHistoricalRefundFXSnapshotMissing
		}
		return snapshot, true, nil
	}

	return currencydomain.OrderFXSnapshot{}, false, fmt.Errorf(
		"%w for order %s in %s",
		ErrHistoricalRefundFXSnapshotMissing,
		strings.TrimSpace(orderRecord.OrderNumber),
		strings.ToUpper(strings.TrimSpace(transactionCurrency)),
	)
}

func validateHistoricalRefundFXCap(
	snapshot currencydomain.OrderFXSnapshot,
	transaction *paymentdomain.Transaction,
	refundAmount float64,
	reservedAmount float64,
) error {
	if transaction == nil {
		return errors.New("transaction is required for FX validation")
	}
	if err := snapshot.Validate(transaction.Currency); err != nil {
		return err
	}
	if refundAmount <= 0 {
		return errors.New("refund amount must be greater than zero")
	}
	if reservedAmount < 0 {
		reservedAmount = 0
	}

	originalBaseAmount, err := snapshot.OrderAmountToBase(transaction.Amount)
	if err != nil {
		return err
	}
	reservedBaseAmount, err := snapshot.OrderAmountToBase(reservedAmount)
	if err != nil {
		return err
	}
	refundBaseAmount, err := snapshot.OrderAmountToBase(refundAmount)
	if err != nil {
		return err
	}
	remainingBaseAmount := originalBaseAmount - reservedBaseAmount
	if refundBaseAmount > remainingBaseAmount+0.01 {
		return fmt.Errorf(
			"refund amount %.2f %s exceeds historical FX cap %.2f %s",
			refundBaseAmount,
			snapshot.BaseCurrency,
			maxFloat(remainingBaseAmount, 0),
			snapshot.BaseCurrency,
		)
	}
	return nil
}

func maxFloat(value, minimum float64) float64 {
	if value < minimum {
		return minimum
	}
	return value
}
