package order

import (
	"errors"
	"fmt"
	"math/big"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/domain/money"
)

type HighValueOrderEvaluation struct {
	OrderTotalUSD money.Money
	IsHighValue   bool
}

// HighValueSignatureThresholdUSD is the human-readable policy amount used by
// transport-facing configuration and fixtures. Comparisons use the exact
// minor-unit threshold below.
const HighValueSignatureThresholdUSD = 750

const HighValueSignatureThresholdUSDMinor int64 = 75000

var highValueSignatureThreshold = money.MustNew(HighValueSignatureThresholdUSDMinor, currency.DefaultPrimaryCurrency)

// OrderAmountToBaseMoney converts an order amount using the immutable FX
// snapshot captured at checkout. The captured decimal is parsed exactly and
// all arithmetic thereafter uses exact minor-unit integers.
func OrderAmountToBaseMoney(amount money.Money, fxSnapshot currency.OrderFXSnapshot) (money.Money, error) {
	if err := fxSnapshot.Validate(fxSnapshot.OrderCurrency); err != nil {
		return money.Money{}, err
	}
	if err := amount.Validate(); err != nil {
		return money.Money{}, err
	}
	if amount.Currency().String() != currency.NormalizeCode(fxSnapshot.OrderCurrency) {
		return money.Money{}, fmt.Errorf("order amount currency %s does not match FX snapshot currency %s", amount.Currency(), currency.NormalizeCode(fxSnapshot.OrderCurrency))
	}
	if amount.AmountMinor() < 0 {
		return money.Money{}, errors.New("order amount cannot be negative")
	}
	rate, err := fxSnapshot.RateRat()
	if err != nil {
		return money.Money{}, errors.New("invalid order FX snapshot rate")
	}
	return amount.ConvertAtRat(new(big.Rat).Inv(rate), fxSnapshot.BaseCurrency)
}

// EvaluateHighValueOrder is the single order-total policy used by signature
// handling and the order evidence snapshot. It never inspects product data.
func EvaluateHighValueOrder(totalAmount money.Money, fxSnapshot currency.OrderFXSnapshot) (HighValueOrderEvaluation, error) {
	if currency.NormalizeCode(fxSnapshot.BaseCurrency) != currency.DefaultPrimaryCurrency {
		return HighValueOrderEvaluation{}, fmt.Errorf(
			"high-value order policy requires a %s FX base currency",
			currency.DefaultPrimaryCurrency,
		)
	}
	orderTotalMoney, err := OrderAmountToBaseMoney(totalAmount, fxSnapshot)
	if err != nil {
		return HighValueOrderEvaluation{}, err
	}
	return HighValueOrderEvaluation{
		OrderTotalUSD: orderTotalMoney,
		IsHighValue:   orderTotalMoney.AmountMinor() >= highValueSignatureThreshold.AmountMinor(),
	}, nil
}
