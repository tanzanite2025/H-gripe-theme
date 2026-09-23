package referral

import (
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"commerce-platform/internal/domain/loyalty"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	referralcookie "commerce-platform/internal/pkg/referral"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/securecookie"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service         referralService
	cookieOptions   securecookie.Options
	redirectBaseURL string
}

type referralService interface {
	CreateAttributionToken(code, source string) (string, int, error)
	ValidateCode(code string) (*service.ReferralValidation, error)
	Dashboard(userID uint) (*service.ReferralDashboard, error)
	History(userID uint, page, pageSize int) (*service.ReferralHistory, error)
	BindFromToken(userID uint, token, clientIP string) (*loyalty.ReferralRecord, error)
}

type referralContextBinder interface {
	BindFromTokenWithContext(userID uint, token string, bindContext service.ReferralBindContext) (*loyalty.ReferralRecord, error)
}

func NewHandler(referralService referralService, options securecookie.Options, redirectBaseURL ...string) *Handler {
	handler := &Handler{
		service:       referralService,
		cookieOptions: securecookie.NormalizeOptions(options),
	}
	if len(redirectBaseURL) > 0 {
		handler.redirectBaseURL = strings.TrimRight(strings.TrimSpace(redirectBaseURL[0]), "/")
	}
	return handler
}

func (h *Handler) Capture(c *gin.Context) {
	token, maxAge, err := h.service.CreateAttributionToken(c.Param("code"), "link")
	if err != nil {
		respondReferralError(c, err)
		return
	}
	h.setReferralCookie(c, token, maxAge)
	redirectPath := safeLocalRedirect(c.Query("next"))
	if h.redirectBaseURL != "" {
		redirectPath = h.redirectBaseURL + redirectPath
	}
	c.Redirect(http.StatusFound, redirectPath)
}

func (h *Handler) Validate(c *gin.Context) {
	var request struct {
		ReferralCode string `json:"referral_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	result, err := h.service.ValidateCode(request.ReferralCode)
	if err != nil {
		respondReferralError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Me(c *gin.Context) {
	userID, exists := authenticatedUserID(c)
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}
	result, err := h.service.Dashboard(userID)
	if err != nil {
		respondReferralError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) History(c *gin.Context) {
	userID, exists := authenticatedUserID(c)
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}
	params := pagination.ParsePagination(c)
	result, err := h.service.History(userID, params.Page, params.PageSize)
	if err != nil {
		respondReferralError(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Bind(c *gin.Context) {
	userID, exists := authenticatedUserID(c)
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}
	// A referral can arrive through the signed attribution cookie created by a
	// share link, or as an explicit code entered during registration. The
	// latter is converted into the same signed token path so both flows share
	// the exact validation, anti-fraud, and idempotency rules.
	var input struct {
		ReferralCode string `json:"referral_code"`
	}
	manualCode := ""
	contentType := strings.ToLower(strings.TrimSpace(c.GetHeader("Content-Type")))
	if strings.HasPrefix(contentType, "application/json") {
		if decodeErr := c.ShouldBindJSON(&input); decodeErr != nil && !errors.Is(decodeErr, io.EOF) {
			apierror.RespondValidationError(c, decodeErr.Error())
			return
		}
		manualCode = strings.TrimSpace(input.ReferralCode)
	}

	token := ""
	var err error
	if manualCode != "" {
		token, _, err = h.service.CreateAttributionToken(manualCode, "manual_input")
		if err != nil {
			respondReferralError(c, err)
			return
		}
	} else {
		token, err = c.Cookie(referralcookie.CookieName)
		if err != nil || strings.TrimSpace(token) == "" {
			apierror.RespondError(c, http.StatusBadRequest, "referral_attribution_missing", "No valid referral attribution is available")
			return
		}
	}
	var record *loyalty.ReferralRecord
	var bindErr error
	bindContext := service.ReferralBindContext{
		ClientIP:          c.ClientIP(),
		DeviceFingerprint: c.GetHeader("X-Device-Fingerprint"),
	}
	if email, ok := c.Get("email"); ok {
		if value, ok := email.(string); ok {
			bindContext.RefereeEmail = value
		}
	}
	if binder, ok := h.service.(referralContextBinder); ok {
		record, bindErr = binder.BindFromTokenWithContext(userID, token, bindContext)
	} else {
		record, bindErr = h.service.BindFromToken(userID, token, c.ClientIP())
	}
	err = bindErr
	if err != nil {
		if isTerminalAttributionError(err) {
			h.setReferralCookie(c, "", -1)
		}
		respondReferralError(c, err)
		return
	}
	h.setReferralCookie(c, "", -1)
	response.Success(c, gin.H{"referral_id": record.ID, "status": record.Status})
}

func isTerminalAttributionError(err error) bool {
	return errors.Is(err, service.ErrReferralProgramDisabled) ||
		errors.Is(err, service.ErrReferralCodeNotFound) ||
		errors.Is(err, service.ErrSelfReferralForbidden) ||
		errors.Is(err, service.ErrRefereeNotEligible) ||
		errors.Is(err, service.ErrRefereeAlreadyAttributed) ||
		errors.Is(err, referralcookie.ErrInvalidCookie)
}

func (h *Handler) setReferralCookie(c *gin.Context, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     referralcookie.CookieName,
		Value:    value,
		Path:     "/",
		Domain:   h.cookieOptions.Domain,
		MaxAge:   maxAge,
		Secure:   h.cookieOptions.Secure,
		HttpOnly: true,
		SameSite: h.cookieOptions.SameSite,
	})
}

func authenticatedUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := value.(uint)
	return userID, ok && userID > 0
}

func safeLocalRedirect(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || !strings.HasPrefix(raw, "/") || strings.HasPrefix(raw, "//") {
		return "/"
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return "/"
	}
	return parsed.RequestURI()
}

func respondReferralError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrReferralProgramDisabled), errors.Is(err, service.ErrReferralCodeNotFound):
		apierror.RespondError(c, http.StatusNotFound, "referral_code_not_found", "Referral code is unavailable")
	case errors.Is(err, service.ErrSelfReferralForbidden):
		apierror.RespondError(c, http.StatusUnprocessableEntity, "self_referral_forbidden", err.Error())
	case errors.Is(err, service.ErrRefereeNotEligible):
		apierror.RespondError(c, http.StatusUnprocessableEntity, "referee_not_eligible", err.Error())
	case errors.Is(err, service.ErrRefereeAlreadyAttributed):
		apierror.RespondError(c, http.StatusConflict, "referee_already_attributed", err.Error())
	case errors.Is(err, referralcookie.ErrInvalidCookie):
		apierror.RespondError(c, http.StatusBadRequest, "referral_attribution_invalid", err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
