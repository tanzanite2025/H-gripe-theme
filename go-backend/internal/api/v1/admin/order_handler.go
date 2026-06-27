package admin

import (
	"encoding/csv"
	"net/http"
	"strconv"
	"tanzanite/internal/repository"
	"tanzanite/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderRepo    *repository.OrderRepository
	orderService *service.OrderService
}

func NewOrderHandler(orderRepo *repository.OrderRepository, orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderRepo:    orderRepo,
		orderService: orderService,
	}
}

// ListOrders 获取订单列表
// GET /api/admin/orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	paymentStatus := c.Query("payment_status")
	shippingStatus := c.Query("shipping_status")
	search := c.Query("search")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := h.orderRepo.FindAllWithFilters(page, pageSize, status, paymentStatus, shippingStatus, search, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetOrder 获取订单详情
// GET /api/admin/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := h.orderRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order": order,
	})
}

// UpdateOrderStatus 更新订单状态
// PATCH /api/admin/orders/:id/status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=pending paid processing shipped completed cancelled refunded"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderService.UpdateOrderStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
	})
}

// UpdatePaymentStatus 更新支付状态
// PATCH /api/admin/orders/:id/payment-status
func (h *OrderHandler) UpdatePaymentStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		PaymentStatus string `json:"payment_status" binding:"required,oneof=unpaid paid refunded"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderRepo.UpdatePaymentStatus(uint(id), req.PaymentStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment status updated successfully",
	})
}

// UpdateShippingStatus 更新物流状态
// PATCH /api/admin/orders/:id/shipping-status
func (h *OrderHandler) UpdateShippingStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		ShippingStatus string `json:"shipping_status" binding:"required,oneof=pending processing shipped delivered"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderRepo.UpdateShippingStatus(uint(id), req.ShippingStatus); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update shipping status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Shipping status updated successfully",
	})
}

// UpdateTrackingInfo 更新物流追踪信息
// PATCH /api/admin/orders/:id/tracking
func (h *OrderHandler) UpdateTrackingInfo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		TrackingNumber string `json:"tracking_number" binding:"required"`
		CarrierCode    string `json:"carrier_code"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderRepo.UpdateTrackingInfo(uint(id), req.TrackingNumber, req.CarrierCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tracking info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tracking info updated successfully",
	})
}

// UpdateAdminNote 更新管理员备注
// PATCH /api/admin/orders/:id/admin-note
func (h *OrderHandler) UpdateAdminNote(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		AdminNote string `json:"admin_note"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingOrder, err := h.orderRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	existingOrder.AdminNote = req.AdminNote
	if err := h.orderRepo.Update(existingOrder); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin note"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin note updated successfully",
	})
}

// GetOrderStats 获取订单统计
// GET /api/admin/orders/stats
func (h *OrderHandler) GetOrderStats(c *gin.Context) {
	stats, err := h.orderRepo.GetStats()
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

	data, err := h.orderRepo.GetSalesByDateRange(startDate, endDate)
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

// sanitizeCSVField 进行安全的单元格防注入过滤
func sanitizeCSVField(val string) string {
	if len(val) > 0 {
		first := val[0]
		if first == '=' || first == '+' || first == '-' || first == '@' {
			return "'" + val
		}
	}
	return val
}

// ExportOrders 导出订单
// GET /api/admin/orders/export
func (h *OrderHandler) ExportOrders(c *gin.Context) {
	// 检查是否有导出权限
	userRole, exists := c.Get("user_role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	role := userRole.(string)
	if role != "admin" && role != "manager" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		return
	}

	status := c.Query("status")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	orders, _, err := h.orderRepo.FindAllWithFilters(1, 10000, status, "", "", "", startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	// 生成 CSV
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=orders.csv")

	// 写入 UTF-8 BOM 以防 Excel 乱码
	_, _ = c.Writer.Write([]byte("\xEF\xBB\xBF"))

	writer := csv.NewWriter(c.Writer)

	// CSV 头部
	header := []string{"Order Number", "Customer", "Status", "Payment Status", "Shipping Status", "Total Amount", "Created At"}
	for i, col := range header {
		header[i] = sanitizeCSVField(col)
	}
	if err := writer.Write(header); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV header"})
		return
	}

	// CSV 数据
	for _, order := range orders {
		customerName := order.ShippingAddress.FirstName + " " + order.ShippingAddress.LastName
		row := []string{
			order.OrderNumber,
			customerName,
			order.Status,
			order.PaymentStatus,
			order.ShippingStatus,
			strconv.FormatFloat(order.TotalAmount, 'f', 2, 64),
			order.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		for i, val := range row {
			row[i] = sanitizeCSVField(val)
		}
		if err := writer.Write(row); err != nil {
			return
		}
	}

	writer.Flush()
}

// BatchUpdateStatus 批量更新订单状态
// POST /api/admin/orders/batch-status
func (h *OrderHandler) BatchUpdateStatus(c *gin.Context) {
	var req struct {
		OrderIDs []uint `json:"order_ids" binding:"required,min=1"`
		Status   string `json:"status" binding:"required,oneof=pending paid processing shipped completed cancelled refunded"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated := 0
	failed := 0

	for _, id := range req.OrderIDs {
		if err := h.orderService.UpdateOrderStatus(id, req.Status); err == nil {
			updated++
		} else {
			failed++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch update completed",
		"updated": updated,
		"failed":  failed,
		"total":   len(req.OrderIDs),
	})
}

// DeleteOrder 删除订单
// DELETE /api/admin/orders/:id
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	if err := h.orderRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order deleted successfully",
	})
}
