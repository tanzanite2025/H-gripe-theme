package feedback

import (
	"errors"
	"net/http"
	"strconv"
	domainfeedback "tanzanite/internal/domain/feedback"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	feedbackService *service.FeedbackService
}

func NewHandler(feedbackService *service.FeedbackService) *Handler {
	return &Handler{feedbackService: feedbackService}
}

type createFeedbackRequest struct {
	Thread  string `json:"thread" binding:"required"`
	Content string `json:"content" binding:"required"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Locale  string `json:"locale"`
}

func (h *Handler) List(c *gin.Context) {
	threadKey := c.Query("thread")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("per_page", c.DefaultQuery("page_size", "20")))
	page, pageSize = normalizePagination(page, pageSize)

	items, total, err := h.feedbackService.List(threadKey, c.Query("status"), c.Query("search"), page, pageSize)
	if err != nil {
		respondFeedbackError(c, err)
		return
	}

	responses := make([]domainfeedback.Response, 0, len(items))
	for _, item := range items {
		responses = append(responses, item.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"data": responses,
		"pagination": gin.H{
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"page":        page,
			"per_page":    pageSize,
		},
	})
}

func (h *Handler) Create(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Please sign in to leave feedback."})
		return
	}

	var req createFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	item := &domainfeedback.Feedback{
		ThreadKey: req.Thread,
		UserID:    userID.(uint),
		Name:      fallbackString(req.Name, contextString(c, "username")),
		Email:     fallbackString(req.Email, contextString(c, "email")),
		Content:   req.Content,
		Status:    "pending",
		Locale:    req.Locale,
	}

	if err := h.feedbackService.Create(item); err != nil {
		respondFeedbackError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Feedback submitted, pending review.",
		"id":      item.ID,
		"status":  item.Status,
	})
}

func (h *Handler) Eligibility(c *gin.Context) {
	_, loggedIn := c.Get("user_id")
	var reason *string
	if !loggedIn {
		message := "Please sign in to leave feedback."
		reason = &message
	}

	c.JSON(http.StatusOK, gin.H{
		"can_post":  loggedIn,
		"logged_in": loggedIn,
		"reason":    reason,
	})
}

func contextString(c *gin.Context, key string) string {
	value, exists := c.Get(key)
	if !exists {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func fallbackString(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func respondFeedbackError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrFeedbackMissingThread), errors.Is(err, service.ErrFeedbackMissingContent):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
	case errors.Is(err, service.ErrFeedbackInvalidStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status", "message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "feedback_error", "message": err.Error()})
	}
}
