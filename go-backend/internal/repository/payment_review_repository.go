package repository

import "commerce-platform/internal/domain/payment"

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

func (r *PaymentRepository) FindPendingPaymentReviewByPaymentIntentID(paymentIntentID string) (*payment.PaymentReview, error) {
	var review payment.PaymentReview
	err := r.db.Where("payment_intent_id = ? AND status = ?", paymentIntentID, "pending").
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
