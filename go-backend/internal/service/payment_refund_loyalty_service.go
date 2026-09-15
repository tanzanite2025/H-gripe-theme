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

	currentRequestedAmount := refund.RequestedAmount
	if currentRequestedAmount <= 0 {
		currentRequestedAmount = refund.Amount
	}
	currentRequestedMoney, err := parseRefundMoney(currentRequestedAmount, currencyCode)
	if err != nil {
		return fmt.Errorf("invalid current refund amount for loyalty settlement: %w", err)
	}
	if currentRequestedMoney.AmountMinor() <= 0 {
		return errors.New("refund amount must be greater than zero for loyalty settlement")
	}

	previousRequestedAmount, err := repos.Payment.SumRefundRequestedAmountByOrderID(orderRecord.ID, "pending", "completed")
	if err != nil {
		return err
	}
	previousRequestedMoney, err := parseRefundMoney(previousRequestedAmount, currencyCode)
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

	orderRefundBaseMoney, err := parseRefundMoney(orderRecord.TotalAmount, currencyCode)
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

	previousClawbackPoints, err := repos.Payment.SumRefundLoyaltyPointsClawbackByOrderID(orderRecord.ID, "pending", "completed")
	if err != nil {
		return err
	}
	previousClawbackPoints -= refund.LoyaltyPointsClawback
	if previousClawbackPoints < 0 {
		previousClawbackPoints = 0
	}
	targetClawbackPoints := proportionalRefundPoints(earnedPoints, cumulativeRequestedMoney, orderRefundBaseMoney)
	pointsToReserve := targetClawbackPoints - previousClawbackPoints
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
	// A refund that covers the entire cash amount actually paid for the order
	// returns every point used on that order. Points are a separate tender and
	// must not be prorated against the merchandise total when the cash leg is
	// fully refunded. Partial cash refunds retain the proportional fallback.
	orderCashMoney, cashErr := orderRecord.PaymentMoney()
	if cashErr != nil || currencydomain.NormalizeCode(string(orderCashMoney.Currency())) != currencyCode || orderCashMoney.AmountMinor() <= 0 {
		orderCashMoney = orderRefundBaseMoney
	}
	var targetReturnedPoints int
	if cumulativeRequestedMoney.AmountMinor() >= orderCashMoney.AmountMinor() {
		targetReturnedPoints = orderRecord.PointsUsed
	} else {
		targetReturnedPoints = proportionalRefundPoints(orderRecord.PointsUsed, cumulativeRequestedMoney, orderCashMoney)
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
	cashDeduction := 0.0
	var cashDeductionMoney domainmoney.Money
	if missingPoints > 0 && refund.Amount > 0 {
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
		refundMoney, err := parseRefundMoney(refund.Amount, currencyCode)
		if err != nil {
			return fmt.Errorf("invalid refund amount for loyalty settlement: %w", err)
		}
		if cashMoney.AmountMinor() >= refundMoney.AmountMinor() {
			cashAmount, _ := cashMoney.MajorFloat()
			refundAmount, _ := refundMoney.MajorFloat()
			return fmt.Errorf(
				"loyalty cash recovery %.2f consumes the full gateway refund %.2f",
				cashAmount,
				refundAmount,
			)
		}
		cashDeduction, err = cashMoney.MajorFloat()
		if err != nil {
			return fmt.Errorf("format loyalty cash deduction: %w", err)
		}
		cashDeductionMoney = cashMoney
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
	refund.LoyaltyCashDeductionAmount = cashDeduction
	if cashDeduction > 0 {
		refundMoney, err := parseRefundMoney(refund.Amount, currencyCode)
		if err != nil {
			return fmt.Errorf("invalid refund amount for loyalty deduction: %w", err)
		}
		remainingMoney, err := refundMoney.Subtract(cashDeductionMoney)
		if err != nil {
			return fmt.Errorf("subtract loyalty cash deduction: %w", err)
		}
		refund.Amount, err = remainingMoney.MajorFloat()
		if err != nil {
			return fmt.Errorf("format loyalty-adjusted refund amount: %w", err)
		}
		if refund.Amount <= 0 {
			return errors.New("loyalty settlement leaves no gateway refund amount")
		}
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
	refundMoney, err := parseRefundMoney(refund.Amount, currencyCode)
	if err != nil {
		return fmt.Errorf("invalid refund amount for loyalty release: %w", err)
	}
	cashMoney, err := parseRefundMoney(refund.LoyaltyCashDeductionAmount, currencyCode)
	if err != nil {
		return fmt.Errorf("invalid loyalty cash deduction for release: %w", err)
	}
	restoredMoney, err := refundMoney.Add(cashMoney)
	if err != nil {
		return fmt.Errorf("restore loyalty cash deduction: %w", err)
	}
	refund.Amount, err = restoredMoney.MajorFloat()
	if err != nil {
		return fmt.Errorf("format restored refund amount: %w", err)
	}
	refund.LoyaltySettlementPrepared = false
	refund.LoyaltyPointsClawback = 0
	refund.LoyaltyPointsReturned = 0
	refund.LoyaltyCashDeductionAmount = 0
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
