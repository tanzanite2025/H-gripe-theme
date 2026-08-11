package repository

import (
	"time"

	"commerce-platform/internal/domain/payment"

	"gorm.io/gorm"
)

func (r *PaymentRepository) UpsertStripeDispute(dispute *payment.StripeDispute) error {
	var existing payment.StripeDispute
	err := r.db.Where("stripe_dispute_id = ?", dispute.StripeDisputeID).First(&existing).Error
	if err == nil {
		dispute.ID = existing.ID
		return r.db.Session(&gorm.Session{SkipHooks: true}).Model(&payment.StripeDispute{}).
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"stripe_charge_id":  dispute.StripeChargeID,
				"payment_intent_id": dispute.PaymentIntentID,
				"order_id":          dispute.OrderID,
				"transaction_id":    dispute.TransactionID,
				"amount":            dispute.Amount,
				"currency":          dispute.Currency,
				"reason":            dispute.Reason,
				"status":            dispute.Status,
				"evidence_due_at":   dispute.EvidenceDueAt,
				"raw_payload":       dispute.RawPayload,
				"updated_at":        time.Now(),
			}).Error
	}
	if !IsRecordNotFound(err) {
		return err
	}
	return r.db.Create(dispute).Error
}

func (r *PaymentRepository) FindStripeDisputeByID(id uint) (*payment.StripeDispute, error) {
	var dispute payment.StripeDispute
	err := r.db.First(&dispute, id).Error
	if err != nil {
		return nil, err
	}
	return &dispute, nil
}

func (r *PaymentRepository) FindStripeDisputeByStripeID(stripeID string) (*payment.StripeDispute, error) {
	var dispute payment.StripeDispute
	err := r.db.Where("stripe_dispute_id = ?", stripeID).First(&dispute).Error
	if err != nil {
		return nil, err
	}
	return &dispute, nil
}

func (r *PaymentRepository) ListStripeDisputes(status string, page, pageSize int) ([]payment.StripeDispute, int64, error) {
	var disputes []payment.StripeDispute
	var total int64
	query := r.db.Model(&payment.StripeDispute{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("evidence_due_at ASC NULLS LAST, created_at DESC").
		Offset(offset).Limit(pageSize).Find(&disputes).Error
	return disputes, total, err
}

func (r *PaymentRepository) UpdateStripeDisputeEvidenceSubmission(id uint, submittedAt *time.Time, payload, errorMessage, status string) error {
	updates := map[string]interface{}{
		"evidence_submission_payload": payload,
		"evidence_submission_error":   errorMessage,
		"updated_at":                  time.Now(),
	}
	if submittedAt != nil {
		updates["evidence_submitted_at"] = submittedAt
	}
	if status != "" {
		updates["status"] = status
	}
	return r.db.Session(&gorm.Session{SkipHooks: true}).Model(&payment.StripeDispute{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PaymentRepository) UpsertPayPalDispute(dispute *payment.PayPalDispute) error {
	var existing payment.PayPalDispute
	err := r.db.Where("paypal_dispute_id = ?", dispute.PayPalDisputeID).First(&existing).Error
	if err == nil {
		dispute.ID = existing.ID
		if dispute.OrderID == nil {
			dispute.OrderID = existing.OrderID
		}
		if dispute.TransactionID == nil {
			dispute.TransactionID = existing.TransactionID
		}
		if dispute.ProviderPaymentID == "" {
			dispute.ProviderPaymentID = existing.ProviderPaymentID
		}
		return r.db.Session(&gorm.Session{SkipHooks: true}).Model(&payment.PayPalDispute{}).
			Where("id = ?", existing.ID).
			Updates(map[string]interface{}{
				"order_id":                 dispute.OrderID,
				"transaction_id":           dispute.TransactionID,
				"provider_payment_id":      dispute.ProviderPaymentID,
				"amount":                   dispute.Amount,
				"currency":                 dispute.Currency,
				"reason":                   dispute.Reason,
				"status":                   dispute.Status,
				"dispute_state":            dispute.DisputeState,
				"dispute_life_cycle_stage": dispute.DisputeLifeCycleStage,
				"raw_payload":              dispute.RawPayload,
				"updated_at":               time.Now(),
			}).Error
	}
	if !IsRecordNotFound(err) {
		return err
	}
	return r.db.Create(dispute).Error
}

func (r *PaymentRepository) FindPayPalDisputeByID(id uint) (*payment.PayPalDispute, error) {
	var dispute payment.PayPalDispute
	err := r.db.First(&dispute, id).Error
	if err != nil {
		return nil, err
	}
	return &dispute, nil
}

func (r *PaymentRepository) FindPayPalDisputeByPayPalID(paypalID string) (*payment.PayPalDispute, error) {
	var dispute payment.PayPalDispute
	err := r.db.Where("paypal_dispute_id = ?", paypalID).First(&dispute).Error
	if err != nil {
		return nil, err
	}
	return &dispute, nil
}

func (r *PaymentRepository) ListPayPalDisputes(status string, page, pageSize int) ([]payment.PayPalDispute, int64, error) {
	var disputes []payment.PayPalDispute
	var total int64
	query := r.db.Model(&payment.PayPalDispute{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&disputes).Error
	return disputes, total, err
}

func (r *PaymentRepository) UpdatePayPalDisputeEvidenceSubmission(id uint, submittedAt *time.Time, payload, errorMessage, status string) error {
	updates := map[string]interface{}{
		"evidence_submission_payload": payload,
		"evidence_submission_error":   errorMessage,
		"updated_at":                  time.Now(),
	}
	if submittedAt != nil {
		updates["evidence_submitted_at"] = submittedAt
	}
	if status != "" {
		updates["status"] = status
	}
	return r.db.Session(&gorm.Session{SkipHooks: true}).Model(&payment.PayPalDispute{}).Where("id = ?", id).Updates(updates).Error
}
