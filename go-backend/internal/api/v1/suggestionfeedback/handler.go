package suggestionfeedback

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"
	domainsuggestion "tanzanite/internal/domain/suggestionfeedback"
	"tanzanite/internal/pkg/storage"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	suggestionService *service.SuggestionFeedbackService
	storageService    storage.StorageService
}

type createSuggestionRequest struct {
	FullName        string                        `json:"fullName"`
	Email           string                        `json:"email"`
	Country         string                        `json:"country"`
	OrderNumber     string                        `json:"orderNumber"`
	ProductCategory string                        `json:"productCategory"`
	RequestType     string                        `json:"requestType"`
	Message         string                        `json:"message" binding:"required"`
	Attachments     []domainsuggestion.Attachment `json:"attachments"`
	ThreadKey       string                        `json:"threadKey"`
}

func NewHandler(suggestionService *service.SuggestionFeedbackService, storageService storage.StorageService) *Handler {
	return &Handler{
		suggestionService: suggestionService,
		storageService:    storageService,
	}
}

func (h *Handler) Eligibility(c *gin.Context) {
	userID, loggedIn := currentUserID(c)
	eligibility, err := h.suggestionService.GetEligibility(userID, loggedIn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "suggestion_feedback_error", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, eligibility)
}

func (h *Handler) Upload(c *gin.Context) {
	_, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Please sign in to upload files."})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_file", "message": "No file uploaded"})
		return
	}

	url, err := h.storageService.Upload(c.Request.Context(), file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload_failed", "message": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"url":  url,
		"name": file.Filename,
		"size": file.Size,
	})
}

func (h *Handler) Create(c *gin.Context) {
	userID, exists := currentUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Please sign in before submitting feedback."})
		return
	}

	var req createSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_payload", "message": err.Error()})
		return
	}

	eligibility, err := h.suggestionService.GetEligibility(userID, true)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "suggestion_feedback_error", "message": err.Error()})
		return
	}

	item := &domainsuggestion.SuggestionFeedback{
		UserID:              userID,
		FullName:            req.FullName,
		Email:               req.Email,
		Country:             req.Country,
		OrderNumber:         req.OrderNumber,
		ProductCategory:     req.ProductCategory,
		RequestType:         req.RequestType,
		Message:             req.Message,
		Meta:                domainsuggestion.JSONFromMeta(requestMeta(c, req.ThreadKey)),
		Status:              "new",
		MemberLevelRequired: eligibility.RequiredLevel,
		MemberLevelMet:      eligibility.CanAttach,
		EligibilityHash:     eligibilityHash(c.GetHeader("User-Agent")),
	}

	if err := h.suggestionService.Create(item, req.Attachments); err != nil {
		respondSuggestionError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      item.ID,
		"status":  item.Status,
		"message": "Feedback submitted. Customer service will review it soon.",
	})
}

func currentUserID(c *gin.Context) (uint, bool) {
	value, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	userID, ok := value.(uint)
	if !ok {
		return 0, false
	}
	return userID, true
}

func requestMeta(c *gin.Context, threadKey string) map[string]string {
	return map[string]string{
		"ip":        c.ClientIP(),
		"agent":     c.GetHeader("User-Agent"),
		"locale":    c.GetHeader("Accept-Language"),
		"threadKey": threadKey,
	}
}

func eligibilityHash(userAgent string) string {
	hash := sha1.Sum([]byte(userAgent))
	return hex.EncodeToString(hash[:])
}

func respondSuggestionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSuggestionFeedbackMissingMessage):
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_field", "message": err.Error()})
	case errors.Is(err, service.ErrSuggestionFeedbackInvalidStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_status", "message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "suggestion_feedback_error", "message": err.Error()})
	}
}
