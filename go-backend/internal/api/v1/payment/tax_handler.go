package payment

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListTaxRates(c *gin.Context) {
	rates, err := h.paymentService.ListTaxRates()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": rates})
}

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

	taxRate, tax, err := h.paymentService.CalculateTax(req.Amount, req.Country, req.State)
	if err != nil {
		response.Success(c, gin.H{
			"amount":   req.Amount,
			"tax_rate": 0.0,
			"tax":      0.0,
			"total":    req.Amount,
		})
		return
	}

	total := req.Amount + tax

	response.Success(c, gin.H{
		"amount":   req.Amount,
		"tax_rate": taxRate,
		"tax":      tax,
		"total":    total,
	})
}
