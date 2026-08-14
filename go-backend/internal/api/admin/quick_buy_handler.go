package admin

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	productapi "commerce-platform/internal/api/v1/product"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type QuickBuyHandler struct {
	quickBuyService *service.QuickBuyService
}

const (
	quickBuyPreviewDefaultPageSize = 12
	quickBuyPreviewMaxPageSize     = 24
)

func NewQuickBuyHandler(quickBuyService *service.QuickBuyService) *QuickBuyHandler {
	return &QuickBuyHandler{quickBuyService: quickBuyService}
}

func (h *QuickBuyHandler) ListFlows(c *gin.Context) {
	flows, err := h.quickBuyService.ListFlows()
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, gin.H{"data": flows})
}

func (h *QuickBuyHandler) GetFlow(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid quick buy flow id")
	if err != nil {
		return
	}

	flow, err := h.quickBuyService.GetFlow(id, c.Query("locale"))
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, gin.H{"data": flow})
}

func (h *QuickBuyHandler) CreateFlow(c *gin.Context) {
	var input service.QuickBuyFlowInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	flow, err := h.quickBuyService.CreateFlow(input)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Created(c, gin.H{"data": flow})
}

func (h *QuickBuyHandler) UpdateFlow(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid quick buy flow id")
	if err != nil {
		return
	}

	var input service.QuickBuyFlowInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	flow, err := h.quickBuyService.UpdateFlow(id, input)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, gin.H{"data": flow})
}

func (h *QuickBuyHandler) SaveFlowConfiguration(c *gin.Context) {
	id, err := parseUintParam(c, "id", "invalid quick buy flow id")
	if err != nil {
		return
	}

	var input service.QuickBuyFlowInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	flow, err := h.quickBuyService.SaveFlowConfiguration(id, input)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, gin.H{"data": flow})
}

func (h *QuickBuyHandler) CreateDraftVersion(c *gin.Context) {
	flowID, err := parseUintParam(c, "id", "invalid quick buy flow id")
	if err != nil {
		return
	}

	var input service.QuickBuyVersionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	flow, err := h.quickBuyService.CreateDraftVersion(flowID, input)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Created(c, gin.H{"data": flow})
}

func (h *QuickBuyHandler) UpdateDraftVersion(c *gin.Context) {
	versionID, err := parseUintParam(c, "version_id", "invalid quick buy version id")
	if err != nil {
		return
	}

	var input service.QuickBuyVersionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	flow, err := h.quickBuyService.UpdateDraftVersion(versionID, input)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, gin.H{"data": flow})
}

func (h *QuickBuyHandler) ValidateVersion(c *gin.Context) {
	versionID, err := parseUintParam(c, "version_id", "invalid quick buy version id")
	if err != nil {
		return
	}

	result, err := h.quickBuyService.ValidateVersion(versionID)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, gin.H{"data": result})
}

func (h *QuickBuyHandler) PreviewVersionStepCandidates(c *gin.Context) {
	versionID, err := parseUintParam(c, "version_id", "invalid quick buy version id")
	if err != nil {
		return
	}

	input, err := quickBuyPreviewInputFromQuery(c)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&input); err != nil {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		input, err = mergeQuickBuyPreviewQuery(c, input)
		if err != nil {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
	}

	result, err := h.quickBuyService.PreviewVersionStepCandidates(versionID, input)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, gin.H{"data": gin.H{
		"step":            result.Step,
		"products":        productapi.PublicProductsFromDomainWithLocaleAndDisplayCurrency(result.Products, result.Currency, result.Locale),
		"flow_id":         result.FlowID,
		"flow_version_id": result.FlowVersionID,
		"locale":          result.Locale,
		"currency":        result.Currency,
		"page":            result.Page,
		"page_size":       result.PageSize,
		"total":           result.Total,
		"has_more":        result.HasMore,
	}})
}

func (h *QuickBuyHandler) PublishVersion(c *gin.Context) {
	versionID, err := parseUintParam(c, "version_id", "invalid quick buy version id")
	if err != nil {
		return
	}

	var publishedBy *uint
	if userID := c.GetUint("user_id"); userID > 0 {
		publishedBy = &userID
	}

	flow, err := h.quickBuyService.PublishVersion(versionID, publishedBy)
	if err != nil {
		respondQuickBuyError(c, err)
		return
	}

	response.Success(c, gin.H{"data": flow})
}

func quickBuyPreviewInputFromQuery(c *gin.Context) (service.QuickBuyCandidateInput, error) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("per_page", c.DefaultQuery("page_size", strconv.Itoa(quickBuyPreviewDefaultPageSize))))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = quickBuyPreviewDefaultPageSize
	}
	if pageSize > quickBuyPreviewMaxPageSize {
		pageSize = quickBuyPreviewMaxPageSize
	}
	specFilters, err := parseQuickBuySpecFilters(c.Query("spec_filters"))
	if err != nil {
		return service.QuickBuyCandidateInput{}, err
	}
	return service.QuickBuyCandidateInput{
		StepKey:     strings.TrimSpace(c.Query("step_key")),
		Keyword:     strings.TrimSpace(c.Query("keyword")),
		Locale:      strings.TrimSpace(c.Query("locale")),
		Currency:    strings.TrimSpace(c.Query("currency")),
		SpecFilters: specFilters,
		Page:        page,
		PageSize:    pageSize,
	}, nil
}

func mergeQuickBuyPreviewQuery(c *gin.Context, input service.QuickBuyCandidateInput) (service.QuickBuyCandidateInput, error) {
	queryInput, err := quickBuyPreviewInputFromQuery(c)
	if err != nil {
		return service.QuickBuyCandidateInput{}, err
	}
	if input.StepKey == "" {
		input.StepKey = queryInput.StepKey
	}
	if input.Keyword == "" {
		input.Keyword = queryInput.Keyword
	}
	if input.Locale == "" {
		input.Locale = queryInput.Locale
	}
	if input.Currency == "" {
		input.Currency = queryInput.Currency
	}
	if input.Page < 1 {
		input.Page = queryInput.Page
	}
	if input.PageSize < 1 {
		input.PageSize = queryInput.PageSize
	}
	if len(input.SpecFilters) == 0 {
		input.SpecFilters = queryInput.SpecFilters
	}
	return input, nil
}

func parseQuickBuySpecFilters(raw string) (map[string][]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var filters map[string][]string
	if err := json.Unmarshal([]byte(raw), &filters); err != nil {
		return nil, errors.New("spec_filters must be a JSON object of string arrays")
	}
	return filters, nil
}

func respondQuickBuyError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrQuickBuyNotFound):
		apierror.RespondNotFound(c, "Quick buy flow")
	case errors.Is(err, service.ErrQuickBuyInvalid):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrQuickBuyNotMutable):
		apierror.RespondConflict(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
