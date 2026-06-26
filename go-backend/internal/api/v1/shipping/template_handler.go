package shipping

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/shipping"

	"github.com/gin-gonic/gin"
)

// ============ Shipping Template 相关接口 ============

// ListTemplates 获取运费模板列表
// @Summary 获取运费模板列表
// @Tags Shipping
// @Produce json
// @Success 200 {array} shipping.ShippingTemplate
// @Router /api/v1/shipping/templates [get]
func (h *Handler) ListTemplates(c *gin.Context) {
	templates, err := h.shippingRepo.FindAllTemplates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": templates})
}

// GetTemplate 获取运费模板详情
// @Summary 获取运费模板详情
// @Tags Shipping
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} shipping.ShippingTemplate
// @Router /api/v1/shipping/templates/{id} [get]
func (h *Handler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	template, err := h.shippingRepo.FindTemplateByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// CalculateShipping 计算运费
// @Summary 计算运费
// @Tags Shipping
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "计算请求"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/shipping/calculate [post]
func (h *Handler) CalculateShipping(c *gin.Context) {
	var req struct {
		TemplateID uint    `json:"template_id" binding:"required"`
		Weight     float64 `json:"weight"`
		Quantity   int     `json:"quantity"`
		Amount     float64 `json:"amount"`
		Country    string  `json:"country" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 获取模板
	template, err := h.shippingRepo.FindTemplateByID(req.TemplateID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "template not found"})
		return
	}

	// 检查是否免运费
	if template.FreeShipping && req.Amount >= template.FreeThreshold {
		c.JSON(http.StatusOK, gin.H{
			"shipping_fee":  0.0,
			"free_shipping": true,
		})
		return
	}

	// 根据模板类型计算运费
	var value float64
	switch template.Type {
	case "weight":
		value = req.Weight
	case "quantity":
		value = float64(req.Quantity)
	case "price":
		value = req.Amount
	default:
		value = req.Weight
	}

	// 查找匹配的规则
	var shippingFee float64
	for _, rule := range template.Rules {
		if value >= rule.MinValue && (rule.MaxValue == 0 || value <= rule.MaxValue) {
			shippingFee = rule.Fee
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"shipping_fee":  shippingFee,
		"free_shipping": false,
	})
}

// CreateTemplate 创建运费模板（管理员）
// @Summary 创建运费模板
// @Tags Shipping
// @Accept json
// @Produce json
// @Param template body shipping.ShippingTemplate true "模板信息"
// @Success 201 {object} shipping.ShippingTemplate
// @Router /api/v1/admin/shipping/templates [post]
func (h *Handler) CreateTemplate(c *gin.Context) {
	var template shipping.ShippingTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.shippingRepo.CreateTemplate(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, template)
}

// UpdateTemplate 更新运费模板（管理员）
// @Summary 更新运费模板
// @Tags Shipping
// @Accept json
// @Produce json
// @Param id path int true "模板ID"
// @Param template body shipping.ShippingTemplate true "模板信息"
// @Success 200 {object} shipping.ShippingTemplate
// @Router /api/v1/admin/shipping/templates/{id} [put]
func (h *Handler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	var template shipping.ShippingTemplate
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template.ID = uint(id)
	if err := h.shippingRepo.UpdateTemplate(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate 删除运费模板（管理员）
// @Summary 删除运费模板
// @Tags Shipping
// @Produce json
// @Param id path int true "模板ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/shipping/templates/{id} [delete]
func (h *Handler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	if err := h.shippingRepo.DeleteTemplate(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "template deleted"})
}
