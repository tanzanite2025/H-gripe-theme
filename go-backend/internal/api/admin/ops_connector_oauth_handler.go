package admin

import (
	"errors"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type opsConnectorOAuthStartRequest struct {
	Provider    string `json:"provider" binding:"required"`
	ConnectorID *uint  `json:"connector_id"`
	Environment string `json:"environment"`
	ReturnPath  string `json:"return_path"`
}

func (h *OpsConnectorHandler) StartOAuth(c *gin.Context) {
	if h == nil || h.oauthService == nil {
		apierror.RespondInternalError(c, errors.New("operations connector OAuth service is not configured"))
		return
	}
	userID, ok := currentOpsConnectorUserID(c)
	if !ok {
		return
	}
	var req opsConnectorOAuthStartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.oauthService.Start(c.Request.Context(), userID, service.OpsConnectorOAuthStartInput{
		Provider:    req.Provider,
		ConnectorID: req.ConnectorID,
		Environment: req.Environment,
		ReturnPath:  req.ReturnPath,
	})
	if err != nil {
		respondOpsConnectorOAuthError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *OpsConnectorHandler) CompleteOAuth(c *gin.Context) {
	returnPath := "/ops/connectors"
	if h != nil && h.oauthService != nil {
		result, err := h.oauthService.Complete(
			c.Request.Context(),
			c.Query("state"),
			c.Query("code"),
			c.Query("error"),
			c.Query("error_description"),
		)
		if result != nil && result.ReturnPath != "" {
			returnPath = result.ReturnPath
		}
		h.redirectOAuthResult(c, returnPath, result, err)
		return
	}
	h.redirectOAuthResult(c, returnPath, nil, errors.New("operations connector OAuth service is not configured"))
}

func (h *OpsConnectorHandler) redirectOAuthResult(
	c *gin.Context,
	returnPath string,
	result *service.OpsConnectorOAuthCallbackResult,
	callbackErr error,
) {
	redirectTarget := strings.TrimSpace(os.Getenv("OPS_CONNECTOR_OAUTH_POST_CONNECT_URL"))
	returnURL, err := url.Parse(returnPath)
	if err != nil || returnURL.IsAbs() || returnURL.Host != "" || !strings.HasPrefix(returnURL.Path, "/") {
		returnURL = &url.URL{Path: "/ops/connectors"}
	}
	parsed, err := url.Parse(redirectTarget)
	if redirectTarget == "" || err != nil || !parsed.IsAbs() || parsed.Host == "" {
		parsed = returnURL
	} else {
		// Keep the configured admin origin, while preserving the relative
		// callback path and query used by chained connections.
		parsed.Path = returnURL.Path
		parsed.RawPath = returnURL.RawPath
		parsed.RawQuery = returnURL.RawQuery
	}
	query := parsed.Query()
	if callbackErr != nil {
		query.Set("ops_oauth_status", "error")
		query.Set("ops_oauth_message", callbackErr.Error())
	} else if result != nil {
		query.Set("ops_oauth_status", result.Status)
		query.Set("ops_oauth_provider", result.Provider)
		query.Set("ops_oauth_connector_id", strconv.FormatUint(uint64(result.ConnectorID), 10))
		query.Set("ops_oauth_bound_vps", strconv.Itoa(result.BoundVPSCount))
		query.Set("ops_oauth_bound_projects", strconv.Itoa(result.BoundProjectCount))
		query.Set("ops_oauth_bound_domains", strconv.Itoa(result.BoundDomainCount))
		query.Set("ops_oauth_binding_warnings", strconv.Itoa(len(result.BindingWarnings)))
		query.Set("ops_oauth_message", result.Message)
	}
	parsed.RawQuery = query.Encode()
	c.Redirect(http.StatusFound, parsed.String())
}

func currentOpsConnectorUserID(c *gin.Context) (uint, bool) {
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

func respondOpsConnectorOAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrOpsConnectorOAuthNotConfigured),
		errors.Is(err, service.ErrOpsConnectorOAuthStateInvalid):
		apierror.RespondBadRequest(c, err.Error())
	case repository.IsRecordNotFound(err):
		apierror.RespondNotFound(c, "Operations connector")
	case errors.Is(err, service.ErrOpsConnectorOAuthExchange):
		apierror.RespondError(c, http.StatusBadGateway, "ops_connector_oauth_exchange_failed", err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
