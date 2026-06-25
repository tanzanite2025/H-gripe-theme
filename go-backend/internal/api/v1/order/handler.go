package order

import (
	"net/http"
	"strconv"
	"tanzanite/internal/domain/order"
	"tanzanite/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	orderService *service.OrderService
	cartService  *service.CartService
}

func NewHandler(orderService *service.OrderService, cartService *service.CartService) *Handler {
	return &Handler{
		orderService: orderService,
		cartService:  cartService,
	}
}

// CreateOrderRequest 创建订单请求
type CreateOrderRequest struct {
	Items           []OrderItemRequest `json:"items" binding:"required,min=1"`
	ShippingAddress AddressRequest     `json:"shipping_address" binding:"required"`
	BillingAddress  AddressRequest     `json:"billing_address"`
	PaymentMethod   string             `json:"payment_method" binding:"required"`
	ShippingMethod  string             `json:"shipping_method" binding:"required"`
	CouponCode      string             `json:"coupon_code"`
	PointsToUse     int                `json:"points_to_use"`
}

type OrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type AddressRequest struct {
	FirstName  string `json:"first_name" binding:"required"`
	LastName   string `json:"last_name" binding:"required"`
	Company    string `json:"company"`
	Address1   string `json:"address1" binding:"required"`
	Address2   string `json:"address2"`
	City       string `json:"city" binding:"required"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code" binding:"required"`
	Country    string `json:"country" binding:"required"`
	Phone      string `json:"phone" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 安全闭环：强制从后端 Cart 拉取，忽略前端的 req.Items
	sessionID, _ := c.Cookie("session_id")
	uid := userID.(uint)
	
	cart, err := h.cartService.GetOrCreateCart(&uid, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve cart session"})
		return
	}

	summary, err := h.cartService.GetCartSummary(&uid, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve cart items"})
		return
	}

	if len(summary.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
		return
	}

	items := make([]order.OrderItem, len(summary.Items))
	for i, item := range summary.Items {
		items[i] = order.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	// 转换地址
	shippingAddr := order.Address{
		FirstName:  req.ShippingAddress.FirstName,
		LastName:   req.ShippingAddress.LastName,
		Company:    req.ShippingAddress.Company,
		Address1:   req.ShippingAddress.Address1,
		Address2:   req.ShippingAddress.Address2,
		City:       req.ShippingAddress.City,
		State:      req.ShippingAddress.State,
		PostalCode: req.ShippingAddress.PostalCode,
		Country:    req.ShippingAddress.Country,
		Phone:      req.ShippingAddress.Phone,
		Email:      req.ShippingAddress.Email,
	}

	// 如果没有提供账单地址，使用配送地址
	billingAddr := shippingAddr
	if req.BillingAddress.FirstName != "" {
		billingAddr = order.Address{
			FirstName:  req.BillingAddress.FirstName,
			LastName:   req.BillingAddress.LastName,
			Company:    req.BillingAddress.Company,
			Address1:   req.BillingAddress.Address1,
			Address2:   req.BillingAddress.Address2,
			City:       req.BillingAddress.City,
			State:      req.BillingAddress.State,
			PostalCode: req.BillingAddress.PostalCode,
			Country:    req.BillingAddress.Country,
			Phone:      req.BillingAddress.Phone,
			Email:      req.BillingAddress.Email,
		}
	}

	// 创建订单
	o, err := h.orderService.CreateOrder(
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 清空购物车
	_ = h.cartService.ClearCart(cart.ID)

	c.JSON(http.StatusCreated, o)
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	o, err := h.orderService.GetOrder(uint(id), userID.(uint))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, o)
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := h.orderService.GetUserOrders(userID.(uint), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": orders,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

func (h *Handler) ListPublicChatOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", c.DefaultQuery("page_size", "10")))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	orders, _, err := h.orderService.GetUserOrders(userID.(uint), 1, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	items := make([]gin.H, 0, len(orders))
	for _, item := range orders {
		items = append(items, makePublicChatOrder(item))
	}

	c.JSON(http.StatusOK, items)
}

// ListAllOrders 获取所有订单（管理员）
// @Summary 获取所有订单
// @Tags Orders
// @Produce json
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "订单状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/orders [get]
func (h *Handler) ListAllOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := h.orderService.GetAllOrders(page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": orders,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_page": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateOrderStatus 更新订单状态
// @Summary 更新订单状态
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "订单ID"
// @Param status body map[string]string true "状态"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/admin/orders/{id}/status [put]
func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	if role, exists := c.Get("role"); !exists || role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderService.UpdateOrderStatus(uint(id), req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	if err := h.orderService.CancelOrder(uint(id), userID.(uint)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order cancelled"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	stats, err := h.orderService.GetOrderStats(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func makePublicChatOrder(item order.Order) gin.H {
	title := "Order #" + item.OrderNumber
	if item.OrderNumber == "" {
		title = "Order #" + strconv.FormatUint(uint64(item.ID), 10)
	}

	return gin.H{
		"id":           item.ID,
		"order_number": item.OrderNumber,
		"title":        title,
		"status":       item.Status,
		"total":        item.TotalAmount,
		"currency":     "USD",
		"date":         item.CreatedAt.Format("2006-01-02"),
		"created_at":   item.CreatedAt.Format(time.RFC3339),
		"url":          "/orders/" + strconv.FormatUint(uint64(item.ID), 10),
		"thumbnail":    "",
	}
}
