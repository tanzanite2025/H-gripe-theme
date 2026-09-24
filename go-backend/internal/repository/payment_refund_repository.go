package repository

import (
	"fmt"
	"strings"
	"time"

	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/domain/payment"

	"gorm.io/gorm"
)

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
		if lineItems[i].Currency == "" {
			lineItems[i].Currency = rf.Currency
		}
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

func (r *PaymentRepository) FindRefundByRefundIDForUpdate(refundID string) (*payment.Refund, error) {
	var rf payment.Refund
	err := r.lockForUpdate(r.db).Preload("LineItems").Where("refund_id = ?", refundID).First(&rf).Error
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

// HasPendingRefundByOrderID reports whether an order has a refund intent that
// may still move money at the gateway. The order row is locked by the caller
// when this is used as a fulfillment gate, so the check and the subsequent
// shipping mutation share the same serialization point.
func (r *PaymentRepository) HasPendingRefundByOrderID(orderID uint) (bool, error) {
	var count int64
	err := r.db.Model(&payment.Refund{}).
		Where("order_id = ? AND status = ?", orderID, "pending").
		Count(&count).Error
	if err != nil {
		// SQLite unit fixtures may intentionally omit optional payment tables;
		// production dialects fail closed so a missing refund migration can never
		// silently allow fulfillment.
		if r.db != nil && r.db.Dialector.Name() == "sqlite" &&
			(strings.Contains(strings.ToLower(err.Error()), "no such table") || strings.Contains(strings.ToLower(err.Error()), "doesn't exist")) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (r *PaymentRepository) FindPendingRefundByTransactionAndAmount(transactionID uint, amount domainmoney.Money) (*payment.Refund, error) {
	if err := amount.Validate(); err != nil {
		return nil, fmt.Errorf("pending refund amount: %w", err)
	}

	var refunds []payment.Refund
	if err := r.lockForUpdate(r.db).
		Preload("LineItems").
		Where("transaction_id = ? AND status = ?", transactionID, "pending").
		Order("created_at ASC").
		Find(&refunds).Error; err != nil {
		return nil, err
	}

	for i := range refunds {
		candidate, err := refunds[i].AmountMoney()
		if err == nil && candidate.Currency() != amount.Currency() {
			err = domainmoney.ErrCurrencyMismatch
		}
		if err != nil {
			return nil, fmt.Errorf("pending refund %d amount: %w", refunds[i].ID, err)
		}
		if candidate.Currency() == amount.Currency() && candidate.AmountMinor() == amount.AmountMinor() {
			return &refunds[i], nil
		}
		requested, err := refunds[i].RequestedAmountMoney()
		if err == nil && requested.Currency() != amount.Currency() {
			err = domainmoney.ErrCurrencyMismatch
		}
		if err != nil {
			return nil, fmt.Errorf("pending refund %d requested amount: %w", refunds[i].ID, err)
		}
		if requested.Currency() == amount.Currency() && requested.AmountMinor() == amount.AmountMinor() {
			return &refunds[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *PaymentRepository) sumRefundMinor(column string, filterColumn string, filterValue interface{}, statuses ...string) (int64, error) {
	var total int64
	query := r.db.Model(&payment.Refund{}).Where(filterColumn+" = ?", filterValue)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if err := query.Select("COALESCE(SUM(" + column + "), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PaymentRepository) SumRefundAmountMinorByTransactionID(transactionID uint, statuses ...string) (int64, error) {
	return r.sumRefundMinor("amount_minor", "transaction_id", transactionID, statuses...)
}

func (r *PaymentRepository) SumRefundAmountMinorByOrderID(orderID uint, statuses ...string) (int64, error) {
	return r.sumRefundMinor("amount_minor", "order_id", orderID, statuses...)
}

func (r *PaymentRepository) SumRefundRequestedAmountMinorByOrderID(orderID uint, statuses ...string) (int64, error) {
	var total int64
	query := r.db.Model(&payment.Refund{}).Where("order_id = ?", orderID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if err := query.Select("COALESCE(SUM(CASE WHEN requested_amount_minor > 0 THEN requested_amount_minor ELSE amount_minor END), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}

func (r *PaymentRepository) SumRefundDiscountClawbackMinorByOrderID(orderID uint, statuses ...string) (int64, error) {
	return r.sumRefundMinor("discount_clawback_amount_minor", "order_id", orderID, statuses...)
}

func (r *PaymentRepository) SumRefundTotalAmountMinorByTransactionID(transactionID uint, statuses ...string) (int64, error) {
	return r.SumRefundAmountMinorByTransactionID(transactionID, statuses...)
}

func (r *PaymentRepository) SumRefundTotalAmountMinorByOrderID(orderID uint, statuses ...string) (int64, error) {
	return r.SumRefundAmountMinorByOrderID(orderID, statuses...)
}

func (r *PaymentRepository) SumRefundedSubtotalMinorAmountByOrderID(orderID uint, statuses ...string) (int64, error) {
	var lineSubtotal int64
	lineQuery := r.db.Model(&payment.RefundLineItem{}).
		Joins("JOIN refunds ON refunds.id = refund_line_items.refund_id").
		Where("refund_line_items.order_id = ?", orderID).
		Where("refunds.deleted_at IS NULL")
	if len(statuses) > 0 {
		lineQuery = lineQuery.Where("refunds.status IN ?", statuses)
	}
	if err := lineQuery.Select("COALESCE(SUM(refund_line_items.line_subtotal_minor), 0)").Scan(&lineSubtotal).Error; err != nil {
		return 0, err
	}
	var amountOnly int64
	amountQuery := r.db.Model(&payment.Refund{}).
		Where("order_id = ?", orderID).
		Where("NOT EXISTS (SELECT 1 FROM refund_line_items WHERE refund_line_items.refund_id = refunds.id)")
	if len(statuses) > 0 {
		amountQuery = amountQuery.Where("status IN ?", statuses)
	}
	if err := amountQuery.Select("COALESCE(SUM(CASE WHEN requested_amount_minor > 0 THEN requested_amount_minor ELSE amount_minor END), 0)").Scan(&amountOnly).Error; err != nil {
		return 0, err
	}
	if amountOnly > 0 && lineSubtotal > int64(^uint64(0)>>1)-amountOnly {
		return 0, fmt.Errorf("refunded subtotal overflows int64")
	}
	return lineSubtotal + amountOnly, nil
}

func (r *PaymentRepository) FindPendingRefundByTransactionIDForUpdate(transactionID uint) (*payment.Refund, error) {
	var rf payment.Refund
	err := r.lockForUpdate(r.db).
		Preload("LineItems").
		Where("transaction_id = ? AND status = ?", transactionID, "pending").
		Order("created_at ASC").
		First(&rf).Error
	if err != nil {
		return nil, err
	}
	return &rf, nil
}

// FindFailedRefundByTransactionIDForUpdate finds a locally failed refund that
// has not yet been associated with a provider refund identifier. Synchronous
// gateway failures leave this retryable intent in that state; a later provider
// failure webhook can then enrich the same refund instead of creating a second
// failed intent for the transaction.
func (r *PaymentRepository) FindFailedRefundByTransactionIDForUpdate(transactionID uint) (*payment.Refund, error) {
	var rf payment.Refund
	err := r.lockForUpdate(r.db).
		Preload("LineItems").
		Where("transaction_id = ? AND status = ? AND refund_id IS NULL", transactionID, "failed").
		Order("created_at ASC").
		First(&rf).Error
	if err != nil {
		return nil, err
	}
	return &rf, nil
}

func (r *PaymentRepository) FindRefundByTransactionIDAndReasonForUpdate(transactionID uint, reason string) (*payment.Refund, error) {
	var rf payment.Refund
	err := r.lockForUpdate(r.db).
		Preload("LineItems").
		Where("transaction_id = ? AND reason = ?", transactionID, reason).
		Order("created_at ASC").
		First(&rf).Error
	if err != nil {
		return nil, err
	}
	return &rf, nil
}

func (r *PaymentRepository) SumRefundLoyaltyPointsClawbackByOrderID(orderID uint, statuses ...string) (int, error) {
	var total int
	query := r.db.Model(&payment.Refund{}).Where("order_id = ?", orderID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(loyalty_points_clawback), 0)").Scan(&total).Error
	return total, err
}

func (r *PaymentRepository) SumRefundLoyaltyPointsSettledByOrderID(orderID uint, statuses ...string) (int, error) {
	var total int
	query := r.db.Model(&payment.Refund{}).Where("order_id = ?", orderID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(loyalty_points_clawback + loyalty_points_debt), 0)").Scan(&total).Error
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
