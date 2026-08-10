package admin

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type GoogleMerchantHandler struct {
	service        *service.GoogleMerchantService
	postConnectURL string
}

func NewGoogleMerchantHandler(service *service.GoogleMerchantService, postConnectURL string) *GoogleMerchantHandler {
	return &GoogleMerchantHandler{service: service, postConnectURL: postConnectURL}
}

func (h *GoogleMerchantHandler) GetConnection(c *gin.Context) {
	connection, err := h.service.GetConnection()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"connection": connection})
}

func (h *GoogleMerchantHandler) UpdateConnection(c *gin.Context) {
	var input service.GoogleMerchantConnectionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	connection, err := h.service.UpdateConnection(input)
	if err != nil {
		respondGoogleMerchantError(c, err)
		return
	}
	response.Success(c, gin.H{"connection": connection})
}

func (h *GoogleMerchantHandler) StartOAuth(c *gin.Context) {
	userID, ok := currentGoogleMerchantUserID(c)
	if !ok {
		return
	}
	authorizationURL, err := h.service.StartOAuth(userID)
	if err != nil {
		respondGoogleMerchantError(c, err)
		return
	}
	response.Success(c, gin.H{"authorization_url": authorizationURL})
}

func (h *GoogleMerchantHandler) CompleteOAuth(c *gin.Context) {
	err := h.service.CompleteOAuth(c.Request.Context(), c.Query("state"), c.Query("code"), c.Query("error"))
	c.Redirect(http.StatusFound, h.oauthReturnURL(err))
}

func (h *GoogleMerchantHandler) Disconnect(c *gin.Context) {
	if err := h.service.Disconnect(); err != nil {
		respondGoogleMerchantError(c, err)
		return
	}
	response.SuccessWithMessage(c, "Google Merchant disconnected", nil)
}

func (h *GoogleMerchantHandler) ListRemoteProducts(c *gin.Context) {
	pageSize := 50
	if raw := c.Query("page_size"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			apierror.RespondBadRequest(c, "invalid remote product page size")
			return
		}
		pageSize = parsed
	}
	page, err := h.service.ListRemoteProducts(c.Request.Context(), pageSize, c.Query("page_token"))
	if err != nil {
		respondGoogleMerchantError(c, err)
		return
	}
	response.Success(c, page)
}

func (h *GoogleMerchantHandler) ListOffers(c *gin.Context) {
	offers, err := h.service.ListOffers()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"offers": offers})
}

func (h *GoogleMerchantHandler) Reconcile(c *gin.Context) {
	result, err := h.service.ReconcileOffers(c.Request.Context())
	if err != nil {
		if len(result.Errors) > 0 {
			apierror.RespondError(c, http.StatusBadGateway, "google_merchant_reconciliation_failed", err.Error())
			return
		}
		respondGoogleMerchantError(c, err)
		return
	}
	response.Success(c, gin.H{"result": result})
}

func (h *GoogleMerchantHandler) CreateOffer(c *gin.Context) {
	var input service.GoogleMerchantOfferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	offer, err := h.service.CreateOffer(input)
	if err != nil {
		respondGoogleMerchantError(c, err)
		return
	}
	response.Created(c, gin.H{"offer": offer})
}

func (h *GoogleMerchantHandler) UpdateOffer(c *gin.Context) {
	id, ok := googleMerchantOfferPathID(c)
	if !ok {
		return
	}
	var input service.GoogleMerchantOfferInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	offer, err := h.service.UpdateOffer(id, input)
	if err != nil {
		respondGoogleMerchantError(c, err)
		return
	}
	response.Success(c, gin.H{"offer": offer})
}

func (h *GoogleMerchantHandler) ValidateOffer(c *gin.Context) {
	id, ok := googleMerchantOfferPathID(c)
	if !ok {
		return
	}
	offer, err := h.service.ValidateOffer(id)
	if err != nil {
		respondGoogleMerchantError(c, err)
		return
	}
	response.Success(c, gin.H{"offer": offer})
}

func (h *GoogleMerchantHandler) SyncOffer(c *gin.Context) {
	id, ok := googleMerchantOfferPathID(c)
	if !ok {
		return
	}
	offer, err := h.service.SyncOffer(c.Request.Context(), id)
	if err != nil {
		respondGoogleMerchantError(c, err)
		return
	}
	response.Success(c, gin.H{"offer": offer})
}

func (h *GoogleMerchantHandler) RemoveRemoteOffer(c *gin.Context) {
	id, ok := googleMerchantOfferPathID(c)
	if !ok {
		return
	}
	offer, err := h.service.RemoveRemoteOffer(c.Request.Context(), id)
	if err != nil {
		respondGoogleMerchantError(c, err)
		return
	}
	response.Success(c, gin.H{"offer": offer})
}

func (h *GoogleMerchantHandler) DeleteOffer(c *gin.Context) {
	id, ok := googleMerchantOfferPathID(c)
	if !ok {
		return
	}
	if err := h.service.DeleteOffer(id); err != nil {
		respondGoogleMerchantError(c, err)
		return
	}
	response.SuccessWithMessage(c, "Google Merchant offer removed", nil)
}

func googleMerchantOfferPathID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid Google Merchant offer id")
		return 0, false
	}
	return uint(id), true
}

func respondGoogleMerchantError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrGoogleMerchantOfferNotFound):
		apierror.RespondNotFound(c, "Google Merchant offer")
	case errors.Is(err, service.ErrGoogleMerchantOfferInvalid):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrGoogleMerchantConnectionInvalid):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrGoogleMerchantOAuthNotConfigured):
		apierror.RespondBadRequest(c, "Google Merchant OAuth is not configured")
	case errors.Is(err, service.ErrGoogleMerchantOAuthStateInvalid):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrGoogleMerchantOAuthExchange):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrGoogleMerchantConnectionNotFound):
		apierror.RespondBadRequest(c, "Google Merchant is not connected")
	case errors.Is(err, service.ErrGoogleMerchantRemoteAPI):
		apierror.RespondError(c, http.StatusBadGateway, "upstream_error", err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}

func currentGoogleMerchantUserID(c *gin.Context) (uint, bool) {
	rawUserID, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return 0, false
	}
	userID, ok := rawUserID.(uint)
	if !ok || userID == 0 {
		apierror.RespondUnauthorized(c)
		return 0, false
	}
	return userID, true
}

func (h *GoogleMerchantHandler) oauthReturnURL(callbackErr error) string {
	target := h.postConnectURL
	if target == "" {
		target = "/google-merchant"
	}
	parsed, err := url.Parse(target)
	if err != nil {
		parsed = &url.URL{Path: "/google-merchant"}
	}

	query := parsed.Query()
	if callbackErr != nil {
		query.Set("google_merchant_status", "error")
		query.Set("google_merchant_message", callbackErr.Error())
	} else {
		query.Set("google_merchant_status", "connected")
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}
