package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	paymentdomain "commerce-platform/internal/domain/payment"
	"commerce-platform/internal/repository"
)

var ErrPaymentReviewNotFound = errors.New("payment review not found")

const (
	stripeRequiresActionReviewReason = "stripe_requires_action"
	stripeRadarReviewReason          = "stripe_review_opened"
)

type CreatePaymentReviewInput struct {
	OrderID         *uint
	TransactionID   *uint
	DisputeID       *uint
	PaymentIntentID string
	StripeReviewID  string
	Status          string
	Reason          string
	Source          string
	Notes           string
	AssignedToID    *uint
}

func (s *PaymentService) CreatePaymentReview(input CreatePaymentReviewInput) (*paymentdomain.PaymentReview, error) {
	if strings.TrimSpace(input.Reason) == "" {
		return nil, errors.New("review reason is required")
	}
	input.PaymentIntentID = strings.TrimSpace(input.PaymentIntentID)
	input.StripeReviewID = strings.TrimSpace(input.StripeReviewID)
	if strings.TrimSpace(input.Source) == "" {
		input.Source = "operator"
	}
	if input.Status == "" {
		input.Status = "pending"
	}
	if !validPaymentReviewStatus(input.Status) {
		return nil, errors.New("invalid payment review status")
	}

	if strings.TrimSpace(input.PaymentIntentID) != "" && input.Status == "pending" &&
		strings.EqualFold(strings.TrimSpace(input.Source), "radar") {
		if existing, err := s.paymentRepo.FindPendingPaymentReviewByPaymentIntentID(input.PaymentIntentID); err == nil {
			if isStripeAutomaticReview(existing) {
				changed := false
				if existing.OrderID == nil && input.OrderID != nil {
					existing.OrderID = input.OrderID
					changed = true
				}
				if existing.TransactionID == nil && input.TransactionID != nil {
					existing.TransactionID = input.TransactionID
					changed = true
				}
				if existing.StripeReviewID == "" && strings.TrimSpace(input.StripeReviewID) != "" {
					existing.StripeReviewID = strings.TrimSpace(input.StripeReviewID)
					changed = true
				}
				if existing.Notes == "" && strings.TrimSpace(input.Notes) != "" {
					existing.Notes = input.Notes
					changed = true
				}
				if changed {
					if err := s.paymentRepo.UpdatePaymentReview(existing); err != nil {
						return nil, err
					}
				}
				return existing, nil
			}
		} else if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}

	record := &paymentdomain.PaymentReview{
		OrderID:         input.OrderID,
		TransactionID:   input.TransactionID,
		DisputeID:       input.DisputeID,
		PaymentIntentID: input.PaymentIntentID,
		StripeReviewID:  input.StripeReviewID,
		Status:          input.Status,
		Reason:          input.Reason,
		Source:          input.Source,
		Notes:           input.Notes,
		AssignedToID:    input.AssignedToID,
	}
	if err := s.attachPaymentReviewOrderReference(record); err != nil {
		return nil, err
	}
	if err := s.paymentRepo.CreatePaymentReview(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *PaymentService) ResolveStripeReview(reviewID, paymentIntentID, closedReason string) error {
	reviewID = strings.TrimSpace(reviewID)
	paymentIntentID = strings.TrimSpace(paymentIntentID)
	note := strings.TrimSpace(fmt.Sprintf("Stripe Radar review %s closed (%s).", reviewID, closedReason))
	return s.resolveStripeAutomaticReview(
		reviewID,
		paymentIntentID,
		[]string{stripeRadarReviewReason, stripeRequiresActionReviewReason},
		stripeReviewClosedStatus(closedReason),
		note,
	)
}

// ResolveStripeRequiresActionReview clears the automatic review created while
// Stripe was waiting for customer authentication. It deliberately does not
// touch other reviews for the same PaymentIntent.
func (s *PaymentService) ResolveStripeRequiresActionReview(paymentIntentID string) error {
	paymentIntentID = strings.TrimSpace(paymentIntentID)
	if paymentIntentID == "" {
		return nil
	}
	return s.resolveStripeAutomaticReview(
		"",
		paymentIntentID,
		[]string{stripeRequiresActionReviewReason},
		"approved",
		"Stripe 3DS authentication completed successfully.",
	)
}

func (s *PaymentService) resolveStripeAutomaticReview(
	reviewID string,
	paymentIntentID string,
	reasons []string,
	status string,
	note string,
) error {
	if s.txManager == nil {
		return errors.New("payment review transaction is not configured")
	}

	return s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		review, err := findStripeAutomaticPaymentReview(repos.Payment, reviewID, paymentIntentID, reasons)
		if repository.IsRecordNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if review == nil || review.Status != "pending" || !isStripeAutomaticReview(review) {
			return nil
		}

		orderID, transactionID, err := paymentReviewOrderReference(repos.Payment, review)
		if err != nil {
			return err
		}
		if orderID != nil {
			if _, err := repos.Order.FindByIDForUpdate(*orderID); err != nil {
				return err
			}
		}
		lockedReview, err := repos.Payment.FindPaymentReviewByIDForUpdate(review.ID)
		if repository.IsRecordNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if lockedReview.Status != "pending" || !isStripeAutomaticReview(lockedReview) {
			return nil
		}
		if orderID != nil {
			if lockedReview.OrderID != nil && *lockedReview.OrderID != *orderID {
				return errors.New("payment review order changed while it was being resolved")
			}
			if lockedReview.OrderID == nil {
				lockedReview.OrderID = orderID
			}
			if lockedReview.TransactionID == nil && transactionID != nil {
				lockedReview.TransactionID = transactionID
			}
		}
		if lockedReview.StripeReviewID == "" {
			lockedReview.StripeReviewID = reviewID
		}
		if strings.TrimSpace(lockedReview.Notes) == "" {
			lockedReview.Notes = note
		} else {
			lockedReview.Notes = strings.TrimSpace(lockedReview.Notes + "\n" + note)
		}
		lockedReview.Status = status
		now := time.Now().UTC()
		lockedReview.ReviewedAt = &now
		if err := repos.Payment.UpdatePaymentReview(lockedReview); err != nil {
			return err
		}
		if lockedReview.OrderID == nil {
			return nil
		}
		return releasePaymentLiabilityReviewHoldIfClear(repos, *lockedReview.OrderID)
	})
}

func findStripeAutomaticPaymentReview(
	repo *repository.PaymentRepository,
	reviewID string,
	paymentIntentID string,
	reasons []string,
) (*paymentdomain.PaymentReview, error) {
	if reviewID != "" {
		review, err := repo.FindPaymentReviewByStripeReviewID(reviewID)
		if err == nil {
			return review, nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}
	if paymentIntentID == "" {
		return nil, repository.ErrRecordNotFound
	}
	for _, reason := range reasons {
		review, err := repo.FindPendingPaymentReviewByPaymentIntentIDAndReason(paymentIntentID, reason)
		if err == nil {
			return review, nil
		}
		if !repository.IsRecordNotFound(err) {
			return nil, err
		}
	}
	return nil, repository.ErrRecordNotFound
}

func paymentReviewOrderReference(
	repo *repository.PaymentRepository,
	review *paymentdomain.PaymentReview,
) (*uint, *uint, error) {
	if review == nil {
		return nil, nil, nil
	}
	if review.OrderID != nil {
		return review.OrderID, review.TransactionID, nil
	}
	if review.TransactionID != nil {
		transaction, err := repo.FindTransactionByID(*review.TransactionID)
		if repository.IsRecordNotFound(err) {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		orderID := transaction.OrderID
		transactionID := transaction.ID
		return &orderID, &transactionID, nil
	}
	if strings.TrimSpace(review.PaymentIntentID) == "" {
		return nil, nil, nil
	}
	transaction, err := repo.FindTransactionByTransactionID(review.PaymentIntentID)
	if repository.IsRecordNotFound(err) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	orderID := transaction.OrderID
	transactionID := transaction.ID
	return &orderID, &transactionID, nil
}

func (s *PaymentService) attachPaymentReviewOrderReference(review *paymentdomain.PaymentReview) error {
	if s == nil || s.paymentRepo == nil || review == nil || review.OrderID != nil {
		return nil
	}
	orderID, transactionID, err := paymentReviewOrderReference(s.paymentRepo, review)
	if err != nil {
		return err
	}
	review.OrderID = orderID
	if review.TransactionID == nil {
		review.TransactionID = transactionID
	}
	return nil
}

func (s *PaymentService) GetPaymentReview(id uint) (*paymentdomain.PaymentReview, error) {
	record, err := s.paymentRepo.FindPaymentReviewByID(id)
	if repository.IsRecordNotFound(err) {
		return nil, ErrPaymentReviewNotFound
	}
	return record, err
}

func (s *PaymentService) ListPaymentReviews(status string, page, pageSize int) ([]paymentdomain.PaymentReview, int64, error) {
	return s.paymentRepo.ListPaymentReviews(status, page, pageSize)
}

func (s *PaymentService) UpdatePaymentReview(id uint, status, notes string, adminID uint) (*paymentdomain.PaymentReview, error) {
	if !validPaymentReviewStatus(status) {
		return nil, errors.New("invalid payment review status")
	}

	if s.txManager == nil {
		return nil, errors.New("payment review transaction is not configured")
	}

	var record *paymentdomain.PaymentReview
	err := s.txManager.WithinTx(func(repos repository.TxRepositories) error {
		var err error
		reviewMeta, err := repos.Payment.FindPaymentReviewByID(id)
		if repository.IsRecordNotFound(err) {
			return ErrPaymentReviewNotFound
		}
		if err != nil {
			return err
		}
		lateReview := isLatePaymentReviewReason(reviewMeta.Reason)
		if (reviewMeta.Reason == highValueLiabilityReviewReason || lateReview) && reviewMeta.OrderID != nil {
			if _, err := repos.Order.FindByIDForUpdate(*reviewMeta.OrderID); err != nil {
				return err
			}
		}
		record, err = repos.Payment.FindPaymentReviewByIDForUpdate(id)
		if repository.IsRecordNotFound(err) {
			return ErrPaymentReviewNotFound
		}
		if err != nil {
			return err
		}
		if reviewMeta.OrderID != nil && (record.OrderID == nil || *reviewMeta.OrderID != *record.OrderID) {
			return errors.New("payment review order changed while it was being updated")
		}
		if record.Status != "pending" && record.Status != status {
			return errors.New("payment review is already finalized")
		}

		record.Status = status
		record.Notes = notes
		if status != "pending" {
			now := time.Now().UTC()
			record.ReviewedAt = &now
			record.ReviewedByID = &adminID
		}
		if err := repos.Payment.UpdatePaymentReview(record); err != nil {
			return err
		}
		if status == "approved" && lateReview && record.OrderID != nil {
			transactionID := record.TransactionID
			if transactionID == nil && strings.TrimSpace(record.PaymentIntentID) != "" {
				transaction, lookupErr := repos.Payment.FindTransactionByTransactionIDForUpdate(record.PaymentIntentID)
				if lookupErr != nil {
					return lookupErr
				}
				transactionID = &transaction.ID
			}
			if transactionID == nil {
				return errors.New("late payment review is missing transaction reference")
			}
			transaction, err := repos.Payment.FindTransactionByIDForUpdate(*transactionID)
			if err != nil {
				return err
			}
			orderRecord, err := repos.Order.FindByIDForUpdate(*record.OrderID)
			if err != nil {
				return err
			}
			if err := createLatePaymentRefundInTx(repos, orderRecord, transaction, latePaymentRefundReason(record.Reason)); err != nil {
				return err
			}
		}

		if status != "approved" ||
			record.Reason != highValueLiabilityReviewReason ||
			record.OrderID == nil {
			return nil
		}

		return releasePaymentLiabilityReviewHoldIfClear(repos, *record.OrderID)
	})
	if err != nil {
		return nil, err
	}
	return record, nil
}

func releasePaymentLiabilityReviewHoldIfClear(repos repository.TxRepositories, orderID uint) error {
	approvedLiabilityReview, err := repos.Payment.HasApprovedPaymentReviewByOrderIDAndReason(
		orderID,
		highValueLiabilityReviewReason,
	)
	if err != nil {
		return err
	}
	if !approvedLiabilityReview {
		return nil
	}
	activePaymentReview, err := repos.Payment.HasActivePaymentReviewByOrderID(orderID)
	if err != nil {
		return err
	}
	activeStripeDispute, err := repos.Payment.HasActiveStripeDisputeByOrderID(orderID)
	if err != nil {
		return err
	}
	activePayPalDispute, err := repos.Payment.HasActivePayPalDisputeByOrderID(orderID)
	if err != nil {
		return err
	}
	if activePaymentReview || activeStripeDispute || activePayPalDispute {
		return nil
	}
	return repos.Order.ReleasePaymentLiabilityReviewHold(orderID)
}

func isStripeAutomaticReview(review *paymentdomain.PaymentReview) bool {
	if review == nil || !strings.EqualFold(strings.TrimSpace(review.Source), "radar") {
		return false
	}
	switch strings.TrimSpace(review.Reason) {
	case stripeRequiresActionReviewReason, stripeRadarReviewReason:
		return true
	default:
		return false
	}
}

func isLatePaymentReviewReason(reason string) bool {
	switch strings.TrimSpace(reason) {
	case "payment_succeeded_after_cancellation", "payment_succeeded_after_expiration", "payment_succeeded_after_refund":
		return true
	default:
		return false
	}
}

func latePaymentRefundReason(reason string) string {
	return "late_payment_refund:" + strings.TrimSpace(reason)
}

func validPaymentReviewStatus(value string) bool {
	switch value {
	case "pending", "approved", "rejected", "cancelled":
		return true
	default:
		return false
	}
}

func stripeReviewClosedStatus(reason string) string {
	switch strings.ToLower(strings.TrimSpace(reason)) {
	case "approved":
		return "approved"
	case "refunded", "canceled", "cancelled", "disputed":
		return "rejected"
	default:
		return "cancelled"
	}
}
