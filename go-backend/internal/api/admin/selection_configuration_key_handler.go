package admin

import (
	"errors"
	"net/http"
	"strconv"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type SelectionConfigurationKeyHandler struct {
	service *service.SelectionConfigurationKeyService
}

type selectionConfigurationKeyRequest struct {
	Kind         string `json:"kind" binding:"required"`
	Code         string `json:"code" binding:"required"`
	DisplayLabel string `json:"display_label" binding:"required"`
	Description  string `json:"description"`
	IsEnabled    *bool  `json:"is_enabled"`
	SortOrder    int    `json:"sort_order"`
}

func NewSelectionConfigurationKeyHandler(selectionConfigurationKeyService *service.SelectionConfigurationKeyService) *SelectionConfigurationKeyHandler {
	return &SelectionConfigurationKeyHandler{service: selectionConfigurationKeyService}
}

func (h *SelectionConfigurationKeyHandler) List(c *gin.Context) {
	keys, err := h.service.ListSelectionConfigurationKeys(c.Query("kind"), c.Query("include_disabled") == "true")
	if err != nil {
		respondSelectionConfigurationKeyError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": keys})
}

func (h *SelectionConfigurationKeyHandler) ListOptions(c *gin.Context) {
	options, err := h.service.ListEnabledSelectionConfigurationKeyOptions(c.Query("kind"))
	if err != nil {
		respondSelectionConfigurationKeyError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": options})
}

func (h *SelectionConfigurationKeyHandler) Create(c *gin.Context) {
	var req selectionConfigurationKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	key, err := h.service.CreateSelectionConfigurationKey(selectionConfigurationKeyInput(req))
	if err != nil {
		respondSelectionConfigurationKeyError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": key})
}

func (h *SelectionConfigurationKeyHandler) Update(c *gin.Context) {
	id, ok := parseSelectionConfigurationKeyID(c)
	if !ok {
		return
	}
	var req selectionConfigurationKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	key, err := h.service.UpdateSelectionConfigurationKey(id, selectionConfigurationKeyInput(req))
	if err != nil {
		respondSelectionConfigurationKeyError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": key})
}

func parseSelectionConfigurationKeyID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid selection configuration key ID"})
		return 0, false
	}
	return uint(id), true
}

func selectionConfigurationKeyInput(req selectionConfigurationKeyRequest) service.SelectionConfigurationKeyInput {
	enabled := true
	if req.IsEnabled != nil {
		enabled = *req.IsEnabled
	}
	return service.SelectionConfigurationKeyInput{
		Kind:         req.Kind,
		Code:         req.Code,
		DisplayLabel: req.DisplayLabel,
		Description:  req.Description,
		IsEnabled:    enabled,
		SortOrder:    req.SortOrder,
	}
}

func respondSelectionConfigurationKeyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSelectionConfigurationKeyNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Selection configuration key not found"})
	case errors.Is(err, service.ErrSelectionConfigurationKeyAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": "Selection configuration key already exists"})
	case errors.Is(err, service.ErrSelectionConfigurationKeyCodeImmutable):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Selection configuration key code and kind are immutable"})
	case errors.Is(err, service.ErrSelectionConfigurationKeyKindUnsupported):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrSelectionConfigurationKeyInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to manage selection configuration keys"})
	}
}
