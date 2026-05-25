package admin

import (
	"net/http"
	"tanzanite/internal/repository"
	"time"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	orderRepo        *repository.OrderRepository
	userRepo         *repository.UserRepository
	ticketRepo       *repository.TicketRepository
	subscriptionRepo *repository.SubscriptionRepository
}

func NewDashboardHandler(
	orderRepo *repository.OrderRepository,
	userRepo *repository.UserRepository,
	ticketRepo *repository.TicketRepository,
	subscriptionRepo *repository.SubscriptionRepository,
) *DashboardHandler {
	return &DashboardHandler{
		orderRepo:        orderRepo,
		userRepo:         userRepo,
		ticketRepo:       ticketRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

// GetStats 获取仪表板统计数据
// GET /api/admin/dashboard/stats
func (h *DashboardHandler) GetStats(c *gin.Context) {
	// 获取今日日期
	today := time.Now().Truncate(24 * time.Hour)

	// 订单统计
	orderStats, err := h.orderRepo.GetStats()
	if err != nil {
		orderStats = make(map[string]interface{})
	}

	// 用户统计
	totalUsers, err := h.userRepo.Count()
	if err != nil {
		totalUsers = 0
	}

	// 今日新用户
	todayUsers, err := h.userRepo.CountByDateRange(today, time.Now())
	if err != nil {
		todayUsers = 0
	}

	// 工单统计
	ticketStats, err := h.ticketRepo.GetStats()
	if err != nil {
		ticketStats = make(map[string]interface{})
	}

	// 订阅统计
	subscriptionStats, err := h.subscriptionRepo.GetStats()
	if err != nil {
		subscriptionStats = make(map[string]interface{})
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": gin.H{
			"total":         orderStats["total"],
			"today":         orderStats["today"],
			"pending":       orderStats["pending"],
			"processing":    orderStats["processing"],
			"completed":     orderStats["completed"],
			"revenue":       orderStats["total_revenue"],
			"today_revenue": orderStats["today_revenue"],
		},
		"users": gin.H{
			"total": totalUsers,
			"today": todayUsers,
		},
		"tickets": gin.H{
			"total":   ticketStats["total"],
			"open":    ticketStats["open"],
			"pending": ticketStats["pending"],
		},
		"subscriptions": gin.H{
			"total":  subscriptionStats["total"],
			"active": subscriptionStats["active"],
		},
		"timestamp": time.Now(),
	})
}

// GetRecentOrders 获取最近订单
// GET /api/admin/dashboard/recent-orders
func (h *DashboardHandler) GetRecentOrders(c *gin.Context) {
	limit := 10
	orders, err := h.orderRepo.FindRecent(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
	})
}

// GetRecentUsers 获取最近注册用户
// GET /api/admin/dashboard/recent-users
func (h *DashboardHandler) GetRecentUsers(c *gin.Context) {
	limit := 10
	users, err := h.userRepo.FindRecent(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent users"})
		return
	}

	// 转换为响应格式（不包含敏感信息）
	userResponses := make([]interface{}, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"users": userResponses,
	})
}

// GetRecentTickets 获取最近工单
// GET /api/admin/dashboard/recent-tickets
func (h *DashboardHandler) GetRecentTickets(c *gin.Context) {
	limit := 10
	tickets, err := h.ticketRepo.FindRecent(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tickets": tickets,
	})
}

// GetSalesChart 获取销售图表数据
// GET /api/admin/dashboard/sales-chart
func (h *DashboardHandler) GetSalesChart(c *gin.Context) {
	// 获取最近30天的销售数据
	days := 30
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	salesData, err := h.orderRepo.GetSalesByDateRange(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sales data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       salesData,
		"start_date": startDate,
		"end_date":   endDate,
	})
}
