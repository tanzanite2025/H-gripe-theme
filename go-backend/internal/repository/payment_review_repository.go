package repository

import (
	"strings"
	"time"

	"commerce-platform/internal/domain/payment"
)

func (r *PaymentRepository) CreatePaymentReview(review *payment.PaymentReview) error {
	return r.db.Create(review).Error
}

func (r *PaymentRepository) FindPaymentReviewByID(id uint) (*payment.PaymentReview, error) {
	var review payment.PaymentReview
	err := r.db.First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *PaymentRepository) FindPendingPaymentReviewByOrderID(orderID uint) (*payment.PaymentReview, error) {
	var review payment.PaymentReview
	err := r.db.Where("order_id = ? AND status = ?", orderID, "pending").
		Order("created_at DESC").First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *PaymentRepository) FindPendingPaymentReviewByOrderIDAndReason(orderID uint, reason string) (*payment.PaymentReview, error) {
	var review payment.PaymentReview
	err := r.db.Where("order_id = ? AND reason = ? AND status = ?", orderID, reason, "pending").
		Order("created_at DESC").First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *PaymentRepository) HasApprovedPaymentReviewByOrderIDAndReason(orderID uint, reason string) (bool, error) {
	var review payment.PaymentReview
	err := r.db.Select("id").
		Where("order_id = ? AND reason = ? AND LOWER(status) = ?", orderID, reason, "approved").
		Limit(1).
		First(&review).Error
	if IsRecordNotFound(err) {
		return false, nil
	}
	return err == nil, err
}

func (r *PaymentRepository) FindPaymentReviewByIDForUpdate(id uint) (*payment.PaymentReview, error) {
	var review payment.PaymentReview
	err := r.lockForUpdate(r.db).First(&review, id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *PaymentRepository) FindPendingPaymentReviewByOrderIDAndReasonForUpdate(orderID uint, reason string) (*payment.PaymentReview, error) {
	var review payment.PaymentReview
	err := r.lockForUpdate(r.db).
		Where("order_id = ? AND reason = ? AND status = ?", orderID, reason, "pending").
		Order("created_at DESC").
		First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *PaymentRepository) HasActivePaymentReviewByOrderID(orderID uint) (bool, error) {
	var review payment.PaymentReview
	err := r.db.Select("id").
		Where("order_id = ? AND LOWER(status) NOT IN ?", orderID, []string{"approved", "rejected", "cancelled"}).
		Limit(1).
		First(&review).Error
	if IsRecordNotFound(err) {
		return false, nil
	}
	return err == nil, err
}

func (r *PaymentRepository) FindPendingPaymentReviewByPaymentIntentID(paymentIntentID string) (*payment.PaymentReview, error) {
	var review payment.PaymentReview
	err := r.db.Where("payment_intent_id = ? AND status = ?", paymentIntentID, "pending").
		Order("created_at DESC").First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *PaymentRepository) FindPendingPaymentReviewByPaymentIntentIDAndReason(paymentIntentID, reason string) (*payment.PaymentReview, error) {
	var review payment.PaymentReview
	err := r.db.Where("payment_intent_id = ? AND reason = ? AND status = ?", paymentIntentID, reason, "pending").
		Order("created_at DESC").First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *PaymentRepository) FindPaymentReviewByStripeReviewID(stripeReviewID string) (*payment.PaymentReview, error) {
	var review payment.PaymentReview
	err := r.db.Where("stripe_review_id = ?", stripeReviewID).First(&review).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *PaymentRepository) ListPaymentReviews(status string, page, pageSize int) ([]payment.PaymentReview, int64, error) {
	var reviews []payment.PaymentReview
	var total int64
	query := r.db.Model(&payment.PaymentReview{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reviews).Error
	return reviews, total, err
}

func (r *PaymentRepository) UpdatePaymentReview(review *payment.PaymentReview) error {
	return r.db.Save(review).Error
}

// ResolvePendingDisputeReview finalizes only the review belonging to the
// provider dispute being resolved.
func (r *PaymentRepository) ResolvePendingDisputeReview(disputeID uint, status, note string) error {
	if disputeID == 0 {
		return nil
	}
	var review payment.PaymentReview
	err := r.lockForUpdate(r.db).
		Where("dispute_id = ? AND source = ? AND status = ?", disputeID, "dispute", "pending").
		First(&review).Error
	if IsRecordNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}
	review.Status = strings.TrimSpace(status)
	if review.Status == "" {
		review.Status = "cancelled"
	}
	if strings.TrimSpace(note) != "" {
		if strings.TrimSpace(review.Notes) == "" {
			review.Notes = strings.TrimSpace(note)
		} else {
			review.Notes = strings.TrimSpace(review.Notes + "\n" + strings.TrimSpace(note))
		}
	}
	now := time.Now().UTC()
	review.ReviewedAt = &now
	return r.db.Model(&payment.PaymentReview{}).Where("id = ?", review.ID).Updates(map[string]interface{}{
		"status":      review.Status,
		"notes":       review.Notes,
		"reviewed_at": review.ReviewedAt,
	}).Error
}
