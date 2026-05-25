package cart

import (
	"net/http"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	cartService *service.CartService
}

func NewHandler(cartService *service.CartService) *Handler {
	return &Handler{
		cartService: cartService,
	}
}

// AddToCartRequest 添加到购物车请求
type AddToCartRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

// UpdateCartItemRequest 更新购物车项目请求
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

// GetCartSummary 获取购物车摘要
func (h *Handler) GetCartSummary(c *gin.Context) {
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uint)
		userID = &id
	}

	// 获取或创建会话ID
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		sessionID = uuid.New().String()
		c.SetCookie("session_id", sessionID, 86400*30, "/", "", false, true)
	}

	summary, err := h.cartService.GetCartSummary(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

// AddToCart 添加商品到购物车
func (h *Handler) AddToCart(c *gin.Context) {
	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		id := uid.(uint)
		userID = &id
	}

	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		sessionID = uuid.New().String()
		c.SetCookie("session_id", sessionID, 86400*30, "/", "", false, true)
	}

	cart, err := h.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.cartService.AddToCart(cart.ID, req.ProductID, req.Quantity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "product added to cart"})
}

// UpdateCartItem 更新购物车项目
func (h *Handler) UpdateCartItem(c *gin.Context) {
	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 实现更新逻辑
	c.JSON(http.StatusOK, gin.H{"message": "cart item updated"})
}

// RemoveFromCart 从购物车移除商品
func (h *Handler) RemoveFromCart(c *gin.Context) {
	// 实现移除逻辑
	c.JSON(http.StatusOK, gin.H{"message": "product removed from cart"})
}
