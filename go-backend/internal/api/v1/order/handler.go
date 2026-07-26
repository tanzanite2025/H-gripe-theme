package order

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"tanzanite/internal/domain/order"
	"tanzanite/internal/pkg/antifraud"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/metrics"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	orderService *service.OrderService
	cartService  *service.CartService
	riskService  *antifraud.Service
}

func NewHandler(orderService *service.OrderService, cartService *service.CartService, riskServices ...*antifraud.Service) *Handler {
	var riskService *antifraud.Service
	if len(riskServices) > 0 {
		riskService = riskServices[0]
	}
	return &Handler{
		orderService: orderService,
		cartService:  cartService,
		riskService:  riskService,
	}
}

// CreateOrder 创建订单
// @Summary 创建订单
// @Tags Orders
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "订单信息"
// @Success 201 {object} order.Order
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	riskKey := fmt.Sprintf("user:%d", userID.(uint))
	if sessionID, err := c.Cookie("session_id"); err == nil && strings.TrimSpace(sessionID) != "" {
		riskKey += ":session:" + strings.TrimSpace(sessionID)
	}
	if isCardLikePaymentMethod(req.PaymentMethod) && h.riskService != nil {
		signals := antifraud.Signals{
			IPCountry:      c.GetHeader("CF-IPCountry"),
			BillingCountry: req.ShippingAddress.Country,
			UserAgent:      c.Request.UserAgent(),
		}
		if req.ClientRisk != nil {
			if req.ClientRisk.IPCountry != "" {
				signals.IPCountry = req.ClientRisk.IPCountry
			}
			if req.ClientRisk.BillingCountry != "" {
				signals.BillingCountry = req.ClientRisk.BillingCountry
			}
			signals.VPNDetected = req.ClientRisk.VPNDetected
			signals.Timezone = req.ClientRisk.Timezone
		}
		decision, err := h.riskService.Evaluate(c.Request.Context(), riskKey, signals)
		if err != nil {
			apierror.RespondError(c, 503, "payment_risk_unavailable", "Payment risk service is temporarily unavailable")
			return
		}
		if decision.Delay > 0 {
			metrics.PaymentRiskDelayed.Inc()
			time.Sleep(decision.Delay)
		}
	}

	// 安全闭环：强制从后端 Cart 拉取，忽略前端的 req.Items
	sessionID, _ := c.Cookie("session_id")
	uid := userID.(uint)

	cart, err := h.cartService.GetOrCreateCart(&uid, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	summary, err := h.cartService.GetCartSummary(&uid, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	if len(summary.Items) == 0 {
		apierror.RespondBadRequest(c, "Cart is empty")
		return
	}

	items := make([]order.OrderItem, len(summary.Items))
	for i, item := range summary.Items {
		items[i] = order.OrderItem{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Quantity:  item.Quantity,
		}
	}

	shippingAddr := addressFromRequest(req.ShippingAddress)
	billingAddr := billingAddressFromRequest(shippingAddr, req.BillingAddress)

	// 创建订单
	o, err := h.orderService.CreateOrder(
		c.Request.Context(),
		userID.(uint),
		items,
		shippingAddr,
		billingAddr,
		req.PaymentMethod,
		req.ShippingMethod,
		req.CouponCode,
		req.PointsToUse,
	)
	if err != nil {
		if isCardLikePaymentMethod(req.PaymentMethod) && h.riskService != nil {
			_ = h.riskService.RecordFailure(c.Request.Context(), riskKey)
		}
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if isCardLikePaymentMethod(req.PaymentMethod) && h.riskService != nil {
		_ = h.riskService.RecordSuccess(c.Request.Context(), riskKey)
	}

	// 清空购物车
	_ = h.cartService.ClearCart(cart.ID)

	response.Created(c, o)
}

func isCardLikePaymentMethod(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "card", "credit_card", "debit_card", "stripe", "alipay", "paypal":
		return true
	default:
		return false
	}
}

// GetOrder 获取订单详情
// @Summary 获取订单详情
// @Tags Orders
// @Produce json
// @Param id path int true "订单ID"
// @Success 200 {object} order.Order
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid order ID")
		return
	}

	o, err := h.orderService.GetOrder(uint(id), userID.(uint))
	if err != nil {
		apierror.RespondNotFound(c, "Order")
		return
	}

	response.Success(c, o)
}

// ListOrders 获取订单列表
// @Summary 获取订单列表
// @Tags Orders
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/orders [get]
func (h *Handler) ListOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	params := pagination.ParsePagination(c)

	orders, total, err := h.orderService.GetUserOrders(userID.(uint), params.Page, params.PageSize)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Paged(c, orders, params.Page, params.PageSize, total)
}

// CancelOrder 取消订单
// @Summary 取消订单
// @Tags Orders
// @Produce json
// @Param id path int true "订单ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/orders/{id}/cancel [post]
func (h *Handler) CancelOrder(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "Invalid order ID")
		return
	}

	if err := h.orderService.CancelOrder(uint(id), userID.(uint)); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Order cancelled", nil)
}

// GetOrderStats 获取订单统计
// @Summary 获取订单统计
// @Tags Orders
// @Produce json
// @Success 200 {object} map[string]int64
// @Router /api/v1/orders/stats [get]
func (h *Handler) GetOrderStats(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}

	stats, err := h.orderService.GetOrderStats(userID.(uint))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, stats)
}
