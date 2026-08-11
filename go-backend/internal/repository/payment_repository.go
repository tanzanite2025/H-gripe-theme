package repository

import (
	"commerce-platform/internal/domain/payment"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) WithTx(tx *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: tx}
}

func (r *PaymentRepository) lockForUpdate(query *gorm.DB) *gorm.DB {
	switch r.db.Dialector.Name() {
	case "postgres", "mysql", "sqlserver":
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	default:
		return query
	}
}

// PaymentMethod 相关方法

// FindPaymentMethodByID 根据ID查找支付方式
func (r *PaymentRepository) FindPaymentMethodByID(id uint) (*payment.PaymentMethod, error) {
	var pm payment.PaymentMethod
	err := r.db.First(&pm, id).Error
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

// FindPaymentMethodByCode 根据代码查找支付方式
func (r *PaymentRepository) FindPaymentMethodByCode(code string) (*payment.PaymentMethod, error) {
	var pm payment.PaymentMethod
	err := r.db.Where("code = ?", code).First(&pm).Error
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

// FindAllPaymentMethods 查找所有支付方式
func (r *PaymentRepository) FindAllPaymentMethods(enabledOnly bool) ([]payment.PaymentMethod, error) {
	var methods []payment.PaymentMethod
	query := r.db.Order("sort_order ASC")

	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}

	err := query.Find(&methods).Error
	return methods, err
}

// CreatePaymentMethod 创建支付方式
func (r *PaymentRepository) CreatePaymentMethod(pm *payment.PaymentMethod) error {
	enabled := pm.Enabled
	if err := r.db.Create(pm).Error; err != nil {
		return err
	}
	if !enabled {
		if err := r.db.Model(pm).Update("enabled", false).Error; err != nil {
			return err
		}
		pm.Enabled = false
	}
	return nil
}

// UpdatePaymentMethod 更新支付方式
func (r *PaymentRepository) UpdatePaymentMethod(pm *payment.PaymentMethod) error {
	return r.db.Save(pm).Error
}

// DeletePaymentMethod 删除支付方式
func (r *PaymentRepository) DeletePaymentMethod(id uint) error {
	return r.db.Delete(&payment.PaymentMethod{}, id).Error
}

// TaxRate 相关方法

// FindTaxRateByID 根据ID查找税率
func (r *PaymentRepository) FindTaxRateByID(id uint) (*payment.TaxRate, error) {
	var tr payment.TaxRate
	err := r.db.First(&tr, id).Error
	if err != nil {
		return nil, err
	}
	return &tr, nil
}

// FindTaxRateByLocation 根据地区查找税率
func (r *PaymentRepository) FindTaxRateByLocation(country, state string) (*payment.TaxRate, error) {
	var tr payment.TaxRate
	err := r.db.Where("country = ? AND state = ? AND enabled = ?", country, state, true).
		First(&tr).Error
	if err != nil {
		return nil, err
	}
	return &tr, nil
}

// FindAllTaxRates 查找所有税率
func (r *PaymentRepository) FindAllTaxRates(enabledOnly bool) ([]payment.TaxRate, error) {
	var rates []payment.TaxRate
	query := r.db.Order("country ASC, state ASC")
	if enabledOnly {
		query = query.Where("enabled = ?", true)
	}
	err := query.Find(&rates).Error
	return rates, err
}

// Transaction 相关方法

// CreateTransaction 创建交易记录
func (r *PaymentRepository) CreateTransaction(t *payment.Transaction) error {
	return r.db.Create(t).Error
}

// FindTransactionByID 根据ID查找交易
func (r *PaymentRepository) FindTransactionByID(id uint) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.db.First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PaymentRepository) FindTransactionByIDForUpdate(id uint) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.lockForUpdate(r.db).First(&t, id).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// FindTransactionByOrderID 根据订单ID查找交易
func (r *PaymentRepository) FindTransactionByOrderID(orderID uint) ([]payment.Transaction, error) {
	var transactions []payment.Transaction
	err := r.db.Where("order_id = ?", orderID).Order("created_at DESC").Find(&transactions).Error
	return transactions, err
}

// FindTransactionByTransactionID 根据交易ID查找
func (r *PaymentRepository) FindTransactionByTransactionID(transactionID string) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.db.Where("transaction_id = ?", transactionID).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *PaymentRepository) FindTransactionByTransactionIDForUpdate(transactionID string) (*payment.Transaction, error) {
	var t payment.Transaction
	err := r.lockForUpdate(r.db).Where("transaction_id = ?", transactionID).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// UpdateTransaction 更新交易
func (r *PaymentRepository) UpdateTransaction(t *payment.Transaction) error {
	return r.db.Save(t).Error
}

func (r *PaymentRepository) ExpireOpenTransactionsByOrderID(orderID uint, expiredAt time.Time) (int64, error) {
	if expiredAt.IsZero() {
		expiredAt = time.Now()
	}
	result := r.db.Session(&gorm.Session{SkipHooks: true}).Model(&payment.Transaction{}).
		Where("order_id = ? AND status IN ?", orderID, []string{"pending", "processing", "requires_action"}).
		Updates(map[string]interface{}{
			"status":        "expired",
			"error_message": "payment attempt expired before completion",
			"updated_at":    expiredAt,
		})
	return result.RowsAffected, result.Error
}

// Refund 相关方法

// CreateRefund 创建退款记录
func (r *PaymentRepository) CreateRefund(rf *payment.Refund) error {
	lineItems := append([]payment.RefundLineItem(nil), rf.LineItems...)
	rf.LineItems = nil
	if err := r.db.Create(rf).Error; err != nil {
		rf.LineItems = lineItems
		return err
	}

	for i := range lineItems {
		lineItems[i].RefundID = rf.ID
		lineItems[i].OrderID = rf.OrderID
		if err := r.db.Create(&lineItems[i]).Error; err != nil {
			rf.LineItems = lineItems
			return err
		}
	}
	rf.LineItems = lineItems
	return nil
}

// FindRefundByID 根据ID查找退款
func (r *PaymentRepository) FindRefundByID(id uint) (*payment.Refund, error) {
	var rf payment.Refund
	err := r.db.Preload("LineItems").First(&rf, id).Error
	if err != nil {
		return nil, err
	}
	return &rf, nil
}

func (r *PaymentRepository) FindRefundByIDForUpdate(id uint) (*payment.Refund, error) {
	var rf payment.Refund
	err := r.lockForUpdate(r.db).Preload("LineItems").First(&rf, id).Error
	if err != nil {
		return nil, err
	}
	return &rf, nil
}

func (r *PaymentRepository) FindRefundByRefundID(refundID string) (*payment.Refund, error) {
	var rf payment.Refund
	err := r.db.Preload("LineItems").Where("refund_id = ?", refundID).First(&rf).Error
	if err != nil {
		return nil, err
	}
	return &rf, nil
}

// FindRefundsByOrderID 根据订单ID查找退款
func (r *PaymentRepository) FindRefundsByOrderID(orderID uint) ([]payment.Refund, error) {
	var refunds []payment.Refund
	err := r.db.Preload("LineItems").Where("order_id = ?", orderID).Order("created_at DESC").Find(&refunds).Error
	return refunds, err
}

func (r *PaymentRepository) FindPendingRefundByTransactionAndAmount(transactionID uint, amount float64) (*payment.Refund, error) {
	var rf payment.Refund
	err := r.lockForUpdate(r.db).
		Preload("LineItems").
		Where("transaction_id = ? AND status = ? AND amount BETWEEN ? AND ?", transactionID, "pending", amount-0.01, amount+0.01).
		Order("created_at ASC").
		First(&rf).Error
	if err != nil {
		return nil, err
	}
	return &rf, nil
}

func (r *PaymentRepository) SumRefundAmountByTransactionID(transactionID uint, statuses ...string) (float64, error) {
	var total float64
	query := r.db.Model(&payment.Refund{}).Where("transaction_id = ?", transactionID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	return total, err
}

func (r *PaymentRepository) SumRefundAmountByOrderID(orderID uint, statuses ...string) (float64, error) {
	var total float64
	query := r.db.Model(&payment.Refund{}).Where("order_id = ?", orderID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	return total, err
}

func (r *PaymentRepository) SumRefundRequestedAmountByOrderID(orderID uint, statuses ...string) (float64, error) {
	var total float64
	query := r.db.Model(&payment.Refund{}).Where("order_id = ?", orderID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(CASE WHEN requested_amount > 0 THEN requested_amount ELSE amount END), 0)").Scan(&total).Error
	return total, err
}

func (r *PaymentRepository) SumRefundDiscountClawbackByOrderID(orderID uint, statuses ...string) (float64, error) {
	var total float64
	query := r.db.Model(&payment.Refund{}).Where("order_id = ?", orderID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(discount_clawback_amount), 0)").Scan(&total).Error
	return total, err
}

func (r *PaymentRepository) SumRefundedSubtotalAmountByOrderID(orderID uint, statuses ...string) (float64, error) {
	lineItemSubtotal, err := r.sumRefundLineItemSubtotalAmountByOrderID(orderID, statuses...)
	if err != nil {
		return 0, err
	}

	var amountOnlySubtotal float64
	query := r.db.Model(&payment.Refund{}).
		Where("order_id = ?", orderID).
		Where("NOT EXISTS (SELECT 1 FROM refund_line_items WHERE refund_line_items.refund_id = refunds.id)")
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if err := query.Select("COALESCE(SUM(CASE WHEN requested_amount > 0 THEN requested_amount ELSE amount END), 0)").Scan(&amountOnlySubtotal).Error; err != nil {
		return 0, err
	}

	return lineItemSubtotal + amountOnlySubtotal, nil
}

func (r *PaymentRepository) sumRefundLineItemSubtotalAmountByOrderID(orderID uint, statuses ...string) (float64, error) {
	var total float64
	query := r.db.Model(&payment.RefundLineItem{}).
		Joins("JOIN refunds ON refunds.id = refund_line_items.refund_id").
		Where("refund_line_items.order_id = ?", orderID).
		Where("refunds.deleted_at IS NULL")
	if len(statuses) > 0 {
		query = query.Where("refunds.status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(refund_line_items.line_subtotal_amount), 0)").Scan(&total).Error
	return total, err
}

func (r *PaymentRepository) SumRefundLineItemQuantitiesByOrderID(orderID uint, statuses ...string) (map[uint]int, error) {
	type quantityRow struct {
		OrderItemID uint
		Quantity    int
	}

	var rows []quantityRow
	query := r.db.Model(&payment.RefundLineItem{}).
		Select("refund_line_items.order_item_id, COALESCE(SUM(refund_line_items.quantity), 0) AS quantity").
		Joins("JOIN refunds ON refunds.id = refund_line_items.refund_id").
		Where("refund_line_items.order_id = ?", orderID).
		Where("refunds.deleted_at IS NULL")
	if len(statuses) > 0 {
		query = query.Where("refunds.status IN ?", statuses)
	}
	if err := query.Group("refund_line_items.order_item_id").Scan(&rows).Error; err != nil {
		return nil, err
	}

	result := make(map[uint]int, len(rows))
	for _, row := range rows {
		result[row.OrderItemID] = row.Quantity
	}
	return result, nil
}

func (r *PaymentRepository) HasAmountOnlyRefundsByOrderID(orderID uint, statuses ...string) (bool, error) {
	var count int64
	query := r.db.Model(&payment.Refund{}).
		Where("order_id = ?", orderID).
		Where("NOT EXISTS (SELECT 1 FROM refund_line_items WHERE refund_line_items.refund_id = refunds.id)")
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *PaymentRepository) HasLineItemRefundsByOrderID(orderID uint, statuses ...string) (bool, error) {
	var count int64
	query := r.db.Model(&payment.RefundLineItem{}).
		Joins("JOIN refunds ON refunds.id = refund_line_items.refund_id").
		Where("refund_line_items.order_id = ?", orderID).
		Where("refunds.deleted_at IS NULL")
	if len(statuses) > 0 {
		query = query.Where("refunds.status IN ?", statuses)
	}
	if err := query.Distinct("refund_line_items.refund_id").Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// UpdateRefund 更新退款
func (r *PaymentRepository) UpdateRefund(rf *payment.Refund) error {
	return r.db.Omit("LineItems").Save(rf).Error
}

func (r *PaymentRepository) MarkRefundLineItemRestocked(id uint, restockedAt time.Time) (bool, error) {
	result := r.db.Model(&payment.RefundLineItem{}).
		Where("id = ? AND restock = ? AND restocked_at IS NULL", id, true).
		Updates(map[string]interface{}{
			"restocked_at": restockedAt,
			"updated_at":   restockedAt,
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

// ClaimStripeWebhookEvent creates an event inbox row once. Processed events
// are acknowledged without re-running side effects; failed events can be
// claimed again on a later Stripe retry.
func (r *PaymentRepository) ClaimStripeWebhookEvent(eventID, eventType, payload string) (bool, error) {
	event := &payment.StripeWebhookEvent{
		EventID:   eventID,
		EventType: eventType,
		Status:    "processing",
		Payload:   payload,
	}
	result := r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(event)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected > 0 {
		return true, nil
	}

	var existing payment.StripeWebhookEvent
	if err := r.db.Where("event_id = ?", eventID).First(&existing).Error; err != nil {
		return false, err
	}
	if existing.Status == "processed" || existing.Status == "processing" {
		return false, nil
	}

	result = r.db.Model(&payment.StripeWebhookEvent{}).
		Where("event_id = ? AND status = ?", eventID, "failed").
		Updates(map[string]interface{}{
			"status":        "processing",
			"payload":       payload,
			"error_message": "",
			"processed_at":  nil,
		})
	return result.RowsAffected > 0, result.Error
}

func (r *PaymentRepository) MarkStripeWebhookEventProcessed(eventID string) error {
	now := time.Now()
	return r.db.Model(&payment.StripeWebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":        "processed",
			"error_message": "",
			"processed_at":  &now,
		}).Error
}

func (r *PaymentRepository) MarkStripeWebhookEventFailed(eventID string, processingErr error) error {
	message := ""
	if processingErr != nil {
		message = processingErr.Error()
	}
	return r.db.Model(&payment.StripeWebhookEvent{}).
		Where("event_id = ?", eventID).
		Updates(map[string]interface{}{
			"status":        "failed",
			"error_message": message,
		}).Error
}

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
