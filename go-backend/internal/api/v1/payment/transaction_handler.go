package payment

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ Transaction 相关接口 ============

// GetTransaction 获取交易详情
// @Summary 获取交易详情
// @Tags Payment
// @Produce json
// @Param id path int true "交易ID"
// @Success 200 {object} payment.Transaction
// @Router /api/v1/payment/transactions/{id} [get]
func (h *Handler) GetTransaction(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid transaction id")
		return
	}

	transaction, err := h.paymentService.GetTransaction(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Transaction")
		return
	}
	if !h.authorizeOrderPaymentRead(c, transaction.OrderID) {
		return
	}

	response.Success(c, transaction)
}

// GetOrderTransactions 获取订单的交易记录
// @Summary 获取订单的交易记录
// @Tags Payment
// @Produce json
// @Param order_id path int true "订单ID"
// @Success 200 {array} payment.Transaction
// @Router /api/v1/payment/orders/{order_id}/transactions [get]
func (h *Handler) GetOrderTransactions(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid order id")
		return
	}
	if !h.authorizeOrderPaymentRead(c, uint(orderID)) {
		return
	}

	transactions, err := h.paymentService.GetOrderTransactions(uint(orderID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": transactions})
}
