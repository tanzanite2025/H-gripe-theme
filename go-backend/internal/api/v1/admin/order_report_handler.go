package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetOrderStats 获取订单统计
// GET /api/admin/orders/stats
func (h *OrderHandler) GetOrderStats(c *gin.Context) {
	stats, err := h.orderService.GetAdminStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get order stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSalesChart 获取销售图表数据
// GET /api/admin/orders/sales-chart
func (h *OrderHandler) GetSalesChart(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	if days < 1 || days > 365 {
		days = 30
	}

	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	data, err := h.orderService.GetSalesByDateRange(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sales chart data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       data,
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
	})
}
