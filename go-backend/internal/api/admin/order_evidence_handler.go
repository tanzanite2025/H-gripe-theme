package admin

import (
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/orderevidence"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type OrderEvidenceHandler struct {
	evidenceService   *service.OrderEvidenceAdminService
	attachmentService *service.OrderEvidenceAttachmentService
	exportService     *service.OrderEvidenceExportSnapshotService
	auditService      adminAuditRecorder
}

func NewOrderEvidenceHandler(evidenceService *service.OrderEvidenceAdminService) *OrderEvidenceHandler {
	return &OrderEvidenceHandler{evidenceService: evidenceService}
}

func (h *OrderEvidenceHandler) ConfigureAttachmentService(
	attachmentService *service.OrderEvidenceAttachmentService,
) {
	if h == nil {
		return
	}
	h.attachmentService = attachmentService
}

func (h *OrderEvidenceHandler) ConfigureExportService(
	exportService *service.OrderEvidenceExportSnapshotService,
) {
	if h == nil {
		return
	}
	h.exportService = exportService
}

func (h *OrderEvidenceHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

func (h *OrderEvidenceHandler) ListOrders(c *gin.Context) {
	if h == nil || h.evidenceService == nil {
		apierror.RespondInternalError(c, errors.New("order evidence service is not configured"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	highValue, err := parseOptionalBoolQuery(c, "high_value")
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}
	spokeTensionQC, err := parseOptionalBoolQuery(c, "spoke_tension_qc")
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	result, err := h.evidenceService.ListOrders(service.OrderEvidenceAdminListInput{
		Page:           page,
		PageSize:       pageSize,
		Search:         c.Query("search"),
		PackageStatus:  strings.TrimSpace(c.Query("package_status")),
		HighValue:      highValue,
		SpokeTensionQC: spokeTensionQC,
	})
	if err != nil {
		respondOrderEvidenceError(c, err)
		return
	}
	response.Paged(c, result.Items, result.Page, result.PageSize, result.Total)
}

func (h *OrderEvidenceHandler) GetPackage(c *gin.Context) {
	if h == nil || h.evidenceService == nil {
		apierror.RespondInternalError(c, errors.New("order evidence service is not configured"))
		return
	}
	orderID, ok := parsePositiveUintParam(c, "id", "invalid order id")
	if !ok {
		return
	}

	result, err := h.evidenceService.GetPackage(orderID)
	if err != nil {
		respondOrderEvidenceError(c, err)
		return
	}
	response.Success(c, result)
}

// ExportSnapshot returns the immutable JSON manifest for the current locked
// evidence package version. It never serializes a fresh read-time assembly.
func (h *OrderEvidenceHandler) ExportSnapshot(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.exportService == nil {
		respondOrderEvidenceError(c, service.ErrOrderEvidenceExportSnapshotUnavailable)
		return
	}
	orderID, ok := parsePositiveUintParam(c, "id", "invalid order id")
	if !ok {
		return
	}

	createdBy := uint(0)
	if actorID, actorOK := currentAdminUserID(c); actorOK {
		createdBy = actorID
	}
	snapshot, err := h.exportService.Export(orderID, createdBy)
	if err != nil {
		recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceOrderOutboundEvidence,
			ResourceID:   orderID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: map[string]interface{}{
				"operation": "export",
			},
		})
		respondOrderEvidenceError(c, err)
		return
	}

	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionExecute,
		Resource:   adminAuditResourceOrderOutboundEvidence,
		ResourceID: orderID,
		Status:     adminAuditStatusSuccess,
		Changes: map[string]interface{}{
			"operation": "export",
		},
		NewValue: orderEvidenceExportSnapshotAuditValue(snapshot),
	})

	filename := "order-evidence-" + strconv.FormatUint(uint64(orderID), 10) +
		"-v" + strconv.Itoa(snapshot.EvidencePackageVersion) + ".json"
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{
		"filename": filename,
	}))
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Order-Evidence-Export-Snapshot-ID", strconv.FormatUint(uint64(snapshot.ID), 10))
	c.Header("X-Order-Evidence-Export-Package-Version", strconv.Itoa(snapshot.EvidencePackageVersion))
	c.Header("X-Order-Evidence-Export-SHA256", strings.TrimSpace(snapshot.SnapshotSHA256))
	c.Data(http.StatusOK, "application/json; charset=utf-8", append([]byte(nil), snapshot.SnapshotData...))
}

func (h *OrderEvidenceHandler) UpdateItem(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.evidenceService == nil {
		apierror.RespondInternalError(c, errors.New("order evidence service is not configured"))
		return
	}
	orderID, ok := parsePositiveUintParam(c, "id", "invalid order id")
	if !ok {
		return
	}
	itemID, ok := parsePositiveUintParam(c, "item_id", "invalid order evidence item id")
	if !ok {
		return
	}
	actorID, ok := currentAdminUserID(c)
	if !ok {
		apierror.RespondUnauthorized(c)
		return
	}

	var req orderEvidenceItemUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOrderOutboundEvidence,
			ResourceID:   orderID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: map[string]interface{}{
				"item_id":       itemID,
				"request_valid": false,
			},
		})
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	before, _ := h.evidenceService.GetPackage(orderID)
	result, err := h.evidenceService.UpdateItem(orderID, req.toServiceInput(itemID, actorID))
	if err != nil {
		recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionUpdate,
			Resource:     adminAuditResourceOrderOutboundEvidence,
			ResourceID:   orderID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes:      orderEvidenceUpdateAuditChanges(itemID, req, nil),
			OldValue:     orderEvidenceAuditItemValue(orderEvidenceResultItem(before, itemID)),
		})
		respondOrderEvidenceError(c, err)
		return
	}
	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionUpdate,
		Resource:   adminAuditResourceOrderOutboundEvidence,
		ResourceID: orderID,
		Status:     adminAuditStatusSuccess,
		Changes:    orderEvidenceUpdateAuditChanges(itemID, req, result),
		OldValue:   orderEvidenceAuditItemValue(orderEvidenceResultItem(before, itemID)),
		NewValue:   orderEvidenceAuditItemValue(orderEvidenceResultItem(result, itemID)),
	})
	response.Success(c, result)
}

func (h *OrderEvidenceHandler) LockPackage(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.evidenceService == nil {
		apierror.RespondInternalError(c, errors.New("order evidence service is not configured"))
		return
	}
	orderID, ok := parsePositiveUintParam(c, "id", "invalid order id")
	if !ok {
		return
	}

	before, _ := h.evidenceService.GetPackage(orderID)
	result, err := h.evidenceService.LockPackage(orderID)
	if err != nil {
		recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceOrderOutboundEvidence,
			ResourceID:   orderID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: map[string]interface{}{
				"operation": "lock",
			},
			OldValue: orderEvidenceAuditPackageValue(evidencePackageFromResult(before)),
		})
		respondOrderEvidenceError(c, err)
		return
	}
	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionExecute,
		Resource:   adminAuditResourceOrderOutboundEvidence,
		ResourceID: orderID,
		Status:     adminAuditStatusSuccess,
		Changes: map[string]interface{}{
			"operation": "lock",
		},
		OldValue: orderEvidenceAuditPackageValue(evidencePackageFromResult(before)),
		NewValue: orderEvidenceAuditPackageValue(evidencePackageFromResult(result)),
	})
	response.Success(c, result)
}

func (h *OrderEvidenceHandler) CreateRevision(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.evidenceService == nil {
		apierror.RespondInternalError(c, errors.New("order evidence service is not configured"))
		return
	}
	orderID, ok := parsePositiveUintParam(c, "id", "invalid order id")
	if !ok {
		return
	}
	actorID, ok := currentAdminUserID(c)
	if !ok {
		apierror.RespondUnauthorized(c)
		return
	}

	before, _ := h.evidenceService.GetPackage(orderID)
	result, err := h.evidenceService.CreateRevision(orderID, actorID)
	if err != nil {
		recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionExecute,
			Resource:     adminAuditResourceOrderOutboundEvidence,
			ResourceID:   orderID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: map[string]interface{}{
				"operation":  "revision",
				"created_by": actorID,
			},
			OldValue: orderEvidenceAuditPackageValue(evidencePackageFromResult(before)),
		})
		respondOrderEvidenceError(c, err)
		return
	}
	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionExecute,
		Resource:   adminAuditResourceOrderOutboundEvidence,
		ResourceID: orderID,
		Status:     adminAuditStatusSuccess,
		Changes: map[string]interface{}{
			"operation":  "revision",
			"created_by": actorID,
		},
		OldValue: orderEvidenceAuditPackageValue(evidencePackageFromResult(before)),
		NewValue: orderEvidenceAuditPackageValue(evidencePackageFromResult(result)),
	})
	response.Success(c, result)
}

func (h *OrderEvidenceHandler) UploadAttachment(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.attachmentService == nil {
		respondOrderEvidenceError(c, service.ErrOrderEvidenceAttachmentUnavailable)
		return
	}
	orderID, ok := parsePositiveUintParam(c, "id", "invalid order id")
	if !ok {
		return
	}
	itemID, ok := parsePositiveUintParam(c, "item_id", "invalid order evidence item id")
	if !ok {
		return
	}
	actorID, ok := currentAdminUserID(c)
	if !ok {
		apierror.RespondUnauthorized(c)
		return
	}

	const maxRequestBytes = 10 << 20
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxRequestBytes)
	if err := c.Request.ParseMultipartForm(2 << 20); err != nil {
		if isRequestBodyTooLarge(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "order evidence attachment is too large",
				"code":  upload.CodeFileTooLarge,
			})
			return
		}
		apierror.RespondBadRequest(c, "invalid multipart upload")
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file is required",
			"code":  upload.CodeEmptyFile,
		})
		return
	}
	attachment, err := h.attachmentService.Upload(
		c.Request.Context(),
		orderID,
		itemID,
		file,
		actorID,
	)
	if err != nil {
		recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionCreate,
			Resource:     adminAuditResourceOrderOutboundEvidence,
			ResourceID:   orderID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: map[string]interface{}{
				"operation": "attachment_upload",
				"item_id":   itemID,
			},
		})
		respondOrderEvidenceError(c, err)
		return
	}
	var after *service.OrderEvidenceAdminPackageResult
	if h.evidenceService != nil {
		after, _ = h.evidenceService.GetPackage(orderID)
	}
	attachmentCount := 0
	if item := orderEvidenceResultItem(after, itemID); item != nil {
		attachmentCount = len(item.Attachments)
	}
	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionCreate,
		Resource:   adminAuditResourceOrderOutboundEvidence,
		ResourceID: orderID,
		Status:     adminAuditStatusSuccess,
		Changes: map[string]interface{}{
			"operation": "attachment_upload",
			"item_id":   itemID,
		},
		NewValue: orderEvidenceAttachmentAuditValue(attachment, attachmentCount),
	})
	response.Created(c, attachment)
}

func (h *OrderEvidenceHandler) ServeAttachment(c *gin.Context) {
	if h == nil || h.attachmentService == nil {
		respondOrderEvidenceError(c, service.ErrOrderEvidenceAttachmentUnavailable)
		return
	}
	orderID, ok := parsePositiveUintParam(c, "id", "invalid order id")
	if !ok {
		return
	}
	itemID, ok := parsePositiveUintParam(c, "item_id", "invalid order evidence item id")
	if !ok {
		return
	}
	attachmentID, ok := parsePositiveUintParam(c, "attachment_id", "invalid order evidence attachment id")
	if !ok {
		return
	}

	attachment, object, err := h.attachmentService.Open(
		c.Request.Context(),
		orderID,
		itemID,
		attachmentID,
	)
	if err != nil {
		respondOrderEvidenceError(c, err)
		return
	}
	if object == nil || object.ReadCloser == nil {
		respondOrderEvidenceError(c, service.ErrOrderEvidenceAttachmentUnavailable)
		return
	}
	defer func() { _ = object.ReadCloser.Close() }()

	contentType := strings.TrimSpace(object.MimeType)
	if contentType == "" {
		contentType = strings.TrimSpace(attachment.MimeType)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "private, no-store")
	c.Header("Content-Disposition", mime.FormatMediaType("inline", map[string]string{
		"filename": strings.TrimSpace(attachment.OriginalFilename),
	}))
	if object.Size >= 0 {
		c.Header("Content-Length", strconv.FormatInt(object.Size, 10))
	}
	if _, err := io.Copy(c.Writer, object.ReadCloser); err != nil {
		return
	}
}

func (h *OrderEvidenceHandler) DeleteAttachment(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	if h == nil || h.attachmentService == nil {
		respondOrderEvidenceError(c, service.ErrOrderEvidenceAttachmentUnavailable)
		return
	}
	orderID, ok := parsePositiveUintParam(c, "id", "invalid order id")
	if !ok {
		return
	}
	itemID, ok := parsePositiveUintParam(c, "item_id", "invalid order evidence item id")
	if !ok {
		return
	}
	attachmentID, ok := parsePositiveUintParam(c, "attachment_id", "invalid order evidence attachment id")
	if !ok {
		return
	}

	err := h.attachmentService.Delete(c.Request.Context(), orderID, itemID, attachmentID)
	if err != nil {
		recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionDelete,
			Resource:     adminAuditResourceOrderOutboundEvidence,
			ResourceID:   orderID,
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: map[string]interface{}{
				"operation":     "attachment_reference_delete",
				"item_id":       itemID,
				"attachment_id": attachmentID,
			},
		})
		respondOrderEvidenceError(c, err)
		return
	}
	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionDelete,
		Resource:   adminAuditResourceOrderOutboundEvidence,
		ResourceID: orderID,
		Status:     adminAuditStatusSuccess,
		Changes: map[string]interface{}{
			"operation":      "attachment_reference_delete",
			"item_id":        itemID,
			"attachment_id":  attachmentID,
			"storage_object": "retained",
		},
	})
	response.Success(c, gin.H{"deleted": true})
}

func parseOptionalBoolQuery(c *gin.Context, name string) (*bool, error) {
	value := strings.TrimSpace(c.Query(name))
	if value == "" || strings.EqualFold(value, "all") {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, errors.New("query parameter " + name + " must be true, false, or all")
	}
	return &parsed, nil
}

func respondOrderEvidenceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrOrderEvidenceAdminPackageNotFound),
		errors.Is(err, service.ErrOrderEvidenceAdminItemNotFound),
		errors.Is(err, service.ErrOrderEvidencePackageNotFound),
		errors.Is(err, service.ErrOrderEvidenceItemNotFound):
		apierror.RespondNotFound(c, "Order evidence resource")
	case errors.Is(err, orderevidence.ErrOrderEvidencePackageLocked),
		errors.Is(err, orderevidence.ErrOrderEvidencePackageSuperseded),
		errors.Is(err, orderevidence.ErrOrderEvidencePackageIncomplete),
		errors.Is(err, orderevidence.ErrOrderEvidenceRevisionSource),
		errors.Is(err, service.ErrOrderEvidenceConfigurationImmutable),
		errors.Is(err, service.ErrOrderEvidenceExportRequiresLocked):
		apierror.RespondConflict(c, err.Error())
	case errors.Is(err, service.ErrOrderEvidenceItemInvalid):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrOrderEvidenceAttachmentRequired):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrOrderEvidenceAdminStoreUnavailable),
		errors.Is(err, service.ErrOrderEvidenceStoreUnavailable),
		errors.Is(err, service.ErrOrderEvidenceAttachmentUnavailable),
		errors.Is(err, service.ErrOrderEvidenceExportSnapshotUnavailable):
		apierror.RespondError(c, http.StatusServiceUnavailable, "order_evidence_unavailable", "Order evidence service is unavailable")
	case errors.Is(err, service.ErrOrderEvidenceExportPackageMissing):
		apierror.RespondNotFound(c, "Order evidence package")
	case errors.Is(err, service.ErrOrderEvidenceExportSnapshotInvalid):
		apierror.RespondConflict(c, err.Error())
	case errors.Is(err, service.ErrOrderEvidenceAttachmentItemNotFound):
		apierror.RespondNotFound(c, "Order evidence attachment")
	case errors.Is(err, service.ErrOrderEvidenceAttachmentOwnership):
		apierror.RespondBadRequest(c, err.Error())
	case upload.ErrorCode(err) != "invalid_upload":
		apierror.RespondError(c, upload.HTTPStatus(err), upload.ErrorCode(err), err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
