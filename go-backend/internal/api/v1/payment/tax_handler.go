package payment

import (
	"strconv"
	"tanzanite/internal/domain/payment"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ Tax Rate 相关接口 ============

// ListTaxRates 获取税率列表
// @Summary 获取税率列表
// @Tags Payment
// @Produce json
// @Success 200 {array} payment.TaxRate
// @Router /api/v1/payment/tax-rates [get]
func (h *Handler) ListTaxRates(c *gin.Context) {
	rates, err := h.paymentService.ListTaxRates()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": rates})
}

// GetTaxRate 获取税率详情
// @Summary 获取税率详情
// @Tags Payment
// @Produce json
// @Param id path int true "税率ID"
// @Success 200 {object} payment.TaxRate
// @Router /api/v1/payment/tax-rates/{id} [get]
func (h *Handler) GetTaxRate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid tax rate id")
		return
	}

	rate, err := h.paymentService.GetTaxRate(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Tax rate")
		return
	}

	response.Success(c, rate)
}

// CalculateTax 计算税费
// @Summary 计算税费
// @Tags Payment
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "计算请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/payment/calculate-tax [post]
func (h *Handler) CalculateTax(c *gin.Context) {
	var req struct {
		Amount  float64 `json:"amount" binding:"required,gt=0"`
		Country string  `json:"country" binding:"required"`
		State   string  `json:"state"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	// 查找税率
	taxRate, tax, err := h.paymentService.CalculateTax(req.Amount, req.Country, req.State)
	if err != nil {
		// 没有找到税率，返回0
		response.Success(c, gin.H{
			"amount":   req.Amount,
			"tax_rate": 0.0,
			"tax":      0.0,
			"total":    req.Amount,
		})
		return
	}

	// 计算税费
	total := req.Amount + tax

	response.Success(c, gin.H{
		"amount":   req.Amount,
		"tax_rate": taxRate,
		"tax":      tax,
		"total":    total,
	})
}

// CreateTaxRate 创建税率（管理员）
// @Summary 创建税率
// @Tags Payment
// @Accept json
// @Produce json
// @Param rate body payment.TaxRate true "税率信息"
// @Success 201 {object} payment.TaxRate
// @Router /api/v1/admin/payment/tax-rates [post]
func (h *Handler) CreateTaxRate(c *gin.Context) {
	var rate payment.TaxRate
	if err := c.ShouldBindJSON(&rate); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.paymentService.CreateTaxRate(&rate); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, rate)
}

// UpdateTaxRate 更新税率（管理员）
// @Summary 更新税率
// @Tags Payment
// @Accept json
// @Produce json
// @Param id path int true "税率ID"
// @Param rate body payment.TaxRate true "税率信息"
// @Success 200 {object} payment.TaxRate
// @Router /api/v1/admin/payment/tax-rates/{id} [put]
func (h *Handler) UpdateTaxRate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid tax rate id")
		return
	}

	var rate payment.TaxRate
	if err := c.ShouldBindJSON(&rate); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	rate.ID = uint(id)
	if err := h.paymentService.UpdateTaxRate(&rate); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, rate)
}

// DeleteTaxRate 删除税率（管理员）
// @Summary 删除税率
// @Tags Payment
// @Produce json
// @Param id path int true "税率ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/payment/tax-rates/{id} [delete]
func (h *Handler) DeleteTaxRate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid tax rate id")
		return
	}

	if err := h.paymentService.DeleteTaxRate(uint(id)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "tax rate deleted", nil)
}
