package admin

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/domain/security"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type GlobalIPBlockHandler struct {
	globalIPBlockService *service.GlobalIPBlockService
	auditService         adminAuditRecorder
}

func NewGlobalIPBlockHandler(globalIPBlockService *service.GlobalIPBlockService) *GlobalIPBlockHandler {
	return &GlobalIPBlockHandler{globalIPBlockService: globalIPBlockService}
}

func (h *GlobalIPBlockHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
	if h.globalIPBlockService != nil {
		h.globalIPBlockService.ConfigureAuditRecorderFactory(ipBlockAuditRecorderFactory(recorder))
	}
}

func (h *GlobalIPBlockHandler) List(c *gin.Context) {
	if h == nil || h.globalIPBlockService == nil {
		apierror.RespondInternalError(c, errors.New("global IP block service is not configured"))
		return
	}

	params := pagination.ParsePagination(c)
	input := service.IPBlockRuleListInput{
		Search: strings.TrimSpace(c.Query("search")),
		Source: strings.TrimSpace(c.Query("source")),
		Status: strings.TrimSpace(c.Query("status")),
	}
	rules, total, err := h.globalIPBlockService.List(c.Request.Context(), params.Page, params.PageSize, input)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	totalPages := 0
	if params.PageSize > 0 {
		totalPages = (int(total) + params.PageSize - 1) / params.PageSize
	}
	response.Success(c, gin.H{
		"rules": rules,
		"pagination": gin.H{
			"page":        params.Page,
			"page_size":   params.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func (h *GlobalIPBlockHandler) Create(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.globalIPBlockService == nil {
		err := errors.New("global IP block service is not configured")
		if h != nil {
			h.recordGlobalIPBlockAudit(c, adminAuditEvent{
				StartedAt:    startedAt,
				Action:       adminAuditActionCreate,
				Resource:     adminAuditResourceGlobalIPBlockRule,
				Status:       adminAuditStatusFailed,
				ErrorMessage: err.Error(),
			})
		}
		apierror.RespondInternalError(c, err)
		return
	}

	var req struct {
		CIDR      string     `json:"cidr"`
		IP        string     `json:"ip"`
		Reason    string     `json:"reason"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordGlobalIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      globalIPBlockAuditDetails(req.CIDR, security.IPBlockRuleSourceManual, "admin", req.Reason, req.ExpiresAt, nil),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminUserID, ok := currentAdminUserID(c)
	if !ok {
		err := errors.New("admin user id is required")
		h.recordGlobalIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      globalIPBlockAuditDetails(req.CIDR, security.IPBlockRuleSourceManual, "admin", req.Reason, req.ExpiresAt, nil),
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	cidr := strings.TrimSpace(req.CIDR)
	if cidr == "" {
		cidr = strings.TrimSpace(req.IP)
	}
	beforeRule, rule, err := h.globalIPBlockService.BlockWithPreviousAndAudit(
		c.Request.Context(),
		service.IPBlockRuleInput{
			CIDR:            cidr,
			Source:          security.IPBlockRuleSourceManual,
			SourceReference: "admin",
			Reason:          req.Reason,
			ExpiresAt:       req.ExpiresAt,
			CreatedBy:       adminUserID,
		},
		func(txBefore *service.IPBlockRuleSnapshot, txRule service.IPBlockRuleSnapshot) (*audit.AuditLog, error) {
			auditAction := adminAuditActionCreate
			var oldValue interface{}
			if txBefore != nil {
				auditAction = adminAuditActionUpdate
				oldValue = globalIPBlockAuditDetails(
					txBefore.CIDR,
					txBefore.Source,
					txBefore.SourceReference,
					txBefore.Reason,
					txBefore.ExpiresAt,
					txBefore,
				)
			}
			return newAdminAuditLog(c, adminAuditEvent{
				StartedAt:  startedAt,
				Action:     auditAction,
				Resource:   adminAuditResourceGlobalIPBlockRule,
				ResourceID: txRule.ID,
				Status:     adminAuditStatusSuccess,
				Changes:    globalIPBlockAuditDetails(cidr, security.IPBlockRuleSourceManual, "admin", req.Reason, req.ExpiresAt, &txRule),
				OldValue:   oldValue,
				NewValue:   globalIPBlockAuditDetails(cidr, security.IPBlockRuleSourceManual, "admin", req.Reason, req.ExpiresAt, &txRule),
			}), nil
		},
	)
	cacheRefreshPending := errors.Is(err, service.ErrIPBlockCacheRefresh) && rule.ID > 0
	if err != nil && !cacheRefreshPending {
		h.recordGlobalIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      globalIPBlockAuditDetails(cidr, security.IPBlockRuleSourceManual, "admin", req.Reason, req.ExpiresAt, nil),
		})
		if respondIPBlockAuditUnavailable(c, err) {
			return
		}
		if errors.Is(err, service.ErrIPBlockRuleInvalid) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	_ = beforeRule
	if cacheRefreshPending {
		c.JSON(http.StatusAccepted, gin.H{
			"code":    0,
			"message": "IP block rule stored; enforcement cache refresh is pending",
			"data": gin.H{
				"rule":    rule,
				"warning": "The durable rule was saved, but this instance has not completed a cache refresh yet.",
			},
		})
		return
	}
	response.Success(c, gin.H{"rule": rule})
}

func (h *GlobalIPBlockHandler) Disable(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 32)
	if err != nil || id == 0 {
		h.recordGlobalIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: "invalid IP block rule id",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid IP block rule id"})
		return
	}
	if h == nil || h.globalIPBlockService == nil {
		err := errors.New("global IP block service is not configured")
		h.recordGlobalIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			ResourceID:   uint(id),
		})
		apierror.RespondInternalError(c, err)
		return
	}

	adminUserID, ok := currentAdminUserID(c)
	if !ok {
		err := errors.New("admin user id is required")
		h.recordGlobalIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			ResourceID:   uint(id),
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	beforeRule, rule, err := h.globalIPBlockService.DisableWithPreviousAndAudit(
		c.Request.Context(),
		uint(id),
		adminUserID,
		func(txBefore *service.IPBlockRuleSnapshot, txRule service.IPBlockRuleSnapshot) (*audit.AuditLog, error) {
			var oldValue interface{}
			if txBefore != nil {
				oldValue = globalIPBlockAuditDetails(
					txBefore.CIDR,
					txBefore.Source,
					txBefore.SourceReference,
					txBefore.Reason,
					txBefore.ExpiresAt,
					txBefore,
				)
			}
			return newAdminAuditLog(c, adminAuditEvent{
				StartedAt:  startedAt,
				Action:     adminAuditActionDelete,
				Resource:   adminAuditResourceGlobalIPBlockRule,
				ResourceID: txRule.ID,
				Status:     adminAuditStatusSuccess,
				Changes:    globalIPBlockAuditDetails(txRule.CIDR, txRule.Source, txRule.SourceReference, txRule.Reason, txRule.ExpiresAt, &txRule),
				OldValue:   oldValue,
				NewValue:   globalIPBlockAuditDetails(txRule.CIDR, txRule.Source, txRule.SourceReference, txRule.Reason, txRule.ExpiresAt, &txRule),
			}), nil
		},
	)
	cacheRefreshPending := errors.Is(err, service.ErrIPBlockCacheRefresh) && rule.ID > 0
	if err != nil && !cacheRefreshPending {
		h.recordGlobalIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			ResourceID:   uint(id),
		})
		if respondIPBlockAuditUnavailable(c, err) {
			return
		}
		if errors.Is(err, service.ErrIPBlockRuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "IP block rule not found"})
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	_ = beforeRule
	if cacheRefreshPending {
		c.JSON(http.StatusAccepted, gin.H{
			"code":    0,
			"message": "IP block rule disabled; enforcement cache refresh is pending",
			"data": gin.H{
				"rule":    rule,
				"warning": "The durable rule was disabled, but this instance has not completed a cache refresh yet.",
			},
		})
		return
	}
	response.Success(c, gin.H{"rule": rule})
}
