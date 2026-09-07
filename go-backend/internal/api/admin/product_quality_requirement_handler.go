package admin

import (
	"errors"
	"net/http"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type ProductQualityRequirementHandler struct {
	requirementService *service.ProductQualityRequirementService
}

func NewProductQualityRequirementHandler(
	requirementService *service.ProductQualityRequirementService,
) *ProductQualityRequirementHandler {
	return &ProductQualityRequirementHandler{requirementService: requirementService}
}

func (h *ProductQualityRequirementHandler) List(c *gin.Context) {
	if h == nil || h.requirementService == nil {
		apierror.RespondInternalError(c, errors.New("product quality requirement service is not configured"))
		return
	}
	productID, ok := parsePositiveUintParam(c, "id", "invalid product id")
	if !ok {
		return
	}

	rules, err := h.requirementService.ListSpokeTensionQCRules(productID)
	if err != nil {
		respondProductQualityRequirementError(c, err)
		return
	}
	response.Success(c, gin.H{
		"product_id": productID,
		"rules":      rules,
	})
}

func (h *ProductQualityRequirementHandler) UpsertSpokeTensionQC(c *gin.Context) {
	if h == nil || h.requirementService == nil {
		apierror.RespondInternalError(c, errors.New("product quality requirement service is not configured"))
		return
	}
	productID, ok := parsePositiveUintParam(c, "id", "invalid product id")
	if !ok {
		return
	}
	actorID, ok := currentAdminUserID(c)
	if !ok {
		apierror.RespondUnauthorized(c)
		return
	}

	var req productQualityRequirementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	rule, _, err := h.requirementService.UpsertSpokeTensionQC(req.toServiceInput(productID, actorID))
	if err != nil {
		respondProductQualityRequirementError(c, err)
		return
	}
	response.Success(c, rule)
}

func respondProductQualityRequirementError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProductQualityRequirementProductNotFound),
		errors.Is(err, service.ErrProductQualityRequirementVariantNotFound):
		apierror.RespondNotFound(c, "Product quality requirement scope")
	case errors.Is(err, service.ErrProductQualityRequirementStoreUnavailable):
		apierror.RespondError(c, http.StatusServiceUnavailable, "product_quality_requirement_unavailable", "Product quality requirement service is unavailable")
	case errors.Is(err, service.ErrProductQualityRequirementInvalid):
		apierror.RespondBadRequest(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
