package admin

import (
	"errors"
	"net/http"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type PublicChatAgentHandler struct {
	publicChatAgentService *service.AdminPublicChatAgentService
}

func NewPublicChatAgentHandler(publicChatAgentService *service.AdminPublicChatAgentService) *PublicChatAgentHandler {
	return &PublicChatAgentHandler{publicChatAgentService: publicChatAgentService}
}

type publicChatAgentUpsertRequest struct {
	UserID       uint   `json:"user_id" binding:"required"`
	AgentID      string `json:"agent_id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Avatar       string `json:"avatar"`
	WhatsApp     string `json:"whatsapp"`
	Status       string `json:"status"`
	OnlineStatus string `json:"online_status"`
}

func (h *PublicChatAgentHandler) ListPublicChatAgents(c *gin.Context) {
	overview, err := h.publicChatAgentService.ListPublicChatAgents(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch public chat agents"})
		return
	}

	c.JSON(http.StatusOK, overview)
}

func (h *PublicChatAgentHandler) ListPublicChatAgentCandidates(c *gin.Context) {
	candidates, err := h.publicChatAgentService.ListPublicChatAgentCandidates(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch public chat agent candidates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"candidates": candidates})
}

func (h *PublicChatAgentHandler) UpsertPublicChatAgent(c *gin.Context) {
	var req publicChatAgentUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	agent, created, err := h.publicChatAgentService.UpsertPublicChatAgentProfile(service.AdminPublicChatAgentUpsertInput{
		UserID:       req.UserID,
		AgentID:      req.AgentID,
		Name:         req.Name,
		Email:        req.Email,
		Avatar:       req.Avatar,
		WhatsApp:     req.WhatsApp,
		Status:       req.Status,
		OnlineStatus: req.OnlineStatus,
	})
	if err != nil {
		respondPublicChatAgentError(c, err)
		return
	}

	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, gin.H{
		"agent":   agent,
		"created": created,
	})
}

func respondPublicChatAgentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPublicChatAgentUserRequired),
		errors.Is(err, service.ErrPublicChatAgentUserNotFound),
		errors.Is(err, service.ErrPublicChatAgentUserInvalid),
		errors.Is(err, service.ErrPublicChatAgentIDInvalid),
		errors.Is(err, service.ErrPublicChatAgentIDTaken),
		errors.Is(err, service.ErrPublicChatAgentStatusInvalid),
		errors.Is(err, service.ErrPublicChatAgentOnlineInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save public chat agent"})
	}
}
