package repository

import "commerce-platform/internal/domain/order"

// FindByUserID 查找用户的订单列表
func (r *OrderRepository) FindByUserID(userID uint, page, pageSize int) ([]order.Order, int64, error) {
	var orders []order.Order
	var total int64

	query := r.db.Model(&order.Order{}).Where("user_id = ?", userID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Items").Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&orders).Error

	return orders, total, err
}

// FindByUserIDForShowcaseUpload returns only the order fields needed to
// display upload eligibility. It deliberately avoids addresses and items.
func (r *OrderRepository) FindByUserIDForShowcaseUpload(userID uint, limit int) ([]order.Order, error) {
	var orders []order.Order
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	err := r.db.Model(&order.Order{}).
		Select("id, user_id, order_number, status, shipping_status, total_amount, currency, completed_at, created_at").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&orders).Error
	return orders, err
}

// FindByIDAndUserIDForShowcaseUpload loads an order only when it belongs to
// the current user, preventing cross-account order probing by ID.
func (r *OrderRepository) FindByIDAndUserIDForShowcaseUpload(id, userID uint) (*order.Order, error) {
	var item order.Order
	err := r.db.Model(&order.Order{}).
		Select("id, user_id, order_number, status, shipping_status, total_amount, currency, completed_at, created_at").
		Where("id = ? AND user_id = ?", id, userID).
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// FindAll 查找所有订单（管理员）
func (r *OrderRepository) FindAll(page, pageSize int, status string) ([]order.Order, int64, error) {
	var orders []order.Order
	var total int64

	query := r.db.Model(&order.Order{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Preload("Items").Order("created_at DESC").
		Offset(offset).Limit(pageSize).Find(&orders).Error

	return orders, total, err
}

// FindRecent 获取最近订单
func (r *OrderRepository) FindRecent(limit int) ([]order.Order, error) {
	var orders []order.Order
	err := r.db.Preload("Items").Order("created_at DESC").Limit(limit).Find(&orders).Error
	return orders, err
}

// FindAllWithFilters 根据筛选条件获取订单列表
func (r *OrderRepository) FindAllWithFilters(page, pageSize int, status, paymentStatus, shippingStatus, search, startDate, endDate string) ([]order.Order, int64, error) {
	var orders []order.Order
	var total int64

	query := r.db.Model(&order.Order{}).Preload("Items")

	// 应用筛选条件
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if paymentStatus != "" {
		query = query.Where("payment_status = ?", paymentStatus)
	}
	if shippingStatus != "" {
		query = query.Where("shipping_status = ?", shippingStatus)
	}
	if search != "" {
		query = query.Where("order_number LIKE ? OR shipping_first_name LIKE ? OR shipping_last_name LIKE ? OR shipping_email LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&orders).Error

	return orders, total, err
}
