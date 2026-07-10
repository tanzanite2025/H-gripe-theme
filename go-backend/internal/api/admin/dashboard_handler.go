package admin

import (
	"net/http"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	c.JSON(http.StatusOK, h.dashboardService.GetStats())
}

func (h *DashboardHandler) GetRecentOrders(c *gin.Context) {
	orders, err := h.dashboardService.GetRecentOrders(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *DashboardHandler) GetRecentUsers(c *gin.Context) {
	users, err := h.dashboardService.GetRecentUsers(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (h *DashboardHandler) GetRecentTickets(c *gin.Context) {
	tickets, err := h.dashboardService.GetRecentTickets(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get recent tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}

func (h *DashboardHandler) GetSalesChart(c *gin.Context) {
	salesChart, err := h.dashboardService.GetSalesChart(30)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sales data"})
		return
	}

	c.JSON(http.StatusOK, salesChart)
}
