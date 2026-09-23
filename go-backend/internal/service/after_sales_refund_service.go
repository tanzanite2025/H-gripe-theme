package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"
)

var (
	ErrAfterSalesRefundReviewNotFound         = errors.New("after-sales refund review not found")
	ErrAfterSalesRefundReviewUnavailable      = errors.New("after-sales case is not eligible for refund review")
	ErrAfterSalesRefundReviewAmountInvalid    = errors.New("proposed refund amount is invalid")
	ErrAfterSalesRefundReviewAmountExceeded   = errors.New("proposed refund amount exceeds selected item value")
	ErrAfterSalesRefundReviewCurrencyInvalid  = errors.New("refund review currency must match the order currency")
	ErrAfterSalesRefundReviewNotesRequired    = errors.New("refund review notes are required")
	ErrAfterSalesRefundReviewDecisionInvalid  = errors.New("invalid after-sales refund review decision")
	ErrAfterSalesRefundReviewFinalized        = errors.New("after-sales refund review is already finalized")
	ErrAfterSalesRefundReviewOperatorRequired = errors.New("admin user id is required for refund review")
	ErrAfterSalesRefundReviewNotApproved      = errors.New("after-sales refund review is not approved")
	ErrAfterSalesRefundTransactionNotFound    = errors.New("completed payment transaction not found")
)

type SaveAfterSalesRefundReviewInput struct {
	CaseID         uint
	ProposedAmount domainmoney.Money
	RequestNotes   string
	UpdatedBy      uint
}

type DecideAfterSalesRefundReviewInput struct {
	CaseID        uint
	Status        string
	DecisionNotes string
	ReviewedBy    uint
}

type CreateAfterSalesPendingRefundInput struct {
	CaseID  uint
	AdminID uint
}

func (s *AfterSalesService) GetRefundReview(caseID uint) (*aftersales.AfterSalesRefundReview, error) {
	if s == nil || s.refundReviewRepo == nil {
		return nil, errors.New("after-sales refund review service is not configured")
	}
	if caseID == 0 {
		return nil, ErrAfterSalesCaseNotFound
	}

	review, err := s.refundReviewRepo.FindByCaseID(caseID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrAfterSalesRefundReviewNotFound
		}
		return nil, err
	}
	s.populateRefundReviewOperatorNames(review)
	return review, nil
}

func (s *AfterSalesService) SaveRefundReview(
	input SaveAfterSalesRefundReviewInput,
) (*aftersales.AfterSalesRefundReview, error) {
	if s == nil || s.caseRepo == nil || s.orderRepo == nil || s.refundReviewRepo == nil {
		return nil, errors.New("after-sales refund review service is not configured")
	}
	if input.CaseID == 0 {
		return nil, ErrAfterSalesCaseNotFound
	}
	if input.UpdatedBy == 0 {
		return nil, ErrAfterSalesRefundReviewOperatorRequired
	}
	input.RequestNotes = strings.TrimSpace(input.RequestNotes)
	if input.RequestNotes == "" {
		return nil, ErrAfterSalesRefundReviewNotesRequired
	}

	caseRecord, err := s.GetCase(input.CaseID)
	if err != nil {
		return nil, err
	}
	if !refundReviewAvailable(caseRecord) {
		return nil, ErrAfterSalesRefundReviewUnavailable
	}
	orderRecord, err := s.orderRepo.FindByID(caseRecord.OrderID)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			return nil, ErrAfterSalesOrderNotFound
		}
		return nil, err
	}
	maximumAmount := refundReviewLimit(caseRecord, orderRecord)
	expectedCurrency := maximumAmount.Currency().String()
	if input.ProposedAmount.Validate() != nil || input.ProposedAmount.AmountMinor() <= 0 {
		return nil, ErrAfterSalesRefundReviewAmountInvalid
	}
	if input.ProposedAmount.Currency().String() != expectedCurrency {
		return nil, ErrAfterSalesRefundReviewCurrencyInvalid
	}
	if maximumAmount.Validate() != nil || input.ProposedAmount.AmountMinor() > maximumAmount.AmountMinor() {
		return nil, ErrAfterSalesRefundReviewAmountExceeded
	}

	var saved *aftersales.AfterSalesRefundReview
	err = s.refundReviewRepo.Transaction(func(txRepo *repository.AfterSalesRefundReviewRepository) error {
		lockedCase, err := txRepo.FindCaseByIDForUpdate(input.CaseID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrAfterSalesCaseNotFound
			}
			return err
		}
		if !refundReviewAvailable(lockedCase) {
			return ErrAfterSalesRefundReviewUnavailable
		}

		existing, err := txRepo.FindByCaseIDForUpdate(input.CaseID)
		if repository.IsRecordNotFound(err) {
			review := &aftersales.AfterSalesRefundReview{
				CaseID:              input.CaseID,
				Status:              aftersales.RefundReviewStatusPending,
				ProposedAmountMinor: input.ProposedAmount.AmountMinor(),
				Currency:            expectedCurrency,
				RequestNotes:        input.RequestNotes,
				CreatedBy:           input.UpdatedBy,
				UpdatedBy:           input.UpdatedBy,
			}
			if err := txRepo.Create(review); err != nil {
				return err
			}
			saved = review
			return nil
		}
		if err != nil {
			return err
		}
		if existing.Status != aftersales.RefundReviewStatusPending {
			return ErrAfterSalesRefundReviewFinalized
		}

		existing.ProposedAmountMinor = input.ProposedAmount.AmountMinor()
		existing.Currency = expectedCurrency
		existing.RequestNotes = input.RequestNotes
		existing.UpdatedBy = input.UpdatedBy
		if err := txRepo.Update(existing); err != nil {
			return err
		}
		saved = existing
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.populateRefundReviewOperatorNames(saved)
	return saved, nil
}

func (s *AfterSalesService) DecideRefundReview(
	input DecideAfterSalesRefundReviewInput,
) (*aftersales.AfterSalesRefundReview, error) {
	if s == nil || s.refundReviewRepo == nil {
		return nil, errors.New("after-sales refund review service is not configured")
	}
	if input.CaseID == 0 {
		return nil, ErrAfterSalesCaseNotFound
	}
	if input.ReviewedBy == 0 {
		return nil, ErrAfterSalesRefundReviewOperatorRequired
	}
	input.Status = strings.TrimSpace(input.Status)
	input.DecisionNotes = strings.TrimSpace(input.DecisionNotes)
	if !aftersales.IsRefundReviewDecisionStatus(input.Status) {
		return nil, ErrAfterSalesRefundReviewDecisionInvalid
	}
	if input.DecisionNotes == "" {
		return nil, ErrAfterSalesRefundReviewNotesRequired
	}

	var saved *aftersales.AfterSalesRefundReview
	err := s.refundReviewRepo.Transaction(func(txRepo *repository.AfterSalesRefundReviewRepository) error {
		lockedCase, err := txRepo.FindCaseByIDForUpdate(input.CaseID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrAfterSalesCaseNotFound
			}
			return err
		}
		if !refundReviewAvailable(lockedCase) {
			return ErrAfterSalesRefundReviewUnavailable
		}

		review, err := txRepo.FindByCaseIDForUpdate(input.CaseID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrAfterSalesRefundReviewNotFound
			}
			return err
		}
		if review.Status != aftersales.RefundReviewStatusPending {
			return ErrAfterSalesRefundReviewFinalized
		}

		now := time.Now().UTC()
		review.Status = input.Status
		review.DecisionNotes = input.DecisionNotes
		review.UpdatedBy = input.ReviewedBy
		review.ReviewedByID = &input.ReviewedBy
		review.ReviewedAt = &now
		if err := txRepo.Update(review); err != nil {
			return err
		}
		saved = review
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.populateRefundReviewOperatorNames(saved)
	return saved, nil
}

func (s *AfterSalesService) CreatePendingRefundFromApprovedReview(
	input CreateAfterSalesPendingRefundInput,
) (*aftersales.AfterSalesRefundReview, *payment.Refund, error) {
	if s == nil || s.refundReviewRepo == nil || s.txManager == nil {
		return nil, nil, errors.New("after-sales pending refund workflow is not configured")
	}
	if input.CaseID == 0 {
		return nil, nil, ErrAfterSalesCaseNotFound
	}
	if input.AdminID == 0 {
		return nil, nil, ErrAfterSalesRefundReviewOperatorRequired
	}

	var savedReview *aftersales.AfterSalesRefundReview
	var savedRefund *payment.Refund
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		if repos.AfterSalesRefund == nil {
			return errors.New("after-sales refund review repository is not configured for transactions")
		}

		caseRecord, err := repos.AfterSalesRefund.FindCaseByIDForUpdate(input.CaseID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrAfterSalesCaseNotFound
			}
			return err
		}
		if !refundReviewAvailable(caseRecord) {
			return ErrAfterSalesRefundReviewUnavailable
		}

		review, err := repos.AfterSalesRefund.FindByCaseIDForUpdate(input.CaseID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrAfterSalesRefundReviewNotFound
			}
			return err
		}
		if review.LinkedRefundID != nil && *review.LinkedRefundID > 0 {
			refund, err := repos.Payment.FindRefundByID(*review.LinkedRefundID)
			if err != nil {
				return fmt.Errorf("linked after-sales refund draft not found: %w", err)
			}
			savedReview = review
			savedRefund = refund
			return nil
		}
		if review.Status != aftersales.RefundReviewStatusApproved {
			return ErrAfterSalesRefundReviewNotApproved
		}

		transaction, err := repos.Payment.FindCompletedTransactionByOrderIDForUpdate(caseRecord.OrderID)
		if err != nil {
			if repository.IsRecordNotFound(err) {
				return ErrAfterSalesRefundTransactionNotFound
			}
			return err
		}
		if !strings.EqualFold(strings.TrimSpace(transaction.Currency), strings.TrimSpace(review.Currency)) {
			return ErrAfterSalesRefundReviewCurrencyInvalid
		}

		orderRecord, err := repos.Order.FindByIDForUpdateWithItems(caseRecord.OrderID)
		if err != nil {
			return normalizeOrderError(err)
		}
		maximumAmount := refundReviewLimit(caseRecord, orderRecord)
		expectedCurrency := maximumAmount.Currency().String()
		if !strings.EqualFold(expectedCurrency, review.Currency) {
			return ErrAfterSalesRefundReviewCurrencyInvalid
		}
		proposedMoney, amountErr := domainmoney.New(review.ProposedAmountMinor, review.Currency)
		if amountErr != nil || maximumAmount.Validate() != nil || proposedMoney.AmountMinor() <= 0 || proposedMoney.AmountMinor() > maximumAmount.AmountMinor() {
			return ErrAfterSalesRefundReviewAmountExceeded
		}

		refund := &payment.Refund{
			OrderID:       caseRecord.OrderID,
			TransactionID: transaction.ID,
			Currency:      transaction.Currency,
			AmountMinor:   review.ProposedAmountMinor,
			Reason:        afterSalesRefundDraftReason(caseRecord.ID, review),
		}
		if proposedMoney.AmountMinor() >= maximumAmount.AmountMinor() {
			refund.LineItems = afterSalesRefundLineItems(caseRecord)
			// Line item snapshots carry the pre-order-discount merchandise total.
			// Let the payment refund service derive the requested amount from those
			// snapshots so the order-level coupon clawback can be applied exactly
			// once. Keeping the review's post-discount amount here would make the
			// line-item amount consistency check reject every fully approved
			// order-level coupon refund (for example, 900 vs. 1000).
		}
		if err := createAdminRefundInTx(repos, refund, input.AdminID); err != nil {
			return err
		}

		review.LinkedRefundID = &refund.ID
		review.UpdatedBy = input.AdminID
		if err := repos.AfterSalesRefund.Update(review); err != nil {
			return err
		}

		savedReview = review
		savedRefund = refund
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	s.populateRefundReviewOperatorNames(savedReview)
	return savedReview, savedRefund, nil
}

func (s *AfterSalesService) populateRefundReviewDetails(record *aftersales.AfterSalesCase) {
	if record == nil || !aftersales.IsRefundReviewCaseType(record.Type) {
		return
	}
	if s != nil && s.orderRepo != nil {
		orderRecord, err := s.orderRepo.FindByID(record.OrderID)
		if err == nil {
			maximumAmount := refundReviewLimit(record, orderRecord)
			record.RefundReviewCurrency = maximumAmount.Currency().String()
			record.RefundReviewMaximumAmountMinor = maximumAmount.AmountMinor()
		}
	}
	s.populateRefundReviewOperatorNames(record.RefundReview)
}

func (s *AfterSalesService) populateRefundReviewOperatorNames(review *aftersales.AfterSalesRefundReview) {
	if review == nil {
		return
	}

	namesByID := map[uint]string{}
	ids := make([]uint, 0, 2)
	if review.CreatedBy > 0 {
		ids = append(ids, review.CreatedBy)
	}
	if review.ReviewedByID != nil && *review.ReviewedByID > 0 {
		ids = append(ids, *review.ReviewedByID)
	}
	if s != nil && s.userRepo != nil && len(ids) > 0 {
		users, err := s.userRepo.FindByIDs(ids)
		if err == nil {
			for _, operator := range users {
				namesByID[operator.ID] = afterSalesOperatorName(operator)
			}
		}
	}

	review.CreatorName = afterSalesOperatorLabel(review.CreatedBy, namesByID)
	if review.ReviewedByID != nil {
		review.ReviewerName = afterSalesOperatorLabel(*review.ReviewedByID, namesByID)
	}
}

func refundReviewAvailable(record *aftersales.AfterSalesCase) bool {
	return record != nil &&
		aftersales.IsRefundReviewCaseType(record.Type) &&
		record.Status == aftersales.StatusResolving
}

func refundReviewLimit(
	caseRecord *aftersales.AfterSalesCase,
	orderRecord *order.Order,
) domainmoney.Money {
	if caseRecord == nil || orderRecord == nil || !aftersales.IsRefundReviewCaseType(caseRecord.Type) {
		return domainmoney.Money{}
	}

	currencyCode := currency.NormalizeCode(orderRecord.Currency)
	quantitiesByOrderItemID := make(map[uint]int, len(caseRecord.Items))
	for _, item := range caseRecord.Items {
		quantitiesByOrderItemID[item.OrderItemID] += item.Quantity
	}

	amountMoney, err := domainmoney.New(0, currencyCode)
	if err != nil {
		return domainmoney.Money{}
	}
	for _, item := range orderRecord.Items {
		quantity := quantitiesByOrderItemID[item.ID]
		if quantity <= 0 || item.Quantity <= 0 {
			continue
		}
		lineMoney, lineErr := item.TotalMoney()
		if lineErr != nil {
			continue
		}
		if len(item.PricingSnapshotData) == 0 || strings.TrimSpace(string(item.PricingSnapshotData)) == "{}" {
			return domainmoney.Money{}
		}
		if lineMoney.Currency().String() != currencyCode {
			return domainmoney.Money{}
		}
		allocated, lineErr := lineMoney.MultiplyRatio(int64(quantity), int64(item.Quantity))
		if lineErr != nil {
			continue
		}
		amountMoney, err = amountMoney.Add(allocated)
		if err != nil {
			return domainmoney.Money{}
		}
	}

	// Order-level discounts are not stored on order items, so allocate them
	// across selected merchandise by the original order subtotal.
	orderSubtotalMoney, subtotalErr := orderRecord.SubtotalMoney()
	if subtotalErr != nil {
		return domainmoney.Money{}
	}
	if orderSubtotalMoney.AmountMinor() <= 0 {
		orderSubtotalMoney, subtotalErr = domainmoney.New(0, currencyCode)
		if subtotalErr != nil {
			return domainmoney.Money{}
		}
		for _, item := range orderRecord.Items {
			lineMoney, lineErr := item.SubtotalMoney()
			if lineErr == nil && lineMoney.AmountMinor() > 0 {
				orderSubtotalMoney, subtotalErr = orderSubtotalMoney.Add(lineMoney)
				continue
			}
			lineMoney, lineErr = item.TotalMoney()
			if lineErr == nil && lineMoney.AmountMinor() > 0 {
				orderSubtotalMoney, subtotalErr = orderSubtotalMoney.Add(lineMoney)
			}
		}
	}
	if subtotalErr != nil {
		return domainmoney.Money{}
	}
	orderDiscountMoney, discountErr := orderRecord.DiscountMoney()
	if amountMoney.AmountMinor() > 0 && orderSubtotalMoney.AmountMinor() > 0 && discountErr == nil && orderDiscountMoney.AmountMinor() > 0 {
		discountMoney := orderDiscountMoney
		if discountMoney.AmountMinor() > orderSubtotalMoney.AmountMinor() {
			discountMoney = orderSubtotalMoney
		}
		netSubtotalMoney, subtractErr := orderSubtotalMoney.Subtract(discountMoney)
		if subtractErr != nil {
			return domainmoney.Money{}
		}
		amountMoney, err = amountMoney.MultiplyRatio(netSubtotalMoney.AmountMinor(), orderSubtotalMoney.AmountMinor())
		if err != nil {
			return domainmoney.Money{}
		}
	}

	orderTotalMoney, totalErr := orderRecord.TotalMoney()
	if totalErr != nil || orderTotalMoney.AmountMinor() <= 0 {
		amountMoney, _ = domainmoney.New(0, currencyCode)
	} else if amountMoney.AmountMinor() > orderTotalMoney.AmountMinor() {
		amountMoney = orderTotalMoney
	}
	return amountMoney
}

func afterSalesRefundDraftReason(
	caseID uint,
	review *aftersales.AfterSalesRefundReview,
) string {
	note := ""
	if review != nil {
		note = strings.TrimSpace(review.DecisionNotes)
		if note == "" {
			note = strings.TrimSpace(review.RequestNotes)
		}
	}
	if note == "" {
		note = "approved by after-sales refund review"
	}
	return fmt.Sprintf("After-sales case #%d refund: %s", caseID, note)
}

func afterSalesRefundLineItems(
	caseRecord *aftersales.AfterSalesCase,
) []payment.RefundLineItem {
	if caseRecord == nil || len(caseRecord.Items) == 0 {
		return nil
	}
	lineItems := make([]payment.RefundLineItem, 0, len(caseRecord.Items))
	for _, item := range caseRecord.Items {
		lineItems = append(lineItems, payment.RefundLineItem{
			OrderID:     item.OrderID,
			OrderItemID: item.OrderItemID,
			ProductID:   item.ProductID,
			VariantID:   item.VariantID,
			ProductName: item.ProductName,
			SKU:         item.SKU,
			Quantity:    item.Quantity,
			Restock:     false,
		})
	}
	return lineItems
}
