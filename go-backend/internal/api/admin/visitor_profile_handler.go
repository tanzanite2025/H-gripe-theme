package admin

import (
	"commerce-platform/internal/domain/audit"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type VisitorProfileHandler struct {
	visitorProfileService *service.VisitorProfileService
	auditService          adminAuditRecorder
}

func NewVisitorProfileHandler(visitorProfileService *service.VisitorProfileService) *VisitorProfileHandler {
	return &VisitorProfileHandler{visitorProfileService: visitorProfileService}
}

func (h *VisitorProfileHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
	if h.visitorProfileService != nil {
		h.visitorProfileService.ConfigureIPBlockAuditRecorderFactory(ipBlockAuditRecorderFactory(recorder))
	}
}

// ListVisitorProfiles returns the read-only visitor profile source used by Public Chat context.
func (h *VisitorProfileHandler) ListVisitorProfiles(c *gin.Context) {
	if h.visitorProfileService == nil {
		apierror.RespondInternalError(c, errors.New("visitor profile service is not configured"))
		return
	}

	params := pagination.ParsePagination(c)
	input := service.VisitorProfileListInput{
		Search:                 strings.TrimSpace(c.Query("search")),
		Identity:               strings.TrimSpace(c.Query("identity")),
		CountryCode:            strings.TrimSpace(c.Query("country_code")),
		Locale:                 strings.TrimSpace(c.Query("locale")),
		Email:                  strings.TrimSpace(c.Query("email")),
		CartSession:            strings.TrimSpace(c.Query("cart_session")),
		CustomerServiceVisitor: strings.TrimSpace(c.Query("customer_service_visitor")),
		LastSeen:               strings.TrimSpace(c.Query("last_seen")),
		LastMeaningful:         strings.TrimSpace(c.Query("last_meaningful")),
		Status:                 strings.TrimSpace(c.Query("status")),
	}

	profiles, total, err := h.visitorProfileService.ListProfilesContext(
		c.Request.Context(),
		params.Page,
		params.PageSize,
		input,
	)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	totalPages := 0
	if params.PageSize > 0 {
		totalPages = (int(total) + params.PageSize - 1) / params.PageSize
	}

	response.Success(c, gin.H{
		"profiles": profiles,
		"pagination": gin.H{
			"page":        params.Page,
			"page_size":   params.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
		"filters": gin.H{
			"search":                   input.Search,
			"identity":                 emptyFilterAsAll(input.Identity),
			"country_code":             input.CountryCode,
			"locale":                   input.Locale,
			"email":                    emptyFilterAsAll(input.Email),
			"cart_session":             emptyFilterAsAll(input.CartSession),
			"customer_service_visitor": emptyFilterAsAll(input.CustomerServiceVisitor),
			"last_seen":                emptyFilterAsAll(input.LastSeen),
			"last_meaningful":          emptyFilterAsAll(input.LastMeaningful),
			"status":                   emptyFilterAsAll(input.Status),
		},
	})
}

// GetVisitorProfileStats returns aggregate capture coverage for the visitor profile fact source.
func (h *VisitorProfileHandler) GetVisitorProfileStats(c *gin.Context) {
	if h.visitorProfileService == nil {
		apierror.RespondInternalError(c, errors.New("visitor profile service is not configured"))
		return
	}

	stats, err := h.visitorProfileService.GetStats()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"stats": stats})
}

func (h *VisitorProfileHandler) BlockVisitorProfileIP(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	profileID, err := parseVisitorProfileID(c)
	if err != nil {
		h.recordVisitorProfileIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorIPBlockAuditDetails(profileID, "", nil, nil),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitor profile id"})
		return
	}
	if h == nil || h.visitorProfileService == nil {
		err := errors.New("visitor profile service is not configured")
		h.recordVisitorProfileIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorIPBlockAuditDetails(profileID, "", nil, nil),
		})
		apierror.RespondInternalError(c, err)
		return
	}

	var req struct {
		Reason    string     `json:"reason"`
		ExpiresAt *time.Time `json:"expires_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordVisitorProfileIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorIPBlockAuditDetails(profileID, req.Reason, req.ExpiresAt, nil),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	adminUserID, ok := currentAdminUserID(c)
	if !ok {
		err := errors.New("admin user id is required")
		h.recordVisitorProfileIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorIPBlockAuditDetails(profileID, req.Reason, req.ExpiresAt, nil),
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	beforeRule, rule, err := h.visitorProfileService.BlockProfileIPWithPreviousAndAudit(
		c.Request.Context(),
		profileID,
		service.IPBlockRuleInput{
			Reason:    req.Reason,
			ExpiresAt: req.ExpiresAt,
		},
		adminUserID,
		func(txBefore *service.IPBlockRuleSnapshot, txRule service.IPBlockRuleSnapshot) (*audit.AuditLog, error) {
			auditAction := adminAuditActionCreate
			var oldValue interface{}
			if txBefore != nil {
				auditAction = adminAuditActionUpdate
				oldValue = visitorIPBlockAuditDetails(
					profileID,
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
				Changes:    visitorIPBlockAuditDetails(profileID, req.Reason, req.ExpiresAt, &txRule),
				OldValue:   oldValue,
				NewValue:   visitorIPBlockAuditDetails(profileID, req.Reason, req.ExpiresAt, &txRule),
			}), nil
		},
	)
	cacheRefreshPending := errors.Is(err, service.ErrIPBlockCacheRefresh) && rule.ID > 0
	if err != nil && !cacheRefreshPending {
		h.recordVisitorProfileIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorIPBlockAuditDetails(profileID, req.Reason, req.ExpiresAt, nil),
		})
		if respondIPBlockAuditUnavailable(c, err) {
			return
		}
		switch {
		case errors.Is(err, service.ErrVisitorProfileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "visitor profile not found"})
		case errors.Is(err, service.ErrVisitorProfileIPUnavailable):
			c.JSON(http.StatusConflict, gin.H{"error": "visitor profile has no retained IP address"})
		case errors.Is(err, service.ErrIPBlockRuleInvalid):
			apierror.RespondBadRequest(c, err.Error())
		default:
			apierror.RespondInternalError(c, err)
		}
		return
	}
	_ = beforeRule
	if cacheRefreshPending {
		c.JSON(http.StatusAccepted, gin.H{
			"code":    0,
			"message": "Visitor IP block rule stored; enforcement cache refresh is pending",
			"data": gin.H{
				"rule":    rule,
				"warning": "The durable rule was saved, but this instance has not completed a cache refresh yet.",
			},
		})
		return
	}
	response.Success(c, gin.H{"rule": rule})
}

func (h *VisitorProfileHandler) UnblockVisitorProfileIP(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	profileID, err := parseVisitorProfileID(c)
	if err != nil {
		h.recordVisitorProfileIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorIPBlockAuditDetails(profileID, "", nil, nil),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitor profile id"})
		return
	}
	if h == nil || h.visitorProfileService == nil {
		err := errors.New("visitor profile service is not configured")
		h.recordVisitorProfileIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorIPBlockAuditDetails(profileID, "", nil, nil),
		})
		apierror.RespondInternalError(c, err)
		return
	}

	adminUserID, ok := currentAdminUserID(c)
	if !ok {
		err := errors.New("admin user id is required")
		h.recordVisitorProfileIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorIPBlockAuditDetails(profileID, "", nil, nil),
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	beforeRules, rules, err := h.visitorProfileService.UnblockProfileIPsWithPreviousAndAudit(
		c.Request.Context(),
		profileID,
		adminUserID,
		func(txBefore, txRules []service.IPBlockRuleSnapshot) (*audit.AuditLog, error) {
			resourceID := uint(0)
			if len(txRules) > 0 {
				resourceID = txRules[0].ID
			}
			return newAdminAuditLog(c, adminAuditEvent{
				StartedAt:  startedAt,
				Action:     adminAuditActionDelete,
				Resource:   adminAuditResourceGlobalIPBlockRule,
				ResourceID: resourceID,
				Status:     adminAuditStatusSuccess,
				Changes:    visitorIPBlockAuditDetailsForRules(profileID, txRules),
				OldValue:   visitorIPBlockAuditDetailsForRules(profileID, txBefore),
				NewValue:   visitorIPBlockAuditDetailsForRules(profileID, txRules),
			}), nil
		},
	)
	var rule service.IPBlockRuleSnapshot
	if len(rules) > 0 {
		rule = rules[0]
	}
	cacheRefreshPending := errors.Is(err, service.ErrIPBlockCacheRefresh) && rule.ID > 0
	if err != nil && !cacheRefreshPending {
		h.recordVisitorProfileIPBlockAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceGlobalIPBlockRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      visitorIPBlockAuditDetails(profileID, "", nil, nil),
		})
		if respondIPBlockAuditUnavailable(c, err) {
			return
		}
		if errors.Is(err, service.ErrIPBlockRuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "active visitor profile IP block not found"})
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	_ = beforeRules
	if cacheRefreshPending {
		c.JSON(http.StatusAccepted, gin.H{
			"code":    0,
			"message": "Visitor IP block rule disabled; enforcement cache refresh is pending",
			"data": gin.H{
				"rule":    rule,
				"warning": "The durable rule was disabled, but this instance has not completed a cache refresh yet.",
			},
		})
		return
	}
	response.Success(c, gin.H{"rule": rule})
}

// CleanupExpiredVisitorProfiles applies the configured retention status fields.
func (h *VisitorProfileHandler) CleanupExpiredVisitorProfiles(c *gin.Context) {
	if h.visitorProfileService == nil {
		apierror.RespondInternalError(c, errors.New("visitor profile service is not configured"))
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

	result, err := h.visitorProfileService.CleanupExpiredProfiles(now)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"cleanup": result})
}

func parseVisitorProfileID(c *gin.Context) (uint, error) {
	if c == nil {
		return 0, errors.New("visitor profile id is required")
	}
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 32)
	if err != nil || id == 0 {
		return 0, errors.New("invalid visitor profile id")
	}
	return uint(id), nil
}

func emptyFilterAsAll(value string) string {
	if strings.TrimSpace(value) == "" {
		return "all"
	}
	return value
}
