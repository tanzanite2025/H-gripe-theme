package urlmanagement

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const adminAuditResourceStorefrontURLIssue = "storefront_url_issue"

type IssuesHandler struct {
	issues  *service.StorefrontURLIssueService
	catalog *service.StorefrontRouteCatalogService
	audit   urlIssueAuditRecorder
}

func NewIssuesHandler(
	issues *service.StorefrontURLIssueService,
	catalog *service.StorefrontRouteCatalogService,
) *IssuesHandler {
	return &IssuesHandler{
		issues:  issues,
		catalog: catalog,
	}
}

func (h *IssuesHandler) ConfigureAuditService(audit urlIssueAuditRecorder) {
	if h == nil {
		return
	}
	h.audit = audit
}

func (h *IssuesHandler) List(c *gin.Context) {
	page, pageSize := urlPagination(c)
	issues, total, err := h.issues.List(repository.StorefrontURLIssueListFilter{
		Page:      page,
		PageSize:  pageSize,
		State:     c.DefaultQuery("state", "active"),
		Severity:  c.Query("severity"),
		IssueType: c.Query("issue_type"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": issues,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *IssuesHandler) Summary(c *gin.Context) {
	stats, err := h.issues.Stats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

func (h *IssuesHandler) Get(c *gin.Context) {
	issue, err := h.getIssue(c)
	if err != nil {
		writeURLIssueError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": issue})
}

func (h *IssuesHandler) Events(c *gin.Context) {
	id, err := parseURLIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	page, pageSize := urlPagination(c)
	events, total, err := h.issues.ListEvents(id, page, pageSize)
	if err != nil {
		writeURLIssueError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items": events,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *IssuesHandler) Acknowledge(c *gin.Context) {
	var input urlmanagementdomain.StorefrontURLIssueActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.runIssueAction(c, "acknowledge", func(id uint) (*urlmanagementdomain.StorefrontURLIssue, error) {
		return h.issues.Acknowledge(id, c.GetUint("user_id"), input)
	})
}

func (h *IssuesHandler) Claim(c *gin.Context) {
	h.runIssueAction(c, "claim", func(id uint) (*urlmanagementdomain.StorefrontURLIssue, error) {
		return h.issues.Claim(id, c.GetUint("user_id"))
	})
}

func (h *IssuesHandler) Comment(c *gin.Context) {
	var input urlmanagementdomain.StorefrontURLIssueActionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.runIssueAction(c, "comment", func(id uint) (*urlmanagementdomain.StorefrontURLIssue, error) {
		return h.issues.AddComment(id, c.GetUint("user_id"), input)
	})
}

func (h *IssuesHandler) LinkRedirect(c *gin.Context) {
	var input urlmanagementdomain.StorefrontURLIssueLinkRedirectInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.runIssueAction(c, "link_redirect", func(id uint) (*urlmanagementdomain.StorefrontURLIssue, error) {
		return h.issues.LinkRedirect(id, c.GetUint("user_id"), input)
	})
}

func (h *IssuesHandler) Resolve(c *gin.Context) {
	var input urlmanagementdomain.StorefrontURLIssueResolutionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.runIssueAction(c, "resolve", func(id uint) (*urlmanagementdomain.StorefrontURLIssue, error) {
		return h.issues.Resolve(id, c.GetUint("user_id"), input)
	})
}

func (h *IssuesHandler) Suppress(c *gin.Context) {
	var input urlmanagementdomain.StorefrontURLIssueSuppressionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.runIssueAction(c, "suppress", func(id uint) (*urlmanagementdomain.StorefrontURLIssue, error) {
		return h.issues.Suppress(id, c.GetUint("user_id"), input)
	})
}

func (h *IssuesHandler) Recheck(c *gin.Context) {
	startedAt := urlIssueAuditStartedAt()
	issue, err := h.getIssue(c)
	if err != nil {
		writeURLIssueError(c, err)
		return
	}
	if issue.RouteEntry == nil || !issue.RouteEntry.IsCheckable {
		c.JSON(http.StatusBadRequest, gin.H{"error": "this URL issue cannot be checked"})
		return
	}
	result, err := h.catalog.CheckEntry(contextOrBackground(c), issue.RouteEntryID)
	if err != nil {
		h.recordIssueAudit(c, startedAt, "probe", issue.ID, "failed", err, nil, nil)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	updatedIssue, err := h.issues.Get(issue.ID)
	if err != nil {
		writeURLIssueError(c, err)
		return
	}
	h.recordIssueAudit(c, startedAt, "probe", issue.ID, "success", nil, map[string]interface{}{
		"route_entry_id": issue.RouteEntryID,
	}, map[string]interface{}{
		"check_result_id": result.ID,
		"status":          result.Status,
	})
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"issue":        updatedIssue,
		"check_result": result,
	}})
}

func (h *IssuesHandler) Verify(c *gin.Context) {
	startedAt := urlIssueAuditStartedAt()
	issue, err := h.getIssue(c)
	if err != nil {
		writeURLIssueError(c, err)
		return
	}
	var checkResult interface{}
	if h.issues.RequiresRouteCheck(issue) {
		if issue.RouteEntry == nil || !issue.RouteEntry.IsCheckable {
			c.JSON(http.StatusBadRequest, gin.H{"error": "this URL issue cannot be checked"})
			return
		}
		result, err := h.catalog.CheckEntry(contextOrBackground(c), issue.RouteEntryID)
		if err != nil {
			h.recordIssueAudit(c, startedAt, "probe", issue.ID, "failed", err, nil, nil)
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		checkResult = result
	}

	updatedIssue, err := h.issues.Verify(issue.ID, c.GetUint("user_id"))
	if err != nil {
		h.recordIssueAudit(c, startedAt, "execute", issue.ID, "failed", err, issue, nil)
		writeURLIssueError(c, err)
		return
	}
	h.recordIssueAudit(c, startedAt, "execute", issue.ID, "success", nil, issue, updatedIssue)
	c.JSON(http.StatusOK, gin.H{"data": gin.H{
		"issue":        updatedIssue,
		"check_result": checkResult,
	}})
}

func (h *IssuesHandler) runIssueAction(
	c *gin.Context,
	action string,
	run func(uint) (*urlmanagementdomain.StorefrontURLIssue, error),
) {
	startedAt := urlIssueAuditStartedAt()
	id, err := parseURLIssueID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	before, err := h.issues.Get(id)
	if err != nil {
		writeURLIssueError(c, err)
		return
	}
	issue, err := run(id)
	if err != nil {
		h.recordIssueAudit(c, startedAt, "update", id, "failed", err, before, nil)
		writeURLIssueError(c, err)
		return
	}
	h.recordIssueAudit(c, startedAt, "update", id, "success", nil, before, issue)
	c.JSON(http.StatusOK, gin.H{"data": issue, "action": action})
}

func (h *IssuesHandler) getIssue(c *gin.Context) (*urlmanagementdomain.StorefrontURLIssue, error) {
	id, err := parseURLIssueID(c)
	if err != nil {
		return nil, err
	}
	return h.issues.Get(id)
}

func (h *IssuesHandler) recordIssueAudit(
	c *gin.Context,
	startedAt time.Time,
	action string,
	issueID uint,
	status string,
	err error,
	oldValue interface{},
	newValue interface{},
) {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	recordURLIssueAudit(h.audit, c, urlIssueAuditEvent{
		StartedAt:    startedAt,
		Action:       action,
		Resource:     adminAuditResourceStorefrontURLIssue,
		ResourceID:   issueID,
		Status:       status,
		ErrorMessage: errorMessage,
		OldValue:     oldValue,
		NewValue:     newValue,
	})
}

func parseURLIssueID(c *gin.Context) (uint, error) {
	parsed, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || parsed == 0 {
		return 0, errors.New("invalid URL issue ID")
	}
	return uint(parsed), nil
}

func writeURLIssueError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status = http.StatusNotFound
	}
	c.JSON(status, gin.H{"error": err.Error()})
}
