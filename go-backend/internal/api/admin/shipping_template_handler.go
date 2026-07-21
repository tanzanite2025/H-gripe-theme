package admin

import (
	"errors"
	"strconv"
	"strings"
	shippingdomain "tanzanite/internal/domain/shipping"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *ShippingHandler) ListTemplates(c *gin.Context) {
	templates, err := h.shippingService.ListTemplates()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"data": templates})
}

func (h *ShippingHandler) GetTemplate(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid template id")
	if err != nil {
		return
	}

	template, err := h.shippingService.GetTemplate(id)
	if err != nil {
		apierror.RespondNotFound(c, "Shipping template")
		return
	}

	response.Success(c, template)
}

func (h *ShippingHandler) CreateTemplate(c *gin.Context) {
	var req shippingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	template := req.toDomain()
	if err := validateShippingTemplate(template); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.CreateTemplate(&template); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, template)
}

func (h *ShippingHandler) UpdateTemplate(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid template id")
	if err != nil {
		return
	}

	var req shippingTemplateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	template := req.toDomain()
	template.ID = id
	if err := validateShippingTemplate(template); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.UpdateTemplate(&template); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, template)
}

func (h *ShippingHandler) DeleteTemplate(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid template id")
	if err != nil {
		return
	}

	if err := h.shippingService.DeleteTemplate(id); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "shipping template deleted", nil)
}

func (h *ShippingHandler) CreateTemplateRule(c *gin.Context) {
	templateID, err := parseUintParam(c, "id", "invalid template id")
	if err != nil {
		return
	}

	var req shippingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	rule := req.toDomain()
	if err := validateShippingRule(rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.CreateTemplateRule(templateID, &rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Created(c, rule)
}

func (h *ShippingHandler) UpdateTemplateRule(c *gin.Context) {
	templateID, err := parseUintParam(c, "id", "invalid template id")
	if err != nil {
		return
	}

	ruleID, err := parseUintParam(c, "ruleId", "invalid rule id")
	if err != nil {
		return
	}

	var req shippingRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	rule := req.toDomain()
	rule.ID = ruleID
	if err := validateShippingRule(rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	if err := h.shippingService.UpdateTemplateRule(templateID, &rule); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.Success(c, rule)
}

func (h *ShippingHandler) DeleteTemplateRule(c *gin.Context) {
	templateID, err := parseUintParam(c, "id", "invalid template id")
	if err != nil {
		return
	}

	ruleID, err := parseUintParam(c, "ruleId", "invalid rule id")
	if err != nil {
		return
	}

	if err := h.shippingService.DeleteTemplateRule(templateID, ruleID); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "shipping rule deleted", nil)
}

func parseUintParam(c *gin.Context, name string, message string) (uint, error) {
	id, err := strconv.ParseUint(c.Param(name), 10, 32)
	if err != nil {
		apierror.RespondBadRequest(c, message)
		return 0, err
	}
	return uint(id), nil
}

func validateShippingTemplate(template shippingdomain.ShippingTemplate) error {
	if strings.TrimSpace(template.Name) == "" {
		return errors.New("template name is required")
	}
	switch template.Type {
	case "weight", "quantity", "price":
	default:
		return errors.New("template type must be weight, quantity or price")
	}
	if template.FreeThreshold < 0 || template.DefaultFee < 0 {
		return errors.New("fees and thresholds cannot be negative")
	}
	for _, rule := range template.Rules {
		if err := validateShippingRule(rule); err != nil {
			return err
		}
	}
	return nil
}

func validateShippingRule(rule shippingdomain.ShippingRule) error {
	if rule.Region == "" {
		return errors.New("rule region is required")
	}
	if rule.MinValue < 0 || rule.MaxValue < 0 || rule.Fee < 0 || rule.Additional < 0 {
		return errors.New("rule values and fees cannot be negative")
	}
	if rule.MaxValue > 0 && rule.MaxValue < rule.MinValue {
		return errors.New("rule max value cannot be less than min value")
	}
	return nil
}
