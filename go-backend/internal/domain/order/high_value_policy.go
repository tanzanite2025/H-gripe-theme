package order

import (
	"fmt"

	"commerce-platform/internal/domain/currency"
)

type HighValueOrderEvaluation struct {
	OrderTotalUSD float64
	IsHighValue   bool
}

// EvaluateHighValueOrder is the single order-total policy used by signature
// handling and the order evidence snapshot. It never inspects product data.
func EvaluateHighValueOrder(totalAmount float64, fxSnapshot currency.OrderFXSnapshot) (HighValueOrderEvaluation, error) {
	if currency.NormalizeCode(fxSnapshot.BaseCurrency) != currency.DefaultPrimaryCurrency {
		return HighValueOrderEvaluation{}, fmt.Errorf(
			"high-value order policy requires a %s FX base currency",
			currency.DefaultPrimaryCurrency,
		)
	}
	orderTotalUSD, err := fxSnapshot.OrderAmountToBase(totalAmount)
	if err != nil {
		return HighValueOrderEvaluation{}, err
	}
	return HighValueOrderEvaluation{
		OrderTotalUSD: orderTotalUSD,
		IsHighValue:   orderTotalUSD >= HighValueSignatureThresholdUSD,
	}, nil
}
