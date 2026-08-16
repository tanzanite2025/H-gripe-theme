package admin

import (
	orderdomain "commerce-platform/internal/domain/order"
	"encoding/csv"
	"net/http"
	"strconv"
	"strings"

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

func sanitizeCSVRow(row []string) []string {
	for i, value := range row {
		row[i] = sanitizeCSVField(value)
	}
	return row
}

func orderCustomsExportRow(record orderdomain.Order, item orderdomain.OrderItem) []string {
	recipient := strings.TrimSpace(strings.Join([]string{
		record.ShippingAddress.FirstName,
		record.ShippingAddress.LastName,
	}, " "))
	declaredValue := ""
	if item.DeclaredValue != nil {
		declaredValue = strconv.FormatFloat(*item.DeclaredValue, 'f', 2, 64)
	}
	confirmed := "pending"
	if item.DeclaredValueConfirmed {
		confirmed = "confirmed"
	}

	return []string{
		record.OrderNumber,
		recipient,
		record.ShippingAddress.Company,
		record.ShippingAddress.Phone,
		record.ShippingAddress.Email,
		record.ShippingAddress.Address1,
		record.ShippingAddress.Address2,
		record.ShippingAddress.City,
		record.ShippingAddress.State,
		record.ShippingAddress.PostalCode,
		record.ShippingAddress.Country,
		strconv.FormatUint(uint64(item.ID), 10),
		item.ProductName,
		item.SKU,
		strconv.FormatUint(uint64(item.Quantity), 10),
		item.HSCode,
		item.CNCode,
		item.CountryOfOrigin,
		item.CustomsDescription,
		declaredValue,
		confirmed,
	}
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
	if err := writer.Write(sanitizeCSVRow(header)); err != nil {
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
		if err := writer.Write(sanitizeCSVRow(row)); err != nil {
			return
		}
	}

	writer.Flush()
}

// ExportOrderCustoms 导出单个订单的清关资料
// GET /api/admin/orders/:id/customs-export
func (h *OrderHandler) ExportOrderCustoms(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	record, err := h.orderService.GetAdminOrder(uint(id))
	if err != nil {
		respondOrderServiceError(c, err, "Failed to fetch order", http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="customs_data.csv"`)
	_, _ = c.Writer.Write([]byte("\xEF\xBB\xBF"))

	writer := csv.NewWriter(c.Writer)
	header := []string{
		"Order Number",
		"Recipient Name",
		"Company",
		"Phone",
		"Email",
		"Address 1",
		"Address 2",
		"City",
		"State / Province",
		"Postal Code",
		"Country",
		"Order Item ID",
		"Product Name",
		"SKU",
		"Quantity",
		"HS Code",
		"CN Code",
		"Country of Origin",
		"Customs Description",
		"Declared Value",
		"Declared Value Status",
	}
	if err := writer.Write(sanitizeCSVRow(header)); err != nil {
		return
	}

	for _, item := range record.Items {
		if err := writer.Write(sanitizeCSVRow(orderCustomsExportRow(*record, item))); err != nil {
			return
		}
	}
	writer.Flush()
}
