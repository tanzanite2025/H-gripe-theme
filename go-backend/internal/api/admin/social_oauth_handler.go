package admin

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type SocialOAuthHandler struct {
	service        *service.SocialOAuthService
	postConnectURL string
}

type socialOAuthStartRequest struct {
	ReturnPath string `json:"return_path"`
}

func NewSocialOAuthHandler(socialOAuthService *service.SocialOAuthService, postConnectURL string) *SocialOAuthHandler {
	return &SocialOAuthHandler{
		service:        socialOAuthService,
		postConnectURL: postConnectURL,
	}
}

func (h *SocialOAuthHandler) ListConnections(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("social OAuth service is not configured"))
		return
	}
	view, err := h.service.ListConnections()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	c.JSON(http.StatusOK, view)
}

func (h *SocialOAuthHandler) Start(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("social OAuth service is not configured"))
		return
	}
	userID, ok := currentSocialOAuthUserID(c)
	if !ok {
		return
	}
	var request socialOAuthStartRequest
	if err := c.ShouldBindJSON(&request); err != nil && !errors.Is(err, io.EOF) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	result, err := h.service.Start(userID, service.SocialOAuthStartInput{
		Provider:   c.Param("provider"),
		ReturnPath: request.ReturnPath,
	})
	if err != nil {
		respondSocialOAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *SocialOAuthHandler) Disconnect(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("social OAuth service is not configured"))
		return
	}
	if err := h.service.Disconnect(c.Param("provider")); err != nil {
		respondSocialOAuthError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "social account disconnected"})
}

func (h *SocialOAuthHandler) CompleteOAuth(c *gin.Context) {
	if h == nil || h.service == nil {
		h.redirectOAuthResult(c, nil, errors.New("social OAuth service is not configured"))
		return
	}
	result, err := h.service.Complete(
		c.Request.Context(),
		c.Query("state"),
		c.Query("code"),
		c.Query("error"),
		c.Query("error_description"),
	)
	h.redirectOAuthResult(c, result, err)
}

func (h *SocialOAuthHandler) redirectOAuthResult(
	c *gin.Context,
	result *service.SocialOAuthCallbackResult,
	callbackErr error,
) {
	returnPath := "/social"
	if result != nil && result.ReturnPath != "" {
		returnPath = result.ReturnPath
	}
	returnURL, err := url.Parse(returnPath)
	if err != nil || returnURL.IsAbs() || returnURL.Host != "" || !strings.HasPrefix(returnURL.Path, "/") {
		returnURL = &url.URL{Path: "/social"}
	}

	redirectTarget := strings.TrimSpace(h.postConnectURL)
	parsed, err := url.Parse(redirectTarget)
	if redirectTarget == "" || err != nil || !parsed.IsAbs() || parsed.Host == "" {
		parsed = returnURL
	} else {
		parsed.Path = returnURL.Path
		parsed.RawPath = returnURL.RawPath
		parsed.RawQuery = returnURL.RawQuery
	}

	query := parsed.Query()
	if callbackErr != nil {
		query.Set("social_oauth_status", "error")
		query.Set("social_oauth_message", callbackErr.Error())
	} else if result != nil {
		query.Set("social_oauth_status", result.Status)
		query.Set("social_oauth_provider", result.Provider)
		query.Set("social_oauth_message", result.Message)
	}
	parsed.RawQuery = query.Encode()
	c.Redirect(http.StatusFound, parsed.String())
}

func currentSocialOAuthUserID(c *gin.Context) (uint, bool) {
	userID, ok := c.Get("user_id")
	if !ok {
		apierror.RespondUnauthorized(c)
		return 0, false
	}
	typedUserID, ok := userID.(uint)
	if !ok || typedUserID == 0 {
		apierror.RespondUnauthorized(c)
		return 0, false
	}
	return typedUserID, true
}

func respondSocialOAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSocialOAuthProviderInvalid),
		errors.Is(err, service.ErrSocialOAuthNotConfigured),
		errors.Is(err, service.ErrSocialOAuthStateInvalid):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrSocialOAuthExchange),
		errors.Is(err, service.ErrSocialOAuthRemoteAPI):
		apierror.RespondError(c, http.StatusBadGateway, "social_oauth_failed", err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
