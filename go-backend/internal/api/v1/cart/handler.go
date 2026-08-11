package cart

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/visitorcookie"
	"commerce-platform/internal/service"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	cartService           *service.CartService
	visitorProfileService *service.VisitorProfileService
	visitorSecret         []byte
}

type Options struct {
	VisitorProfileService *service.VisitorProfileService
	VisitorSecret         string
}

func NewHandler(cartService *service.CartService, opts ...Options) *Handler {
	options := Options{}
	if len(opts) > 0 {
		options = opts[0]
	}
	return &Handler{
		cartService:           cartService,
		visitorProfileService: options.VisitorProfileService,
		visitorSecret:         []byte(options.VisitorSecret),
	}
}

// getUserIDAndSession reads the user and ensures a cart session for meaningful write paths.
func (h *Handler) getUserIDAndSession(c *gin.Context) (*uint, string) {
	userID := currentUserID(c)

	sessionID := existingCartSessionID(c)
	if sessionID == "" {
		sessionID = uuid.New().String()
		c.SetCookie("session_id", sessionID, 86400*30, "/", "", false, true)
	}

	return userID, sessionID
}

func currentUserID(c *gin.Context) *uint {
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(uint); ok && id > 0 {
			userID = &id
		}
	}
	return userID
}

func existingCartSessionID(c *gin.Context) string {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(sessionID)
}

func (h *Handler) touchVisitorProfile(c *gin.Context, userID *uint, sessionID string) {
	if h.visitorProfileService == nil {
		return
	}
	input := h.visitorProfileTouchInput(c, userID, sessionID)
	input.MeaningfulAction = service.VisitorProfileActionCart
	input.QualityScoreDelta = service.VisitorProfileQualityCartAction
	if _, err := h.visitorProfileService.TouchMeaningfulAction(input); err != nil {
		return
	}
}

func (h *Handler) touchPassiveVisitorProfile(c *gin.Context, userID *uint, sessionID string) {
	if h.visitorProfileService == nil {
		return
	}
	if _, err := h.visitorProfileService.TouchPassiveSeen(h.visitorProfileTouchInput(c, userID, sessionID)); err != nil {
		return
	}
}

func (h *Handler) visitorProfileTouchInput(c *gin.Context, userID *uint, sessionID string) service.VisitorProfileTouchInput {
	visitorHash, _ := visitorcookie.ExistingCustomerServiceVisitorHash(c, h.visitorSecret)
	return service.VisitorProfileTouchInput{
		UserID:                     userID,
		CustomerServiceVisitorHash: visitorHash,
		CartSessionID:              sessionID,
		Locale:                     firstNonEmptyHeader(c, "X-Locale", "Accept-Language"),
		LocaleSource:               "accept_language",
		CountryCode:                middleware.TrustedEdgeCountry(c),
		Region:                     firstNonEmptyHeader(c, "CF-Region", "X-Region"),
		City:                       firstNonEmptyHeader(c, "CF-IPCity", "X-City"),
		Timezone:                   firstNonEmptyHeader(c, "CF-Timezone", "X-Timezone"),
		IPAddress:                  c.ClientIP(),
		UserAgent:                  c.GetHeader("User-Agent"),
	}
}

func firstNonEmptyHeader(c *gin.Context, keys ...string) string {
	for _, key := range keys {
		if value := c.GetHeader(key); value != "" {
			return value
		}
	}
	return ""
}

// AddToCartRequest 添加到购物车请求
func parseOptionalUintQuery(c *gin.Context, key string) *uint {
	raw := c.Query(key)
	if raw == "" {
		return nil
	}
	parsed, err := strconv.ParseUint(raw, 10, 32)
	if err != nil {
		return nil
	}
	value := uint(parsed)
	return &value
}

type AddToCartRequest struct {
	ProductID uint  `json:"product_id" binding:"required"`
	VariantID *uint `json:"variant_id"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

// UpdateCartItemRequest 更新购物车项目请求
type UpdateCartItemRequest struct {
	VariantID *uint `json:"variant_id"`
	Quantity  int   `json:"quantity" binding:"required,min=1"`
}

// GetCartSummary 获取购物车摘要
func (h *Handler) GetCartSummary(c *gin.Context) {
	userID := currentUserID(c)
	sessionID := existingCartSessionID(c)

	summary, err := h.cartService.GetCartSummary(userID, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	if userID != nil || len(summary.Items) > 0 {
		h.touchPassiveVisitorProfile(c, userID, sessionID)
	}

	response.Success(c, PublicCartSummaryFromDomain(summary))
}

// AddToCart 添加商品到购物车
func (h *Handler) AddToCart(c *gin.Context) {
	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	if err := h.cartService.ValidateAddToCart(req.ProductID, req.VariantID, req.Quantity); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	userID, sessionID := h.getUserIDAndSession(c)

	cart, err := h.cartService.GetOrCreateCart(userID, sessionID)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	if err := h.cartService.AddToCart(cart.ID, req.ProductID, req.VariantID, req.Quantity); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	h.touchVisitorProfile(c, userID, sessionID)
	response.SuccessWithMessage(c, "Product added to cart", nil)
}

// UpdateCartItem 更新购物车项目
func (h *Handler) UpdateCartItem(c *gin.Context) {
	var req UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	userID := currentUserID(c)
	sessionID := existingCartSessionID(c)

	cart, err := h.cartService.FindCart(userID, sessionID)
	if err != nil {
		if errors.Is(err, service.ErrCartNotFound) {
			apierror.RespondBadRequest(c, "Cart is empty")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}

	productID := c.Param("id")
	parsedProductID, err := strconv.ParseUint(productID, 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid product id")
		return
	}
	pID := uint(parsedProductID)

	if err := h.cartService.UpdateCartItem(cart.ID, pID, req.VariantID, req.Quantity); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	h.touchVisitorProfile(c, userID, sessionID)
	response.SuccessWithMessage(c, "Cart item updated", nil)
}

// RemoveFromCart 从购物车移除商品
func (h *Handler) RemoveFromCart(c *gin.Context) {
	userID := currentUserID(c)
	sessionID := existingCartSessionID(c)

	cart, err := h.cartService.FindCart(userID, sessionID)
	if err != nil {
		if errors.Is(err, service.ErrCartNotFound) {
			response.SuccessWithMessage(c, "Product removed from cart", nil)
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}

	productID := c.Param("id")
	parsedProductID, err := strconv.ParseUint(productID, 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid product id")
		return
	}
	pID := uint(parsedProductID)

	variantID := parseOptionalUintQuery(c, "variant_id")
	if err := h.cartService.RemoveFromCart(cart.ID, pID, variantID); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if len(cart.Items) > 0 {
		h.touchVisitorProfile(c, userID, sessionID)
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
	if !h.cartService.HasPurchasableSyncItems(items) {
		userID := currentUserID(c)
		sessionID := existingCartSessionID(c)
		summary, err := h.cartService.GetCartSummary(userID, sessionID)
		if err != nil {
			apierror.RespondInternalError(c, err)
			return
		}
		response.Success(c, PublicCartSummaryFromDomain(summary))
		return
	}

	userID, sessionID := h.getUserIDAndSession(c)

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
	if len(summary.Items) > 0 {
		h.touchVisitorProfile(c, userID, sessionID)
	}

	response.Success(c, PublicCartSummaryFromDomain(summary))
}

// ClearCart 清空购物车
func (h *Handler) ClearCart(c *gin.Context) {
	userID := currentUserID(c)
	sessionID := existingCartSessionID(c)

	cart, err := h.cartService.FindCart(userID, sessionID)
	if err != nil {
		if errors.Is(err, service.ErrCartNotFound) {
			response.SuccessWithMessage(c, "Cart cleared", nil)
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}

	if err := h.cartService.ClearCart(cart.ID); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if len(cart.Items) > 0 {
		h.touchVisitorProfile(c, userID, sessionID)
	}
	response.SuccessWithMessage(c, "Cart cleared", nil)
}
