package shipping

import (
	"strconv"
	"tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ============ Packaging Rule 相关接口 ============

// ListPackagingRules 获取所有包装规则
func (h *Handler) ListPackagingRules(c *gin.Context) {
	rules, err := h.shippingRepo.FindAllPackagingRules()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": rules})
}

// GetPackagingRule 获取包装规则详情
func (h *Handler) GetPackagingRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid rule id")
		return
	}

	rule, err := h.shippingRepo.FindPackagingRuleByID(uint(id))
	if err != nil {
		apierror.RespondNotFound(c, "Packaging rule")
		return
	}

	response.Success(c, rule)
}

// CreatePackagingRule 创建包装规则
func (h *Handler) CreatePackagingRule(c *gin.Context) {
	var rule shipping.PackagingRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingRepo.CreatePackagingRule(&rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, rule)
}

// UpdatePackagingRule 更新包装规则
func (h *Handler) UpdatePackagingRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid rule id")
		return
	}

	var rule shipping.PackagingRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	rule.ID = uint(id)
	if err := h.shippingRepo.UpdatePackagingRule(&rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, rule)
}

// DeletePackagingRule 删除包装规则
func (h *Handler) DeletePackagingRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid rule id")
		return
	}

	if err := h.shippingRepo.DeletePackagingRule(uint(id)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "packaging rule deleted successfully", nil)
}

// CreatePackagingRuleApply 新增规则适用商品
func (h *Handler) CreatePackagingRuleApply(c *gin.Context) {
	var apply shipping.PackagingRuleApply
	if err := c.ShouldBindJSON(&apply); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingRepo.CreatePackagingRuleApply(&apply); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, apply)
}

// DeletePackagingRuleApply 删除规则适用商品
func (h *Handler) DeletePackagingRuleApply(c *gin.Context) {
	applyID, err := strconv.ParseUint(c.Param("applyId"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid apply id")
		return
	}

	if err := h.shippingRepo.DeletePackagingRuleApply(uint(applyID)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "packaging rule apply record deleted", nil)
}

// GetProductPackagingRules 获取某产品关联的包装规则
func (h *Handler) GetProductPackagingRules(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid product id")
		return
	}

	rules, err := h.shippingRepo.FindPackagingRulesByProductID(uint(productID))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, rules)
}
