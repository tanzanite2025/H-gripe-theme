package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	sitequalitydomain "commerce-platform/internal/domain/sitequality"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SiteQualityHandler struct {
	siteQualityService *service.LighthouseRunnerService
	siteQualityEngine  *service.SiteQualityEngineService
	auditService       adminAuditRecorder
}

const (
	adminAuditResourceSiteQualityRun     = "preflight_site_quality_run"
	adminAuditResourceSiteQualityFinding = "preflight_site_quality_finding"
)

func NewSiteQualityHandler(
	siteQualityService *service.LighthouseRunnerService,
	siteQualityEngine *service.SiteQualityEngineService,
) *SiteQualityHandler {
	return &SiteQualityHandler{
		siteQualityService: siteQualityService,
		siteQualityEngine:  siteQualityEngine,
	}
}

func (h *SiteQualityHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}
func (h *SiteQualityHandler) ListSiteQualityRuns(c *gin.Context) {
	if h == nil || h.siteQualityService == nil {
		apierror.RespondInternalError(c, errors.New("Lighthouse runner service is not configured"))
		return
	}
	page, pageSize := siteQualityPagination(c)
	result, err := h.siteQualityService.List(repository.SiteQualityRunListFilter{
		Page:      page,
		PageSize:  pageSize,
		TargetURL: strings.TrimSpace(c.Query("url")),
		Strategy:  strings.TrimSpace(c.Query("strategy")),
	})
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	summary := h.buildSiteQualityOperationalSummary(result)
	response.Success(c, gin.H{
		"runner_configured": result.RunnerConfigured,
		"default_url":       result.DefaultURL,
		"summary":           summary,
		"items":             result.Items,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       result.Total,
			"total_pages": (int(result.Total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *SiteQualityHandler) ListSiteQualityTargets(c *gin.Context) {
	if h == nil || h.siteQualityEngine == nil {
		apierror.RespondInternalError(c, errors.New("site quality engine is not configured"))
		return
	}
	targets, err := h.siteQualityEngine.ListTargetOptions()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, targets)
}

func (h *SiteQualityHandler) CreateSiteQualityJob(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.siteQualityEngine == nil {
		err := errors.New("site quality engine is not configured")
		h.recordSiteQualityRunAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceSiteQualityRun,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	var req struct {
		URL      string `json:"url"`
		Strategy string `json:"strategy"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordSiteQualityRunAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceSiteQualityRun,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	job, err := h.siteQualityEngine.EnqueueManualTarget(
		c.Request.Context(),
		req.URL,
		req.Strategy,
		c.GetUint("user_id"),
		sitequalitydomain.SiteQualityJobKindManual,
	)
	if err != nil {
		h.recordSiteQualityRunAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceSiteQualityRun,
			ResourceID:   0,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      gin.H{"url": strings.TrimSpace(req.URL), "strategy": strings.TrimSpace(req.Strategy)},
			NewValue:     job,
		})
		if errors.Is(err, service.ErrInvalidSiteQualityRun) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondError(c, http.StatusBadGateway, "site_quality_runner_failed", err.Error())
		return
	}
	h.recordSiteQualityRunAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionExecute,
		Resource:   adminAuditResourceSiteQualityRun,
		ResourceID: job.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    gin.H{"url": strings.TrimSpace(req.URL), "strategy": strings.TrimSpace(req.Strategy), "job_id": job.ID},
		NewValue:   job,
	})
	c.JSON(http.StatusAccepted, response.Response{
		Code: 0,
		Data: gin.H{
			"job_id": job.ID,
			"job":    job,
		},
	})
}

func (h *SiteQualityHandler) GetSiteQualityJob(c *gin.Context) {
	if h == nil || h.siteQualityEngine == nil {
		apierror.RespondInternalError(c, errors.New("site quality engine is not configured"))
		return
	}
	id, err := siteQualityJobID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	job, err := h.siteQualityEngine.GetJob(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			apierror.RespondNotFound(c, "site quality job was not found")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, job)
}

func (h *SiteQualityHandler) ListSiteQualityFindings(c *gin.Context) {
	if h == nil || h.siteQualityService == nil {
		apierror.RespondInternalError(c, errors.New("Lighthouse runner service is not configured"))
		return
	}
	page, pageSize := siteQualityPagination(c)
	findings, total, err := h.siteQualityService.ListFindings(repository.SiteQualityFindingListFilter{
		Page:        page,
		PageSize:    pageSize,
		State:       strings.TrimSpace(c.DefaultQuery("state", "active")),
		Severity:    strings.TrimSpace(c.Query("severity")),
		TargetURL:   strings.TrimSpace(c.Query("url")),
		Strategy:    strings.TrimSpace(c.Query("strategy")),
		FindingKind: strings.TrimSpace(c.Query("kind")),
	})
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	stats, err := h.siteQualityService.FindingStats()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{
		"items": findings,
		"stats": stats,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (int(total) + pageSize - 1) / pageSize,
		},
	})
}

func (h *SiteQualityHandler) GetSiteQualityFinding(c *gin.Context) {
	if h == nil || h.siteQualityService == nil {
		apierror.RespondInternalError(c, errors.New("Lighthouse runner service is not configured"))
		return
	}
	id, err := siteQualityFindingID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	finding, err := h.siteQualityService.GetFinding(id)
	if err != nil {
		h.respondSiteQualityFindingError(c, err)
		return
	}
	response.Success(c, finding)
}

func (h *SiteQualityHandler) ListSiteQualityFindingEvents(c *gin.Context) {
	if h == nil || h.siteQualityService == nil {
		apierror.RespondInternalError(c, errors.New("Lighthouse runner service is not configured"))
		return
	}
	id, err := siteQualityFindingID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	page, pageSize := siteQualityPagination(c)
	events, total, err := h.siteQualityService.ListFindingEvents(id, page, pageSize)
	if err != nil {
		h.respondSiteQualityFindingError(c, err)
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

func (h *SiteQualityHandler) AcknowledgeSiteQualityFinding(c *gin.Context) {
	var req struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	h.runSiteQualityFindingAction(c, "acknowledge", func(id uint) (interface{}, error) {
		return h.siteQualityService.AcknowledgeFinding(id, c.GetUint("user_id"), sitequalitydomain.SiteQualityFindingActionInput{
			Note: req.Note,
		})
	})
}

func (h *SiteQualityHandler) ResolveSiteQualityFinding(c *gin.Context) {
	var req struct {
		ResolutionNote string `json:"resolution_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	h.runSiteQualityFindingAction(c, "resolve", func(id uint) (interface{}, error) {
		return h.siteQualityService.ResolveFinding(id, c.GetUint("user_id"), sitequalitydomain.SiteQualityFindingResolutionInput{
			ResolutionNote: req.ResolutionNote,
		})
	})
}

func (h *SiteQualityHandler) RecheckSiteQualityFinding(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.siteQualityService == nil || h.siteQualityEngine == nil {
		err := errors.New("site quality engine is not configured")
		h.recordSiteQualityFindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceSiteQualityFinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	id, err := siteQualityFindingID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	before, err := h.siteQualityService.GetFinding(id)
	if err != nil {
		h.respondSiteQualityFindingError(c, err)
		return
	}
	job, err := h.siteQualityEngine.EnqueueRecheckFinding(before, c.GetUint("user_id"))
	if err != nil {
		h.recordSiteQualityFindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceSiteQualityFinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			OldValue:     before,
			NewValue:     job,
		})
		if errors.Is(err, service.ErrInvalidSiteQualityRun) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondError(c, http.StatusBadGateway, "site_quality_runner_failed", err.Error())
		return
	}
	h.recordSiteQualityFindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionExecute,
		Resource:   adminAuditResourceSiteQualityFinding,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		OldValue:   before,
		NewValue:   job,
		Changes:    gin.H{"job_id": job.ID, "finding_id": id},
	})
	c.JSON(http.StatusAccepted, response.Response{
		Code: 0,
		Data: gin.H{
			"job_id":     job.ID,
			"finding_id": id,
			"job":        job,
		},
	})
}

func (h *SiteQualityHandler) recordSiteQualityRunAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}

func (h *SiteQualityHandler) recordSiteQualityFindingAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}

func (h *SiteQualityHandler) runSiteQualityFindingAction(
	c *gin.Context,
	action string,
	run func(uint) (interface{}, error),
) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.siteQualityService == nil {
		err := errors.New("Lighthouse runner service is not configured")
		h.recordSiteQualityFindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceSiteQualityFinding,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
		})
		apierror.RespondInternalError(c, err)
		return
	}
	id, err := siteQualityFindingID(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	before, err := h.siteQualityService.GetFinding(id)
	if err != nil {
		h.respondSiteQualityFindingError(c, err)
		return
	}
	result, err := run(id)
	if err != nil {
		h.recordSiteQualityFindingAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceSiteQualityFinding,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			OldValue:     before,
		})
		h.respondSiteQualityFindingError(c, err)
		return
	}
	h.recordSiteQualityFindingAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceSiteQualityFinding,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		OldValue:   before,
		NewValue:   result,
		Changes:    gin.H{"action": action},
	})
	response.Success(c, gin.H{"finding": result, "action": action})
}

func (h *SiteQualityHandler) respondSiteQualityFindingError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		apierror.RespondNotFound(c, "Site Quality finding")
		return
	}
	apierror.RespondBadRequest(c, err.Error())
}

func (h *SiteQualityHandler) buildSiteQualityOperationalSummary(result *service.LighthouseRunnerListResult) *service.SiteQualityOperationalSummary {
	summary := &service.SiteQualityOperationalSummary{
		GeneratedAt:      time.Now().UTC(),
		Status:           "not_configured",
		RunnerConfigured: result != nil && result.RunnerConfigured,
		DefaultURL:       "",
		RunCount:         0,
	}
	if result != nil {
		summary.DefaultURL = result.DefaultURL
		summary.RunCount = result.Total
		if len(result.Items) > 0 {
			latest := result.Items[0]
			summary.LatestRun = &latest
			if latest.Status != sitequalitydomain.SiteQualityRunStatusSuccess {
				summary.Warnings = append(summary.Warnings, "latest Lighthouse sample failed")
			}
		}
		if h != nil && h.siteQualityService != nil {
			latestSuccessAt, err := h.siteQualityService.LatestSuccessfulAt()
			if err != nil {
				summary.Warnings = append(summary.Warnings, err.Error())
			} else {
				summary.LatestSuccessAt = latestSuccessAt
			}
		}
	}
	if h == nil || h.siteQualityEngine == nil {
		if summary.RunnerConfigured {
			summary.Status = "healthy"
		}
		if len(summary.Warnings) > 0 {
			summary.Status = "degraded"
		}
		return summary
	}
	engineSummary, err := h.siteQualityEngine.Summary(summary.GeneratedAt)
	if err != nil || engineSummary == nil {
		if err != nil {
			summary.Warnings = append(summary.Warnings, err.Error())
		} else {
			summary.Warnings = append(summary.Warnings, "site quality operational summary is unavailable")
		}
		if summary.RunnerConfigured {
			summary.Status = "degraded"
		}
		return summary
	}
	return engineSummary
}

func (h *SiteQualityHandler) latestSiteQualityRunForFinding(
	finding *sitequalitydomain.SiteQualityFinding,
) *service.LighthouseRunnerRunView {
	if h == nil || h.siteQualityService == nil || finding == nil {
		return nil
	}
	result, err := h.siteQualityService.List(repository.SiteQualityRunListFilter{
		Page:      1,
		PageSize:  1,
		TargetURL: finding.TargetURL,
		Strategy:  finding.Strategy,
	})
	if err != nil || result == nil || len(result.Items) == 0 {
		return nil
	}
	return &result.Items[0]
}

func siteQualityRunAuditID(run *service.LighthouseRunnerRunView) uint {
	if run == nil {
		return 0
	}
	return run.ID
}

func siteQualityPagination(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func siteQualityFindingID(c *gin.Context) (uint, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || value == 0 {
		return 0, errors.New("invalid SiteQuality finding ID")
	}
	return uint(value), nil
}

func siteQualityJobID(c *gin.Context) (uint, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || value == 0 {
		return 0, errors.New("invalid SiteQuality job ID")
	}
	return uint(value), nil
}
