package admin

import (
	"errors"
	"io"
	"net/http"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type NotificationTemplateHandler struct {
	service *service.TransactionalNotificationTemplateService
}

func NewNotificationTemplateHandler(templateService *service.TransactionalNotificationTemplateService) *NotificationTemplateHandler {
	return &NotificationTemplateHandler{service: templateService}
}

func (h *NotificationTemplateHandler) List(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("notification template service is not configured"))
		return
	}
	templates, err := h.service.List()
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"templates": templates})
}

func (h *NotificationTemplateHandler) Get(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("notification template service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid notification template id")
	if err != nil {
		return
	}
	templateRecord, err := h.service.Get(id)
	if err != nil {
		respondNotificationTemplateError(c, err)
		return
	}
	response.Success(c, templateRecord)
}

func (h *NotificationTemplateHandler) Versions(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("notification template service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid notification template id")
	if err != nil {
		return
	}
	versions, err := h.service.Versions(id)
	if err != nil {
		respondNotificationTemplateError(c, err)
		return
	}
	response.Success(c, gin.H{"versions": versions})
}

func (h *NotificationTemplateHandler) Rollback(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("notification template service is not configured"))
		return
	}
	id, err := parseUintParam(c, "id", "invalid notification template id")
	if err != nil {
		return
	}
	version, err := parseIntParam(c, "version", "invalid notification template version")
	if err != nil {
		return
	}
	var input struct {
		ChangeReason string `json:"change_reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	templateRecord, err := h.service.Rollback(id, version, adminUserID(c), input.ChangeReason)
	if err != nil {
		respondNotificationTemplateError(c, err)
		return
	}
	response.Success(c, templateRecord)
}

func (h *NotificationTemplateHandler) Save(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("notification template service is not configured"))
		return
	}
	var input struct {
		Code              string   `json:"code"`
		Locale            string   `json:"locale"`
		Category          string   `json:"category"`
		Name              string   `json:"name"`
		SubjectTemplate   string   `json:"subject_template"`
		BodyHTML          string   `json:"body_html"`
		BodyText          string   `json:"body_text"`
		AllowedVariables  []string `json:"allowed_variables"`
		RequiredVariables []string `json:"required_variables"`
		IsEnabled         bool     `json:"is_enabled"`
		Version           int      `json:"version"`
		ChangeReason      string   `json:"change_reason"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	adminID := adminUserID(c)
	templateRecord, err := h.service.Save(service.SaveTransactionalNotificationTemplateInput{
		Code: input.Code, Locale: input.Locale, Category: input.Category, Name: input.Name,
		SubjectTemplate: input.SubjectTemplate, BodyHTML: input.BodyHTML, BodyText: input.BodyText,
		AllowedVariables: input.AllowedVariables, RequiredVariables: input.RequiredVariables,
		IsEnabled: input.IsEnabled, Version: input.Version, ChangedByUserID: adminID,
		ChangeReason: input.ChangeReason,
	})
	if err != nil {
		respondNotificationTemplateError(c, err)
		return
	}
	response.Success(c, templateRecord)
}

func (h *NotificationTemplateHandler) Preview(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("notification template service is not configured"))
		return
	}
	var input struct {
		Code      string            `json:"code"`
		Locale    string            `json:"locale"`
		Variables map[string]string `json:"variables"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	rendered, err := h.service.Render(input.Code, input.Locale, input.Variables)
	if err != nil {
		respondNotificationTemplateError(c, err)
		return
	}
	response.Success(c, rendered)
}

func adminUserID(c *gin.Context) *uint {
	for _, key := range []string{"admin_user_id", "user_id", "userID"} {
		if value, ok := c.Get(key); ok {
			switch typed := value.(type) {
			case uint:
				return &typed
			case uint64:
				converted := uint(typed)
				return &converted
			case int:
				if typed > 0 {
					converted := uint(typed)
					return &converted
				}
			}
		}
	}
	return nil
}

func respondNotificationTemplateError(c *gin.Context, err error) {
	if repository.IsRecordNotFound(err) {
		apierror.RespondNotFound(c, "Notification template")
		return
	}
	if errors.Is(err, service.ErrNotificationTemplateDefinitionMismatch) ||
		errors.Is(err, service.ErrNotificationTemplateLocaleUnavailable) ||
		errors.Is(err, service.ErrNotificationTemplateUnknown) ||
		errors.Is(err, service.ErrNotificationTemplatePlaceholderInvalid) ||
		errors.Is(err, service.ErrNotificationTemplateVariableMissing) ||
		errors.Is(err, service.ErrNotificationTemplateVariableUnknown) ||
		errors.Is(err, service.ErrNotificationTemplateSubjectUnsafe) {
		apierror.RespondError(c, http.StatusBadRequest, "NOTIFICATION_TEMPLATE_INVALID", err.Error())
		return
	}
	apierror.RespondInternalError(c, err)
}

func registerNotificationTemplateRoutes(authenticated *gin.RouterGroup, handler *NotificationTemplateHandler) {
	group := authenticated.Group("/email/templates")
	group.Use(middleware.RequirePermission(auth.PermSettingsView))
	{
		group.GET("", handler.List)
		group.GET("/:id", handler.Get)
		group.GET("/:id/versions", handler.Versions)
		group.POST("/:id/versions/:version/rollback", middleware.RequirePermission(auth.PermSettingsEdit), handler.Rollback)
		group.POST("/preview", middleware.RequirePermission(auth.PermSettingsEdit), handler.Preview)
		group.PUT("/:id", middleware.RequirePermission(auth.PermSettingsEdit), handler.Save)
	}
}
