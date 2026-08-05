package admin

import (
	"errors"
	"net/http"
	"strconv"
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
	GroupIDs     []uint `json:"group_ids"`
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
		GroupIDs:     req.GroupIDs,
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

type publicChatGroupUpsertRequest struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	SortOrder   int    `json:"sort_order"`
}

func (h *PublicChatAgentHandler) ListPublicChatGroups(c *gin.Context) {
	groups, err := h.publicChatAgentService.ListPublicChatGroups(500)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch public chat groups"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"groups": groups})
}

func (h *PublicChatAgentHandler) UpsertPublicChatGroup(c *gin.Context) {
	var req publicChatGroupUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group, created, err := h.publicChatAgentService.UpsertPublicChatGroup(service.AdminPublicChatGroupUpsertInput{
		ID:          req.ID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		respondPublicChatGroupError(c, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, gin.H{"group": group, "created": created})
}

func (h *PublicChatAgentHandler) UpdatePublicChatGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public chat group id"})
		return
	}

	var req publicChatGroupUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.ID = uint(id)

	group, _, err := h.publicChatAgentService.UpsertPublicChatGroup(service.AdminPublicChatGroupUpsertInput{
		ID:          req.ID,
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		SortOrder:   req.SortOrder,
	})
	if err != nil {
		respondPublicChatGroupError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"group": group, "created": false})
}

func (h *PublicChatAgentHandler) DeletePublicChatGroup(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid public chat group id"})
		return
	}
	if err := h.publicChatAgentService.DeletePublicChatGroup(uint(id)); err != nil {
		respondPublicChatGroupError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"deleted": true})
}

func respondPublicChatAgentError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPublicChatAgentUserRequired),
		errors.Is(err, service.ErrPublicChatAgentUserNotFound),
		errors.Is(err, service.ErrPublicChatAgentUserInvalid),
		errors.Is(err, service.ErrPublicChatAgentIDInvalid),
		errors.Is(err, service.ErrPublicChatAgentIDTaken),
		errors.Is(err, service.ErrPublicChatAgentStatusInvalid),
		errors.Is(err, service.ErrPublicChatAgentOnlineInvalid),
		errors.Is(err, service.ErrPublicChatAgentEmailRequired),
		errors.Is(err, service.ErrPublicChatAgentWhatsAppRequired),
		errors.Is(err, service.ErrPublicChatAgentGroupInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save public chat agent"})
	}
}

func respondPublicChatGroupError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPublicChatGroupNameRequired),
		errors.Is(err, service.ErrPublicChatGroupCodeInvalid),
		errors.Is(err, service.ErrPublicChatGroupCodeTaken),
		errors.Is(err, service.ErrPublicChatGroupStatusInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrPublicChatGroupNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save public chat group"})
	}
}
