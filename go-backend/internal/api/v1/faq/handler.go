package faq

import (
	"commerce-platform/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	faqService *service.FAQService
}

func NewHandler(faqService *service.FAQService) *Handler {
	return &Handler{
		faqService: faqService,
	}
}

func (h *Handler) IncrementFAQView(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.faqService.IncrementPublicViewCount(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "faq not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "View count incremented",
	})
}
