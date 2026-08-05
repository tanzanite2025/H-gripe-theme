package admin

import (
	"strings"
	"time"
	"unicode/utf8"

	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

const adminAuditResourceVisitorRiskDecision = "visitor_risk_decision"

func (h *VisitorRiskHandler) recordVisitorRiskDecisionAudit(c *gin.Context, event adminAuditEvent) {
	if h == nil {
		return
	}
	recordAdminAudit(h.auditService, c, event)
}

func visitorRiskDecisionAuditDetails(
	factID uint,
	action string,
	reason string,
	expiresAt *time.Time,
	decision *service.VisitorRiskDecisionSnapshot,
) map[string]interface{} {
	details := map[string]interface{}{
		"fact_id":          factID,
		"requested_action": strings.ToLower(strings.TrimSpace(action)),
		"reason_present":   strings.TrimSpace(reason) != "",
		"reason_length":    utf8.RuneCountInString(strings.TrimSpace(reason)),
		"expires":          expiresAt != nil,
	}
	if expiresAt != nil {
		details["expires_at"] = expiresAt.UTC()
	}
	if decision == nil {
		return details
	}
	details["decision_id"] = decision.ID
	details["scope"] = strings.TrimSpace(decision.Scope)
	details["action"] = strings.TrimSpace(decision.Action)
	details["reason_present"] = strings.TrimSpace(decision.Reason) != ""
	details["reason_length"] = utf8.RuneCountInString(strings.TrimSpace(decision.Reason))
	details["expires"] = decision.ExpiresAt != nil
	if decision.ExpiresAt != nil {
		details["expires_at"] = decision.ExpiresAt.UTC()
	}
	if decision.CreatedBy != nil {
		details["created_by_id"] = *decision.CreatedBy
	}
	return details
}
