package repository

import (
	"tanzanite/internal/domain/order"
	"time"
)

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
