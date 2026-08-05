package repository

import (
	"tanzanite/internal/domain/order"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository struct {
	db *gorm.DB
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

// FindOrderItemByID 根据 ID 查找订单商品项
func (r *OrderRepository) FindOrderItemByID(id uint) (*order.OrderItem, error) {
	var item order.OrderItem
	err := r.db.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Update 更新订单
func (r *OrderRepository) Update(o *order.Order) error {
	return r.db.Save(o).Error
}

// UpdateStatus 更新订单状态
func (r *OrderRepository) UpdateStatus(id uint, status string) error {
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

	return r.db.Model(&order.Order{}).Where("id = ?", id).Updates(updates).Error
}

func (r *OrderRepository) MarkPaymentExpired(id uint, expiredAt time.Time) error {
	if expiredAt.IsZero() {
		expiredAt = time.Now()
	}
	return r.db.Model(&order.Order{}).
		Where("id = ? AND status = ? AND payment_status = ?", id, "pending", "unpaid").
		Updates(map[string]interface{}{
			"status":         "payment_expired",
			"payment_status": "expired",
			"cancelled_at":   expiredAt,
			"updated_at":     expiredAt,
		}).Error
}

// Delete 删除订单
func (r *OrderRepository) Delete(id uint) error {
	return r.db.Delete(&order.Order{}, id).Error
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
