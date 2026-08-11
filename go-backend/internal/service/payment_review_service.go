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
	if strings.TrimSpace(input.Source) == "" {
		input.Source = "operator"
	}
	if input.Status == "" {
		input.Status = "pending"
	}
	if !validPaymentReviewStatus(input.Status) {
		return nil, errors.New("invalid payment review status")
	}

	if strings.TrimSpace(input.PaymentIntentID) != "" && input.Status == "pending" {
		if existing, err := s.paymentRepo.FindPendingPaymentReviewByPaymentIntentID(input.PaymentIntentID); err == nil {
			changed := false
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
	if err := s.paymentRepo.CreatePaymentReview(record); err != nil {
		return nil, err
	}
	return record, nil
}

func (s *PaymentService) ResolveStripeReview(reviewID, paymentIntentID, closedReason string) error {
	var review *paymentdomain.PaymentReview
	var err error

	if strings.TrimSpace(reviewID) != "" {
		review, err = s.paymentRepo.FindPaymentReviewByStripeReviewID(reviewID)
	}
	if repository.IsRecordNotFound(err) || review == nil {
		if strings.TrimSpace(paymentIntentID) == "" {
			return nil
		}
		review, err = s.paymentRepo.FindPendingPaymentReviewByPaymentIntentID(paymentIntentID)
	}
	if repository.IsRecordNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if review.Status != "pending" {
		return nil
	}

	review.Status = stripeReviewClosedStatus(closedReason)
	if review.StripeReviewID == "" {
		review.StripeReviewID = reviewID
	}
	note := strings.TrimSpace(fmt.Sprintf("Stripe Radar review %s closed (%s).", reviewID, closedReason))
	if review.Notes == "" {
		review.Notes = note
	} else {
		review.Notes = strings.TrimSpace(review.Notes + "\n" + note)
	}
	now := time.Now()
	review.ReviewedAt = &now
	return s.paymentRepo.UpdatePaymentReview(review)
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
	record, err := s.GetPaymentReview(id)
	if err != nil {
		return nil, err
	}
	if !validPaymentReviewStatus(status) {
		return nil, errors.New("invalid payment review status")
	}
	if record.Status != "pending" && record.Status != status {
		return nil, errors.New("payment review is already finalized")
	}

	record.Status = status
	record.Notes = notes
	if status != "pending" {
		now := time.Now()
		record.ReviewedAt = &now
		record.ReviewedByID = &adminID
	}
	if err := s.paymentRepo.UpdatePaymentReview(record); err != nil {
		return nil, err
	}
	return record, nil
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
