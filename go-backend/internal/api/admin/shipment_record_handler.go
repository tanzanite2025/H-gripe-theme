package admin

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/repository"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	shipmentRecordUploadMaxRequestBytes = 82 << 20
	shipmentRecordUploadMultipartMemory = 16 << 20
)

type ShipmentRecordHandler struct {
	shipmentRecordService *service.ShipmentRecordService
	storageService        storage.StorageService
}

func NewShipmentRecordHandler(
	shipmentRecordService *service.ShipmentRecordService,
	storageService storage.StorageService,
) *ShipmentRecordHandler {
	return &ShipmentRecordHandler{
		shipmentRecordService: shipmentRecordService,
		storageService:        storageService,
	}
}

func (h *ShipmentRecordHandler) List(c *gin.Context) {
	params := pagination.ParsePagination(c)
	records, total, err := h.shipmentRecordService.ListAdmin(service.ShipmentRecordListInput{
		Page:     params.Page,
		PageSize: params.PageSize,
		Keyword:  c.Query("keyword"),
		Status:   c.Query("status"),
	})
	if err != nil {
		respondShipmentRecordError(c, err)
		return
	}

	response.Paged(c, records, params.Page, params.PageSize, total)
}

func (h *ShipmentRecordHandler) Get(c *gin.Context) {
	id, ok := parseShipmentRecordID(c)
	if !ok {
		return
	}

	record, err := h.shipmentRecordService.GetAdmin(id)
	if err != nil {
		respondShipmentRecordError(c, err)
		return
	}
	response.Success(c, record)
}

func (h *ShipmentRecordHandler) Stats(c *gin.Context) {
	stats, err := h.shipmentRecordService.GetStats()
	if err != nil {
		respondShipmentRecordError(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *ShipmentRecordHandler) Update(c *gin.Context) {
	id, ok := parseShipmentRecordID(c)
	if !ok {
		return
	}

	var req struct {
		ShippingNote   string   `json:"shipping_note"`
		ShippingImages []string `json:"shipping_images"`
		ProductCodes   []string `json:"product_codes"`
		WarrantyMonths int      `json:"warranty_months"`
		WarrantyStart  string   `json:"warranty_start_at"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	warrantyStart, err := parseShipmentRecordDate(req.WarrantyStart)
	if err != nil {
		apierror.RespondBadRequest(c, "invalid warranty_start_at")
		return
	}

	record, err := h.shipmentRecordService.Update(id, service.ShipmentRecordUpdateInput{
		ShippingNote:   req.ShippingNote,
		ShippingImages: req.ShippingImages,
		ProductCodes:   req.ProductCodes,
		WarrantyMonths: req.WarrantyMonths,
		WarrantyStart:  warrantyStart,
	})
	if err != nil {
		respondShipmentRecordError(c, err)
		return
	}
	response.Success(c, record)
}

func (h *ShipmentRecordHandler) UploadImages(c *gin.Context) {
	id, ok := parseShipmentRecordID(c)
	if !ok {
		return
	}
	if h.storageService == nil {
		apierror.RespondError(c, http.StatusServiceUnavailable, "storage_unavailable", "file storage is unavailable")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, shipmentRecordUploadMaxRequestBytes)
	if err := c.Request.ParseMultipartForm(shipmentRecordUploadMultipartMemory); err != nil {
		if isRequestBodyTooLarge(err) {
			apierror.RespondError(c, http.StatusRequestEntityTooLarge, upload.CodeFileTooLarge, "shipment images are too large")
			return
		}
		apierror.RespondBadRequest(c, "invalid multipart upload")
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}

	form, err := c.MultipartForm()
	if err != nil {
		apierror.RespondBadRequest(c, "invalid multipart upload")
		return
	}
	files := make([]*multipart.FileHeader, 0, len(form.File["images[]"])+len(form.File["images"])+len(form.File["files"]))
	files = append(files, form.File["images[]"]...)
	files = append(files, form.File["images"]...)
	files = append(files, form.File["files"]...)
	if err := upload.ValidateSpecFiles(files, string(upload.SpecWarrantyEvidence)); err != nil {
		apierror.RespondError(c, upload.HTTPStatus(err), upload.ErrorCode(err), err.Error())
		return
	}

	urls := make([]string, 0, len(files))
	for _, file := range files {
		url, err := h.storageService.UploadWithPrefix(c.Request.Context(), file, "warranty/shipment")
		if err != nil {
			apierror.RespondInternalError(c, err)
			return
		}
		urls = append(urls, url)
	}

	record, err := h.shipmentRecordService.AddImages(id, urls)
	if err != nil {
		respondShipmentRecordError(c, err)
		return
	}
	response.Success(c, gin.H{"images": urls, "record": record})
}

func parseShipmentRecordID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		apierror.RespondBadRequest(c, "invalid order ID")
		return 0, false
	}
	return uint(id), true
}

func parseShipmentRecordDate(value string) (time.Time, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return time.Time{}, nil
	}
	if parsed, err := time.Parse(time.RFC3339, normalized); err == nil {
		return parsed, nil
	}
	parsed, err := time.Parse("2006-01-02", normalized)
	return parsed, err
}

func respondShipmentRecordError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrShipmentRecordUnavailable):
		apierror.RespondError(c, http.StatusServiceUnavailable, "shipment_record_unavailable", err.Error())
	case repository.IsRecordNotFound(err):
		apierror.RespondNotFound(c, "Shipped order")
	default:
		apierror.RespondBadRequest(c, err.Error())
	}
}
