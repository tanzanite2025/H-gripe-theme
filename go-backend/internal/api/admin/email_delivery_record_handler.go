package admin

import (
	"errors"
	"strconv"

	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type EmailDeliveryRecordHandler struct {
	service *service.TransactionalNotificationDeliveryRecordService
}

func NewEmailDeliveryRecordHandler(recordService *service.TransactionalNotificationDeliveryRecordService) *EmailDeliveryRecordHandler {
	return &EmailDeliveryRecordHandler{service: recordService}
}

func (h *EmailDeliveryRecordHandler) ListEmailDeliveryRecords(c *gin.Context) {
	if h == nil || h.service == nil {
		apierror.RespondInternalError(c, errors.New("email delivery record service is not configured"))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	records, total, err := h.service.ListEmailDeliveryRecordViews(repository.EmailDeliveryRecordListFilters{
		Status:       c.Query("status"),
		EventType:    c.Query("event_type"),
		TemplateCode: c.Query("template_code"),
		Search:       c.Query("search"),
		Page:         page,
		PageSize:     pageSize,
	})
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	totalPages := 0
	if pageSize > 0 {
		totalPages = int((total + int64(pageSize) - 1) / int64(pageSize))
	}
	response.Success(c, gin.H{
		"records": records,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

func registerEmailDeliveryRecordRoutes(authenticated *gin.RouterGroup, handler *EmailDeliveryRecordHandler) {
	group := authenticated.Group("/email/deliveries")
	group.Use(middleware.RequirePermission(auth.PermSettingsView))
	group.GET("", handler.ListEmailDeliveryRecords)
}
