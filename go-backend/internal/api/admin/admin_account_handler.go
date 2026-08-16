package admin

import (
	"errors"
	"fmt"
	"strings"

	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type AdminAccountHandler struct {
	maintenanceService *service.AdminAccountMaintenanceService
}

type ensureAdminAccountRequest struct {
	Email     string `json:"email" binding:"required"`
	Username  string `json:"username"`
	Password  string `json:"password" binding:"required"`
	Role      string `json:"role"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Locale    string `json:"locale"`
}

func NewAdminAccountHandler(maintenanceService *service.AdminAccountMaintenanceService) *AdminAccountHandler {
	return &AdminAccountHandler{maintenanceService: maintenanceService}
}

func (h *AdminAccountHandler) List(c *gin.Context) {
	if h == nil || h.maintenanceService == nil {
		apierror.RespondInternalError(c, errors.New("admin account maintenance service is not configured"))
		return
	}
	_, actorRole, ok := currentAdminActor(c)
	if !ok {
		apierror.RespondUnauthorized(c)
		return
	}
	if auth.NormalizeRole(actorRole) != auth.RoleAdmin {
		apierror.RespondForbidden(c)
		return
	}

	accounts, err := h.maintenanceService.ListBackofficeAccounts(c.Query("search"))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}

	response.Success(c, gin.H{"accounts": accounts})
}

func (h *AdminAccountHandler) Ensure(c *gin.Context) {
	if h == nil || h.maintenanceService == nil {
		apierror.RespondInternalError(c, errors.New("admin account maintenance service is not configured"))
		return
	}

	actorID, actorRole, ok := currentAdminActor(c)
	if !ok {
		apierror.RespondUnauthorized(c)
		return
	}
	if auth.NormalizeRole(actorRole) != auth.RoleAdmin {
		apierror.RespondForbidden(c)
		return
	}

	var req ensureAdminAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	result, err := h.maintenanceService.EnsureBackofficeAccount(service.AdminAccountMaintenanceInput{
		Email:       req.Email,
		Username:    req.Username,
		Password:    req.Password,
		Role:        req.Role,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Locale:      req.Locale,
		OperatorID:  actorID,
		Operator:    adminAccountOperatorName(c, actorID),
		AuditMethod: c.Request.Method,
		AuditPath:   c.Request.URL.Path,
		AuditIP:     c.ClientIP(),
		AuditAgent:  c.Request.UserAgent(),
	})
	if err != nil {
		respondAdminAccountMaintenanceError(c, err)
		return
	}

	message := "Admin account reset successfully"
	if result.Created {
		message = "Admin account created successfully"
	}
	response.SuccessWithMessage(c, message, result)
}

func adminAccountOperatorName(c *gin.Context, actorID uint) string {
	if c != nil {
		if username := strings.TrimSpace(c.GetString("username")); username != "" {
			return username
		}
		if email := strings.TrimSpace(c.GetString("email")); email != "" {
			return email
		}
	}
	return fmt.Sprintf("admin:%d", actorID)
}

func respondAdminAccountMaintenanceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAdminAccountEmailRequired),
		errors.Is(err, service.ErrAdminAccountEmailInvalid),
		errors.Is(err, service.ErrAdminAccountUsernameInvalid),
		errors.Is(err, service.ErrAdminAccountPasswordRequired),
		errors.Is(err, service.ErrAdminAccountWeakPassword),
		errors.Is(err, service.ErrUnsupportedLocale):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrUsernameExists):
		apierror.RespondConflict(c, "Username already exists")
	case errors.Is(err, service.ErrAdminAccountRoleForbidden),
		errors.Is(err, service.ErrAdminAccountSelfRoleChange):
		apierror.RespondForbidden(c)
	default:
		apierror.RespondInternalError(c, err)
	}
}
