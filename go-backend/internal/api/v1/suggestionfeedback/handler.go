package suggestionfeedback

import (
	domainsuggestion "commerce-platform/internal/domain/suggestionfeedback"
	"commerce-platform/internal/pkg/honeypot"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const suggestionFeedbackMaxRequestBytes = 6 << 20

const (
	suggestionFeedbackUploadMaxPerUserPerDay       = 20
	suggestionFeedbackUploadBytesPerUserDay  int64 = 50 << 20
)

type Handler struct {
	suggestionService *service.SuggestionFeedbackService
	storageService    storage.StorageService
	mediaService      *service.MediaService
	honeypotPolicy    honeypot.Policy
	quotaMu           sync.Mutex
	quotas            map[uint]*suggestionUploadQuota
}

type suggestionUploadQuota struct {
	day   time.Time
	count int
	bytes int64
}

type createSuggestionRequest struct {
	FullName        string                        `json:"fullName"`
	Email           string                        `json:"email"`
	Country         string                        `json:"country"`
	OrderNumber     string                        `json:"orderNumber"`
	ProductCategory string                        `json:"productCategory"`
	RequestType     string                        `json:"requestType"`
	Message         string                        `json:"message" binding:"required"`
	CompanyTaxID    string                        `json:"company_tax_id"`
	Attachments     []domainsuggestion.Attachment `json:"attachments"`
	ThreadKey       string                        `json:"threadKey"`
}

func NewHandler(suggestionService *service.SuggestionFeedbackService, storageService storage.StorageService, mediaServices ...*service.MediaService) *Handler {
	var mediaService *service.MediaService
	if len(mediaServices) > 0 {
		mediaService = mediaServices[0]
	}
	return &Handler{
		suggestionService: suggestionService,
		storageService:    storageService,
		mediaService:      mediaService,
		honeypotPolicy:    honeypot.NewPolicy(honeypot.ModeEnforce),
		quotas:            make(map[uint]*suggestionUploadQuota),
	}
}

func (h *Handler) ConfigureHoneypot(policy honeypot.Policy) {
	if h != nil {
		h.honeypotPolicy = policy
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
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, suggestionFeedbackMaxRequestBytes)

	userID, exists := currentUserID(c)
	if !exists || userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Please sign in to upload files."})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": upload.CodeFileTooLarge, "message": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing_file", "message": "No file uploaded"})
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}
	if err := upload.ValidateSpecFile(file, string(upload.SpecSuggestionAttachment)); err != nil {
		c.JSON(upload.HTTPStatus(err), gin.H{"error": upload.ErrorCode(err), "message": err.Error()})
		return
	}
	if h.storageService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "storage_unavailable", "message": "File storage is unavailable."})
		return
	}
	if !h.reserveSuggestionUpload(userID, file.Size) {
		c.Header("Retry-After", "86400")
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "upload_quota_exceeded", "message": "Daily upload quota exceeded."})
		return
	}

	url, err := h.storageService.Upload(c.Request.Context(), file)
	if err != nil {
		h.releaseSuggestionUpload(userID, file.Size)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload_failed", "message": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"url":  h.publicMediaURL(url),
		"name": file.Filename,
		"size": file.Size,
	})
}

func (h *Handler) reserveSuggestionUpload(userID uint, size int64) bool {
	if h == nil || userID == 0 || size <= 0 {
		return false
	}
	h.quotaMu.Lock()
	defer h.quotaMu.Unlock()
	today := time.Now().UTC().Truncate(24 * time.Hour)
	quota := h.quotas[userID]
	if quota == nil || !quota.day.Equal(today) {
		quota = &suggestionUploadQuota{day: today}
		h.quotas[userID] = quota
	}
	if quota.count >= suggestionFeedbackUploadMaxPerUserPerDay || quota.bytes > suggestionFeedbackUploadBytesPerUserDay-size {
		return false
	}
	quota.count++
	quota.bytes += size
	return true
}

func (h *Handler) releaseSuggestionUpload(userID uint, size int64) {
	h.quotaMu.Lock()
	defer h.quotaMu.Unlock()
	if quota := h.quotas[userID]; quota != nil {
		if quota.count > 0 {
			quota.count--
		}
		if quota.bytes >= size {
			quota.bytes -= size
		} else {
			quota.bytes = 0
		}
	}
}

func (h *Handler) publicMediaURL(value string) string {
	if h == nil || h.mediaService == nil {
		return value
	}
	return h.mediaService.CanonicalPublicMediaURL(value)
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
	if h.honeypotPolicy.ShouldDrop(req.CompanyTaxID, "suggestion_feedback", "company_tax_id", c.Request.URL.Path) {
		c.JSON(http.StatusCreated, gin.H{
			"status":  "new",
			"message": "Feedback submitted. Customer service will review it soon.",
		})
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
