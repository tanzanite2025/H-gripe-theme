package order

import (
	"encoding/json"
	"strings"
	orderdomain "tanzanite/internal/domain/order"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListPublicChatOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	limit := pagination.ParseLimit(c)

	orders, _, err := h.orderService.GetUserOrders(userID.(uint), 1, limit)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	items := make([]gin.H, 0, len(orders))
	for _, item := range orders {
		items = append(items, makePublicChatOrder(item))
	}

	c.JSON(200, items)
}

func makePublicChatOrder(item orderdomain.Order) gin.H {
	title := "Order #" + item.OrderNumber
	if item.OrderNumber == "" {
		title = "Order"
	}

	return gin.H{
		"order_number":    item.OrderNumber,
		"title":           title,
		"status":          item.Status,
		"payment_status":  item.PaymentStatus,
		"shipping_status": item.ShippingStatus,
		"total":           item.TotalAmount,
		"currency":        item.Currency,
		"date":            item.CreatedAt.Format("2006-01-02"),
		"created_at":      item.CreatedAt.Format(time.RFC3339),
		"url":             "",
		"thumbnail":       "",
		"item_count":      publicChatOrderItemCount(item.Items),
		"items":           makePublicChatOrderItems(item.Items),
	}
}

func publicChatOrderItemCount(items []orderdomain.OrderItem) int {
	count := 0
	for _, item := range items {
		if item.Quantity > 0 {
			count += item.Quantity
		}
	}
	return count
}

func makePublicChatOrderItems(items []orderdomain.OrderItem) []gin.H {
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, gin.H{
			"product_id":   item.ProductID,
			"variant_id":   item.VariantID,
			"product_name": item.ProductName,
			"sku":          item.SKU,
			"quantity":     item.Quantity,
			"price":        item.Price,
			"subtotal":     item.Subtotal,
			"total":        item.Total,
			"attributes":   parsePublicChatOrderItemAttributes(item.Attributes),
		})
	}
	return result
}

func parsePublicChatOrderItemAttributes(value string) interface{} {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return value
	}
	return payload
}
