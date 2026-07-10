package order

import (
	"strconv"
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
		title = "Order #" + strconv.FormatUint(uint64(item.ID), 10)
	}

	return gin.H{
		"id":           item.ID,
		"order_number": item.OrderNumber,
		"title":        title,
		"status":       item.Status,
		"total":        item.TotalAmount,
		"currency":     "USD",
		"date":         item.CreatedAt.Format("2006-01-02"),
		"created_at":   item.CreatedAt.Format(time.RFC3339),
		"url":          "/orders/" + strconv.FormatUint(uint64(item.ID), 10),
		"thumbnail":    "",
	}
}
