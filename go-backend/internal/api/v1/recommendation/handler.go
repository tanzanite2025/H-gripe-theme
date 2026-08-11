package recommendation

import (
	"net/http"
	"strings"
	"time"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	recommendationService *service.RecommendationService
}

func NewHandler(recommendationService *service.RecommendationService) *Handler {
	return &Handler{recommendationService: recommendationService}
}

type recommendationRequest struct {
	Surface           string                `json:"surface"`
	Locale            string                `json:"locale"`
	AnonymousID       string                `json:"anonymous_id"`
	SessionID         string                `json:"session_id"`
	Context           recommendationContext `json:"context"`
	Limit             int                   `json:"limit"`
	ExcludeProductIDs []uint                `json:"exclude_product_ids"`
}

type recommendationContext struct {
	ProductID  *uint  `json:"product_id"`
	CategoryID *uint  `json:"category_id"`
	Query      string `json:"query"`
	Route      string `json:"route"`
}

func (h *Handler) GetRecommendations(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32*1024)

	var req recommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	locale := strings.TrimSpace(req.Locale)
	if locale == "" {
		locale = middleware.GetLocale(c)
	}

	result, err := h.recommendationService.Recommend(service.RecommendationRequest{
		Surface:           req.Surface,
		Locale:            locale,
		AnonymousID:       req.AnonymousID,
		SessionID:         req.SessionID,
		ProductID:         req.Context.ProductID,
		CategoryID:        req.Context.CategoryID,
		Query:             req.Context.Query,
		Route:             req.Context.Route,
		Limit:             req.Limit,
		ExcludeProductIDs: req.ExcludeProductIDs,
	})
	if err != nil {
		if service.IsRecommendationValidationError(err) {
			apierror.RespondValidationError(c, err.Error())
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}

	c.Header("Cache-Control", "private, max-age=60")
	c.Header("Expires", result.ExpiresAt.UTC().Format(time.RFC1123))
	response.Success(c, result)
}
