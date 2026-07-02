package checkout

import (
	"tanzanite/internal/domain/order"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	checkoutService *service.CheckoutService
	cartService     *service.CartService
}

func NewHandler(checkoutService *service.CheckoutService, cartService *service.CartService) *Handler {
	return &Handler{
		checkoutService: checkoutService,
		cartService:     cartService,
	}
}

type QuoteRequest struct {
	ShippingAddress AddressRequest `json:"shipping_address"`
	CouponCode      string         `json:"coupon_code"`
	PointsToUse     int            `json:"points_to_use"`
}

type AddressRequest struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	Company    string `json:"company"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
}

func (h *Handler) Quote(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}
	userID := userIDValue.(uint)

	var req QuoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	sessionID, _ := c.Cookie("session_id")
	summary, err := h.cartService.GetCartSummary(&userID, sessionID)
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
			Quantity:  item.Quantity,
		}
	}

	quote, err := h.checkoutService.Quote(service.CheckoutQuoteInput{
		UserID:          userID,
		Items:           items,
		ShippingAddress: toOrderAddress(req.ShippingAddress),
		CouponCode:      req.CouponCode,
		PointsToUse:     req.PointsToUse,
	})
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, quote)
}

func toOrderAddress(req AddressRequest) order.Address {
	return order.Address{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Company:    req.Company,
		Address1:   req.Address1,
		Address2:   req.Address2,
		City:       req.City,
		State:      req.State,
		PostalCode: req.PostalCode,
		Country:    req.Country,
		Phone:      req.Phone,
		Email:      req.Email,
	}
}
