package urlmanagement

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	urlmanagementdomain "commerce-platform/internal/domain/urlmanagement"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RedirectsHandler struct {
	rules *service.StorefrontRedirectRuleService
}

func NewRedirectsHandler(rules *service.StorefrontRedirectRuleService) *RedirectsHandler {
	return &RedirectsHandler{rules: rules}
}

func (h *RedirectsHandler) List(c *gin.Context) {
	rules, err := h.rules.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rules})
}

func (h *RedirectsHandler) Create(c *gin.Context) {
	var input urlmanagementdomain.StorefrontRedirectRuleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startedAt := time.Now().UTC()
	rule, err := h.rules.Create(input, c.GetUint("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Header("X-Operation-Started-At", startedAt.Format(time.RFC3339Nano))
	c.JSON(http.StatusCreated, gin.H{"data": rule})
}

func (h *RedirectsHandler) Publish(c *gin.Context) {
	id, err := redirectRuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.rules.Publish(id, c.GetUint("user_id"))
	if err != nil {
		writeRedirectRuleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}

func (h *RedirectsHandler) Disable(c *gin.Context) {
	id, err := redirectRuleID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.rules.Disable(id)
	if err != nil {
		writeRedirectRuleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rule})
}

func redirectRuleID(c *gin.Context) (uint, error) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		return 0, errors.New("invalid redirect rule ID")
	}
	return uint(value), nil
}

func writeRedirectRuleError(c *gin.Context, err error) {
	status := http.StatusBadRequest
	if errors.Is(err, gorm.ErrRecordNotFound) {
		status = http.StatusNotFound
	}
	c.JSON(status, gin.H{"error": err.Error()})
}
