package admin

import (
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

func (h *PublicChatAgentHandler) ListPublicChatAgents(c *gin.Context) {
	overview, err := h.publicChatAgentService.ListPublicChatAgents(100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch public chat agents"})
		return
	}

	c.JSON(http.StatusOK, overview)
}
