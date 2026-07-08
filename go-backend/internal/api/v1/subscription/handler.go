package subscription

import (
	"net/http"

	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	subscriptionService *service.SubscriptionService
}

func NewHandler(subscriptionService *service.SubscriptionService) *Handler {
	return &Handler{
		subscriptionService: subscriptionService,
	}
}

// Subscribe 鐠併垽妲?// POST /api/v1/subscriptions
func (h *Handler) Subscribe(c *gin.Context) {
	var req struct {
		Email  string   `json:"email" binding:"required,email"`
		Source string   `json:"source"`
		Locale string   `json:"locale"`
		Tags   []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 鐠佸墽鐤嗘妯款吇閸?
	if req.Source == "" {
		req.Source = "website"
	}
	if req.Locale == "" {
		req.Locale = "en"
	}

	sub, err := h.subscriptionService.Subscribe(req.Email, req.Source, req.Locale, req.Tags)
	if err != nil {
		if err.Error() == "email already subscribed" {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already subscribed"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Subscribed successfully",
		"data":    sub,
	})
}

// Unsubscribe 闁偓鐠?// GET /api/v1/subscriptions/unsubscribe/:token
func (h *Handler) Unsubscribe(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token is required"})
		return
	}

	if err := h.subscriptionService.Unsubscribe(token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
}

// UnsubscribeByEmail 闁俺绻冮柇顔绢唸闁偓鐠?// POST /api/v1/subscriptions/unsubscribe
func (h *Handler) UnsubscribeByEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.subscriptionService.UnsubscribeByEmail(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
}

// Resubscribe 闁插秵鏌婄拋銏ゆ
// POST /api/v1/subscriptions/resubscribe
func (h *Handler) Resubscribe(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.subscriptionService.Resubscribe(req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resubscribed successfully"})
}

// GetSubscription 閼惧嘲褰囩拋銏ゆ閻樿埖鈧?// GET /api/v1/subscriptions/status/:email
func (h *Handler) GetSubscription(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email is required"})
		return
	}

	sub, err := h.subscriptionService.GetSubscription(email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sub})
}
