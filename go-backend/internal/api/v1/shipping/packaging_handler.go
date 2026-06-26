package shipping

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/shipping"

	"github.com/gin-gonic/gin"
)

// ============ Packaging Rule 相关接口 ============

// ListPackagingRules 获取所有包装规则
func (h *Handler) ListPackagingRules(c *gin.Context) {
	rules, err := h.shippingRepo.FindAllPackagingRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rules})
}

// GetPackagingRule 获取包装规则详情
func (h *Handler) GetPackagingRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	rule, err := h.shippingRepo.FindPackagingRuleByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// CreatePackagingRule 创建包装规则
func (h *Handler) CreatePackagingRule(c *gin.Context) {
	var rule shipping.PackagingRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.shippingRepo.CreatePackagingRule(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// UpdatePackagingRule 更新包装规则
func (h *Handler) UpdatePackagingRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	var rule shipping.PackagingRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule.ID = uint(id)
	if err := h.shippingRepo.UpdatePackagingRule(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeletePackagingRule 删除包装规则
func (h *Handler) DeletePackagingRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rule id"})
		return
	}

	if err := h.shippingRepo.DeletePackagingRule(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "packaging rule deleted successfully"})
}

// CreatePackagingRuleApply 新增规则适用商品
func (h *Handler) CreatePackagingRuleApply(c *gin.Context) {
	var apply shipping.PackagingRuleApply
	if err := c.ShouldBindJSON(&apply); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.shippingRepo.CreatePackagingRuleApply(&apply); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, apply)
}

// DeletePackagingRuleApply 删除规则适用商品
func (h *Handler) DeletePackagingRuleApply(c *gin.Context) {
	applyID, err := strconv.ParseUint(c.Param("applyId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid apply id"})
		return
	}

	if err := h.shippingRepo.DeletePackagingRuleApply(uint(applyID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "packaging rule apply record deleted"})
}

// GetProductPackagingRules 获取某产品关联的包装规则
func (h *Handler) GetProductPackagingRules(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	rules, err := h.shippingRepo.FindPackagingRulesByProductID(uint(productID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}
