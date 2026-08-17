package wheelsetfit

import (
	"errors"
	"net/http"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.WheelsetFitQuestionnaireService
}

func NewHandler(questionnaireService *service.WheelsetFitQuestionnaireService) *Handler {
	return &Handler{service: questionnaireService}
}

func (h *Handler) GetCurrentFlow(c *gin.Context) {
	flow, err := h.service.GetPublishedFlow()
	if err != nil {
		respondWheelsetFitError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": flow})
}

func respondWheelsetFitError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrWheelsetFitQuestionnaireNotFound),
		errors.Is(err, service.ErrWheelsetFitQuestionnaireVersionNotFound):
		apierror.RespondNotFound(c, "Wheelset fit questionnaire")
	case errors.Is(err, service.ErrWheelsetFitQuestionnaireInvalid):
		apierror.RespondBadRequest(c, err.Error())
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load wheelset fit questionnaire"})
	}
}
