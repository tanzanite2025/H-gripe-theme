package payment

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/payment"

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
	rates, err := h.paymentRepo.FindAllTaxRates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": rates})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tax rate id"})
		return
	}

	rate, err := h.paymentRepo.FindTaxRateByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rate)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查找税率
	taxRate, err := h.paymentRepo.FindTaxRateByLocation(req.Country, req.State)
	if err != nil {
		// 没有找到税率，返回0
		c.JSON(http.StatusOK, gin.H{
			"amount":   req.Amount,
			"tax_rate": 0.0,
			"tax":      0.0,
			"total":    req.Amount,
		})
		return
	}

	// 计算税费
	tax := req.Amount * taxRate.Rate / 100
	total := req.Amount + tax

	c.JSON(http.StatusOK, gin.H{
		"amount":   req.Amount,
		"tax_rate": taxRate.Rate,
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.paymentRepo.CreateTaxRate(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rate)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tax rate id"})
		return
	}

	var rate payment.TaxRate
	if err := c.ShouldBindJSON(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rate.ID = uint(id)
	if err := h.paymentRepo.UpdateTaxRate(&rate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rate)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tax rate id"})
		return
	}

	if err := h.paymentRepo.DeleteTaxRate(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tax rate deleted"})
}
