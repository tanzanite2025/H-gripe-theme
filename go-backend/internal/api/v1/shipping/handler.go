package shipping

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/shipping"
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	shippingRepo *repository.ShippingRepository
}

func NewHandler(shippingRepo *repository.ShippingRepository) *Handler {
	return &Handler{
		shippingRepo: shippingRepo,
	}
}

// Shipping Template 相关接口

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
			"shipping_fee": 0.0,
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

// Carrier 相关接口

// ListCarriers 获取物流公司列表
// @Summary 获取物流公司列表
// @Tags Shipping
// @Produce json
// @Param enabled query bool false "只显示启用的"
// @Success 200 {array} shipping.Carrier
// @Router /api/v1/shipping/carriers [get]
func (h *Handler) ListCarriers(c *gin.Context) {
	enabledOnly := c.Query("enabled") == "true"

	carriers, err := h.shippingRepo.FindAllCarriers(enabledOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": carriers})
}

// GetCarrier 获取物流公司详情
// @Summary 获取物流公司详情
// @Tags Shipping
// @Produce json
// @Param id path int true "物流公司ID"
// @Success 200 {object} shipping.Carrier
// @Router /api/v1/shipping/carriers/{id} [get]
func (h *Handler) GetCarrier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid carrier id"})
		return
	}

	carrier, err := h.shippingRepo.FindCarrierByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, carrier)
}

// CreateCarrier 创建物流公司（管理员）
// @Summary 创建物流公司
// @Tags Shipping
// @Accept json
// @Produce json
// @Param carrier body shipping.Carrier true "物流公司信息"
// @Success 201 {object} shipping.Carrier
// @Router /api/v1/admin/shipping/carriers [post]
func (h *Handler) CreateCarrier(c *gin.Context) {
	var carrier shipping.Carrier
	if err := c.ShouldBindJSON(&carrier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.shippingRepo.CreateCarrier(&carrier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, carrier)
}

// UpdateCarrier 更新物流公司（管理员）
// @Summary 更新物流公司
// @Tags Shipping
// @Accept json
// @Produce json
// @Param id path int true "物流公司ID"
// @Param carrier body shipping.Carrier true "物流公司信息"
// @Success 200 {object} shipping.Carrier
// @Router /api/v1/admin/shipping/carriers/{id} [put]
func (h *Handler) UpdateCarrier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid carrier id"})
		return
	}

	var carrier shipping.Carrier
	if err := c.ShouldBindJSON(&carrier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	carrier.ID = uint(id)
	if err := h.shippingRepo.UpdateCarrier(&carrier); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, carrier)
}

// DeleteCarrier 删除物流公司（管理员）
// @Summary 删除物流公司
// @Tags Shipping
// @Produce json
// @Param id path int true "物流公司ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/shipping/carriers/{id} [delete]
func (h *Handler) DeleteCarrier(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid carrier id"})
		return
	}

	if err := h.shippingRepo.DeleteCarrier(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "carrier deleted"})
}

// Tracking 相关接口

// TrackShipment 追踪物流
// @Summary 追踪物流
// @Tags Shipping
// @Produce json
// @Param tracking_number path string true "追踪号"
// @Success 200 {array} shipping.TrackingEvent
// @Router /api/v1/shipping/track/{tracking_number} [get]
func (h *Handler) TrackShipment(c *gin.Context) {
	trackingNumber := c.Param("tracking_number")
	if trackingNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tracking number is required"})
		return
	}

	events, err := h.shippingRepo.FindTrackingEventsByTrackingNumber(trackingNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": events})
}

// GetOrderTracking 获取订单物流追踪
// @Summary 获取订单物流追踪
// @Tags Shipping
// @Produce json
// @Param order_id path int true "订单ID"
// @Success 200 {array} shipping.TrackingEvent
// @Router /api/v1/shipping/orders/{order_id}/tracking [get]
func (h *Handler) GetOrderTracking(c *gin.Context) {
	orderID, err := strconv.ParseUint(c.Param("order_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	events, err := h.shippingRepo.FindTrackingEventsByOrderID(uint(orderID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": events})
}

// CreateTrackingEvent 创建物流追踪事件（管理员）
// @Summary 创建物流追踪事件
// @Tags Shipping
// @Accept json
// @Produce json
// @Param event body shipping.TrackingEvent true "追踪事件"
// @Success 201 {object} shipping.TrackingEvent
// @Router /api/v1/admin/shipping/tracking [post]
func (h *Handler) CreateTrackingEvent(c *gin.Context) {
	var event shipping.TrackingEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.shippingRepo.CreateTrackingEvent(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// Shipping Zone 相关接口

// ListZones 获取配送区域列表
// @Summary 获取配送区域列表
// @Tags Shipping
// @Produce json
// @Success 200 {array} shipping.ShippingZone
// @Router /api/v1/shipping/zones [get]
func (h *Handler) ListZones(c *gin.Context) {
	zones, err := h.shippingRepo.FindAllZones()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": zones})
}

// GetZone 获取配送区域详情
// @Summary 获取配送区域详情
// @Tags Shipping
// @Produce json
// @Param id path int true "区域ID"
// @Success 200 {object} shipping.ShippingZone
// @Router /api/v1/shipping/zones/{id} [get]
func (h *Handler) GetZone(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}

	zone, err := h.shippingRepo.FindZoneByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, zone)
}

// CreateZone 创建配送区域（管理员）
// @Summary 创建配送区域
// @Tags Shipping
// @Accept json
// @Produce json
// @Param zone body shipping.ShippingZone true "区域信息"
// @Success 201 {object} shipping.ShippingZone
// @Router /api/v1/admin/shipping/zones [post]
func (h *Handler) CreateZone(c *gin.Context) {
	var zone shipping.ShippingZone
	if err := c.ShouldBindJSON(&zone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.shippingRepo.CreateZone(&zone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, zone)
}

// UpdateZone 更新配送区域（管理员）
// @Summary 更新配送区域
// @Tags Shipping
// @Accept json
// @Produce json
// @Param id path int true "区域ID"
// @Param zone body shipping.ShippingZone true "区域信息"
// @Success 200 {object} shipping.ShippingZone
// @Router /api/v1/admin/shipping/zones/{id} [put]
func (h *Handler) UpdateZone(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}

	var zone shipping.ShippingZone
	if err := c.ShouldBindJSON(&zone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	zone.ID = uint(id)
	if err := h.shippingRepo.UpdateZone(&zone); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, zone)
}

// DeleteZone 删除配送区域（管理员）
// @Summary 删除配送区域
// @Tags Shipping
// @Produce json
// @Param id path int true "区域ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/shipping/zones/{id} [delete]
func (h *Handler) DeleteZone(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zone id"})
		return
	}

	if err := h.shippingRepo.DeleteZone(uint(id)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "zone deleted"})
}

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
