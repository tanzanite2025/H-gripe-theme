package payment

import (
	"strconv"
	"tanzanite/internal/domain/payment"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ Refund 相关接口 ============

// CreateRefund 创建退款
// @Summary 创建退款
// @Tags Payment
// @Accept json
// @Produce json
// @Param refund body payment.Refund true "退款信息"
// @Success 201 {object} payment.Refund
// @Router /api/v1/payment/refunds [post]
func (h *Handler) CreateRefund(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	var req struct {
		OrderID       uint    `json:"order_id" binding:"required"`
		TransactionID uint    `json:"transaction_id" binding:"required"`
		Amount        float64 `json:"amount" binding:"required,gt=0"`
		Reason        string  `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	refund := payment.Refund{
		OrderID:       req.OrderID,
		TransactionID: req.TransactionID,
		Amount:        req.Amount,
		Reason:        req.Reason,
	}

	if err := h.paymentService.CreateAdminRefund(&refund, userIDValue.(uint)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, refund)
}

// GetRefund 获取退款详情
// @Summary 获取退款详情
// @Tags Payment
// @Produce json
// @Param id path int true "退款ID"
// @Success 200 {object} payment.Refund
// @Router /api/v1/payment/refunds/{id} [get]
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

// GetOrderRefunds 获取订单的退款记录
// @Summary 获取订单的退款记录
// @Tags Payment
// @Produce json
// @Param order_id path int true "订单ID"
// @Success 200 {array} payment.Refund
// @Router /api/v1/payment/orders/{order_id}/refunds [get]
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
