package admin

import (
	"errors"
	"strings"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/pagination"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type VisitorProfileHandler struct {
	visitorProfileService *service.VisitorProfileService
}

func NewVisitorProfileHandler(visitorProfileService *service.VisitorProfileService) *VisitorProfileHandler {
	return &VisitorProfileHandler{visitorProfileService: visitorProfileService}
}

// ListVisitorProfiles returns the read-only visitor profile source used by Public Chat context.
func (h *VisitorProfileHandler) ListVisitorProfiles(c *gin.Context) {
	if h.visitorProfileService == nil {
		apierror.RespondInternalError(c, errors.New("visitor profile service is not configured"))
		return
	}

	params := pagination.ParsePagination(c)
	input := service.VisitorProfileListInput{
		Search:                 strings.TrimSpace(c.Query("search")),
		Identity:               strings.TrimSpace(c.Query("identity")),
		CountryCode:            strings.TrimSpace(c.Query("country_code")),
		Locale:                 strings.TrimSpace(c.Query("locale")),
		Email:                  strings.TrimSpace(c.Query("email")),
		CartSession:            strings.TrimSpace(c.Query("cart_session")),
		CustomerServiceVisitor: strings.TrimSpace(c.Query("customer_service_visitor")),
		LastSeen:               strings.TrimSpace(c.Query("last_seen")),
	}

	profiles, total, err := h.visitorProfileService.ListProfiles(params.Page, params.PageSize, input)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	totalPages := 0
	if params.PageSize > 0 {
		totalPages = (int(total) + params.PageSize - 1) / params.PageSize
	}

	response.Success(c, gin.H{
		"profiles": profiles,
		"pagination": gin.H{
			"page":        params.Page,
			"page_size":   params.PageSize,
			"total":       total,
			"total_pages": totalPages,
		},
		"filters": gin.H{
			"search":                   input.Search,
			"identity":                 emptyFilterAsAll(input.Identity),
			"country_code":             input.CountryCode,
			"locale":                   input.Locale,
			"email":                    emptyFilterAsAll(input.Email),
			"cart_session":             emptyFilterAsAll(input.CartSession),
			"customer_service_visitor": emptyFilterAsAll(input.CustomerServiceVisitor),
			"last_seen":                emptyFilterAsAll(input.LastSeen),
		},
	})
}

// GetVisitorProfileStats returns aggregate capture coverage for the visitor profile fact source.
func (h *VisitorProfileHandler) GetVisitorProfileStats(c *gin.Context) {
	if h.visitorProfileService == nil {
		apierror.RespondInternalError(c, errors.New("visitor profile service is not configured"))
		return
	}

	stats, err := h.visitorProfileService.GetStats()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"stats": stats})
}

func emptyFilterAsAll(value string) string {
	if strings.TrimSpace(value) == "" {
		return "all"
	}
	return value
}
