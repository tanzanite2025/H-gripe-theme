package admin

import (
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const adminAuditResourceGlobalIPBlockRule = "global_ip_block_rule"

func respondIPBlockAuditUnavailable(c *gin.Context, err error) bool {
	if err == nil ||
		(!errors.Is(err, service.ErrIPBlockAuditWrite) &&
			!errors.Is(err, service.ErrAuditRepoUnavailable) &&
			!errors.Is(err, service.ErrAuditLogRequired)) {
		return false
	}
	apierror.RespondError(
		c,
		http.StatusServiceUnavailable,
		"ip_block_audit_unavailable",
		"IP block operation is temporarily unavailable because its audit record could not be saved",
	)
	return true
}

func ipBlockAuditRecorderFactory(recorder adminAuditRecorder) service.IPBlockAuditRecorderFactory {
	if recorder == nil {
		return nil
	}
	if txFactory, ok := recorder.(interface {
		NewIPBlockAuditRecorderForTx(*gorm.DB) service.IPBlockAuditRecorder
	}); ok {
		return txFactory.NewIPBlockAuditRecorderForTx
	}
	return func(_ *gorm.DB) service.IPBlockAuditRecorder {
		return recorder
	}
}

func (h *VisitorProfileHandler) recordVisitorProfileIPBlockAudit(c *gin.Context, event adminAuditEvent) error {
	if h == nil {
		return nil
	}
	return recordAdminAudit(h.auditService, c, event)
}

func (h *GlobalIPBlockHandler) recordGlobalIPBlockAudit(c *gin.Context, event adminAuditEvent) error {
	if h == nil {
		return nil
	}
	return recordAdminAudit(h.auditService, c, event)
}

func visitorIPBlockAuditDetails(
	profileID uint,
	reason string,
	expiresAt interface{},
	rule *service.IPBlockRuleSnapshot,
) map[string]interface{} {
	details := map[string]interface{}{
		"profile_id":     profileID,
		"reason_present": strings.TrimSpace(reason) != "",
		"reason_length":  utf8.RuneCountInString(strings.TrimSpace(reason)),
	}
	if expiresAt != nil {
		details["expires_at"] = expiresAt
	}
	if rule == nil {
		return details
	}
	details["rule_id"] = rule.ID
	details["cidr"] = rule.CIDR
	details["source"] = rule.Source
	details["source_reference"] = rule.SourceReference
	details["status"] = rule.Status
	details["expires"] = rule.ExpiresAt != nil
	if rule.ExpiresAt != nil {
		details["expires_at"] = rule.ExpiresAt.UTC()
	}
	if rule.CreatedBy != nil {
		details["created_by_id"] = *rule.CreatedBy
	}
	if rule.DisabledBy != nil {
		details["disabled_by_id"] = *rule.DisabledBy
	}
	return details
}

func visitorIPBlockAuditDetailsForRules(profileID uint, rules []service.IPBlockRuleSnapshot) map[string]interface{} {
	if len(rules) == 0 {
		return visitorIPBlockAuditDetails(profileID, "", nil, nil)
	}
	if len(rules) == 1 {
		return visitorIPBlockAuditDetails(profileID, "", nil, &rules[0])
	}

	details := map[string]interface{}{
		"profile_id":     profileID,
		"reason_present": false,
		"reason_length":  0,
		"rule_count":     len(rules),
	}
	ruleDetails := make([]map[string]interface{}, 0, len(rules))
	for index := range rules {
		ruleDetails = append(ruleDetails, visitorIPBlockAuditDetails(profileID, "", nil, &rules[index]))
	}
	details["rules"] = ruleDetails
	return details
}

func globalIPBlockAuditDetails(
	cidr string,
	source string,
	sourceReference string,
	reason string,
	expiresAt interface{},
	rule *service.IPBlockRuleSnapshot,
) map[string]interface{} {
	details := map[string]interface{}{
		"cidr":             strings.TrimSpace(cidr),
		"source":           strings.TrimSpace(source),
		"source_reference": strings.TrimSpace(sourceReference),
		"reason_present":   strings.TrimSpace(reason) != "",
		"reason_length":    utf8.RuneCountInString(strings.TrimSpace(reason)),
	}
	if expiresAt != nil {
		details["expires_at"] = expiresAt
	}
	if rule == nil {
		return details
	}
	details["rule_id"] = rule.ID
	details["cidr"] = rule.CIDR
	details["source"] = rule.Source
	details["source_reference"] = rule.SourceReference
	details["status"] = rule.Status
	details["expires"] = rule.ExpiresAt != nil
	if rule.ExpiresAt != nil {
		details["expires_at"] = rule.ExpiresAt.UTC()
	}
	if rule.CreatedBy != nil {
		details["created_by_id"] = *rule.CreatedBy
	}
	if rule.DisabledBy != nil {
		details["disabled_by_id"] = *rule.DisabledBy
	}
	return details
}
