package subscription

import (
	"errors"
	"net/http"
	"strings"

	domainsubscription "commerce-platform/internal/domain/subscription"
	"commerce-platform/internal/pkg/antibot"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	subscriptionService *service.SubscriptionService
	antiBot             *antibot.Service
}

func NewHandler(subscriptionService *service.SubscriptionService, antiBotServices ...*antibot.Service) *Handler {
	var antiBot *antibot.Service
	if len(antiBotServices) > 0 {
		antiBot = antiBotServices[0]
	}
	return &Handler{
		subscriptionService: subscriptionService,
		antiBot:             antiBot,
	}
}

func acceptedSubscriptionResponse(c *gin.Context) {
	c.JSON(http.StatusAccepted, gin.H{"message": "If the email can be subscribed, the request has been accepted."})
}

func (h *Handler) Subscribe(c *gin.Context) {
	var req struct {
		Email        string   `json:"email" binding:"required,email"`
		Source       string   `json:"source"`
		Locale       string   `json:"locale"`
		Tags         []string `json:"tags"`
		CaptchaToken string   `json:"captcha_token"`
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
	if h.subscriptionService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "subscription service is unavailable"})
		return
	}
	if !h.allowDelivery(c, req.Email, req.CaptchaToken) {
		return
	}

	_, err := h.subscriptionService.Subscribe(req.Email, req.Source, req.Locale, req.Tags)
	if err != nil {
		if err.Error() == "email already subscribed" {
			acceptedSubscriptionResponse(c)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if _, err := h.subscriptionService.IssueSubscriptionConfirmation(req.Email); err != nil {
		if h.antiBot != nil {
			h.antiBot.RecordDeliveryResult("email", false)
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to send subscription confirmation"})
		return
	}
	if h.antiBot != nil {
		h.antiBot.RecordDeliveryResult("email", true)
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
			if h.antiBot != nil {
				h.antiBot.RecordDeliveryResult("email", false)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to send subscription action email"})
			return
		}
		if h.antiBot != nil {
			h.antiBot.RecordDeliveryResult("email", true)
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
			if h.antiBot != nil {
				h.antiBot.RecordDeliveryResult("email", false)
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to send subscription action email"})
			return
		}
		if h.antiBot != nil {
			h.antiBot.RecordDeliveryResult("email", true)
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
		if err := h.subscriptionService.RequestStatus(email); err != nil {
			if h.antiBot != nil {
				h.antiBot.RecordDeliveryResult("email", false)
			}
		} else if h.antiBot != nil {
			h.antiBot.RecordDeliveryResult("email", true)
		}
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
