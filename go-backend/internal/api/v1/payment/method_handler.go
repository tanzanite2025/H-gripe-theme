package payment

import (
	"strconv"
	"tanzanite/internal/domain/payment"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ Payment Method 相关接口 ============

// ListPaymentMethods 获取支付方式列表
// @Summary 获取支付方式列表
// @Tags Payment
// @Produce json
// @Param enabled query bool false "只显示启用的"
// @Success 200 {array} payment.PaymentMethod
// @Router /api/v1/payment/methods [get]
func (h *Handler) ListPaymentMethods(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	methods, err := h.paymentService.ListPaymentMethods(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": methods})
}

// GetPaymentMethod 获取支付方式详情
// @Summary 获取支付方式详情
// @Tags Payment
// @Produce json
// @Param id path int true "支付方式ID"
// @Success 200 {object} payment.PaymentMethod
// @Router /api/v1/payment/methods/{id} [get]
func (h *Handler) GetPaymentMethod(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid payment method id")
		return
	}

	method, err := h.paymentService.GetPaymentMethod(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Payment method")
		return
	}

	response.Success(c, method)
}

// CreatePaymentMethod 创建支付方式（管理员）
// @Summary 创建支付方式
// @Tags Payment
// @Accept json
// @Produce json
// @Param method body payment.PaymentMethod true "支付方式信息"
// @Success 201 {object} payment.PaymentMethod
// @Router /api/v1/admin/payment/methods [post]
func (h *Handler) CreatePaymentMethod(c *gin.Context) {
	var method payment.PaymentMethod
	if err := c.ShouldBindJSON(&method); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.paymentService.CreatePaymentMethod(&method); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, method)
}

// UpdatePaymentMethod 更新支付方式（管理员）
// @Summary 更新支付方式
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path int true "支付方式ID"
// @Param method body payment.PaymentMethod true "支付方式信息"
// @Success 200 {object} payment.PaymentMethod
// @Router /api/v1/admin/payment/methods/{id} [put]
func (h *Handler) UpdatePaymentMethod(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid payment method id")
		return
	}

	var method payment.PaymentMethod
	if err := c.ShouldBindJSON(&method); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	method.ID = uint(id)
	if err := h.paymentService.UpdatePaymentMethod(&method); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, method)
}

// DeletePaymentMethod 删除支付方式（管理员）
// @Summary 删除支付方式
// @Tags Payment
// @Produce json
// @Param id path int true "支付方式ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/payment/methods/{id} [delete]
func (h *Handler) DeletePaymentMethod(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid payment method id")
		return
	}

	if err := h.paymentService.DeletePaymentMethod(uint(id)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "payment method deleted", nil)
}
