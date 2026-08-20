package service

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"commerce-platform/internal/domain/aftersales"
	"commerce-platform/internal/domain/currency"
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
	ProposedAmount float64
	Currency       string
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
	input.Currency = currency.NormalizeCode(input.Currency)
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
	maximumAmount, expectedCurrency := refundReviewLimit(caseRecord, orderRecord)
	amount, err := normalizeRefundReviewAmount(input.ProposedAmount, expectedCurrency)
	if err != nil {
		return nil, err
	}
	if input.Currency != expectedCurrency {
		return nil, ErrAfterSalesRefundReviewCurrencyInvalid
	}
	if amount > maximumAmount+0.000001 {
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
				CaseID:         input.CaseID,
				Status:         aftersales.RefundReviewStatusPending,
				ProposedAmount: amount,
				Currency:       expectedCurrency,
				RequestNotes:   input.RequestNotes,
				CreatedBy:      input.UpdatedBy,
				UpdatedBy:      input.UpdatedBy,
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

		existing.ProposedAmount = amount
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
		maximumAmount, expectedCurrency := refundReviewLimit(caseRecord, orderRecord)
		if !strings.EqualFold(expectedCurrency, review.Currency) {
			return ErrAfterSalesRefundReviewCurrencyInvalid
		}
		if review.ProposedAmount <= 0 || review.ProposedAmount > maximumAmount+0.01 {
			return ErrAfterSalesRefundReviewAmountExceeded
		}

		refund := &payment.Refund{
			OrderID:       caseRecord.OrderID,
			TransactionID: transaction.ID,
			Amount:        review.ProposedAmount,
			Reason:        afterSalesRefundDraftReason(caseRecord.ID, review),
		}
		if review.ProposedAmount+0.01 >= maximumAmount {
			refund.LineItems = afterSalesRefundLineItems(caseRecord)
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
			record.RefundReviewMaximumAmount, record.RefundReviewCurrency = refundReviewLimit(record, orderRecord)
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
) (float64, string) {
	if caseRecord == nil || orderRecord == nil || !aftersales.IsRefundReviewCaseType(caseRecord.Type) {
		return 0, ""
	}

	quantitiesByOrderItemID := make(map[uint]int, len(caseRecord.Items))
	for _, item := range caseRecord.Items {
		quantitiesByOrderItemID[item.OrderItemID] += item.Quantity
	}

	amount := 0.0
	for _, item := range orderRecord.Items {
		quantity := quantitiesByOrderItemID[item.ID]
		if quantity <= 0 || item.Quantity <= 0 {
			continue
		}
		amount += item.Total * float64(quantity) / float64(item.Quantity)
	}

	currencyCode := currency.NormalizeCode(orderRecord.Currency)
	return roundRefundReviewMoney(amount, currencyCode), currencyCode
}

func normalizeRefundReviewAmount(amount float64, currencyCode string) (float64, error) {
	if math.IsNaN(amount) || math.IsInf(amount, 0) || amount <= 0 {
		return 0, ErrAfterSalesRefundReviewAmountInvalid
	}
	if !currency.IsValidCode(currencyCode) || !currency.IsCatalogCode(currencyCode) {
		return 0, ErrAfterSalesRefundReviewCurrencyInvalid
	}

	rounded := roundRefundReviewMoney(amount, currencyCode)
	if math.Abs(amount-rounded) > 0.000001 {
		return 0, ErrAfterSalesRefundReviewAmountInvalid
	}
	return rounded, nil
}

func roundRefundReviewMoney(amount float64, currencyCode string) float64 {
	minorUnits, ok := currency.MinorUnits(currencyCode)
	if !ok {
		return amount
	}
	factor := math.Pow10(minorUnits)
	return math.Round(amount*factor) / factor
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
