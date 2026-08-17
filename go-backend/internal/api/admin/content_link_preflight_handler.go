package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContentLinkPreflightHandler struct {
	service *service.PreflightContentLinkService
}

func NewContentLinkPreflightHandler(
	contentLinkService *service.PreflightContentLinkService,
) *ContentLinkPreflightHandler {
	return &ContentLinkPreflightHandler{service: contentLinkService}
}

func (h *ContentLinkPreflightHandler) ListTargets(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrContentLinkPreflightUnavailable)
		return
	}
	targets, err := h.service.ListTargetOptions()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, targets)
}

func (h *ContentLinkPreflightHandler) Run(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrContentLinkPreflightUnavailable)
		return
	}
	var req struct {
		URL string `json:"url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.RunCheck(c.Request.Context(), service.ContentLinkRunInput{
		TargetURL:   req.URL,
		ActorUserID: c.GetUint("user_id"),
	})
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *ContentLinkPreflightHandler) ListIssues(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrContentLinkPreflightUnavailable)
		return
	}
	page, pageSize := contentLinkPagination(c)
	var fixable *bool
	switch strings.TrimSpace(c.Query("fixable")) {
	case "true", "1", "yes":
		value := true
		fixable = &value
	case "false", "0", "no":
		value := false
		fixable = &value
	}
	issues, total, err := h.service.ListIssues(repository.PreflightContentLinkIssueListFilter{
		Page:      page,
		PageSize:  pageSize,
		State:     strings.TrimSpace(c.DefaultQuery("state", "active")),
		TargetURL: strings.TrimSpace(c.Query("url")),
		Search:    strings.TrimSpace(c.Query("search")),
		Fixable:   fixable,
	})
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	stats, err := h.service.Stats()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{
		"items": issues,
		"stats": stats,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *ContentLinkPreflightHandler) Stats(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrContentLinkPreflightUnavailable)
		return
	}
	stats, err := h.service.Stats()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *ContentLinkPreflightHandler) GetIssue(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrContentLinkPreflightUnavailable)
		return
	}
	id, err := contentLinkIssueID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	issue, err := h.service.GetIssue(id)
	if err != nil {
		respondContentLinkPreflightError(c, err)
		return
	}
	response.Success(c, issue)
}

func (h *ContentLinkPreflightHandler) ListIssueEvents(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrContentLinkPreflightUnavailable)
		return
	}
	id, err := contentLinkIssueID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	page, pageSize := contentLinkPagination(c)
	events, total, err := h.service.ListIssueEvents(id, page, pageSize)
	if err != nil {
		respondContentLinkPreflightError(c, err)
		return
	}
	response.Success(c, gin.H{
		"items": events,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *ContentLinkPreflightHandler) ApplySuggestion(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrContentLinkPreflightUnavailable)
		return
	}
	id, err := contentLinkIssueID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	issue, err := h.service.ApplySuggestion(c.Request.Context(), id, c.GetUint("user_id"))
	if err != nil {
		respondContentLinkPreflightError(c, err)
		return
	}
	response.Success(c, gin.H{"issue": issue})
}

func (h *ContentLinkPreflightHandler) Resolve(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrContentLinkPreflightUnavailable)
		return
	}
	id, err := contentLinkIssueID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	var req struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	issue, err := h.service.ResolveIssue(id, c.GetUint("user_id"), req.Note)
	if err != nil {
		respondContentLinkPreflightError(c, err)
		return
	}
	response.Success(c, gin.H{"issue": issue})
}

func (h *ContentLinkPreflightHandler) Recheck(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrContentLinkPreflightUnavailable)
		return
	}
	id, err := contentLinkIssueID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.RecheckIssue(c.Request.Context(), id, c.GetUint("user_id"))
	if err != nil {
		respondContentLinkPreflightError(c, err)
		return
	}
	response.Success(c, result)
}

func respondContentLinkPreflightError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		apierror.RespondNotFound(c, "content link issue")
	case errors.Is(err, service.ErrContentLinkIssueNotFixable),
		errors.Is(err, service.ErrContentLinkSourceStale):
		apierror.RespondBadRequest(c, err.Error())
	default:
		apierror.RespondError(c, http.StatusBadRequest, "content_link_preflight_error", err.Error())
	}
}

func contentLinkPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	return page, pageSize
}

func contentLinkIssueID(c *gin.Context) (uint, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || value == 0 {
		return 0, errors.New("invalid content link issue ID")
	}
	return uint(value), nil
}
