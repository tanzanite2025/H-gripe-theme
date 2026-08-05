package admin

import (
	"net/http"

	"tanzanite/internal/api/middleware"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type CommercialCrawlerProtectionHandler struct {
	orderService *service.OrderService
}

func NewCommercialCrawlerProtectionHandler(orderServices ...*service.OrderService) *CommercialCrawlerProtectionHandler {
	var orderService *service.OrderService
	if len(orderServices) > 0 {
		orderService = orderServices[0]
	}
	return &CommercialCrawlerProtectionHandler{
		orderService: orderService,
	}
}

func (h *CommercialCrawlerProtectionHandler) GetStatus(c *gin.Context) {
	snapshot := middleware.CommercialCrawlerProtectionSnapshot()
	if h.orderService != nil {
		snapshot["order_number_protection"] = h.orderService.ProtectedOrderNumberStatus()
	}
	c.JSON(http.StatusOK, snapshot)
}
