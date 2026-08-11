package repository

import (
	"time"

	"commerce-platform/internal/domain/payment"
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
