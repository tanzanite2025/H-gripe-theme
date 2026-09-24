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
	previousSettledPoints -= refund.LoyaltyPointsClawback + refund.LoyaltyPointsDebt
	if previousSettledPoints < 0 {
		previousSettledPoints = 0
	}
	targetClawbackPoints := proportionalRefundPoints(earnedPoints, cumulativeRequestedMoney, orderRefundBaseMoney)
	pointsToReserve := targetClawbackPoints - previousSettledPoints
	if pointsToReserve < 0 {
		pointsToReserve = 0
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

	// Earned points are a loyalty balance, never a payment tender. If some
	// earned points were already spent, record the unrecovered amount as a
	// points debt; do not convert it into a cash deduction or consult an
	// exchange-rate configuration.
	var loyaltyPointsDebt = missingPoints

	if reservedPoints > 0 {
		if _, err := repos.Loyalty.AdjustUserPointsInCurrentTxWithConfig(
			orderRecord.UserID,
			-reservedPoints,
			"refund",
			refundLoyaltyClawbackSource,
			refund.ID,
			fmt.Sprintf("Reserved %d earned points for refund #%d clawback", reservedPoints, refund.ID),
			nil,
		); err != nil {
			return fmt.Errorf("failed to reserve refund loyalty clawback: %w", err)
		}
	}

	refund.LoyaltySettlementPrepared = true
	refund.LoyaltyPointsClawback = reservedPoints
	refund.LoyaltyPointsDebt = loyaltyPointsDebt
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
	refund.LoyaltySettlementPrepared = false
	refund.LoyaltyPointsClawback = 0
	refund.LoyaltyPointsDebt = 0
	return repos.Payment.UpdateRefund(refund)
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
