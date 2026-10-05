package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// Yanwen waybill creation, official synchronization, labels, and cancellation admin endpoints.

type YanwenWaybillHandler struct {
	service *service.YanwenWaybillOperationsService
}

func NewYanwenWaybillHandler(waybillService *service.YanwenWaybillOperationsService) *YanwenWaybillHandler {
	return &YanwenWaybillHandler{service: waybillService}
}

func (h *YanwenWaybillHandler) ListYanwenWaybills(c *gin.Context) {
	environment := strings.ToLower(strings.TrimSpace(c.DefaultQuery("environment", "")))
	if environment == "test" {
		environment = "fat"
	}
	if environment != "" && environment != "fat" && environment != "production" {
		apierror.RespondBadRequest(c, "Yanwen environment must be fat or production")
		return
	}
	waybills, err := h.service.ListYanwenWaybills(environment, c.Query("keyword"), c.Query("status"), c.Query("warehouse_code"))
	if err != nil {
		apierror.RespondInternalError(c, err)
		return
	}
	response.Success(c, gin.H{"data": waybills})
}

func (h *YanwenWaybillHandler) CreateYanwenWaybill(c *gin.Context) {
	var input service.YanwenCreateWaybillInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	waybill, err := h.service.CreateYanwenWaybill(c.Request.Context(), input)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Created(c, waybill)
}

func (h *YanwenWaybillHandler) CreateYanwenWaybills(c *gin.Context) {
	var input service.YanwenBatchWaybillCreationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.CreateYanwenWaybills(c.Request.Context(), input.Requests)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *YanwenWaybillHandler) SyncYanwenWaybillOfficialDetails(c *gin.Context) {
	waybillID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || waybillID == 0 {
		apierror.RespondBadRequest(c, "Yanwen waybill id must be a positive integer")
		return
	}
	waybill, err := h.service.SyncYanwenWaybillOfficialDetails(c.Request.Context(), uint(waybillID))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, waybill)
}

func (h *YanwenWaybillHandler) SyncYanwenWaybillsOfficialDetails(c *gin.Context) {
	var input service.YanwenBatchWaybillOfficialSyncInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	waybills, err := h.service.SyncYanwenWaybillsOfficialDetails(c.Request.Context(), input.WaybillIDs)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"data": waybills})
}

func (h *YanwenWaybillHandler) DownloadYanwenWaybillLabel(c *gin.Context) {
	waybillID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || waybillID == 0 {
		apierror.RespondBadRequest(c, "Yanwen waybill id must be a positive integer")
		return
	}
	label, err := h.service.DownloadYanwenWaybillLabel(c.Request.Context(), uint(waybillID))
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, label)
}

func (h *YanwenWaybillHandler) DownloadYanwenWaybillLabelsArchive(c *gin.Context) {
	var input service.YanwenBatchWaybillLabelInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	archive, err := h.service.DownloadYanwenWaybillLabelsArchive(c.Request.Context(), input.WaybillIDs)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+archive.FileName+`"`)
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Yanwen-Label-Succeeded", strconv.Itoa(archive.Summary.Succeeded))
	c.Header("X-Yanwen-Label-Failed", strconv.Itoa(archive.Summary.Failed))
	c.Data(http.StatusOK, archive.ContentType, archive.Data)
}

func (h *YanwenWaybillHandler) CancelYanwenWaybill(c *gin.Context) {
	waybillID, err := strconv.ParseUint(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || waybillID == 0 {
		apierror.RespondBadRequest(c, "Yanwen waybill id must be a positive integer")
		return
	}
	var input service.YanwenCancelWaybillInput
	if err := c.ShouldBindJSON(&input); err != nil && !errors.Is(err, io.EOF) {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.CancelYanwenWaybill(c.Request.Context(), uint(waybillID), input.Note)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *YanwenWaybillHandler) CancelYanwenWaybills(c *gin.Context) {
	var input service.YanwenBatchWaybillCancellationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	result, err := h.service.CancelYanwenWaybills(c.Request.Context(), input.WaybillIDs, input.Note)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	response.Success(c, result)
}

func registerYanwenWaybillRoutes(authenticated *gin.RouterGroup, handler *YanwenWaybillHandler) {
	group := authenticated.Group("/logistics/yanwen/waybills")
	group.Use(middleware.RequirePermission(auth.PermYanwenView))
	group.GET("", handler.ListYanwenWaybills)
	group.POST("", middleware.RequirePermission(auth.PermYanwenShip), handler.CreateYanwenWaybill)
	group.POST("/batch", middleware.RequirePermission(auth.PermYanwenShip), handler.CreateYanwenWaybills)
	group.POST("/batch-sync", middleware.RequirePermission(auth.PermYanwenShip), handler.SyncYanwenWaybillsOfficialDetails)
	group.POST("/batch-labels", middleware.RequirePermission(auth.PermYanwenShip), handler.DownloadYanwenWaybillLabelsArchive)
	group.POST("/batch-cancel", middleware.RequirePermission(auth.PermYanwenCancel), handler.CancelYanwenWaybills)
	group.POST("/:id/sync", middleware.RequirePermission(auth.PermYanwenShip), handler.SyncYanwenWaybillOfficialDetails)
	group.POST("/:id/label", middleware.RequirePermission(auth.PermYanwenShip), handler.DownloadYanwenWaybillLabel)
	group.POST("/:id/cancel", middleware.RequirePermission(auth.PermYanwenCancel), handler.CancelYanwenWaybill)
}
