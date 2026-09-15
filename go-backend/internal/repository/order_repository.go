package repository

import (
	"errors"
	"time"

	"commerce-platform/internal/domain/order"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ErrOrderStatusConflict indicates that an order status transition lost a
// compare-and-swap race (or that the supplied expected status was stale).
var ErrOrderStatusConflict = errors.New("order status transition conflict")

type OrderRepository struct {
	db *gorm.DB
}

// ShippingIdentitySignal contains only the normalized fields needed for
// referral risk comparison. It deliberately omits payment and unrelated order
// data so callers cannot accidentally build a broad PII export query.
type ShippingIdentitySignal struct {
	Address1   string `gorm:"column:shipping_address1"`
	Address2   string `gorm:"column:shipping_address2"`
	City       string `gorm:"column:shipping_city"`
	State      string `gorm:"column:shipping_state"`
	PostalCode string `gorm:"column:shipping_postal_code"`
	Country    string `gorm:"column:shipping_country"`
	Phone      string `gorm:"column:shipping_phone"`
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// WithTx 复用事务 db 实例
func (r *OrderRepository) WithTx(tx *gorm.DB) *OrderRepository {
	return &OrderRepository{db: tx}
}

func (r *OrderRepository) lockForUpdate(query *gorm.DB) *gorm.DB {
	switch r.db.Dialector.Name() {
	case "postgres", "mysql", "sqlserver":
		return query.Clauses(clause.Locking{Strength: "UPDATE"})
	default:
		return query
	}
}

// Create 创建订单
func (r *OrderRepository) Create(o *order.Order) error {
	return r.db.Create(o).Error
}

// FindByID 根据ID查找订单
func (r *OrderRepository) FindByID(id uint) (*order.Order, error) {
	var o order.Order
	err := r.db.Preload("Items").First(&o, id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// FindByIDBasic reads the order record without loading order items. It is
// useful for cross-cutting operational paths that only need order metadata.
func (r *OrderRepository) FindByIDBasic(id uint) (*order.Order, error) {
	var o order.Order
	err := r.db.First(&o, id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) FindByIDsBasic(ids []uint) ([]order.Order, error) {
	ids = uniqueUintValues(ids)
	if len(ids) == 0 {
		return []order.Order{}, nil
	}
	var orders []order.Order
	if err := r.db.Where("id IN ?", ids).Find(&orders).Error; err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) FindByIDForUpdate(id uint) (*order.Order, error) {
	var o order.Order
	err := r.lockForUpdate(r.db).First(&o, id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) FindByIDForUpdateWithItems(id uint) (*order.Order, error) {
	var o order.Order
	err := r.lockForUpdate(r.db).Preload("Items").First(&o, id).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// FindByOrderNumber 根据订单号查找订单
func (r *OrderRepository) FindByOrderNumber(orderNumber string) (*order.Order, error) {
	var o order.Order
	err := r.db.Preload("Items").
		Where("order_number = ?", orderNumber).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) FindByOrderNumberForVerification(orderNumber string) (*order.Order, error) {
	var o order.Order
	err := r.db.Where("order_number = ?", orderNumber).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// FindByOrderNumberForVerificationForUpdate locks the order row while a
// verified gateway payment decides whether it is the first payment or a
// duplicate payment that must be refunded.
func (r *OrderRepository) FindByOrderNumberForVerificationForUpdate(orderNumber string) (*order.Order, error) {
	var o order.Order
	err := r.lockForUpdate(r.db).Where("order_number = ?", orderNumber).First(&o).Error
	if err != nil {
		return nil, err
	}
	return &o, nil
}

// FindOrderItemByID 根据 ID 查找订单商品项
func (r *OrderRepository) FindOrderItemByID(id uint) (*order.OrderItem, error) {
	var item order.OrderItem
	err := r.db.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *OrderRepository) UpdateOrderItemCustoms(orderID, orderItemID uint, declaredValue *float64, confirmed bool) error {
	return r.db.Model(&order.OrderItem{}).
		Where("id = ? AND order_id = ?", orderItemID, orderID).
		Updates(map[string]interface{}{
			"declared_value":           declaredValue,
			"declared_value_confirmed": confirmed,
		}).Error
}

// Update 更新订单
func (r *OrderRepository) Update(o *order.Order) error {
	return r.db.Save(o).Error
}

// UpdateStatus updates an order status using an atomic compare-and-swap.
// Callers must provide the status they observed before attempting the
// transition. A stale observation leaves the row untouched and returns
// ErrOrderStatusConflict.
func (r *OrderRepository) UpdateStatus(id uint, expectedCurrentStatus, status string) error {
	updates := map[string]interface{}{
		"status": status,
	}

	// 根据状态更新时间戳
	switch status {
	case "paid":
		updates["paid_at"] = time.Now()
	case "shipped":
		updates["shipped_at"] = time.Now()
	case "completed":
		updates["completed_at"] = time.Now()
	case "cancelled", "payment_expired":
		updates["cancelled_at"] = time.Now()
	}

	result := r.db.Model(&order.Order{}).
		Where("id = ? AND status = ?", id, expectedCurrentStatus).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrOrderStatusConflict
	}
	return nil
}

// MarkDisputed freezes fulfillment as soon as a payment dispute is linked to
// the order. The caller should hold the order row lock when used in a larger
// transaction.
func (r *OrderRepository) MarkDisputed(id uint) error {
	var current order.Order
	if err := r.lockForUpdate(r.db).First(&current, id).Error; err != nil {
		return err
	}
	updates := map[string]interface{}{"fulfillment_hold": true}
	if current.Status != "disputed" {
		updates["dispute_previous_status"] = current.Status
		updates["dispute_previous_hold"] = current.FulfillmentHold
		updates["status"] = "disputed"
	}
	return r.db.Model(&order.Order{}).Where("id = ?", id).Updates(updates).Error
}

// RestoreDisputeProjection restores the state captured when the first active
// dispute opened. The status predicate prevents stale events from rewinding a
// newer fulfilment transition.
func (r *OrderRepository) RestoreDisputeProjection(id uint) error {
	var current order.Order
	if err := r.lockForUpdate(r.db).First(&current, id).Error; err != nil {
		return err
	}
	if current.Status != "disputed" || current.DisputePreviousStatus == "" {
		return nil
	}
	return r.db.Model(&order.Order{}).
		Where("id = ? AND status = ?", id, "disputed").
		Updates(map[string]interface{}{
			"status":                  current.DisputePreviousStatus,
			"fulfillment_hold":        current.DisputePreviousHold,
			"dispute_previous_status": "",
			"dispute_previous_hold":   false,
		}).Error
}

// MarkPaymentLiabilityReviewHold records a paid order that must be reviewed
// before fulfillment because the gateway did not transfer fraud liability.
func (r *OrderRepository) MarkPaymentLiabilityReviewHold(id uint) error {
	return r.db.Model(&order.Order{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":           "needs_review",
			"fulfillment_hold": true,
		}).Error
}

// ReleasePaymentLiabilityReviewHold resumes fulfillment after the dedicated
// liability review is approved. The status predicate prevents a dispute or a
// different operational hold from being cleared accidentally.
func (r *OrderRepository) ReleasePaymentLiabilityReviewHold(id uint) error {
	const highValueLiabilityReviewReason = "high_value_liability_shift_not_transferred"

	return r.db.Model(&order.Order{}).
		Where("id = ? AND status = ? AND payment_status = ? AND fulfillment_hold = ?", id, "needs_review", "paid", true).
		Where(
			`EXISTS (
				SELECT 1
				FROM payment_reviews approved_liability_review
				WHERE approved_liability_review.order_id = ?
				  AND approved_liability_review.reason = ?
				  AND LOWER(approved_liability_review.status) = 'approved'
			)`,
			id,
			highValueLiabilityReviewReason,
		).
		Where(
			`NOT EXISTS (
				SELECT 1
				FROM payment_reviews active_payment_review
				WHERE active_payment_review.order_id = ?
				  AND LOWER(active_payment_review.status) NOT IN ('approved', 'rejected', 'cancelled')
			)`,
			id,
		).
		Where(
			`NOT EXISTS (
				SELECT 1
				FROM stripe_disputes active_stripe_dispute
				WHERE active_stripe_dispute.order_id = ?
				  AND active_stripe_dispute.deleted_at IS NULL
				  AND LOWER(COALESCE(active_stripe_dispute.status, '')) NOT IN ('won', 'lost', 'closed', 'resolved', 'cancelled', 'canceled', 'denied', 'rejected', 'withdrawn', 'refunded')
			)`,
			id,
		).
		Where(
			`NOT EXISTS (
				SELECT 1
				FROM paypal_disputes active_paypal_dispute
				WHERE active_paypal_dispute.order_id = ?
				  AND active_paypal_dispute.deleted_at IS NULL
				  AND LOWER(COALESCE(active_paypal_dispute.status, '')) NOT IN ('won', 'lost', 'closed', 'resolved', 'cancelled', 'canceled', 'denied', 'rejected', 'withdrawn', 'refunded')
				  AND LOWER(COALESCE(active_paypal_dispute.dispute_state, '')) NOT IN ('won', 'lost', 'closed', 'resolved', 'cancelled', 'canceled', 'denied', 'rejected', 'withdrawn', 'refunded')
			)`,
			id,
		).
		Updates(map[string]interface{}{
			"status":           "processing",
			"fulfillment_hold": false,
		}).Error
}

// MarkCancelledIfPendingUnpaid atomically claims cancellation for an unpaid
// pending order. The affected-row check prevents duplicate rollback work when
// cancellation requests race.
func (r *OrderRepository) MarkCancelledIfPendingUnpaid(id uint, cancelledAt time.Time) (bool, error) {
	if cancelledAt.IsZero() {
		cancelledAt = time.Now().UTC()
	}
	result := r.db.Model(&order.Order{}).
		Where("id = ? AND status = ? AND payment_status = ?", id, "pending", "unpaid").
		Updates(map[string]interface{}{
			"status":       "cancelled",
			"cancelled_at": cancelledAt,
			"updated_at":   cancelledAt,
		})
	return result.RowsAffected == 1, result.Error
}

// MarkPaymentExpired atomically claims payment expiration for an unpaid
// pending order. The boolean is false when another terminal transition
// already won the race; callers must not perform expiration side effects then.
func (r *OrderRepository) MarkPaymentExpired(id uint, expiredAt time.Time) (bool, error) {
	if expiredAt.IsZero() {
		expiredAt = time.Now().UTC()
	}
	result := r.db.Model(&order.Order{}).
		Where("id = ? AND status = ? AND payment_status = ?", id, "pending", "unpaid").
		Updates(map[string]interface{}{
			"status":         "payment_expired",
			"payment_status": "expired",
			"cancelled_at":   expiredAt,
			"updated_at":     expiredAt,
		})
	return result.RowsAffected == 1, result.Error
}

// SoftDeleteUnpaidCancelledOrPaymentExpiredOrderRecord hides only an unpaid
// terminal order from default queries. The financial-state predicate is a
// repository-level guard in addition to the service transaction lock.
func (r *OrderRepository) SoftDeleteUnpaidCancelledOrPaymentExpiredOrderRecord(id uint) (bool, error) {
	result := r.db.Where(
		"id = ? AND ((status = ? AND payment_status = ?) OR (status = ? AND payment_status = ?))",
		id,
		"cancelled",
		"unpaid",
		"payment_expired",
		"expired",
	).Delete(&order.Order{})
	return result.RowsAffected == 1, result.Error
}

// UpdatePaymentStatus 更新支付状态
func (r *OrderRepository) UpdatePaymentStatus(id uint, paymentStatus string) error {
	updates := map[string]interface{}{
		"payment_status": paymentStatus,
	}

	if paymentStatus == "paid" {
		updates["paid_at"] = time.Now()
	}

	return r.db.Model(&order.Order{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateShippingStatus 更新物流状态
func (r *OrderRepository) UpdateShippingStatus(id uint, shippingStatus string) error {
	updates := map[string]interface{}{
		"shipping_status": shippingStatus,
	}

	if shippingStatus == "shipped" {
		updates["shipped_at"] = time.Now()
	}

	return r.db.Model(&order.Order{}).Where("id = ?", id).Updates(updates).Error
}

// UpdateShippingStatusIfDifferent changes the shipping status only when the
// stored value is different. The affected-row result makes one-way events
// such as delivery auditable without duplicate records under concurrent syncs.
func (r *OrderRepository) UpdateShippingStatusIfDifferent(id uint, shippingStatus string) (bool, error) {
	updates := map[string]interface{}{
		"shipping_status": shippingStatus,
	}
	if shippingStatus == "shipped" {
		updates["shipped_at"] = time.Now()
	}

	result := r.db.Model(&order.Order{}).
		Where("id = ? AND (shipping_status IS NULL OR shipping_status <> ?)", id, shippingStatus).
		Updates(updates)
	return result.RowsAffected > 0, result.Error
}

// MarkDeliveredAtIfNeeded preserves the first authoritative delivery time and
// returns whether this call created a new delivery fact.
func (r *OrderRepository) MarkDeliveredAtIfNeeded(id uint, deliveredAt time.Time) (bool, error) {
	if deliveredAt.IsZero() {
		deliveredAt = time.Now().UTC()
	} else {
		deliveredAt = deliveredAt.UTC()
	}
	result := r.db.Model(&order.Order{}).
		Where("id = ? AND (shipping_status IS NULL OR shipping_status <> ? OR delivered_at IS NULL)", id, "delivered").
		Updates(map[string]interface{}{
			"shipping_status": "delivered",
			"delivered_at":    deliveredAt,
			"updated_at":      deliveredAt,
		})
	return result.RowsAffected == 1, result.Error
}

func (r *OrderRepository) MarkProductionStarted(id uint, startedAt time.Time) (bool, error) {
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	result := r.db.Model(&order.Order{}).
		Where(
			"id = ? AND fulfillment_mode IN ? AND payment_status = ? AND status IN ? AND production_status = ?",
			id,
			[]string{order.FulfillmentModeMadeToOrder, order.FulfillmentModeMixed},
			"paid",
			[]string{"paid", "processing"},
			order.ProductionStatusNotStarted,
		).
		Updates(map[string]interface{}{
			"production_status":     order.ProductionStatusStarted,
			"production_started_at": startedAt,
			"updated_at":            startedAt,
		})
	return result.RowsAffected == 1, result.Error
}

func (r *OrderRepository) MarkProductionCompleted(id uint, completedAt time.Time) (bool, error) {
	if completedAt.IsZero() {
		completedAt = time.Now().UTC()
	}
	result := r.db.Model(&order.Order{}).
		Where(
			"id = ? AND fulfillment_mode IN ? AND payment_status = ? AND status IN ? AND production_status = ?",
			id,
			[]string{order.FulfillmentModeMadeToOrder, order.FulfillmentModeMixed},
			"paid",
			[]string{"paid", "processing"},
			order.ProductionStatusStarted,
		).
		Updates(map[string]interface{}{
			"production_status":       order.ProductionStatusCompleted,
			"production_completed_at": completedAt,
			"updated_at":              completedAt,
		})
	return result.RowsAffected == 1, result.Error
}

// UpdateTrackingInfo 更新物流追踪信息
func (r *OrderRepository) UpdateTrackingInfo(id uint, info order.TrackingInfoUpdate) error {
	updates := map[string]interface{}{
		"tracking_number":             info.TrackingNumber,
		"tracking_provider_id":        info.TrackingProviderID,
		"carrier_id":                  info.CarrierID,
		"carrier_service_id":          info.CarrierServiceID,
		"tracking_carrier_mapping_id": info.TrackingCarrierMappingID,
		"provider_carrier_code":       info.ProviderCarrierCode,
		"provider_carrier_name":       info.ProviderCarrierName,
	}

	return r.db.Model(&order.Order{}).Where("id = ?", id).Updates(updates).Error
}

func (r *OrderRepository) FindPaymentExpirationCandidates(cutoff time.Time, limit int) ([]order.Order, error) {
	var orders []order.Order
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query := r.db.Model(&order.Order{}).
		Preload("Items").
		Where("orders.status = ? AND orders.payment_status = ?", "pending", "unpaid").
		Where("NOT EXISTS (SELECT 1 FROM transactions WHERE transactions.order_id = orders.id AND transactions.status = ?)", "completed").
		Where("COALESCE((SELECT MAX(transactions.updated_at) FROM transactions WHERE transactions.order_id = orders.id), orders.created_at) <= ?", cutoff)

	err := query.Order("created_at ASC").Limit(limit).Find(&orders).Error
	return orders, err
}

func (r *OrderRepository) CountPaidOrdersForUserBefore(userID uint, excludeOrderID uint) (int64, error) {
	if r == nil || r.db == nil || userID == 0 {
		return 0, nil
	}

	query := r.db.Model(&order.Order{}).
		Where("user_id = ? AND payment_status = ?", userID, "paid")
	if excludeOrderID > 0 {
		query = query.Where("id <> ?", excludeOrderID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// CountEverPaidOrdersForUserBefore is stricter than the current-payment-state
// query used by 3DS. A refunded or cancelled first purchase still means the
// customer is not new for referral eligibility.
func (r *OrderRepository) CountEverPaidOrdersForUserBefore(userID uint, excludeOrderID uint) (int64, error) {
	if r == nil || r.db == nil || userID == 0 {
		return 0, nil
	}
	query := r.db.Model(&order.Order{}).
		Where("user_id = ? AND paid_at IS NOT NULL", userID)
	if excludeOrderID > 0 {
		query = query.Where("id <> ?", excludeOrderID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// FindHistoricalShippingIdentitySignals returns paid-order address/phone
// comparison inputs for one user. Raw values stay in process memory and are
// immediately HMACed by the referral service; they are never copied into the
// referral tables.
func (r *OrderRepository) FindHistoricalShippingIdentitySignals(userID uint, excludeOrderID uint) ([]ShippingIdentitySignal, error) {
	if r == nil || r.db == nil || userID == 0 {
		return []ShippingIdentitySignal{}, nil
	}
	query := r.db.Model(&order.Order{}).
		Select("shipping_address1, shipping_address2, shipping_city, shipping_state, shipping_postal_code, shipping_country, shipping_phone").
		Where("user_id = ? AND paid_at IS NOT NULL", userID)
	if excludeOrderID > 0 {
		query = query.Where("id <> ?", excludeOrderID)
	}
	var signals []ShippingIdentitySignal
	if err := query.Find(&signals).Error; err != nil {
		return nil, err
	}
	return signals, nil
}
