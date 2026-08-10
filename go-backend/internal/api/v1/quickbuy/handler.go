package quickbuy

import (
	"errors"
	"strconv"
	"strings"

	"tanzanite/internal/api/middleware"
	productapi "tanzanite/internal/api/v1/product"
	"tanzanite/internal/pkg/apierror"
	"tanzanite/internal/pkg/response"
	"tanzanite/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	quickBuyService *service.QuickBuyService
}

const (
	quickBuyCandidateDefaultPageSize = 12
	quickBuyCandidateMaxPageSize     = 24
)

func NewHandler(quickBuyService *service.QuickBuyService) *Handler {
	return &Handler{quickBuyService: quickBuyService}
}

func (h *Handler) GetCurrentFlow(c *gin.Context) {
	if h.quickBuyService == nil {
		c.JSON(200, gin.H{"code": 0, "data": nil})
		return
	}

	locale := c.DefaultQuery("locale", middleware.GetLocale(c))
	flow, err := h.quickBuyService.CurrentFlow(
		strings.TrimSpace(c.DefaultQuery("surface", "dock")),
		locale,
	)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	c.JSON(200, gin.H{"code": 0, "data": flow})
}

func (h *Handler) CreateSession(c *gin.Context) {
	if h.quickBuyService == nil {
		apierror.RespondNotFound(c, "Quick buy flow")
		return
	}

	var input service.QuickBuySessionInput
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&input); err != nil {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
	}
	h.enrichSessionInput(c, &input)

	session, err := h.quickBuyService.CreateSession(input)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Created(c, session)
}

func (h *Handler) GetSession(c *gin.Context) {
	session, err := h.quickBuyService.GetSession(c.Param("token"))
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, session)
}

func (h *Handler) UpdateSelections(c *gin.Context) {
	var input service.QuickBuySelectionUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	session, err := h.quickBuyService.UpdateSessionSelections(c.Param("token"), input)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, session)
}

func (h *Handler) ValidateSession(c *gin.Context) {
	session, err := h.quickBuyService.ValidateSession(c.Param("token"))
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, session)
}

func (h *Handler) ListStepCandidates(c *gin.Context) {
	if h.quickBuyService == nil {
		apierror.RespondNotFound(c, "Quick buy")
		return
	}

	input := quickBuyCandidateInputFromQuery(c)
	input.StepKey = c.Param("step_key")

	result, err := h.quickBuyService.ListSessionStepCandidates(c.Param("token"), input)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"code":            0,
		"data":            productapi.PublicProductsFromDomainWithLocaleAndDisplayCurrency(result.Products, result.Currency, result.Locale),
		"step":            result.Step,
		"flow_id":         result.FlowID,
		"flow_version_id": result.FlowVersionID,
		"page":            result.Page,
		"page_size":       result.PageSize,
		"total":           result.Total,
		"has_more":        result.HasMore,
	})
}

func (h *Handler) enrichSessionInput(c *gin.Context, input *service.QuickBuySessionInput) {
	if input.Surface == "" {
		input.Surface = strings.TrimSpace(c.DefaultQuery("surface", "dock"))
	}
	if input.Locale == "" {
		input.Locale = c.DefaultQuery("locale", middleware.GetLocale(c))
	}
	if input.MarketCountry == "" {
		input.MarketCountry = c.DefaultQuery("country", c.DefaultQuery("market_country", c.GetHeader("X-Market-Country")))
	}
	if input.Currency == "" {
		input.Currency = c.DefaultQuery("currency", c.GetHeader("X-Currency"))
	}
	if input.AnonymousID == "" {
		input.AnonymousID = c.GetHeader("X-Anonymous-ID")
	}
	if userID := currentUserID(c); userID != nil {
		input.UserID = userID
	}
}

func quickBuyCandidateInputFromQuery(c *gin.Context) service.QuickBuyCandidateInput {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("per_page", c.DefaultQuery("page_size", strconv.Itoa(quickBuyCandidateDefaultPageSize))))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = quickBuyCandidateDefaultPageSize
	}
	if pageSize > quickBuyCandidateMaxPageSize {
		pageSize = quickBuyCandidateMaxPageSize
	}
	return service.QuickBuyCandidateInput{
		Keyword:  strings.TrimSpace(c.Query("keyword")),
		Locale:   c.DefaultQuery("locale", middleware.GetLocale(c)),
		Currency: c.DefaultQuery("currency", c.GetHeader("X-Display-Currency")),
		Page:     page,
		PageSize: pageSize,
	}
}

func currentUserID(c *gin.Context) *uint {
	var userID *uint
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(uint); ok && id > 0 {
			userID = &id
		}
	}
	return userID
}

func respondQuickBuyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrQuickBuyNotFound), errors.Is(err, service.ErrQuickBuySessionNotFound):
		apierror.RespondNotFound(c, "Quick buy")
	case errors.Is(err, service.ErrQuickBuyInvalid):
		apierror.RespondBadRequest(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
