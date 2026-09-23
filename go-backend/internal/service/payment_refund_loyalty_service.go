package service

import (
	"errors"
	"fmt"
	"math/big"

	currencydomain "commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	orderdomain "commerce-platform/internal/domain/order"
	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"
)

const (
	refundLoyaltyClawbackSource         = "refund_loyalty_clawback"
	refundLoyaltyClawbackReversalSource = "refund_loyalty_clawback_reversal"
	refundLoyaltyPointsReturnSource     = "refund_loyalty_points_return"
	refundLoyaltyCashRecoveryDebtSource = "refund_loyalty_cash_recovery_debt"
)

// prepareRefundLoyaltySettlementInTx reserves the points that can be clawed
// back before a gateway call. The reservation is the final negative ledger
// entry; a failed gateway call reverses it in the same refund workflow.
func prepareRefundLoyaltySettlementInTx(
	repos repository.TxRepositories,
	orderRecord *orderdomain.Order,
	refund *paymentdomain.Refund,
) error {
	if refund == nil || isDuplicatePaidRefund(refund) || refund.LoyaltySettlementPrepared {
		return nil
	}
	if repos.Loyalty == nil || orderRecord == nil || orderRecord.UserID == 0 {
		refund.LoyaltySettlementPrepared = true
		return repos.Payment.UpdateRefund(refund)
	}

	currencyCode := currencydomain.NormalizeCode(orderRecord.Currency)
	if currencyCode == "" {
		currencyCode = currencydomain.DefaultPrimaryCurrency
	}

	currentRequestedMoney, err := refund.RequestedAmountMoney()
	if err != nil || currentRequestedMoney.AmountMinor() <= 0 || currentRequestedMoney.Currency().String() != currencyCode {
		// Historical pending rows may not have carried Currency yet. Resolve
		// their legacy projection only at this boundary, then keep all further
		// calculations in Money.
		currentRequestedMoney, err = refund.RequestedAmountMoney()
		if err != nil || currentRequestedMoney.AmountMinor() <= 0 {
			currentRequestedMoney, err = refund.AmountMoney()
		}
	}
	if err != nil {
		return fmt.Errorf("invalid current refund amount for loyalty settlement: %w", err)
	}
	if currentRequestedMoney.AmountMinor() <= 0 {
		return errors.New("refund amount must be greater than zero for loyalty settlement")
	}

	previousRequestedAmountMinor, err := repos.Payment.SumRefundRequestedAmountMinorByOrderID(orderRecord.ID, "pending", "completed")
	if err != nil {
		return err
	}
	previousRequestedMoney, err := domainmoney.New(previousRequestedAmountMinor, currencyCode)
	if err != nil {
		return fmt.Errorf("invalid previous refund amount for loyalty settlement: %w", err)
	}

	// The repository sum includes the refund currently being prepared. Remove
	// it using exact minor-unit arithmetic before calculating the cumulative
	// refund allocation. This avoids float subtraction drift at cent and
	// zero-decimal currency boundaries.
	previousRequestedMoney, err = previousRequestedMoney.Subtract(currentRequestedMoney)
	if err != nil {
		return fmt.Errorf("subtract current refund from previous requested amount: %w", err)
	}
	if previousRequestedMoney.AmountMinor() < 0 {
		previousRequestedMoney = zeroRefundMoney(currencyCode)
	}

	orderRefundBaseMoney, err := orderRecord.TotalMoney()
	if err != nil {
		return fmt.Errorf("invalid order refund base amount: %w", err)
	}
	if orderRefundBaseMoney.AmountMinor() <= 0 {
		return nil
	}
	cumulativeRequestedMoney, err := previousRequestedMoney.Add(currentRequestedMoney)
	if err != nil {
		return fmt.Errorf("add cumulative refund amount: %w", err)
	}
	cumulativeRequestedMoney = clampRefundMoneyValue(
		cumulativeRequestedMoney,
		zeroRefundMoney(currencyCode),
		orderRefundBaseMoney,
	)

	earnedPoints, err := repos.Loyalty.SumTransactionPointsByUserTypeSourceAndSourceID(
		orderRecord.UserID,
		"earn",
		"order",
		orderRecord.ID,
	)
	if err != nil {
		return err
	}
	if earnedPoints < 0 {
		earnedPoints = 0
	}
	// Only points earned by this order participate in the optional cash
	// recovery calculation. Referral credits (source=referral_referee),
	// referral reversals, and other account adjustments are intentionally
	// excluded: they are settled in their own ledger flow and must never
	// change the cash amount returned by the payment provider.

	previousSettledPoints, err := repos.Payment.SumRefundLoyaltyPointsSettledByOrderID(orderRecord.ID, "pending", "completed")
	if err != nil {
		return err
	}
	previousSettledPoints -= refund.LoyaltyPointsClawback + refund.LoyaltyPointsCashRecovered + refund.LoyaltyPointsDebt
	if previousSettledPoints < 0 {
		previousSettledPoints = 0
	}
	targetClawbackPoints := proportionalRefundPoints(earnedPoints, cumulativeRequestedMoney, orderRefundBaseMoney)
	pointsToReserve := targetClawbackPoints - previousSettledPoints
	if pointsToReserve < 0 {
		pointsToReserve = 0
	}

	previousReturnedPoints, err := repos.Payment.SumRefundLoyaltyPointsReturnedByOrderID(orderRecord.ID, "pending", "completed")
	if err != nil {
		return err
	}
	previousReturnedPoints -= refund.LoyaltyPointsReturned
	if previousReturnedPoints < 0 {
		previousReturnedPoints = 0
	}
	// Points are a separate tender from cash. A successful refund returns the
	// order's actual PointsUsed as a whole, regardless of the gateway amount or
	// whether the cash leg is refunded in multiple requests. Never prorate the
	// points against merchandise/cash amounts: that would couple the unified
	// points balance to refund-money allocation. The cumulative/previous values
	// below make this idempotent, so a later refund request cannot return the
	// same order points twice.
	targetReturnedPoints := orderRecord.PointsUsed
	if targetReturnedPoints < 0 {
		targetReturnedPoints = 0
	}
	pointsToReturn := targetReturnedPoints - previousReturnedPoints
	if pointsToReturn < 0 {
		pointsToReturn = 0
	}

	userLoyalty, err := repos.Loyalty.FindOrCreateUserLoyaltyForUpdate(orderRecord.UserID)
	if err != nil {
		return err
	}
	reservedPoints := pointsToReserve
	if reservedPoints > userLoyalty.AvailablePoints {
		reservedPoints = userLoyalty.AvailablePoints
	}
	if reservedPoints < 0 {
		reservedPoints = 0
	}
	missingPoints := pointsToReserve - reservedPoints

	var programConfigID *uint
	var cashDeductionMinor int64
	var cashDeductionMoney domainmoney.Money
	var loyaltyPointsCashRecovered int
	var loyaltyPointsDebt int
	refundMoneyForRecovery, refundMoneyErr := refund.AmountMoney()
	if refundMoneyErr != nil || refundMoneyForRecovery.Currency().String() != currencyCode {
		refundMoneyForRecovery, refundMoneyErr = refund.AmountMoney()
	}
	if refundMoneyErr != nil {
		return fmt.Errorf("invalid refund amount for loyalty settlement: %w", refundMoneyErr)
	}
	if missingPoints > 0 && refundMoneyForRecovery.AmountMinor() > 0 {
		if repos.Program == nil {
			return errors.New("loyalty program repository is required to value unrecovered refund points")
		}
		config, err := repos.Program.FindActive()
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return errors.New("active loyalty program is required to value unrecovered refund points")
			}
			return err
		}
		if config.ExchangeRatePoints <= 0 {
			return errors.New("loyalty exchange rate must be greater than zero")
		}
		programConfigID = &config.ID
		cashMajor := new(big.Rat).SetFrac(big.NewInt(int64(missingPoints)), big.NewInt(int64(config.ExchangeRatePoints)))
		cashMoney, err := domainmoney.FromMajorRat(cashMajor, currencyCode)
		if err != nil {
			return fmt.Errorf("invalid loyalty cash deduction: %w", err)
		}
		refundMoney := refundMoneyForRecovery
		if cashMoney.AmountMinor() >= refundMoney.AmountMinor() {
			cashDeductionMoney = refundMoney
			cashDeductionMinor = refundMoney.AmountMinor()
		} else {
			cashDeductionMoney = cashMoney
			cashDeductionMinor = cashMoney.AmountMinor()
		}
		loyaltyPointsCashRecovered = loyaltyPointsCoveredByCash(cashDeductionMoney, config.ExchangeRatePoints, missingPoints)
		loyaltyPointsDebt = missingPoints - loyaltyPointsCashRecovered
	}

	if reservedPoints > 0 {
		if _, err := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
			orderRecord.UserID,
			-reservedPoints,
			"refund",
			refundLoyaltyClawbackSource,
			refund.ID,
			fmt.Sprintf("Reserved %d earned points for refund #%d clawback", reservedPoints, refund.ID),
			programConfigID,
		); err != nil {
			return fmt.Errorf("failed to reserve refund loyalty clawback: %w", err)
		}
	}

	refund.LoyaltySettlementPrepared = true
	refund.LoyaltyPointsClawback = reservedPoints
	refund.LoyaltyPointsReturned = pointsToReturn
	refund.LoyaltyPointsCashRecovered = loyaltyPointsCashRecovered
	refund.LoyaltyPointsDebt = loyaltyPointsDebt
	refund.LoyaltyCashDeductionAmountMinor = cashDeductionMinor
	if cashDeductionMinor > 0 {
		refundMoney, err := refund.AmountMoney()
		if err != nil {
			return fmt.Errorf("invalid refund amount for loyalty deduction: %w", err)
		}
		remainingMoney, err := refundMoney.Subtract(cashDeductionMoney)
		if err != nil {
			return fmt.Errorf("subtract loyalty cash deduction: %w", err)
		}
		refund.AmountMinor = remainingMoney.AmountMinor()
	}
	return repos.Payment.UpdateRefund(refund)
}

func finalizeRefundLoyaltySettlementInTx(
	repos repository.TxRepositories,
	orderRecord *orderdomain.Order,
	refund *paymentdomain.Refund,
) error {
	if refund == nil || isDuplicatePaidRefund(refund) || orderRecord == nil || orderRecord.UserID == 0 || repos.Loyalty == nil {
		return nil
	}
	if !refund.LoyaltySettlementPrepared {
		if err := prepareRefundLoyaltySettlementInTx(repos, orderRecord, refund); err != nil {
			return err
		}
	}

	if refund.LoyaltyPointsClawback > 0 {
		count, err := repos.Loyalty.CountTransactionsByUserTypeSourceAndSourceID(
			orderRecord.UserID,
			"refund",
			refundLoyaltyClawbackSource,
			refund.ID,
		)
		if err != nil {
			return err
		}
		if count == 0 {
			return errors.New("refund loyalty clawback reservation is missing")
		}
	}

	if refund.LoyaltyPointsReturned > 0 {
		count, err := repos.Loyalty.CountTransactionsByUserTypeSourceAndSourceID(
			orderRecord.UserID,
			"refund",
			refundLoyaltyPointsReturnSource,
			refund.ID,
		)
		if err != nil {
			return err
		}
		if count == 0 {
			if _, err := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
				orderRecord.UserID,
				refund.LoyaltyPointsReturned,
				"refund",
				refundLoyaltyPointsReturnSource,
				refund.ID,
				fmt.Sprintf("Returned %d points used on refund #%d", refund.LoyaltyPointsReturned, refund.ID),
				nil,
			); err != nil {
				return fmt.Errorf("failed to return points used on refund: %w", err)
			}
		}
	}

	if refund.LoyaltyPointsDebt > 0 {
		count, err := repos.Loyalty.CountTransactionsByUserTypeSourceAndSourceID(
			orderRecord.UserID,
			"refund",
			refundLoyaltyCashRecoveryDebtSource,
			refund.ID,
		)
		if err != nil {
			return err
		}
		if count == 0 {
			if _, err := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
				orderRecord.UserID,
				-refund.LoyaltyPointsDebt,
				"refund",
				refundLoyaltyCashRecoveryDebtSource,
				refund.ID,
				fmt.Sprintf("Recorded %d unrecovered earned points as refund debt #%d", refund.LoyaltyPointsDebt, refund.ID),
				nil,
			); err != nil {
				return fmt.Errorf("failed to record refund loyalty points debt: %w", err)
			}
		}
	}

	return repos.Payment.UpdateRefund(refund)
}

func releaseRefundLoyaltyReservationInTx(
	repos repository.TxRepositories,
	orderRecord *orderdomain.Order,
	refund *paymentdomain.Refund,
) error {
	if refund == nil || isDuplicatePaidRefund(refund) || orderRecord == nil || orderRecord.UserID == 0 || repos.Loyalty == nil {
		return nil
	}
	if !refund.LoyaltySettlementPrepared {
		return nil
	}

	if refund.LoyaltyPointsClawback > 0 {
		count, err := repos.Loyalty.CountTransactionsByUserTypeSourceAndSourceID(
			orderRecord.UserID,
			"refund",
			refundLoyaltyClawbackReversalSource,
			refund.ID,
		)
		if err != nil {
			return err
		}
		if count == 0 {
			if _, err := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
				orderRecord.UserID,
				refund.LoyaltyPointsClawback,
				"refund",
				refundLoyaltyClawbackReversalSource,
				refund.ID,
				fmt.Sprintf("Released %d reserved refund points for refund #%d", refund.LoyaltyPointsClawback, refund.ID),
				nil,
			); err != nil {
				return fmt.Errorf("failed to release refund loyalty clawback: %w", err)
			}
		}
	}

	currencyCode := currencydomain.NormalizeCode(orderRecord.Currency)
	if currencyCode == "" {
		currencyCode = currencydomain.DefaultPrimaryCurrency
	}
	refundMoney, err := refund.AmountMoney()
	if err != nil || refundMoney.Currency().String() != currencyCode {
		refundMoney, err = refund.AmountMoney()
	}
	if err != nil {
		return fmt.Errorf("invalid refund amount for loyalty release: %w", err)
	}
	cashMoney, err := refund.LoyaltyCashDeductionMoney()
	if err != nil {
		return fmt.Errorf("invalid loyalty cash deduction for release: %w", err)
	}
	restoredMoney, err := refundMoney.Add(cashMoney)
	if err != nil {
		return fmt.Errorf("restore loyalty cash deduction: %w", err)
	}
	refund.AmountMinor = restoredMoney.AmountMinor()
	refund.LoyaltySettlementPrepared = false
	refund.LoyaltyPointsClawback = 0
	refund.LoyaltyPointsReturned = 0
	refund.LoyaltyPointsCashRecovered = 0
	refund.LoyaltyPointsDebt = 0
	refund.LoyaltyCashDeductionAmountMinor = 0
	return repos.Payment.UpdateRefund(refund)
}

func loyaltyPointsCoveredByCash(amount domainmoney.Money, exchangeRatePoints, maxPoints int) int {
	if amount.AmountMinor() <= 0 || exchangeRatePoints <= 0 || maxPoints <= 0 {
		return 0
	}
	minorUnits, ok := currencydomain.MinorUnits(amount.Currency().String())
	if !ok {
		return 0
	}
	scale := big.NewInt(1)
	for i := 0; i < minorUnits; i++ {
		scale.Mul(scale, big.NewInt(10))
	}
	covered := new(big.Int).Mul(big.NewInt(amount.AmountMinor()), big.NewInt(int64(exchangeRatePoints)))
	covered.Quo(covered, scale)
	max := big.NewInt(int64(maxPoints))
	if covered.Cmp(max) >= 0 {
		return maxPoints
	}
	return int(covered.Int64())
}

func proportionalRefundPoints(totalPoints int, refundedMoney, orderMoney domainmoney.Money) int {
	if totalPoints <= 0 || refundedMoney.AmountMinor() <= 0 || orderMoney.AmountMinor() <= 0 {
		return 0
	}
	if refundedMoney.Currency() != orderMoney.Currency() {
		return 0
	}
	if refundedMoney.AmountMinor() >= orderMoney.AmountMinor() {
		return totalPoints
	}
	product := new(big.Int).Mul(big.NewInt(int64(totalPoints)), big.NewInt(refundedMoney.AmountMinor()))
	quotient := new(big.Int).Quo(product, big.NewInt(orderMoney.AmountMinor()))
	if !quotient.IsInt64() {
		return totalPoints
	}
	points := int(quotient.Int64())
	if points < 0 {
		return 0
	}
	if points > totalPoints {
		return totalPoints
	}
	return points
}
