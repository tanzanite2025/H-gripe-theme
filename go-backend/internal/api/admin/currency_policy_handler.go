package admin

import (
	"errors"
	"net/http"

	"tanzanite/internal/domain/currency"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type CurrencyPolicyHandler struct {
	policyService *service.CurrencyPolicyService
	auditService  adminAuditRecorder
}

func NewCurrencyPolicyHandler(policyService *service.CurrencyPolicyService) *CurrencyPolicyHandler {
	return &CurrencyPolicyHandler{policyService: policyService}
}

func (h *CurrencyPolicyHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *CurrencyPolicyHandler) GetPolicy(c *gin.Context) {
	policy, err := h.policyService.GetPolicy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"policy": policy})
}

func (h *CurrencyPolicyHandler) UpdatePolicy(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	var req currency.Policy
	if err := c.ShouldBindJSON(&req); err != nil {
		h.recordCurrencyPolicyAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceCurrencyPolicy,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      currencyPolicyAuditDetails(currency.Policy{}),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if h == nil || h.policyService == nil {
		err := errors.New("currency policy service is not configured")
		h.recordCurrencyPolicyAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceCurrencyPolicy,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      currencyPolicyAuditDetails(req),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	oldPolicy, _ := h.policyService.GetPolicy()

	policy, err := h.policyService.UpdatePolicy(req)
	if err != nil {
		h.recordCurrencyPolicyAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceCurrencyPolicy,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      currencyPolicyAuditDetails(req),
			OldValue:     currencyPolicyOldValue(oldPolicy),
		})
		if errors.Is(err, service.ErrInvalidCurrencyPolicy) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.recordCurrencyPolicyAudit(c, adminAuditEvent{
		StartedAt: startedAt,
		Action:    adminAuditActionUpdate,
		Resource:  adminAuditResourceCurrencyPolicy,
		Status:    adminAuditStatusSuccess,
		Changes:   currencyPolicyAuditDetails(*policy),
		OldValue:  currencyPolicyOldValue(oldPolicy),
		NewValue:  currencyPolicyAuditDetails(*policy),
	})
	c.JSON(http.StatusOK, gin.H{"policy": policy})
}
