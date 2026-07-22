package admin

import (
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

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
