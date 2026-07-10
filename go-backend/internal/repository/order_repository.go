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
	case "cancelled":
		updates["cancelled_at"] = time.Now()
	}

	return r.db.Model(&order.Order{}).Where("id = ?", id).Updates(updates).Error
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
func (r *OrderRepository) UpdateTrackingInfo(id uint, trackingNumber, carrierCode string) error {
	updates := map[string]interface{}{
		"tracking_number": trackingNumber,
		"carrier_code":    carrierCode,
	}

	return r.db.Model(&order.Order{}).Where("id = ?", id).Updates(updates).Error
}
