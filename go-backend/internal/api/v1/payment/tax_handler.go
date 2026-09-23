package payment

import (
	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListTaxRates(c *gin.Context) {
	rates, err := h.paymentService.ListPublicTaxRates()
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

	rate, err := h.paymentService.GetPublicTaxRate(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Tax rate")
		return
	}

	response.Success(c, rate)
}

func (h *Handler) CalculateTax(c *gin.Context) {
	var req struct {
		AmountMinor int64  `json:"amount_minor" binding:"required,gt=0"`
		Country     string `json:"country" binding:"required"`
		State       string `json:"state"`
		PostalCode  string `json:"postal_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	amountMoney, err := domainmoney.New(req.AmountMinor, currency.DefaultPrimaryCurrency)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid amount")
		return
	}
	taxRateDecimal, taxMoney, err := h.paymentService.CalculateTaxMoney(amountMoney, req.Country, req.State, req.PostalCode)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	totalMoney, err := amountMoney.Add(taxMoney)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{
		"amount_minor":     req.AmountMinor,
		"tax_rate_decimal": taxRateDecimal,
		"tax_minor":        taxMoney.AmountMinor(),
		"total_minor":      totalMoney.AmountMinor(),
		"currency":         currency.DefaultPrimaryCurrency,
	})
}
