package admin

import (
	"net/http"
	"strconv"

	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	userService *service.UserService
}

func NewCustomerHandler(userService *service.UserService) *CustomerHandler {
	return &CustomerHandler{userService: userService}
}

// ListCustomers returns storefront accounts only. Customer records are kept
// separate from staff account administration so they cannot be assigned a
// backoffice role from this read-only view.
func (h *CustomerHandler) ListCustomers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	customers, total, err := h.userService.ListCustomerAccounts(page, pageSize, status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}

	responses := make([]interface{}, len(customers))
	for index, customer := range customers {
		responses[index] = customer.ToResponse()
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	c.JSON(http.StatusOK, gin.H{
		"customers": responses,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}
