package payment

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/payment"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transaction id"})
		return
	}

	transaction, err := h.paymentRepo.FindTransactionByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transaction)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	transactions, err := h.paymentRepo.FindTransactionByOrderID(uint(orderID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": transactions})
}

// CreateTransaction 创建交易记录
// @Summary 创建交易记录
// @Tags Payment
// @Accept json
// @Produce json
// @Param transaction body payment.Transaction true "交易信息"
// @Success 201 {object} payment.Transaction
// @Router /api/v1/payment/transactions [post]
func (h *Handler) CreateTransaction(c *gin.Context) {
	var transaction payment.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.paymentRepo.CreateTransaction(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, transaction)
}
