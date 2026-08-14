package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	showcasedomain "commerce-platform/internal/domain/showcase"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type ShowcaseHandler struct {
	showcaseService *service.ShowcaseService
	auditService    adminAuditRecorder
}

func NewShowcaseHandler(showcaseService *service.ShowcaseService) *ShowcaseHandler {
	return &ShowcaseHandler{showcaseService: showcaseService}
}

func (h *ShowcaseHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *ShowcaseHandler) List(c *gin.Context) {
	params := pagination.ParsePagination(c)
	kind := strings.TrimSpace(c.DefaultQuery("type", showcasedomain.KindUser))
	status := normalizeAdminShowcaseStatus(c.DefaultQuery("status", showcasedomain.StatusPending))

	items, err := h.showcaseService.List(kind, status, params.Page, params.PageSize)
	if err != nil {
		respondShowcaseError(c, err)
		return
	}

	total, err := h.showcaseService.Count(kind, status)
	if err != nil {
		respondShowcaseError(c, err)
		return
	}

	response.Paged(c, gin.H{"items": buildAdminShowcaseListItems(items)}, params.Page, params.PageSize, total)
}

func (h *ShowcaseHandler) Approve(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, ok := parseShowcaseID(c)
	if !ok {
		return
	}

	item, err := h.showcaseService.Get(id)
	if err != nil {
		h.recordModerationAudit(c, startedAt, id, "", showcasedomain.StatusApproved, "", err)
		respondShowcaseError(c, err)
		return
	}

	if err := h.showcaseService.Approve(c.Request.Context(), id); err != nil {
		h.recordModerationAudit(c, startedAt, id, item.Status, showcasedomain.StatusApproved, "", err)
		respondShowcaseError(c, err)
		return
	}

	h.recordModerationAudit(c, startedAt, id, item.Status, showcasedomain.StatusApproved, "", nil)
	response.Success(c, gin.H{"message": "Showcase approved"})
}

func (h *ShowcaseHandler) Reject(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, ok := parseShowcaseID(c)
	if !ok {
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason is required"})
		return
	}
	req.Reason = strings.TrimSpace(req.Reason)
	if req.Reason == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason is required"})
		return
	}

	item, err := h.showcaseService.Get(id)
	if err != nil {
		h.recordModerationAudit(c, startedAt, id, "", showcasedomain.StatusRejected, req.Reason, err)
		respondShowcaseError(c, err)
		return
	}
	if err := h.showcaseService.Reject(c.Request.Context(), id, req.Reason); err != nil {
		h.recordModerationAudit(c, startedAt, id, item.Status, showcasedomain.StatusRejected, req.Reason, err)
		respondShowcaseError(c, err)
		return
	}

	h.recordModerationAudit(c, startedAt, id, item.Status, showcasedomain.StatusRejected, req.Reason, nil)
	response.Success(c, gin.H{"message": "Showcase rejected"})
}

func (h *ShowcaseHandler) recordModerationAudit(
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
	if reason != "" {
		changes["reason"] = reason
	}
	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt:    startedAt,
		Action:       adminAuditActionUpdate,
		Resource:     "showcase",
		ResourceID:   id,
		Status:       status,
		ErrorMessage: errorMessage,
		Changes:      changes,
		OldValue:     gin.H{"status": oldStatus},
		NewValue:     gin.H{"status": newStatus},
	})
}

func (h *ShowcaseHandler) ServeImageFile(c *gin.Context) {
	id, ok := parseShowcaseID(c)
	if !ok {
		return
	}
	imageIndex, ok := parseShowcaseImageIndex(c)
	if !ok {
		return
	}

	file, err := h.showcaseService.OpenImageFile(c.Request.Context(), id, imageIndex)
	if err != nil {
		respondShowcaseError(c, err)
		return
	}
	if strings.TrimSpace(file.RedirectURL) != "" {
		c.Header("Cache-Control", "private, no-store")
		c.Header("Pragma", "no-cache")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Redirect(http.StatusTemporaryRedirect, file.RedirectURL)
		return
	}
	if file.ReadCloser == nil {
		respondShowcaseError(c, service.ErrShowcaseStorageUnavailable)
		return
	}
	defer func() { _ = file.ReadCloser.Close() }()

	mimeType := strings.TrimSpace(file.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	headers := map[string]string{
		"Cache-Control":          "private, no-store",
		"Content-Disposition":    fmt.Sprintf("inline; filename=%q", file.Filename),
		"X-Content-Type-Options": "nosniff",
	}
	c.DataFromReader(http.StatusOK, file.Size, mimeType, file.ReadCloser, headers)
}

func parseShowcaseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return 0, false
	}
	return uint(id), true
}

func parseShowcaseImageIndex(c *gin.Context) (int, bool) {
	index, err := strconv.Atoi(c.Param("image_index"))
	if err != nil || index < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image index"})
		return 0, false
	}
	return index, true
}

func normalizeAdminShowcaseStatus(value string) string {
	status := strings.ToLower(strings.TrimSpace(value))
	switch status {
	case "", showcasedomain.StatusPending:
		return showcasedomain.StatusPending
	case showcasedomain.StatusApproved, showcasedomain.StatusRejected:
		return status
	case "all", "*":
		return ""
	default:
		return showcasedomain.StatusPending
	}
}

func respondShowcaseError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrShowcaseNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrShowcaseImageNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrShowcaseStorageUnavailable):
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrShowcaseImagesInvalid):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrShowcaseInvalidTransition):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "showcase operation failed"})
	}
}
