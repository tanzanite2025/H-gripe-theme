package repository

import (
	"commerce-platform/internal/domain/order"
	"sort"
	"time"

	"gorm.io/gorm"
)

// RevenueByCurrency is an exact revenue aggregate. AmountMinor is always
// expressed in the smallest unit of Currency; values from different
// currencies must never be added together.
type RevenueByCurrency struct {
	Currency    string `json:"currency"`
	AmountMinor int64  `json:"amount_minor"`
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

	// Revenue is grouped by currency and summed from the exact minor-unit
	// column. Summing total_amount here would both reintroduce floating-point
	// drift and silently produce nonsense for mixed-currency orders.
	totalRevenue, err := r.revenueByCurrency(r.db.Model(&order.Order{}).Where("status != ?", "cancelled"))
	if err != nil {
		return nil, err
	}
	stats["total_revenue_by_currency"] = totalRevenue

	todayRevenue, err := r.revenueByCurrency(r.db.Model(&order.Order{}).Where("created_at >= ? AND status != ?", today, "cancelled"))
	if err != nil {
		return nil, err
	}
	stats["today_revenue_by_currency"] = todayRevenue

	return stats, nil
}

// GetSalesByDateRange 获取日期范围内的销售数据
func (r *OrderRepository) GetSalesByDateRange(startDate, endDate time.Time) ([]map[string]interface{}, error) {
	var results []struct {
		Date        string
		Count       int64
		Currency    string
		AmountMinor int64
	}

	err := r.db.Model(&order.Order{}).
		Select("DATE(created_at) as date, currency, COUNT(*) as count, COALESCE(SUM(total_amount_minor), 0) as amount_minor").
		Where("created_at BETWEEN ? AND ? AND status != ?", startDate, endDate, "cancelled").
		Group("DATE(created_at), currency").
		Order("date ASC, currency ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	// Merge one row per date while retaining a separate exact amount for each
	// currency. This keeps the chart contract stable (one point per date)
	// without ever adding incompatible currencies.
	type dateAggregate struct {
		count   int64
		revenue []RevenueByCurrency
	}
	aggregates := make(map[string]*dateAggregate)
	for _, result := range results {
		aggregate := aggregates[result.Date]
		if aggregate == nil {
			aggregate = &dateAggregate{}
			aggregates[result.Date] = aggregate
		}
		aggregate.count += result.Count
		aggregate.revenue = append(aggregate.revenue, RevenueByCurrency{
			Currency:    result.Currency,
			AmountMinor: result.AmountMinor,
		})
	}

	dates := make([]string, 0, len(aggregates))
	for date := range aggregates {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	data := make([]map[string]interface{}, 0, len(dates))
	for _, date := range dates {
		aggregate := aggregates[date]
		data = append(data, map[string]interface{}{
			"date":                date,
			"count":               aggregate.count,
			"revenue_by_currency": aggregate.revenue,
		})
	}

	return data, nil
}

func (r *OrderRepository) revenueByCurrency(query *gorm.DB) ([]RevenueByCurrency, error) {
	var rows []RevenueByCurrency
	if err := query.
		Select("currency, COALESCE(SUM(total_amount_minor), 0) as amount_minor").
		Group("currency").
		Order("currency ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	if rows == nil {
		rows = []RevenueByCurrency{}
	}
	return rows, nil
}
