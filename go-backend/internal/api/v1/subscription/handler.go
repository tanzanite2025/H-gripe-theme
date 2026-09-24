package subscription

import (
	"errors"
	"net/http"
	"strings"
	"time"

	domainsubscription "commerce-platform/internal/domain/subscription"
	"commerce-platform/internal/pkg/antibot"
	"commerce-platform/internal/pkg/honeypot"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	subscriptionService *service.SubscriptionService
	antiBot             *antibot.Service
	honeypotPolicy      honeypot.Policy
	timingTokenKey      string
	timingTokenTTL      time.Duration
	timingReplay        *honeypot.TimingReplayObserver
	now                 func() time.Time
}

func NewHandler(subscriptionService *service.SubscriptionService, antiBotServices ...*antibot.Service) *Handler {
	var antiBot *antibot.Service
	if len(antiBotServices) > 0 {
		antiBot = antiBotServices[0]
	}
	return &Handler{
		subscriptionService: subscriptionService,
		antiBot:             antiBot,
		honeypotPolicy:      honeypot.NewPolicy(honeypot.ModeEnforce),
		timingTokenTTL:      honeypot.DefaultTimingTokenTTL(),
		now:                 time.Now,
	}
}

func (h *Handler) ConfigureHoneypot(policy honeypot.Policy) {
	if h != nil {
		h.honeypotPolicy = policy
	}
}

func (h *Handler) ConfigureTimingToken(key string, ttl time.Duration) {
	if h == nil {
		return
	}
	h.timingTokenKey = strings.TrimSpace(key)
	if ttl > 0 {
		h.timingTokenTTL = ttl
	}
}

func (h *Handler) ConfigureTimingReplay(observer *honeypot.TimingReplayObserver) {
	if h != nil {
		h.timingReplay = observer
	}
}

func acceptedSubscriptionResponse(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{"message": "If the email can be subscribed, the request has been accepted."})
}

func (h *Handler) IssueTimingToken(c *gin.Context) {
	if h == nil || h.timingTokenKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "form timing service is unavailable"})
		return
	}
	token, err := honeypot.IssueTimingToken("newsletter", h.timingTokenKey, h.now())
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "form timing service is unavailable"})
		return
	}
	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{
		"token":      token,
		"expires_in": int64(h.timingTokenTTL / time.Second),
	})
}

func (h *Handler) Subscribe(c *gin.Context) {
	var req struct {
		Email              string   `json:"email" binding:"required,email"`
		Source             string   `json:"source"`
		Locale             string   `json:"locale"`
		Tags               []string `json:"tags"`
		CaptchaToken       string   `json:"captcha_token"`
		CorporateTaxNumber string   `json:"corporate_tax_number"`
		TimingToken        string   `json:"timing_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Source == "" {
		req.Source = "website"
	}
	if req.Locale == "" {
		req.Locale = "en"
	}
	now := h.now()
	timingObservation := honeypot.ObserveTimingTokenDetailed(req.TimingToken, "newsletter", h.timingTokenKey, now, h.timingTokenTTL)
	if timingObservation.Valid && h.timingReplay != nil {
		h.timingReplay.Observe(c.Request.Context(), timingObservation.Claims.Form, timingObservation.Claims, now, h.timingTokenTTL)
	}
	if h.honeypotPolicy.ShouldDrop(req.CorporateTaxNumber, "newsletter", "corporate_tax_number", c.Request.URL.Path) {
		acceptedSubscriptionResponse(c)
		return
	}
	if h.subscriptionService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "subscription service is unavailable"})
		return
	}
	if !h.allowDelivery(c, req.Email, req.CaptchaToken) {
		return
	}

	_, _, err := h.subscriptionService.Subscribe(req.Email, req.Source, req.Locale, req.Tags)
	if err != nil {
		if err.Error() == "email already subscribed" {
			acceptedSubscriptionResponse(c)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	acceptedSubscriptionResponse(c)
}

func (h *Handler) ConfirmSubscription(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}
	if h.subscriptionService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "subscription service is unavailable"})
		return
	}
	if err := h.subscriptionService.ConfirmSubscription(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired confirmation token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed successfully"})
}

func (h *Handler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}
	if h.subscriptionService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "subscription service is unavailable"})
		return
	}
	if err := h.subscriptionService.Unsubscribe(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired unsubscribe token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
}

func (h *Handler) UnsubscribeByEmail(c *gin.Context) {
	var req struct {
		Email        string `json:"email" binding:"required,email"`
		CaptchaToken string `json:"captcha_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.allowDelivery(c, req.Email, req.CaptchaToken) {
		return
	}
	if h.subscriptionService != nil {
		if err := h.subscriptionService.UnsubscribeByEmail(req.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to send subscription action email"})
			return
		}
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "If the subscription exists, the request has been accepted."})
}

func (h *Handler) Resubscribe(c *gin.Context) {
	var req struct {
		Email        string `json:"email" binding:"required,email"`
		CaptchaToken string `json:"captcha_token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !h.allowDelivery(c, req.Email, req.CaptchaToken) {
		return
	}
	if h.subscriptionService != nil {
		if err := h.subscriptionService.Resubscribe(req.Email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to send subscription action email"})
			return
		}
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "If the subscription exists, the request has been accepted."})
}

func (h *Handler) ResubscribeByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}
	if h.subscriptionService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "subscription service is unavailable"})
		return
	}

	if err := h.subscriptionService.ResubscribeByToken(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired subscription token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Subscription resumed successfully"})
}

func (h *Handler) GetSubscription(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	if token := c.Query("token"); token != "" && h.subscriptionService != nil {
		sub, err := h.subscriptionService.GetSubscriptionByToken(token)
		if err != nil || !strings.EqualFold(strings.TrimSpace(email), sub.Email) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired subscription token"})
			return
		}
		c.JSON(http.StatusOK, publicSubscriptionResponse(*sub))
		return
	}

	if h.subscriptionService != nil {
		if !h.allowDelivery(c, email, c.Query("captcha_token")) {
			return
		}
		_ = h.subscriptionService.RequestStatus(email)
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "If the subscription exists, the request has been accepted."})
}

func (h *Handler) allowDelivery(c *gin.Context, destination, challengeToken string) bool {
	if h.antiBot == nil {
		return true
	}
	err := h.antiBot.Guard(c.Request.Context(), "email", destination, c.ClientIP(), challengeToken)
	switch {
	case err == nil:
		return true
	case errors.Is(err, antibot.ErrChallengeRequired), errors.Is(err, antibot.ErrChallengeInvalid):
		c.JSON(http.StatusForbidden, gin.H{"error": "verification challenge required"})
	case errors.Is(err, antibot.ErrRateLimited):
		c.Header("Retry-After", "60")
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many verification requests"})
	case errors.Is(err, antibot.ErrBudgetExceeded), errors.Is(err, antibot.ErrCircuitOpen):
		c.Header("Retry-After", "300")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Verification delivery is temporarily paused"})
	default:
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Verification service is temporarily unavailable"})
	}
	return false
}

func (h *Handler) GetSubscriptionByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}
	if h.subscriptionService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "subscription service is unavailable"})
		return
	}

	sub, err := h.subscriptionService.GetSubscriptionByToken(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired subscription token"})
		return
	}
	c.JSON(http.StatusOK, publicSubscriptionResponse(*sub))
}

type subscriptionResponse struct {
	ID     uint   `json:"id"`
	Email  string `json:"email"`
	Status string `json:"status"`
	Locale string `json:"locale"`
}

func publicSubscriptionResponse(item domainsubscription.Subscription) subscriptionResponse {
	return subscriptionResponse{
		ID:     item.ID,
		Email:  item.Email,
		Status: item.Status,
		Locale: item.Locale,
	}
}
