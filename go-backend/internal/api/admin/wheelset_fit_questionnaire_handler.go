package admin

import (
	"errors"

	wheelsetfit "commerce-platform/internal/domain/wheelsetfit"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type WheelsetFitQuestionnaireHandler struct {
	service        *service.WheelsetFitQuestionnaireService
	productService *service.ProductService
}

func NewWheelsetFitQuestionnaireHandler(questionnaireService *service.WheelsetFitQuestionnaireService, productService *service.ProductService) *WheelsetFitQuestionnaireHandler {
	return &WheelsetFitQuestionnaireHandler{service: questionnaireService, productService: productService}
}

func (h *WheelsetFitQuestionnaireHandler) GetCurrentVersion(c *gin.Context) {
	version, err := h.service.GetCurrentVersion()
	if err != nil {
		if errors.Is(err, service.ErrWheelsetFitQuestionnaireNotFound) || errors.Is(err, service.ErrWheelsetFitQuestionnaireVersionNotFound) {
			response.Success(c, gin.H{"data": nil})
			return
		}
		respondWheelsetFitQuestionnaireError(c, err)
		return
	}
	response.Success(c, gin.H{"data": version})
}

func (h *WheelsetFitQuestionnaireHandler) GetProductFilterOptions(c *gin.Context) {
	if h.productService == nil {
		apierror.RespondInternalError(c, errors.New("product service is not configured"))
		return
	}
	options, err := h.productService.ListFilterableSpecificationsWithDynamicValuesForCategory(wheelsetfit.WheelsetProductCategorySlug)
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": options})
}

func (h *WheelsetFitQuestionnaireHandler) CreateDraft(c *gin.Context) {
	version, err := h.service.CreateDraft()
	if err != nil {
		respondWheelsetFitQuestionnaireError(c, err)
		return
	}
	response.Created(c, gin.H{"data": version})
}

func (h *WheelsetFitQuestionnaireHandler) CreateQuestion(c *gin.Context) {
	var input service.WheelsetFitQuestionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	input.ID = 0

	version, err := h.service.SaveQuestion(input)
	if err != nil {
		respondWheelsetFitQuestionnaireError(c, err)
		return
	}
	response.Created(c, gin.H{"data": version})
}

func (h *WheelsetFitQuestionnaireHandler) UpdateQuestion(c *gin.Context) {
	questionID, err := parseUintParam(c, "id", "invalid wheelset fit question id")
	if err != nil {
		return
	}
	var input service.WheelsetFitQuestionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	input.ID = questionID

	version, err := h.service.SaveQuestion(input)
	if err != nil {
		respondWheelsetFitQuestionnaireError(c, err)
		return
	}
	response.Success(c, gin.H{"data": version})
}

func (h *WheelsetFitQuestionnaireHandler) DeleteQuestion(c *gin.Context) {
	questionID, err := parseUintParam(c, "id", "invalid wheelset fit question id")
	if err != nil {
		return
	}
	version, err := h.service.DeleteQuestion(questionID)
	if err != nil {
		respondWheelsetFitQuestionnaireError(c, err)
		return
	}
	response.Success(c, gin.H{"data": version})
}

func (h *WheelsetFitQuestionnaireHandler) ReorderQuestions(c *gin.Context) {
	var input service.WheelsetFitQuestionOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	version, err := h.service.ReorderQuestions(input)
	if err != nil {
		respondWheelsetFitQuestionnaireError(c, err)
		return
	}
	response.Success(c, gin.H{"data": version})
}

func (h *WheelsetFitQuestionnaireHandler) ValidateVersion(c *gin.Context) {
	versionID, err := parseUintParam(c, "version_id", "invalid wheelset fit questionnaire version id")
	if err != nil {
		return
	}
	result, err := h.service.ValidateVersion(versionID)
	if err != nil {
		respondWheelsetFitQuestionnaireError(c, err)
		return
	}
	response.Success(c, gin.H{"data": result})
}

func (h *WheelsetFitQuestionnaireHandler) PublishVersion(c *gin.Context) {
	versionID, err := parseUintParam(c, "version_id", "invalid wheelset fit questionnaire version id")
	if err != nil {
		return
	}
	var publishedBy *uint
	if userID := c.GetUint("user_id"); userID > 0 {
		publishedBy = &userID
	}
	version, err := h.service.PublishVersion(versionID, publishedBy)
	if err != nil {
		respondWheelsetFitQuestionnaireError(c, err)
		return
	}
	response.Success(c, gin.H{"data": version})
}

func respondWheelsetFitQuestionnaireError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWheelsetFitQuestionnaireNotFound),
		errors.Is(err, service.ErrWheelsetFitQuestionnaireVersionNotFound),
		errors.Is(err, service.ErrWheelsetFitQuestionNotFound):
		apierror.RespondNotFound(c, "Wheelset fit questionnaire")
	case errors.Is(err, service.ErrWheelsetFitQuestionnaireInvalid):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrWheelsetFitQuestionKeyNotRegistered),
		errors.Is(err, service.ErrWheelsetFitQuestionKeyDisabled),
		errors.Is(err, service.ErrWheelsetFitAnswerKeyNotRegistered),
		errors.Is(err, service.ErrWheelsetFitAnswerKeyDisabled):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrWheelsetFitQuestionnaireNotMutable):
		apierror.RespondConflict(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
