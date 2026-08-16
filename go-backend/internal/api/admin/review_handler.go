package admin

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	reviewdomain "commerce-platform/internal/domain/review"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const reviewAuditResource = "review"

type ReviewModerationHandler struct {
	reviewService *service.ReviewModerationService
	auditService  adminAuditRecorder
}

func NewReviewModerationHandler(reviewService *service.ReviewModerationService) *ReviewModerationHandler {
	return &ReviewModerationHandler{reviewService: reviewService}
}

func (h *ReviewModerationHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *ReviewModerationHandler) List(c *gin.Context) {
	params := pagination.ParsePagination(c)
	productID := parseOptionalUint(c.Query("product_id"))

	items, total, err := h.reviewService.List(
		c.DefaultQuery("status", reviewdomain.StatusPending),
		c.Query("search"),
		productID,
		params.Page,
		params.PageSize,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "review list failed"})
		return
	}

	adminItems := make([]AdminReview, 0, len(items))
	for index := range items {
		adminItems = append(adminItems, buildAdminReview(&items[index]))
	}
	response.Paged(c, adminItems, params.Page, params.PageSize, total)
}

func (h *ReviewModerationHandler) Get(c *gin.Context) {
	id, ok := parseReviewID(c)
	if !ok {
		return
	}

	item, err := h.reviewService.Get(id)
	if err != nil {
		respondReviewModerationError(c, err)
		return
	}
	response.Success(c, buildAdminReview(item))
}

func (h *ReviewModerationHandler) UpdateStatus(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, ok := parseReviewID(c)
	if !ok {
		return
	}

	current, err := h.reviewService.Get(id)
	if err != nil {
		h.recordReviewAudit(c, startedAt, id, "", "", "", err)
		respondReviewModerationError(c, err)
		return
	}

	adminID, ok := currentAdminUserID(c)
	if !ok {
		err := errors.New("admin user id is required")
		h.recordReviewAudit(c, startedAt, id, current.Status, "", "", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req UpdateReviewStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordReviewAudit(c, startedAt, id, current.Status, "", req.Reason, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.reviewService.UpdateStatus(id, req.Status, req.Reason, adminID)
	if err != nil {
		h.recordReviewAudit(c, startedAt, id, current.Status, req.Status, req.Reason, err)
		respondReviewModerationError(c, err)
		return
	}

	h.recordReviewAudit(c, startedAt, id, current.Status, updated.Status, req.Reason, nil)
	response.Success(c, buildAdminReview(updated))
}

type UpdateReviewStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Reason string `json:"reason"`
}

type AdminReview struct {
	ID               uint                `json:"id"`
	Product          *AdminReviewProduct `json:"product,omitempty"`
	User             *AdminReviewUser    `json:"user,omitempty"`
	OrderID          uint                `json:"order_id"`
	Rating           int                 `json:"rating"`
	Title            string              `json:"title"`
	Content          string              `json:"content"`
	Images           []string            `json:"images"`
	Pros             string              `json:"pros"`
	Cons             string              `json:"cons"`
	Status           string              `json:"status"`
	Featured         bool                `json:"featured"`
	Verified         bool                `json:"verified"`
	HelpfulCount     int                 `json:"helpful_count"`
	ReplyContent     string              `json:"reply_content"`
	RepliedAt        *time.Time          `json:"replied_at"`
	RepliedBy        uint                `json:"replied_by"`
	ModeratedAt      *time.Time          `json:"moderated_at"`
	ModeratedBy      *uint               `json:"moderated_by"`
	ModerationReason string              `json:"moderation_reason"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type AdminReviewProduct struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	SKU  string `json:"sku"`
}

type AdminReviewUser struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

func buildAdminReview(item *reviewdomain.Review) AdminReview {
	result := AdminReview{
		ID:               item.ID,
		OrderID:          item.OrderID,
		Rating:           item.Rating,
		Title:            item.Title,
		Content:          item.Content,
		Images:           decodeReviewImages(item.Images),
		Pros:             item.Pros,
		Cons:             item.Cons,
		Status:           item.Status,
		Featured:         item.Featured,
		Verified:         item.Verified,
		HelpfulCount:     item.HelpfulCount,
		ReplyContent:     item.ReplyContent,
		RepliedAt:        item.RepliedAt,
		RepliedBy:        item.RepliedBy,
		ModeratedAt:      item.ModeratedAt,
		ModeratedBy:      item.ModeratedBy,
		ModerationReason: item.ModerationReason,
		CreatedAt:        item.CreatedAt,
		UpdatedAt:        item.UpdatedAt,
	}
	if item.Product != nil {
		result.Product = &AdminReviewProduct{
			ID:   item.Product.ID,
			Name: item.Product.Name,
			SKU:  item.Product.SKU,
		}
	}
	if item.User != nil {
		displayName := strings.TrimSpace(item.User.Username)
		if displayName == "" {
			displayName = strings.TrimSpace(item.User.Email)
		}
		result.User = &AdminReviewUser{
			ID:          item.User.ID,
			Username:    item.User.Username,
			Email:       item.User.Email,
			DisplayName: displayName,
		}
	}
	return result
}

func decodeReviewImages(value string) []string {
	var images []string
	if err := json.Unmarshal([]byte(value), &images); err != nil || images == nil {
		return []string{}
	}
	return images
}

func (h *ReviewModerationHandler) recordReviewAudit(
	c *gin.Context,
	startedAt time.Time,
	id uint,
	oldStatus string,
	newStatus string,
	reason string,
	operationErr error,
) {
	status := adminAuditStatusSuccess
	errorMessage := ""
	if operationErr != nil {
		status = adminAuditStatusFailed
		errorMessage = operationErr.Error()
	}
	changes := gin.H{
		"old_status": oldStatus,
		"new_status": newStatus,
	}
	if strings.TrimSpace(reason) != "" {
		changes["reason"] = strings.TrimSpace(reason)
	}
	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt:    startedAt,
		Action:       adminAuditActionUpdate,
		Resource:     reviewAuditResource,
		ResourceID:   id,
		Status:       status,
		ErrorMessage: errorMessage,
		Changes:      changes,
		OldValue:     gin.H{"status": oldStatus},
		NewValue:     gin.H{"status": newStatus},
	})
}

func parseReviewID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid review id"})
		return 0, false
	}
	return uint(id), true
}

func parseOptionalUint(value string) *uint {
	id, err := strconv.ParseUint(strings.TrimSpace(value), 10, 32)
	if err != nil || id == 0 {
		return nil
	}
	parsed := uint(id)
	return &parsed
}

func respondReviewModerationError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidReviewStatus):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrReviewModerationReason):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidReviewTransition):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrReviewModerationNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "review not found"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "review operation failed"})
	}
}
