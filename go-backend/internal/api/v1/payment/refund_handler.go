package payment

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/payment"

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
	var refund payment.Refund
	if err := c.ShouldBindJSON(&refund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认状态
	refund.Status = "pending"

	if err := h.paymentRepo.CreateRefund(&refund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, refund)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund id"})
		return
	}

	refund, err := h.paymentRepo.FindRefundByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, refund)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	refunds, err := h.paymentRepo.FindRefundsByOrderID(uint(orderID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": refunds})
}

// UpdateRefundStatus 更新退款状态（管理员）
// @Summary 更新退款状态
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path int true "退款ID"
// @Param request body map[string]string true "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/payment/refunds/{id}/status [put]
func (h *Handler) UpdateRefundStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refund id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refund, err := h.paymentRepo.FindRefundByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	refund.Status = req.Status
	if err := h.paymentRepo.UpdateRefund(refund); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "refund status updated"})
}
