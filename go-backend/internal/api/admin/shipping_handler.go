package admin

import (
	"strconv"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type ShippingHandler struct {
	shippingService *service.ShippingService
}

func NewShippingHandler(shippingService *service.ShippingService) *ShippingHandler {
	return &ShippingHandler{
		shippingService: shippingService,
	}
}

func (h *ShippingHandler) ListPackagingRules(c *gin.Context) {
	rules, err := h.shippingService.ListPackagingRules()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": rules})
}

func (h *ShippingHandler) GetPackagingRule(c *gin.Context) {
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

func (h *ShippingHandler) CreatePackagingRule(c *gin.Context) {
	var rule shippingdomain.PackagingRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.CreatePackagingRule(&rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, rule)
}

func (h *ShippingHandler) UpdatePackagingRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid rule id")
		return
	}

	var rule shippingdomain.PackagingRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	rule.ID = uint(id)
	if err := h.shippingService.UpdatePackagingRule(&rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, rule)
}

func (h *ShippingHandler) DeletePackagingRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid rule id")
		return
	}

	if err := h.shippingService.DeletePackagingRule(uint(id)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "packaging rule deleted successfully", nil)
}

func (h *ShippingHandler) CreatePackagingRuleApply(c *gin.Context) {
	var apply shippingdomain.PackagingRuleApply
	if err := c.ShouldBindJSON(&apply); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.CreatePackagingRuleApply(&apply); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, apply)
}

func (h *ShippingHandler) DeletePackagingRuleApply(c *gin.Context) {
	applyID, err := strconv.ParseUint(c.Param("applyId"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid apply id")
		return
	}

	if err := h.shippingService.DeletePackagingRuleApply(uint(applyID)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "packaging rule apply record deleted", nil)
}

func (h *ShippingHandler) ListCarriers(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	carriers, err := h.shippingService.ListCarriers(enabledOnly)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": carriers})
}

func (h *ShippingHandler) GetCarrier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid carrier id")
		return
	}

	carrier, err := h.shippingService.GetCarrier(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Carrier")
		return
	}

	response.Success(c, carrier)
}

func (h *ShippingHandler) CreateCarrier(c *gin.Context) {
	var carrier shippingdomain.Carrier
	if err := c.ShouldBindJSON(&carrier); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.CreateCarrier(&carrier); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, carrier)
}

func (h *ShippingHandler) UpdateCarrier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid carrier id")
		return
	}

	var carrier shippingdomain.Carrier
	if err := c.ShouldBindJSON(&carrier); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	carrier.ID = uint(id)
	if err := h.shippingService.UpdateCarrier(&carrier); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, carrier)
}

func (h *ShippingHandler) DeleteCarrier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid carrier id")
		return
	}

	if err := h.shippingService.DeleteCarrier(uint(id)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "carrier deleted", nil)
}
