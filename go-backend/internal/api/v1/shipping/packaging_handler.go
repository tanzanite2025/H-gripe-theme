package shipping

import (
	"strconv"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) ListPackagingRules(c *gin.Context) {
	rules, err := h.shippingService.ListPackagingRules()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": rules})
}

func (h *Handler) GetPackagingRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid rule id")
		return
	}

	rule, err := h.shippingService.GetPackagingRule(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Packaging rule")
		return
	}

	response.Success(c, rule)
}

func (h *Handler) GetProductPackagingRules(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid product id")
		return
	}

	rules, err := h.shippingService.GetProductPackagingRules(uint(productID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, rules)
}
