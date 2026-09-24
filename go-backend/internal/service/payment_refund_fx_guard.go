package service

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"

	currencydomain "commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
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
			Version:       currencydomain.OrderFXSnapshotVersion,
			BaseCurrency:  currencydomain.DefaultPrimaryCurrency,
			OrderCurrency: currencydomain.DefaultPrimaryCurrency,
			RateDecimal:   "1",
			Source:        "legacy_same_currency",
			CapturedAt:    orderRecord.CreatedAt.UTC(),
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
	refundAmount domainmoney.Money,
	reservedAmount domainmoney.Money,
) error {
	if transaction == nil {
		return errors.New("transaction is required for FX validation")
	}
	if err := snapshot.Validate(transaction.Currency); err != nil {
		return err
	}
	if err := refundAmount.Validate(); err != nil {
		return fmt.Errorf("invalid refund amount: %w", err)
	}
	if err := reservedAmount.Validate(); err != nil {
		return fmt.Errorf("invalid reserved refund amount: %w", err)
	}
	transactionCurrency, err := currencydomain.ParseCode(transaction.Currency)
	if err != nil {
		return fmt.Errorf("invalid transaction currency: %w", err)
	}
	if refundAmount.Currency() != transactionCurrency || reservedAmount.Currency() != transactionCurrency {
		return domainmoney.ErrCurrencyMismatch
	}
	if refundAmount.AmountMinor() < 0 {
		return errors.New("refund amount cannot be negative")
	}
	if reservedAmount.AmountMinor() < 0 {
		reservedAmount, err = domainmoney.New(0, transaction.Currency)
		if err != nil {
			return err
		}
	}

	transactionMoney, err := transaction.AmountMoney()
	if err != nil {
		return fmt.Errorf("invalid transaction amount: %w", err)
	}
	remainingMoney, err := transactionMoney.Subtract(reservedAmount)
	if err != nil {
		return fmt.Errorf("calculate remaining refund amount: %w", err)
	}
	if remainingMoney.AmountMinor() < 0 {
		remainingMoney, _ = domainmoney.New(0, transaction.Currency)
	}
	rate, err := snapshot.RateRat()
	if err != nil {
		return errors.New("order FX snapshot rate is invalid")
	}
	orderToBase := new(big.Rat).Inv(rate)
	remainingBase, err := remainingMoney.ConvertAtRat(orderToBase, snapshot.BaseCurrency)
	if err != nil {
		return fmt.Errorf("convert remaining refund amount to base currency: %w", err)
	}
	refundBase, err := refundAmount.ConvertAtRat(orderToBase, snapshot.BaseCurrency)
	if err != nil {
		return fmt.Errorf("convert refund amount to base currency: %w", err)
	}
	if refundBase.AmountMinor() > remainingBase.AmountMinor() {
		remainingFormatted, _ := remainingBase.FormatMajor()
		refundFormatted, _ := refundBase.FormatMajor()
		return fmt.Errorf(
			"refund amount %s %s exceeds historical FX cap %s %s",
			refundFormatted,
			snapshot.BaseCurrency,
			remainingFormatted,
			snapshot.BaseCurrency,
		)
	}
	return nil
}

// calculateRefundFXGainLoss compares the provider's actual settlement
// deduction with the historical base-currency value used by the local refund
// cap. The signed result is intentionally loss-positive: a positive value
// means the provider deducted more base-currency minor units than the books
// expected, while a negative value is an FX gain. When a provider settles in
// another currency, the fact is still stored but no incomparable adjustment is
// posted here; that account requires a separate settlement-currency ledger.
func calculateRefundFXGainLoss(
	snapshot currencydomain.OrderFXSnapshot,
	refundAmount domainmoney.Money,
	settlementAmountMinor int64,
	settlementCurrency string,
) (int64, string, error) {
	if settlementAmountMinor == 0 || strings.TrimSpace(settlementCurrency) == "" {
		return 0, "", nil
	}
	if settlementAmountMinor == math.MinInt64 {
		return 0, "", errors.New("provider settlement amount overflows")
	}
	if err := snapshot.Validate(refundAmount.Currency().String()); err != nil {
		return 0, "", err
	}
	settlementCode, err := currencydomain.ParseCode(settlementCurrency)
	if err != nil {
		return 0, "", fmt.Errorf("invalid settlement currency: %w", err)
	}
	base := currencydomain.NormalizeCode(snapshot.BaseCurrency)
	if settlementCode.String() != base {
		return 0, "", nil
	}
	rate, err := snapshot.RateRat()
	if err != nil {
		return 0, "", errors.New("order FX snapshot rate is invalid")
	}
	historicalBase, err := refundAmount.ConvertAtRat(new(big.Rat).Inv(rate), base)
	if err != nil {
		return 0, "", fmt.Errorf("convert historical refund amount to base currency: %w", err)
	}
	if settlementAmountMinor < 0 {
		settlementAmountMinor = -settlementAmountMinor
	}
	return settlementAmountMinor - historicalBase.AmountMinor(), base, nil
}
