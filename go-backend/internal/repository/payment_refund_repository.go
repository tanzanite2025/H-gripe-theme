package repository

import (
	"fmt"
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
		candidate, err := domainmoney.FromMajorFloat(refunds[i].Amount, amount.Currency().String())
		if err != nil {
			return nil, fmt.Errorf("pending refund %d amount: %w", refunds[i].ID, err)
		}
		if candidate.AmountMinor() == amount.AmountMinor() {
			return &refunds[i], nil
		}
		requested, err := domainmoney.FromMajorFloat(refunds[i].RequestedAmount, amount.Currency().String())
		if err != nil {
			return nil, fmt.Errorf("pending refund %d requested amount: %w", refunds[i].ID, err)
		}
		if requested.AmountMinor() == amount.AmountMinor() {
			return &refunds[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
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

func (r *PaymentRepository) SumRefundTotalAmountByTransactionID(transactionID uint, currencyCode string, statuses ...string) (float64, error) {
	var amount float64
	var giftCardAmount float64
	query := r.db.Model(&payment.Refund{}).Where("transaction_id = ?", transactionID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if err := query.Select("COALESCE(SUM(amount), 0)").Scan(&amount).Error; err != nil {
		return 0, err
	}
	if err := query.Select("COALESCE(SUM(gift_card_refund_amount), 0)").Scan(&giftCardAmount).Error; err != nil {
		return 0, err
	}
	amountMoney, err := domainmoney.FromMajorFloat(amount, currencyCode)
	if err != nil {
		return 0, err
	}
	giftCardMoney, err := domainmoney.FromMajorFloat(giftCardAmount, currencyCode)
	if err != nil {
		return 0, err
	}
	totalMoney, err := amountMoney.Add(giftCardMoney)
	if err != nil {
		return 0, err
	}
	return totalMoney.MajorFloat()
}

func (r *PaymentRepository) SumRefundTotalAmountByOrderID(orderID uint, currencyCode string, statuses ...string) (float64, error) {
	var amount float64
	var giftCardAmount float64
	query := r.db.Model(&payment.Refund{}).Where("order_id = ?", orderID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if err := query.Select("COALESCE(SUM(amount), 0)").Scan(&amount).Error; err != nil {
		return 0, err
	}
	if err := query.Select("COALESCE(SUM(gift_card_refund_amount), 0)").Scan(&giftCardAmount).Error; err != nil {
		return 0, err
	}
	amountMoney, err := domainmoney.FromMajorFloat(amount, currencyCode)
	if err != nil {
		return 0, err
	}
	giftCardMoney, err := domainmoney.FromMajorFloat(giftCardAmount, currencyCode)
	if err != nil {
		return 0, err
	}
	totalMoney, err := amountMoney.Add(giftCardMoney)
	if err != nil {
		return 0, err
	}
	return totalMoney.MajorFloat()
}

func (r *PaymentRepository) SumRefundGiftCardAmountByOrderID(orderID uint, statuses ...string) (float64, error) {
	var total float64
	query := r.db.Model(&payment.Refund{}).Where("order_id = ?", orderID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(gift_card_refund_amount), 0)").Scan(&total).Error
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

func (r *PaymentRepository) SumRefundLoyaltyPointsClawbackByOrderID(orderID uint, statuses ...string) (int, error) {
	var total int
	query := r.db.Model(&payment.Refund{}).Where("order_id = ?", orderID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(loyalty_points_clawback), 0)").Scan(&total).Error
	return total, err
}

func (r *PaymentRepository) SumRefundLoyaltyPointsReturnedByOrderID(orderID uint, statuses ...string) (int, error) {
	var total int
	query := r.db.Model(&payment.Refund{}).Where("order_id = ?", orderID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(loyalty_points_returned), 0)").Scan(&total).Error
	return total, err
}

func (r *PaymentRepository) SumRefundedSubtotalAmountByOrderID(orderID uint, currencyCode string, statuses ...string) (float64, error) {
	lineItemSubtotal, err := r.sumRefundLineItemSubtotalAmountByOrderID(orderID, currencyCode, statuses...)
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

	lineItemMoney, err := domainmoney.FromMajorFloat(lineItemSubtotal, currencyCode)
	if err != nil {
		return 0, err
	}
	amountOnlyMoney, err := domainmoney.FromMajorFloat(amountOnlySubtotal, currencyCode)
	if err != nil {
		return 0, err
	}
	totalMoney, err := lineItemMoney.Add(amountOnlyMoney)
	if err != nil {
		return 0, err
	}
	return totalMoney.MajorFloat()
}

func (r *PaymentRepository) sumRefundLineItemSubtotalAmountByOrderID(orderID uint, currencyCode string, statuses ...string) (float64, error) {
	var totalMinor int64
	query := r.db.Model(&payment.RefundLineItem{}).
		Joins("JOIN refunds ON refunds.id = refund_line_items.refund_id").
		Where("refund_line_items.order_id = ?", orderID).
		Where("UPPER(refund_line_items.currency) = UPPER(?)", currencyCode).
		Where("refunds.deleted_at IS NULL")
	if len(statuses) > 0 {
		query = query.Where("refunds.status IN ?", statuses)
	}
	err := query.Select("COALESCE(SUM(refund_line_items.line_subtotal_minor), 0)").Scan(&totalMinor).Error
	if err != nil {
		return 0, err
	}
	money, err := domainmoney.New(totalMinor, currencyCode)
	if err != nil {
		return 0, err
	}
	return money.MajorFloat()
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
