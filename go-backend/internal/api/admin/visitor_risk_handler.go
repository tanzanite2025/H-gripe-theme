package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type VisitorRiskHandler struct {
	visitorRiskService *service.VisitorRiskService
	auditService       adminAuditRecorder
}

func NewVisitorRiskHandler(visitorRiskService *service.VisitorRiskService) *VisitorRiskHandler {
	return &VisitorRiskHandler{visitorRiskService: visitorRiskService}
}

func (h *VisitorRiskHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *VisitorRiskHandler) ListVisitorRiskFacts(c *gin.Context) {
	if h.visitorRiskService == nil {
		apierror.RespondInternalError(c, errors.New("visitor risk service is not configured"))
		return
	}

	params := pagination.ParsePagination(c)
	input := service.VisitorRiskFactListInput{
		Search:       strings.TrimSpace(c.Query("search")),
		RiskLevel:    strings.TrimSpace(c.Query("risk_level")),
		DayRange:     strings.TrimSpace(c.Query("day_range")),
		MinRiskScore: strings.TrimSpace(c.Query("min_risk_score")),
	}

	facts, total, err := h.visitorRiskService.ListFacts(params.Page, params.PageSize, input)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	totalPages := 0
	if params.PageSize > 0 {
		totalPages = (int(total) + params.PageSize - 1) / params.PageSize
	}

	response.Success(c, gin.H{
		"facts": facts,
		"pagination": gin.H{
			"page":        params.Page,
			"page_size":   params.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
		"filters": gin.H{
			"search":         input.Search,
			"risk_level":     emptyFilterAsAll(input.RiskLevel),
			"day_range":      emptyFilterAsAll(input.DayRange),
			"min_risk_score": input.MinRiskScore,
		},
	})
}

func (h *VisitorRiskHandler) GetVisitorRiskStats(c *gin.Context) {
	if h.visitorRiskService == nil {
		apierror.RespondInternalError(c, errors.New("visitor risk service is not configured"))
		return
	}

	stats, err := h.visitorRiskService.GetStats(strings.TrimSpace(c.Query("day_range")))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"stats": stats})
}

func (h *VisitorRiskHandler) CleanupExpiredVisitorRiskFacts(c *gin.Context) {
	if h.visitorRiskService == nil {
		apierror.RespondInternalError(c, errors.New("visitor risk service is not configured"))
		return
	}

	now := time.Now().UTC()
	if rawNow := strings.TrimSpace(c.Query("now")); rawNow != "" {
		parsed, err := time.Parse(time.RFC3339, rawNow)
		if err != nil {
			apierror.RespondBadRequest(c, "invalid cleanup reference timestamp")
			return
		}
		now = parsed
	}

	result, err := h.visitorRiskService.CleanupExpiredFacts(now)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"cleanup": result})
}

func (h *VisitorRiskHandler) CreateVisitorRiskDecision(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.visitorRiskService == nil {
		err := errors.New("visitor risk service is not configured")
		h.recordVisitorRiskDecisionAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceVisitorRiskDecision,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorRiskDecisionAuditDetails(0, "", "", nil, nil),
		})
		apierror.RespondInternalError(c, err)
		return
	}

	factID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 32)
	if err != nil || factID == 0 {
		h.recordVisitorRiskDecisionAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceVisitorRiskDecision,
			Status:       adminAuditStatusFailed,
			ErrorMessage: "invalid visitor risk fact id",
			Changes: map[string]interface{}{
				"raw_fact_id": strings.TrimSpace(c.Param("id")),
			},
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitor risk fact id"})
		return
	}
	factIDUint := uint(factID)

	var req struct {
		Action    string     `json:"action"`
		Reason    string     `json:"reason"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordVisitorRiskDecisionAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceVisitorRiskDecision,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorRiskDecisionAuditDetails(factIDUint, "", "", nil, nil),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminUserID, ok := currentAdminUserID(c)
	if !ok {
		h.recordVisitorRiskDecisionAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceVisitorRiskDecision,
			Status:       adminAuditStatusFailed,
			ErrorMessage: "admin user id is required",
			Changes:      visitorRiskDecisionAuditDetails(factIDUint, req.Action, req.Reason, req.ExpiresAt, nil),
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	decision, err := h.visitorRiskService.CreateDecision(
		factIDUint,
		service.VisitorRiskDecisionInput{
			Action:    req.Action,
			Reason:    req.Reason,
			ExpiresAt: req.ExpiresAt,
		},
		adminUserID,
	)
	if err != nil {
		h.recordVisitorRiskDecisionAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceVisitorRiskDecision,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorRiskDecisionAuditDetails(factIDUint, req.Action, req.Reason, req.ExpiresAt, nil),
		})
		switch {
		case errors.Is(err, service.ErrVisitorRiskFactNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "visitor risk fact not found"})
		case errors.Is(err, service.ErrVisitorRiskDecisionInvalid),
			errors.Is(err, service.ErrVisitorRiskDecisionNoIdentity):
			apierror.RespondBadRequest(c, err.Error())
		default:
			apierror.RespondInternalError(c, err)
		}
		return
	}

	h.recordVisitorRiskDecisionAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionCreate,
		Resource:   adminAuditResourceVisitorRiskDecision,
		ResourceID: decision.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    visitorRiskDecisionAuditDetails(factIDUint, req.Action, req.Reason, req.ExpiresAt, &decision),
		NewValue:   visitorRiskDecisionAuditDetails(factIDUint, req.Action, req.Reason, req.ExpiresAt, &decision),
	})

	response.Success(c, gin.H{"decision": decision})
}

func (h *VisitorRiskHandler) GetVisitorRiskDecision(c *gin.Context) {
	if h.visitorRiskService == nil {
		apierror.RespondInternalError(c, errors.New("visitor risk service is not configured"))
		return
	}

	factID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 32)
	if err != nil || factID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitor risk fact id"})
		return
	}

	decision, err := h.visitorRiskService.GetDecision(uint(factID))
	if err != nil {
		if errors.Is(err, service.ErrVisitorRiskFactNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "visitor risk fact not found"})
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"decision": decision})
}
