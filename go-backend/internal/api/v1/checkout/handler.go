package checkout

import (
	"errors"
	"net/http"
	"strings"

	"commerce-platform/internal/domain/order"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

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
	ShippingAddress     AddressRequest `json:"shipping_address"`
	DisplayCurrency     string         `json:"display_currency"`
	PaymentMethod       string         `json:"payment_method"`
	ShippingQuoteID     string         `json:"shipping_quote_id"`
	SelectedQuotePlanID string         `json:"selected_quote_plan_id"`
	CouponCode          string         `json:"coupon_code"`
	GiftCardCode        string         `json:"gift_card_code"`
	PointsToUse         int            `json:"points_to_use"`
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
		priceMoney, priceErr := item.PriceMoney()
		if priceErr != nil {
			apierror.RespondBadRequest(c, priceErr.Error())
			return
		}
		price, priceErr := priceMoney.MajorFloat()
		if priceErr != nil {
			apierror.RespondBadRequest(c, priceErr.Error())
			return
		}
		items[i] = order.OrderItem{
			ProductID:         item.ProductID,
			VariantID:         item.VariantID,
			Quantity:          item.Quantity,
			Currency:          priceMoney.Currency().String(),
			Price:             price,
			ConfigurationData: append([]byte(nil), item.ConfigurationData...),
			ConfigurationHash: item.ConfigurationHash,
		}
	}

	quote, err := h.checkoutService.Quote(service.CheckoutQuoteInput{
		UserID:              userID,
		Items:               items,
		ShippingAddress:     toOrderAddress(req.ShippingAddress),
		DisplayCurrency:     req.DisplayCurrency,
		PaymentMethod:       req.PaymentMethod,
		ShippingQuoteID:     req.ShippingQuoteID,
		SelectedQuotePlanID: req.SelectedQuotePlanID,
		CouponCode:          req.CouponCode,
		GiftCardCode:        req.GiftCardCode,
		PointsToUse:         req.PointsToUse,
	})
	if err != nil {
		if errors.Is(err, service.ErrProductConfigurationPriceChanged) {
			apierror.RespondError(c, http.StatusConflict, "product_configuration_price_changed", err.Error())
			return
		}
		if errors.Is(err, service.ErrProductConfigurationConflict) || errors.Is(err, service.ErrProductConfigurationRequired) {
			apierror.RespondError(c, http.StatusConflict, "product_configuration_conflict", err.Error())
			return
		}
		if errors.Is(err, service.ErrCountryNotSupported) {
			apierror.RespondError(c, http.StatusUnprocessableEntity, "country_not_supported", err.Error())
			return
		}
		if errors.Is(err, service.ErrShippingQuoteExpired) || errors.Is(err, service.ErrShippingQuoteStale) {
			apierror.RespondError(c, http.StatusConflict, "shipping_quote_stale", err.Error())
			return
		}
		if errors.Is(err, service.ErrShippingQuotePlanUnavailable) {
			apierror.RespondError(c, http.StatusUnprocessableEntity, "shipping_quote_plan_unavailable", err.Error())
			return
		}
		if errors.Is(err, service.ErrShippingRateConfigurationInvalid) {
			apierror.RespondInternalError(c, err)
			return
		}
		if errors.Is(err, service.ErrShippingRateUnavailable) {
			apierror.RespondError(c, http.StatusUnprocessableEntity, "shipping_rate_unavailable", err.Error())
			return
		}
		if strings.Contains(err.Error(), "exchange rate unavailable") {
			apierror.RespondError(c, http.StatusUnprocessableEntity, "exchange_rate_unavailable", err.Error())
			return
		}
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
