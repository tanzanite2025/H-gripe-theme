package payment

import (
	"errors"
	"fmt"
	"strconv"

	"commerce-platform/internal/domain/currency"
	domainmoney "commerce-platform/internal/domain/money"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListTaxRates(c *gin.Context) {
	if h == nil || h.taxRateService == nil {
		apierror.RespondInternalError(c, errors.New("tax rate service is not configured"))
		return
	}
	rates, err := h.taxRateService.ListPublicTaxRates()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": rates})
}

func (h *Handler) GetTaxRate(c *gin.Context) {
	if h == nil || h.taxRateService == nil {
		apierror.RespondInternalError(c, errors.New("tax rate service is not configured"))
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid tax rate id")
		return
	}

	rate, err := h.taxRateService.GetPublicTaxRate(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Tax rate")
		return
	}

	response.Success(c, rate)
}

func (h *Handler) CalculateTax(c *gin.Context) {
	if h == nil || h.taxRateService == nil {
		apierror.RespondInternalError(c, errors.New("tax rate service is not configured"))
		return
	}
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
	taxRateDecimal, taxMoney, err := h.taxRateService.CalculateTaxMoney(amountMoney, req.Country, req.State, req.PostalCode)
	if err != nil {
		if errors.Is(err, service.ErrTaxRateUnavailable) {
			apierror.RespondError(c, 422, "tax_rate_unavailable", err.Error())
			return
		}
		apierror.RespondInternalError(c, fmt.Errorf("calculate tax: %w", err))
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
