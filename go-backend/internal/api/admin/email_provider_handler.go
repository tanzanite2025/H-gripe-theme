package admin

import (
	"errors"
	"net/http"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type EmailProviderHandler struct {
	service *service.EmailProviderService
}

func NewEmailProviderHandler(providerService *service.EmailProviderService) *EmailProviderHandler {
	return &EmailProviderHandler{service: providerService}
}

func (h *EmailProviderHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrEmailProviderServiceNotConfigured)
		return
	}
	providers, err := h.service.List()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"providers": providers})
}

func (h *EmailProviderHandler) Get(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrEmailProviderServiceNotConfigured)
		return
	}
	id, err := parseUintParam(c, "id", "invalid email provider id")
	if err != nil {
		return
	}
	provider, err := h.service.Get(id)
	if err != nil {
		if repository.IsRecordNotFound(err) {
			apierror.RespondNotFound(c, "Email provider")
			return
		}
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, provider)
}

func (h *EmailProviderHandler) Create(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrEmailProviderServiceNotConfigured)
		return
	}
	var input service.SaveEmailProviderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	provider, err := h.service.Create(input)
	if err != nil {
		respondEmailProviderError(c, err)
		return
	}
	response.Created(c, provider)
}

func (h *EmailProviderHandler) Update(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrEmailProviderServiceNotConfigured)
		return
	}
	id, err := parseUintParam(c, "id", "invalid email provider id")
	if err != nil {
		return
	}
	var input service.SaveEmailProviderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	provider, err := h.service.Update(id, input)
	if err != nil {
		respondEmailProviderError(c, err)
		return
	}
	response.Success(c, provider)
}

func (h *EmailProviderHandler) SetActive(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrEmailProviderServiceNotConfigured)
		return
	}
	id, err := parseUintParam(c, "id", "invalid email provider id")
	if err != nil {
		return
	}
	var input struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	provider, err := h.service.SetActive(id, input.Active)
	if err != nil {
		respondEmailProviderError(c, err)
		return
	}
	response.Success(c, provider)
}

func (h *EmailProviderHandler) SetDefault(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrEmailProviderServiceNotConfigured)
		return
	}
	id, err := parseUintParam(c, "id", "invalid email provider id")
	if err != nil {
		return
	}
	provider, err := h.service.SetDefault(id)
	if err != nil {
		respondEmailProviderError(c, err)
		return
	}
	response.Success(c, provider)
}

func (h *EmailProviderHandler) Test(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, service.ErrEmailProviderServiceNotConfigured)
		return
	}
	id, err := parseUintParam(c, "id", "invalid email provider id")
	if err != nil {
		return
	}
	var input struct {
		TargetEmail string `json:"target_email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	provider, err := h.service.Test(id, input.TargetEmail)
	if err != nil {
		if errors.Is(err, service.ErrInvalidEmailProvider) {
			apierror.RespondBadRequest(c, err.Error())
			return
		}
		if errors.Is(err, service.ErrEmailProviderTestFailed) {
			apierror.RespondError(c, http.StatusBadGateway, "EMAIL_PROVIDER_TEST_FAILED", err.Error())
			return
		}
		respondEmailProviderError(c, err)
		return
	}
	response.Success(c, provider)
}

func respondEmailProviderError(c *gin.Context, err error) {
	if errors.Is(err, service.ErrEmailProviderMasterKeyRequired) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	if repository.IsRecordNotFound(err) {
		apierror.RespondNotFound(c, "Email provider")
		return
	}
	if errors.Is(err, service.ErrInvalidEmailProvider) || errors.Is(err, service.ErrEmailProviderCodeConflict) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	apierror.RespondError(c, http.StatusInternalServerError, "EMAIL_PROVIDER_ERROR", "email provider operation failed")
}

func registerEmailProviderRoutes(authenticated *gin.RouterGroup, handler *EmailProviderHandler) {
	group := authenticated.Group("/email/providers")
	group.Use(middleware.RequirePermission(auth.PermSettingsView))
	{
		group.GET("", handler.List)
		group.GET("/:id", handler.Get)
		group.POST("", middleware.RequirePermission(auth.PermSettingsEdit), handler.Create)
		group.PUT("/:id", middleware.RequirePermission(auth.PermSettingsEdit), handler.Update)
		group.PATCH("/:id/active", middleware.RequirePermission(auth.PermSettingsEdit), handler.SetActive)
		group.POST("/:id/default", middleware.RequirePermission(auth.PermSettingsEdit), handler.SetDefault)
		group.POST("/:id/test", middleware.RequirePermission(auth.PermSettingsEdit), handler.Test)
	}
}
