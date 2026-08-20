package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/api/middleware"
	domainfeedback "commerce-platform/internal/domain/feedback"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const pageFeedbackAuditResource = "page_feedback"

type PageFeedbackHandler struct {
	feedbackService *service.FeedbackService
	auditService    adminAuditRecorder
	redisClient     redis.UniversalClient
}

type AdminFeedbackUpdateRequest struct {
	Status       string `json:"status"`
	ReplyContent string `json:"reply_content"`
}

type AdminFeedback struct {
	ID                uint       `json:"id"`
	ThreadKey         string     `json:"thread_key"`
	PagePath          string     `json:"page_path"`
	PageTitle         string     `json:"page_title"`
	UserID            uint       `json:"user_id"`
	Name              string     `json:"name"`
	Email             string     `json:"email"`
	SourceHashPreview string     `json:"source_hash_preview"`
	Content           string     `json:"content"`
	Status            string     `json:"status"`
	Locale            string     `json:"locale"`
	ReplyContent      string     `json:"reply_content"`
	RepliedAt         *time.Time `json:"replied_at"`
	RepliedBy         uint       `json:"replied_by"`
	ReviewedAt        *time.Time `json:"reviewed_at"`
	ReviewedBy        uint       `json:"reviewed_by"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func NewPageFeedbackHandler(feedbackService *service.FeedbackService) *PageFeedbackHandler {
	return &PageFeedbackHandler{feedbackService: feedbackService}
}

func (h *PageFeedbackHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *PageFeedbackHandler) ConfigureRedisClient(redisClient redis.UniversalClient) {
	if h == nil {
		return
	}
	h.redisClient = redisClient
}

func (h *PageFeedbackHandler) List(c *gin.Context) {
	params := pagination.ParsePagination(c)
	items, total, err := h.feedbackService.ListAdmin(service.FeedbackAdminListFilter{
		Status:    c.DefaultQuery("status", "pending"),
		ThreadKey: c.Query("thread_key"),
		PagePath:  c.Query("page_path"),
		Search:    c.Query("search"),
		Page:      params.Page,
		PageSize:  params.PageSize,
	})
	if err != nil {
		respondPageFeedbackError(c, err)
		return
	}

	adminItems := make([]AdminFeedback, 0, len(items))
	for index := range items {
		adminItems = append(adminItems, buildAdminFeedback(&items[index]))
	}
	response.Paged(c, adminItems, params.Page, params.PageSize, total)
}

func (h *PageFeedbackHandler) Get(c *gin.Context) {
	id, ok := parseFeedbackID(c)
	if !ok {
		return
	}

	item, err := h.feedbackService.GetAdmin(id)
	if err != nil {
		respondPageFeedbackError(c, err)
		return
	}
	response.Success(c, buildAdminFeedback(item))
}

func (h *PageFeedbackHandler) RiskOverview(c *gin.Context) {
	windowHours := riskOverviewWindowHours(c)
	rateLimit := service.FeedbackRateLimitSnapshot{WindowHours: windowHours}
	localRateLimit := middleware.FeedbackLocalRateLimitBlockedCounts(windowHours)
	if h.redisClient == nil {
		rateLimit.Unavailable = true
	} else {
		blocked, err := middleware.FeedbackRateLimitBlockedCounts(c.Request.Context(), h.redisClient, windowHours)
		if err != nil {
			rateLimit.Unavailable = true
		} else {
			rateLimit = service.FeedbackRateLimitSnapshot{
				WindowHours: blocked.WindowHours,
				Total:       blocked.Total,
				ReadIP:      blocked.ReadIP,
				WriteIP:     blocked.WriteIP,
				WriteUser:   blocked.WriteUser,
			}
		}
	}
	rateLimit.FallbackTotal = localRateLimit.Total
	rateLimit.FallbackReadIP = localRateLimit.ReadIP
	rateLimit.FallbackWriteIP = localRateLimit.WriteIP
	rateLimit.FallbackWriteUser = localRateLimit.WriteUser
	rateLimit.RedisUnavailable = localRateLimit.RedisUnavailable
	rateLimit.Total += localRateLimit.Total
	rateLimit.ReadIP += localRateLimit.ReadIP
	rateLimit.WriteIP += localRateLimit.WriteIP
	rateLimit.WriteUser += localRateLimit.WriteUser

	overview, err := h.feedbackService.RiskOverview(service.FeedbackRiskOverviewInput{
		WindowHours: windowHours,
		GeneratedAt: time.Now().UTC(),
		RateLimit:   rateLimit,
	})
	if err != nil {
		respondPageFeedbackError(c, err)
		return
	}

	response.Success(c, overview)
}

func (h *PageFeedbackHandler) Update(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, ok := parseFeedbackID(c)
	if !ok {
		return
	}

	current, err := h.feedbackService.GetAdmin(id)
	if err != nil {
		h.recordPageFeedbackAudit(c, startedAt, id, nil, nil, err)
		respondPageFeedbackError(c, err)
		return
	}

	adminID, ok := currentAdminUserID(c)
	if !ok {
		err := errors.New("admin user id is required")
		h.recordPageFeedbackAudit(c, startedAt, id, current, nil, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req AdminFeedbackUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordPageFeedbackAudit(c, startedAt, id, current, nil, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.feedbackService.UpdateAdmin(id, service.FeedbackAdminUpdateInput{
		Status:       req.Status,
		ReplyContent: req.ReplyContent,
		AdminID:      adminID,
	})
	if err != nil {
		h.recordPageFeedbackAudit(c, startedAt, id, current, nil, err)
		respondPageFeedbackError(c, err)
		return
	}

	h.recordPageFeedbackAudit(c, startedAt, id, current, updated, nil)
	response.Success(c, buildAdminFeedback(updated))
}

func buildAdminFeedback(item *domainfeedback.Feedback) AdminFeedback {
	if item == nil {
		return AdminFeedback{}
	}
	return AdminFeedback{
		ID:                item.ID,
		ThreadKey:         item.ThreadKey,
		PagePath:          item.PagePath,
		PageTitle:         item.PageTitle,
		UserID:            item.UserID,
		Name:              item.Name,
		Email:             item.Email,
		SourceHashPreview: pageFeedbackSourceHashPreview(item.SourceHash),
		Content:           item.Content,
		Status:            item.Status,
		Locale:            item.Locale,
		ReplyContent:      item.ReplyContent,
		RepliedAt:         item.RepliedAt,
		RepliedBy:         item.RepliedBy,
		ReviewedAt:        item.ReviewedAt,
		ReviewedBy:        item.ReviewedBy,
		CreatedAt:         item.CreatedAt,
		UpdatedAt:         item.UpdatedAt,
	}
}

func pageFeedbackSourceHashPreview(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}

func (h *PageFeedbackHandler) recordPageFeedbackAudit(
	c *gin.Context,
	startedAt time.Time,
	id uint,
	oldItem *domainfeedback.Feedback,
	newItem *domainfeedback.Feedback,
	operationErr error,
) {
	status := adminAuditStatusSuccess
	errorMessage := ""
	if operationErr != nil {
		status = adminAuditStatusFailed
		errorMessage = operationErr.Error()
	}

	changes := gin.H{}
	if oldItem != nil && newItem != nil {
		if oldItem.Status != newItem.Status {
			changes["old_status"] = oldItem.Status
			changes["new_status"] = newItem.Status
		}
		if oldItem.ReplyContent != newItem.ReplyContent {
			changes["reply_updated"] = true
		}
	}

	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt:    startedAt,
		Action:       adminAuditActionUpdate,
		Resource:     pageFeedbackAuditResource,
		ResourceID:   id,
		Status:       status,
		ErrorMessage: errorMessage,
		Changes:      changes,
		OldValue:     pageFeedbackAuditValue(oldItem),
		NewValue:     pageFeedbackAuditValue(newItem),
	})
}

func pageFeedbackAuditValue(item *domainfeedback.Feedback) gin.H {
	if item == nil {
		return nil
	}
	return gin.H{
		"id":            item.ID,
		"thread_key":    item.ThreadKey,
		"page_path":     item.PagePath,
		"status":        item.Status,
		"reply_content": item.ReplyContent,
	}
}

func parseFeedbackID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid feedback id"})
		return 0, false
	}
	return uint(id), true
}

func riskOverviewWindowHours(c *gin.Context) int {
	windowHours, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("window_hours", "24")))
	if err != nil || windowHours < 1 || windowHours > 168 {
		return 24
	}
	return windowHours
}

func respondPageFeedbackError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrFeedbackNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "feedback not found"})
	case errors.Is(err, service.ErrFeedbackInvalidStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrFeedbackContentTooLong):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "feedback operation failed"})
	}
}
