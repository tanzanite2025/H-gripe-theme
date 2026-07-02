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
	err := r.db.Preload("Items").Preload("ShippingAddress").Preload("BillingAddress").First(&o, id).Error
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
	err := r.db.Preload("Items").Preload("ShippingAddress").Preload("BillingAddress").
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

// GetOrderStats 获取订单统计
func (r *OrderRepository) GetOrderStats(userID uint) (map[string]int64, error) {
	stats := make(map[string]int64)

	query := r.db.Model(&order.Order{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// 统计各状态订单数量
	statuses := []string{"pending", "paid", "shipped", "completed", "cancelled"}
	for _, status := range statuses {
		var count int64
		if err := query.Where("status = ?", status).Count(&count).Error; err != nil {
			return nil, err
		}
		stats[status] = count
	}

	return stats, nil
}

// GetStats 获取订单统计（管理员仪表板）
func (r *OrderRepository) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	today := time.Now().Truncate(24 * time.Hour)

	// 总订单数
	var total int64
	if err := r.db.Model(&order.Order{}).Count(&total).Error; err != nil {
		return nil, err
	}
	stats["total"] = total

	// 今日订单数
	var todayCount int64
	if err := r.db.Model(&order.Order{}).Where("created_at >= ?", today).Count(&todayCount).Error; err != nil {
		return nil, err
	}
	stats["today"] = todayCount

	// 按状态统计
	var statusStats []struct {
		Status string
		Count  int64
	}
	if err := r.db.Model(&order.Order{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusStats).Error; err != nil {
		return nil, err
	}

	for _, stat := range statusStats {
		stats[stat.Status] = stat.Count
	}

	// 总销售额
	var totalRevenue float64
	if err := r.db.Model(&order.Order{}).Where("status != ?", "cancelled").Select("COALESCE(SUM(total_amount), 0)").Scan(&totalRevenue).Error; err != nil {
		return nil, err
	}
	stats["total_revenue"] = totalRevenue

	// 今日销售额
	var todayRevenue float64
	if err := r.db.Model(&order.Order{}).Where("created_at >= ? AND status != ?", today, "cancelled").Select("COALESCE(SUM(total_amount), 0)").Scan(&todayRevenue).Error; err != nil {
		return nil, err
	}
	stats["today_revenue"] = todayRevenue

	return stats, nil
}

// FindRecent 获取最近订单
func (r *OrderRepository) FindRecent(limit int) ([]order.Order, error) {
	var orders []order.Order
	err := r.db.Preload("Items").Order("created_at DESC").Limit(limit).Find(&orders).Error
	return orders, err
}

// GetSalesByDateRange 获取日期范围内的销售数据
func (r *OrderRepository) GetSalesByDateRange(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []struct {
		Date   string
		Count  int64
		Amount float64
	}

	err := r.db.Model(&order.Order{}).
		Select("DATE(created_at) as date, COUNT(*) as count, COALESCE(SUM(total_amount), 0) as amount").
		Where("created_at BETWEEN ? AND ? AND status != ?", startDate, endDate, "cancelled").
		Group("DATE(created_at)").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// 转换为 map 格式
	data := make([]map[string]interface{}, len(results))
	for i, result := range results {
		data[i] = map[string]interface{}{
			"date":   result.Date,
			"count":  result.Count,
			"amount": result.Amount,
		}
	}

	return data, nil
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
