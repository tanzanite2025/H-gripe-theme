package order

import (
	"encoding/json"
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"commerce-platform/internal/domain/aftersales"
	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/pkg/storage"
	"commerce-platform/internal/pkg/upload"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

const (
	customerAfterSalesMaxRequestBytes        int64 = 82 << 20
	customerAfterSalesMaxTotalAttachmentSize int64 = 80 << 20
)

var errAfterSalesEvidenceStorageUnavailable = errors.New("after-sales evidence storage is unavailable")

// CreateAfterSalesRequest accepts a neutral customer support request. The
// request deliberately has no customer-selectable resolution type.
func (h *Handler) CreateAfterSalesRequest(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok || userID == 0 {
		apierror.RespondUnauthorized(c)
		return
	}
	if h == nil || h.orderService == nil || h.afterSalesService == nil {
		apierror.RespondInternalError(c, errors.New("after-sales service is not configured"))
		return
	}

	orderNumber := strings.TrimSpace(c.Param("order_number"))
	if orderNumber == "" {
		apierror.RespondBadRequest(c, "Invalid order number")
		return
	}

	// Resolve the order before accepting files so an order owned by another
	// customer cannot cause uploads or reveal whether it exists.
	orderRecord, err := h.orderService.GetOrderByNumber(orderNumber, userID)
	if err != nil {
		apierror.RespondNotFound(c, "Order")
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, customerAfterSalesMaxRequestBytes)
	if err := c.Request.ParseMultipartForm(16 << 20); err != nil {
		status := http.StatusBadRequest
		code := apierror.ErrCodeBadRequest
		if strings.Contains(strings.ToLower(err.Error()), "too large") {
			status = http.StatusRequestEntityTooLarge
			code = upload.CodeFileTooLarge
		}
		apierror.RespondError(c, status, code, err.Error())
		return
	}
	if c.Request.MultipartForm != nil {
		defer func() { _ = c.Request.MultipartForm.RemoveAll() }()
	}

	reason := strings.TrimSpace(c.PostForm("reason"))
	description := strings.TrimSpace(c.PostForm("description"))
	if reason == "" || description == "" {
		apierror.RespondBadRequest(c, "Reason and description are required")
		return
	}
	items, err := parseCustomerAfterSalesItems(c, orderRecord)
	if err != nil {
		apierror.RespondBadRequest(c, err.Error())
		return
	}

	attachments, uploadedURLs, err := h.uploadAfterSalesEvidence(c)
	if err != nil {
		respondAfterSalesRequestUploadError(c, err)
		return
	}

	record, err := h.afterSalesService.CreateCustomerRequest(service.CreateCustomerAfterSalesRequestInput{
		OrderID:     orderRecord.ID,
		Reason:      reason,
		Description: description,
		Items:       items,
		Attachments: attachments,
		CreatedBy:   userID,
	})
	if err != nil {
		h.cleanupAfterSalesEvidence(c, uploadedURLs)
		respondAfterSalesRequestError(c, err)
		return
	}

	response.Created(c, record)
}

func parseCustomerAfterSalesItems(c *gin.Context, orderRecord *orderdomain.Order) ([]service.AfterSalesCaseItemInput, error) {
	var rawItems []struct {
		OrderItemID uint `json:"order_item_id"`
		ItemIndex   *int `json:"item_index"`
		Quantity    int  `json:"quantity"`
	}
	if value := strings.TrimSpace(c.PostForm("items")); value != "" {
		if err := json.Unmarshal([]byte(value), &rawItems); err != nil {
			return nil, errors.New("items must be valid JSON")
		}
	} else if value := strings.TrimSpace(c.PostForm("item_index")); value != "" {
		index, err := strconv.Atoi(value)
		if err != nil {
			return nil, errors.New("item_index is invalid")
		}
		quantity := 1
		if rawQuantity := strings.TrimSpace(c.PostForm("quantity")); rawQuantity != "" {
			quantity, err = strconv.Atoi(rawQuantity)
			if err != nil {
				return nil, errors.New("quantity is invalid")
			}
		}
		rawItems = append(rawItems, struct {
			OrderItemID uint `json:"order_item_id"`
			ItemIndex   *int `json:"item_index"`
			Quantity    int  `json:"quantity"`
		}{ItemIndex: &index, Quantity: quantity})
	} else if value := strings.TrimSpace(c.PostForm("order_item_id")); value != "" {
		id, err := strconv.ParseUint(value, 10, 64)
		if err != nil || id == 0 {
			return nil, errors.New("order_item_id is invalid")
		}
		quantity := 1
		if rawQuantity := strings.TrimSpace(c.PostForm("quantity")); rawQuantity != "" {
			quantity, err = strconv.Atoi(rawQuantity)
			if err != nil {
				return nil, errors.New("quantity is invalid")
			}
		}
		rawItems = append(rawItems, struct {
			OrderItemID uint `json:"order_item_id"`
			ItemIndex   *int `json:"item_index"`
			Quantity    int  `json:"quantity"`
		}{OrderItemID: uint(id), Quantity: quantity})
	}
	items := make([]service.AfterSalesCaseItemInput, 0, len(rawItems))
	for _, item := range rawItems {
		orderItemID := item.OrderItemID
		if item.ItemIndex != nil {
			if orderRecord == nil || *item.ItemIndex < 0 || *item.ItemIndex >= len(orderRecord.Items) {
				return nil, errors.New("item_index is invalid")
			}
			orderItemID = orderRecord.Items[*item.ItemIndex].ID
		}
		if orderItemID == 0 {
			return nil, errors.New("order item is required")
		}
		items = append(items, service.AfterSalesCaseItemInput{OrderItemID: orderItemID, Quantity: item.Quantity})
	}
	return items, nil
}

// ListAfterSalesRequests returns the authenticated customer's after-sales
// cases for an order, including the current resolution and status history.
// GET /api/v1/orders/:order_number/after-sales
func (h *Handler) ListAfterSalesRequests(c *gin.Context) {
	userIDValue, exists := c.Get("user_id")
	if !exists {
		apierror.RespondUnauthorized(c)
		return
	}
	userID, ok := userIDValue.(uint)
	if !ok || userID == 0 {
		apierror.RespondUnauthorized(c)
		return
	}
	if h == nil || h.orderService == nil || h.afterSalesService == nil {
		apierror.RespondInternalError(c, errors.New("after-sales service is not configured"))
		return
	}

	orderNumber := strings.TrimSpace(c.Param("order_number"))
	if orderNumber == "" {
		apierror.RespondBadRequest(c, "Invalid order number")
		return
	}

	// Resolve ownership before querying cases so an order number cannot be used
	// to enumerate another customer's after-sales history.
	orderRecord, err := h.orderService.GetOrderByNumber(orderNumber, userID)
	if err != nil {
		apierror.RespondNotFound(c, "Order")
		return
	}

	cases, err := h.afterSalesService.ListCustomerCasesByOrder(orderRecord.ID, c.Query("status"))
	if err != nil {
		respondAfterSalesRequestError(c, err)
		return
	}

	response.Success(c, gin.H{"cases": publicAfterSalesCasesFromDomain(cases)})
}

func (h *Handler) uploadAfterSalesEvidence(
	c *gin.Context,
) ([]aftersales.AfterSalesCaseAttachment, []string, error) {
	form, err := c.MultipartForm()
	if err != nil || form == nil {
		return nil, nil, nil
	}

	imageFiles := make([]*multipart.FileHeader, 0, len(form.File["images[]"])+len(form.File["images"]))
	imageFiles = append(imageFiles, form.File["images[]"]...)
	imageFiles = append(imageFiles, form.File["images"]...)
	videoFiles := form.File["video"]

	if len(imageFiles) == 0 && len(videoFiles) == 0 {
		return []aftersales.AfterSalesCaseAttachment{}, []string{}, nil
	}
	if h.storageService == nil {
		return nil, nil, errAfterSalesEvidenceStorageUnavailable
	}
	if _, ok := h.storageService.(storage.PrivateObjectUploader); !ok {
		return nil, nil, errAfterSalesEvidenceStorageUnavailable
	}
	if err := upload.ValidateSpecFiles(imageFiles, string(upload.SpecAfterSalesEvidence)); err != nil {
		return nil, nil, err
	}
	if len(videoFiles) > 1 {
		return nil, nil, errors.New("too_many_files: maximum 1 video allowed")
	}
	if len(videoFiles) == 1 {
		if err := upload.ValidateFile(videoFiles[0], upload.WarrantyVideoRule); err != nil {
			return nil, nil, err
		}
	}

	allFiles := make([]*multipart.FileHeader, 0, len(imageFiles)+len(videoFiles))
	allFiles = append(allFiles, imageFiles...)
	allFiles = append(allFiles, videoFiles...)
	if err := upload.ValidateTotalSize(allFiles, customerAfterSalesMaxTotalAttachmentSize); err != nil {
		return nil, nil, err
	}

	attachments := make([]aftersales.AfterSalesCaseAttachment, 0, len(allFiles))
	uploadedURLs := make([]string, 0, len(allFiles))
	uploadFile := func(file *multipart.FileHeader, kind string) error {
		uploader := h.storageService.(storage.PrivateObjectUploader)
		url, err := uploader.UploadWithPrefixPrivate(
			c.Request.Context(),
			file,
			service.AfterSalesEvidenceUploadPrefix,
		)
		if err != nil {
			return err
		}
		uploadedURLs = append(uploadedURLs, url)
		attachments = append(attachments, aftersales.AfterSalesCaseAttachment{
			Kind:        kind,
			StorageURL:  url,
			Filename:    afterSalesEvidenceFilename(file.Filename),
			ContentType: afterSalesEvidenceContentType(file.Filename, kind),
			SizeBytes:   file.Size,
		})
		return nil
	}

	for _, file := range imageFiles {
		if err := uploadFile(file, aftersales.AttachmentKindImage); err != nil {
			h.cleanupAfterSalesEvidence(c, uploadedURLs)
			return nil, nil, err
		}
	}
	for _, file := range videoFiles {
		if err := uploadFile(file, aftersales.AttachmentKindVideo); err != nil {
			h.cleanupAfterSalesEvidence(c, uploadedURLs)
			return nil, nil, err
		}
	}

	return attachments, uploadedURLs, nil
}

func (h *Handler) cleanupAfterSalesEvidence(c *gin.Context, urls []string) {
	if h == nil || h.storageService == nil {
		return
	}
	for _, url := range urls {
		if strings.TrimSpace(url) == "" {
			continue
		}
		if err := h.storageService.Delete(c.Request.Context(), url); err != nil {
			continue
		}
	}
}

func afterSalesEvidenceFilename(filename string) string {
	cleaned := filepath.Base(strings.ReplaceAll(strings.TrimSpace(filename), "\\", "/"))
	if cleaned == "" || cleaned == "." {
		return "attachment"
	}
	if len(cleaned) > 255 {
		return cleaned[len(cleaned)-255:]
	}
	return cleaned
}

func afterSalesEvidenceContentType(filename, kind string) string {
	if contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))); contentType != "" {
		return contentType
	}
	if kind == aftersales.AttachmentKindVideo {
		return "video/mp4"
	}
	return "image/jpeg"
}

func respondAfterSalesRequestUploadError(c *gin.Context, err error) {
	if errors.Is(err, errAfterSalesEvidenceStorageUnavailable) {
		apierror.RespondError(c, http.StatusServiceUnavailable, "after_sales_storage_unavailable", "After-sales evidence storage is temporarily unavailable")
		return
	}
	if upload.ErrorCode(err) != "invalid_upload" {
		apierror.RespondError(c, upload.HTTPStatus(err), upload.ErrorCode(err), err.Error())
		return
	}
	apierror.RespondBadRequest(c, err.Error())
}

func respondAfterSalesRequestError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAfterSalesOrderNotFound):
		apierror.RespondNotFound(c, "Order")
	case errors.Is(err, service.ErrAfterSalesOrderNotEligible),
		errors.Is(err, service.ErrAfterSalesReturnWindowExpired),
		errors.Is(err, service.ErrAfterSalesReturnNotAllowed),
		errors.Is(err, service.ErrAfterSalesDescriptionRequired),
		errors.Is(err, service.ErrAfterSalesItemsRequired),
		errors.Is(err, service.ErrAfterSalesAttachmentKindInvalid),
		errors.Is(err, service.ErrAfterSalesStatusInvalid):
		apierror.RespondBadRequest(c, err.Error())
	case errors.Is(err, service.ErrAfterSalesRequestAlreadyExists):
		apierror.RespondConflict(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
