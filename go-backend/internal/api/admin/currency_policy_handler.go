package admin

import (
	"errors"
	"net/http"

	"commerce-platform/internal/domain/currency"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type CurrencyPolicyHandler struct {
	policyService  *service.CurrencyPolicyService
	productService *service.ProductService
	auditService   adminAuditRecorder
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

func (h *CurrencyPolicyHandler) ConfigureProductService(productService *service.ProductService) {
	if h == nil {
		return
	}
	h.productService = productService
}

func (h *CurrencyPolicyHandler) GetPolicy(c *gin.Context) {
	policy, err := h.policyService.GetPolicy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"policy": policy})
}

func (h *CurrencyPolicyHandler) GetBackendEntryCurrencyAudit(c *gin.Context) {
	if h == nil || h.policyService == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "currency policy service is not configured"})
		return
	}
	policy, err := h.policyService.GetPolicy()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	audit, err := h.backendEntryCurrencyAudit(policy.PrimaryCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"audit": audit})
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
	payload := gin.H{"policy": policy}
	if audit, err := h.backendEntryCurrencyAudit(policy.PrimaryCurrency); err == nil && audit != nil {
		payload["audit"] = audit
	}
	c.JSON(http.StatusOK, payload)
}

func (h *CurrencyPolicyHandler) backendEntryCurrencyAudit(expectedCurrency string) (*service.BackendEntryCurrencyAudit, error) {
	if h == nil || h.productService == nil {
		return nil, nil
	}
	audit, err := h.productService.AuditBackendEntryCurrencyConsistency(expectedCurrency, 10)
	if err != nil {
		return nil, err
	}
	return &audit, nil
}
