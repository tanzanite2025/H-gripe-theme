package admin

import (
	"errors"
	"net/http"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type TaxRateRuleHandler struct {
	taxRateService *service.TaxRateService
	auditService   adminAuditRecorder
}

func NewTaxRateRuleHandler(taxRateService *service.TaxRateService) *TaxRateRuleHandler {
	return &TaxRateRuleHandler{taxRateService: taxRateService}
}

func (h *TaxRateRuleHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *TaxRateRuleHandler) ListRules(c *gin.Context) {
	if h == nil || h.taxRateService == nil {
		apierror.RespondInternalError(c, errors.New("tax rate service is not configured"))
		return
	}
	rules, err := h.taxRateService.ListTaxRates()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"rules": rules})
}

func (h *TaxRateRuleHandler) CreateRule(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	var input service.TaxRateRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.recordTaxRateRuleAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceTaxRateRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      input,
		})
		apierror.RespondBadRequest(c, "invalid tax rate rule")
		return
	}
	if h == nil || h.taxRateService == nil {
		err := errors.New("tax rate service is not configured")
		h.recordTaxRateRuleAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceTaxRateRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      input,
		})
		apierror.RespondInternalError(c, err)
		return
	}
	rule, err := h.taxRateService.CreateTaxRateRule(input)
	if err != nil {
		h.recordTaxRateRuleAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceTaxRateRule,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      input,
		})
		respondTaxRateRuleError(c, err)
		return
	}
	h.recordTaxRateRuleAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionCreate,
		Resource:   adminAuditResourceTaxRateRule,
		ResourceID: rule.ID,
		Status:     adminAuditStatusSuccess,
		Changes:    input,
		NewValue:   rule,
	})
	response.Created(c, rule)
}

func (h *TaxRateRuleHandler) UpdateRule(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.taxRateService == nil {
		apierror.RespondInternalError(c, errors.New("tax rate service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid tax rate rule id")
	if err != nil {
		return
	}
	oldRule, err := h.taxRateService.GetTaxRate(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Tax rate rule")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	var input service.TaxRateRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		h.recordTaxRateRuleAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceTaxRateRule,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      input,
			OldValue:     oldRule,
		})
		apierror.RespondBadRequest(c, "invalid tax rate rule")
		return
	}
	updatedRule, err := h.taxRateService.UpdateTaxRateRule(id, input)
	if err != nil {
		h.recordTaxRateRuleAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceTaxRateRule,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      input,
			OldValue:     oldRule,
		})
		respondTaxRateRuleError(c, err)
		return
	}
	h.recordTaxRateRuleAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceTaxRateRule,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		Changes:    input,
		OldValue:   oldRule,
		NewValue:   updatedRule,
	})
	response.Success(c, updatedRule)
}

func (h *TaxRateRuleHandler) DeleteRule(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.taxRateService == nil {
		apierror.RespondInternalError(c, errors.New("tax rate service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid tax rate rule id")
	if err != nil {
		return
	}
	oldRule, err := h.taxRateService.GetTaxRate(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Tax rate rule")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	if err := h.taxRateService.DeleteTaxRateRule(id); err != nil {
		h.recordTaxRateRuleAudit(c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceTaxRateRule,
			ResourceID:   id,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			OldValue:     oldRule,
		})
		respondTaxRateRuleError(c, err)
		return
	}
	h.recordTaxRateRuleAudit(c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionDelete,
		Resource:   adminAuditResourceTaxRateRule,
		ResourceID: id,
		Status:     adminAuditStatusSuccess,
		OldValue:   oldRule,
	})
	response.Success(c, gin.H{"deleted": true})
}

func respondTaxRateRuleError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrTaxRateRuleInvalid) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if repository.IsRecordNotFound(err) {
		apierror.RespondNotFound(c, "Tax rate rule")
		return
	}
	apierror.RespondError(c, http.StatusInternalServerError, "tax_rate_rule_failed", err.Error())
}
