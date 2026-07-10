package repository

import "tanzanite/internal/domain/order"

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
