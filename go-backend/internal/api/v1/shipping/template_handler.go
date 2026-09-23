package shipping

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.shippingService.ListPublicTemplates()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": templates})
}

func (h *Handler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid template id")
		return
	}

	template, err := h.shippingService.GetPublicTemplate(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Template")
		return
	}

	response.Success(c, template)
}

func (h *Handler) CalculateShipping(c *gin.Context) {
	var req struct {
		TemplateID  uint    `json:"template_id" binding:"required"`
		Weight      float64 `json:"weight"`
		Quantity    int     `json:"quantity"`
		AmountMinor int64   `json:"amount_minor"`
		Country     string  `json:"country" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	quote, err := h.shippingService.CalculateShipping(service.ShippingCalculationInput{
		TemplateID:  req.TemplateID,
		Weight:      req.Weight,
		Quantity:    req.Quantity,
		AmountMinor: req.AmountMinor,
		Country:     req.Country,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidShippingDestination) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrCountryNotSupported) {
			apierror.RespondError(c, http.StatusUnprocessableEntity, "country_not_supported", err.Error())
			return
		}
		if errors.Is(err, service.ErrShippingRateConfigurationInvalid) {
			apierror.RespondInternalError(c, err)
			return
		}
		if errors.Is(err, service.ErrShippingRateUnavailable) {
			apierror.RespondError(c, http.StatusUnprocessableEntity, "shipping_rate_unavailable", err.Error())
			return
		}
		apierror.RespondNotFound(c, "Template")
		return
	}

	response.Success(c, quote)
}

func (h *Handler) QuoteShipping(c *gin.Context) {
	var req service.ShippingQuoteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	quote, err := h.shippingService.QuoteCart(req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidShippingDestination) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrCountryNotSupported) {
			apierror.RespondError(c, http.StatusUnprocessableEntity, "country_not_supported", err.Error())
			return
		}
		if errors.Is(err, service.ErrShippingQuoteExpired) || errors.Is(err, service.ErrShippingQuoteStale) {
			apierror.RespondError(c, http.StatusConflict, "shipping_quote_stale", err.Error())
			return
		}
		if errors.Is(err, service.ErrShippingQuotePlanUnavailable) {
			apierror.RespondError(c, http.StatusUnprocessableEntity, "shipping_quote_plan_unavailable", err.Error())
			return
		}
		if errors.Is(err, service.ErrShippingRateConfigurationInvalid) {
			apierror.RespondInternalError(c, err)
			return
		}
		if errors.Is(err, service.ErrShippingRateUnavailable) {
			apierror.RespondError(c, http.StatusUnprocessableEntity, "shipping_rate_unavailable", err.Error())
			return
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, quote)
}
