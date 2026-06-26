package cart

import (
	"fmt"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
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

// getUserIDAndSession 从context获取用户ID和session ID
// 统一的辅助方法，减少重复代码
func getUserIDAndSession(c *gin.Context) (*uint, string) {
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

	return userID, sessionID
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
	userID, sessionID := getUserIDAndSession(c)

	summary, err := h.cartService.GetCartSummary(userID, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, summary)
}

// AddToCart 添加商品到购物车
func (h *Handler) AddToCart(c *gin.Context) {
	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	userID, sessionID := getUserIDAndSession(c)

	cart, err := h.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	if err := h.cartService.AddToCart(cart.ID, req.ProductID, req.Quantity); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Product added to cart", nil)
}

// UpdateCartItem 更新购物车项目
func (h *Handler) UpdateCartItem(c *gin.Context) {
	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	userID, sessionID := getUserIDAndSession(c)

	cart, err := h.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	productID := c.Param("id")
	var pID uint
	fmt.Sscanf(productID, "%d", &pID)

	if err := h.cartService.UpdateCartItem(cart.ID, pID, req.Quantity); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Cart item updated", nil)
}

// RemoveFromCart 从购物车移除商品
func (h *Handler) RemoveFromCart(c *gin.Context) {
	userID, sessionID := getUserIDAndSession(c)

	cart, err := h.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	productID := c.Param("id")
	var pID uint
	fmt.Sscanf(productID, "%d", &pID)

	if err := h.cartService.RemoveFromCart(cart.ID, pID); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Product removed from cart", nil)
}

// SyncCart 同步本地购物车到云端
func (h *Handler) SyncCart(c *gin.Context) {
	var items []service.SyncCartItemReq
	if err := c.ShouldBindJSON(&items); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	userID, sessionID := getUserIDAndSession(c)

	cart, err := h.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	if err := h.cartService.SyncCart(cart.ID, items); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	// 重新获取同步后的摘要
	summary, err := h.cartService.GetCartSummary(userID, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, summary)
}

// ClearCart 清空购物车
func (h *Handler) ClearCart(c *gin.Context) {
	userID, sessionID := getUserIDAndSession(c)

	cart, err := h.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	if err := h.cartService.ClearCart(cart.ID); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "Cart cleared", nil)
}
