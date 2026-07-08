package payment

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetRefund(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid refund id")
		return
	}

	refund, err := h.paymentService.GetRefund(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Refund")
		return
	}
	if !h.authorizeOrderPaymentRead(c, refund.OrderID) {
		return
	}

	response.Success(c, refund)
}

func (h *Handler) GetOrderRefunds(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid order id")
		return
	}
	if !h.authorizeOrderPaymentRead(c, uint(orderID)) {
		return
	}

	refunds, err := h.paymentService.GetOrderRefunds(uint(orderID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": refunds})
}
