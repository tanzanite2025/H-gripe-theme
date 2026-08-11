package admin

import (
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

func (h *ShippingHandler) QuoteShipping(c *gin.Context) {
	var req service.ShippingQuoteInput
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	quote, err := h.shippingService.QuoteCart(req)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, quote)
}
