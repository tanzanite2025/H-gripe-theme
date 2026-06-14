package spoke

import (
	"net/http"
	"strconv"
	domainspoke "tanzanite/internal/domain/spoke"
	"tanzanite/internal/repository"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	spokeRepo *repository.SpokeRepository
}

func NewHandler(spokeRepo *repository.SpokeRepository) *Handler {
	return &Handler{spokeRepo: spokeRepo}
}

func (h *Handler) GetExport(c *gin.Context) {
	c.JSON(http.StatusOK, domainspoke.DefaultExport())
}

func (h *Handler) ListHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("per_page", "5"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 5
	}

	items, total, err := h.spokeRepo.ListHistory(c.Query("search"), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "spoke_history_error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"meta": gin.H{
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
			"page":        page,
			"per_page":    pageSize,
		},
	})
}
