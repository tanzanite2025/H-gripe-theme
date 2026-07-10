package admin

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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

	orders, _, err := h.orderService.ListAdminOrders(1, 10000, status, "", "", "", startDate, endDate)
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
