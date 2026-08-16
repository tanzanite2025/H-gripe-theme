package admin

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type SelectionAssistantHandler struct {
	service *service.SelectionAssistantService
}

func NewSelectionAssistantHandler(selectionAssistantService *service.SelectionAssistantService) *SelectionAssistantHandler {
	return &SelectionAssistantHandler{service: selectionAssistantService}
}

func (h *SelectionAssistantHandler) ListFlows(c *gin.Context) {
	flows, err := h.service.ListFlows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch selection assistants"})
		return
	}
	response.Success(c, gin.H{"data": flows})
}

func (h *SelectionAssistantHandler) GetFlow(c *gin.Context) {
	id, ok := parseSelectionAssistantID(c)
	if !ok {
		return
	}
	flow, err := h.service.GetFlow(id)
	if err != nil {
		respondSelectionAssistantError(c, err)
		return
	}
	response.Success(c, gin.H{"data": flow})
}

func (h *SelectionAssistantHandler) CreateFlow(c *gin.Context) {
	var input service.SelectionAssistantFlowInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	flow, err := h.service.CreateFlow(input)
	if err != nil {
		respondSelectionAssistantError(c, err)
		return
	}
	response.Created(c, gin.H{"data": flow})
}

func (h *SelectionAssistantHandler) SaveFlowConfiguration(c *gin.Context) {
	id, ok := parseSelectionAssistantID(c)
	if !ok {
		return
	}
	var input service.SelectionAssistantFlowInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	flow, err := h.service.SaveFlowConfiguration(id, input)
	if err != nil {
		respondSelectionAssistantError(c, err)
		return
	}
	response.Success(c, gin.H{"data": flow})
}

func (h *SelectionAssistantHandler) ValidateVersion(c *gin.Context) {
	id, ok := parseSelectionAssistantVersionID(c)
	if !ok {
		return
	}
	result, err := h.service.ValidateVersion(id)
	if err != nil {
		respondSelectionAssistantError(c, err)
		return
	}
	response.Success(c, gin.H{"data": result})
}

func (h *SelectionAssistantHandler) PublishVersion(c *gin.Context) {
	id, ok := parseSelectionAssistantVersionID(c)
	if !ok {
		return
	}
	var publishedBy *uint
	if userID := c.GetUint("user_id"); userID > 0 {
		publishedBy = &userID
	}
	flow, err := h.service.PublishVersion(id, publishedBy)
	if err != nil {
		respondSelectionAssistantError(c, err)
		return
	}
	response.Success(c, gin.H{"data": flow})
}

func parseSelectionAssistantID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid selection assistant id"})
		return 0, false
	}
	return uint(id), true
}

func parseSelectionAssistantVersionID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("version_id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid selection assistant version id"})
		return 0, false
	}
	return uint(id), true
}

func respondSelectionAssistantError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSelectionAssistantNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "selection assistant not found"})
	case errors.Is(err, service.ErrSelectionAssistantVersionFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "selection assistant version not found"})
	case errors.Is(err, service.ErrSelectionAssistantInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrSelectionAssistantNotMutable):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to manage selection assistant"})
	}
}
