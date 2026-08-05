package admin

import (
	"errors"
	"strconv"
	"strings"
	"tanzanite/internal/domain/ticket"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/locales"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type AutoReplyHandler struct {
	ticketService *service.TicketService
	faqService    *service.FAQService
}

func NewAutoReplyHandler(ticketService *service.TicketService, faqService *service.FAQService) *AutoReplyHandler {
	return &AutoReplyHandler{
		ticketService: ticketService,
		faqService:    faqService,
	}
}

type autoReplyRuleRequest struct {
	Type            string `json:"type"`
	TriggerKeyword  string `json:"trigger_keyword"`
	ReplyMessage    string `json:"reply_message"`
	AgentID         string `json:"agent_id"`
	GroupID         *uint  `json:"group_id"`
	Locale          string `json:"locale"`
	MessageType     string `json:"message_type"`
	Metadata        string `json:"metadata"`
	Attachments     string `json:"attachments"`
	IsActive        *bool  `json:"is_active"`
	Priority        int    `json:"priority"`
	MatchType       string `json:"match_type"`
	CooldownSeconds int    `json:"cooldown_seconds"`
}

func (h *AutoReplyHandler) ListRules(c *gin.Context) {
	rules, err := h.ticketService.ListAutoReplyRules()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"rules": rules})
}

func (h *AutoReplyHandler) GetRule(c *gin.Context) {
	id, ok := parseAutoReplyRuleID(c)
	if !ok {
		return
	}
	rule, err := h.ticketService.GetAutoReplyRule(id)
	if err != nil {
		respondAutoReplyRuleError(c, err)
		return
	}
	response.Success(c, gin.H{"rule": rule})
}

func (h *AutoReplyHandler) CreateRule(c *gin.Context) {
	var req autoReplyRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	rule, err := h.ticketService.CreateAutoReplyRule(req.toInput(nil))
	if err != nil {
		respondAutoReplyRuleError(c, err)
		return
	}
	response.Created(c, gin.H{"rule": rule})
}

func (h *AutoReplyHandler) UpdateRule(c *gin.Context) {
	id, ok := parseAutoReplyRuleID(c)
	if !ok {
		return
	}

	existing, err := h.ticketService.GetAutoReplyRule(id)
	if err != nil {
		respondAutoReplyRuleError(c, err)
		return
	}

	var req autoReplyRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	rule, err := h.ticketService.UpdateAutoReplyRule(id, req.toInput(existing))
	if err != nil {
		respondAutoReplyRuleError(c, err)
		return
	}
	response.Success(c, gin.H{"rule": rule})
}

func (h *AutoReplyHandler) DeleteRule(c *gin.Context) {
	id, ok := parseAutoReplyRuleID(c)
	if !ok {
		return
	}

	if err := h.ticketService.DeleteAutoReplyRule(id); err != nil {
		respondAutoReplyRuleError(c, err)
		return
	}
	response.SuccessWithMessage(c, "Automatic reply rule deleted", nil)
}

func (h *AutoReplyHandler) ListPublishedFAQs(c *gin.Context) {
	if h.faqService == nil {
		apierror.RespondInternalError(c, errors.New("FAQ service is not configured"))
		return
	}

	locale := strings.TrimSpace(c.Query("locale"))
	if locale == "" {
		apierror.RespondBadRequest(c, "FAQ locale is required")
		return
	}
	locale = locales.ResolveSupported(locale)
	if locale == "" {
		apierror.RespondBadRequest(c, "Unsupported FAQ locale")
		return
	}

	pages, total, err := h.faqService.ListAdminGrouped(
		locale,
		c.Query("page_id"),
		c.Query("category"),
		"published",
		c.Query("search"),
	)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{
		"pages": pages,
		"total": total,
	})
}

func (r autoReplyRuleRequest) toInput(existing *ticket.AutoReplyRule) service.AutoReplyRuleInput {
	isActive := true
	if existing != nil {
		isActive = existing.IsActive
	}
	if r.IsActive != nil {
		isActive = *r.IsActive
	}
	return service.AutoReplyRuleInput{
		Type:            r.Type,
		TriggerKeyword:  r.TriggerKeyword,
		ReplyMessage:    r.ReplyMessage,
		AgentID:         r.AgentID,
		GroupID:         r.GroupID,
		Locale:          r.Locale,
		MessageType:     r.MessageType,
		Metadata:        r.Metadata,
		Attachments:     r.Attachments,
		IsActive:        isActive,
		Priority:        r.Priority,
		MatchType:       r.MatchType,
		CooldownSeconds: r.CooldownSeconds,
	}
}

func parseAutoReplyRuleID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 32)
	if err != nil || id == 0 {
		apierror.RespondBadRequest(c, "Invalid automatic-reply rule ID")
		return 0, false
	}
	return uint(id), true
}

func respondAutoReplyRuleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidAutoReplyRule):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrAutoReplyRuleNotFound), service.IsRecordNotFound(err):
		apierror.RespondNotFound(c, "Automatic reply rule")
	default:
		apierror.RespondInternalError(c, err)
	}
}
