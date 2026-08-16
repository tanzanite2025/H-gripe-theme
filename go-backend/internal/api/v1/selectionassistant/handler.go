package selectionassistant

import (
	"net/http"
	"strings"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.SelectionAssistantService
}

func NewHandler(selectionAssistantService *service.SelectionAssistantService) *Handler {
	return &Handler{service: selectionAssistantService}
}

func (h *Handler) GetPublishedFlow(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	flow, err := h.service.GetPublishedFlowBySlug(slug)
	if err != nil {
		switch err {
		case service.ErrSelectionAssistantNotFound, service.ErrSelectionAssistantNotPublished:
			c.JSON(http.StatusNotFound, gin.H{"code": http.StatusNotFound, "message": "selection assistant not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": http.StatusInternalServerError, "message": "failed to load selection assistant"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": flow,
	})
}
