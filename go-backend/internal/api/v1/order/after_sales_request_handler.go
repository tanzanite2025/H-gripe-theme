package order

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"commerce-platform/internal/api/privateattachment"

	orderdomain "commerce-platform/internal/domain/order"
	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/response"
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
	privateattachment.SetPrivateAttachmentResponseHeaders(c)
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

	files, err := collectAndValidateCustomerAfterSalesAttachmentFiles(c)
	if err != nil {
		respondAfterSalesRequestUploadError(c, err)
		return
	}

	record, err := h.afterSalesService.CreateCustomerAfterSalesRequestWithPrivateAttachmentFiles(c.Request.Context(), service.CreateCustomerAfterSalesRequestInput{
		OrderID:     orderRecord.ID,
		Reason:      reason,
		Description: description,
		Items:       items,
		CreatedBy:   userID,
	}, files)
	if err != nil {
		respondAfterSalesRequestError(c, err)
		return
	}

	response.Created(c, publicAfterSalesCaseFromDomain(*record, orderRecord.OrderNumber))
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
	privateattachment.SetPrivateAttachmentResponseHeaders(c)
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

	response.Success(c, gin.H{"cases": publicAfterSalesCasesFromDomain(cases, orderRecord.OrderNumber)})
}

func collectAndValidateCustomerAfterSalesAttachmentFiles(c *gin.Context) (service.PrivateOrderBusinessAttachmentUploadFiles, error) {
	result := service.PrivateOrderBusinessAttachmentUploadFiles{}
	form, err := c.MultipartForm()
	if err != nil {
		return result, err
	}
	result.Images = append(append([]*multipart.FileHeader{}, form.File["images[]"]...), form.File["images"]...)
	if err := upload.ValidateSpecFiles(result.Images, string(upload.SpecAfterSalesEvidence)); err != nil {
		return result, err
	}
	videos := form.File["video"]
	if len(videos) > 1 {
		return result, errors.New("too_many_files: maximum 1 video allowed")
	}
	all := append([]*multipart.FileHeader{}, result.Images...)
	if len(videos) == 1 {
		result.Video = videos[0]
		if err := upload.ValidateFile(result.Video, upload.WarrantyVideoRule); err != nil {
			return result, err
		}
		all = append(all, result.Video)
	}
	return result, upload.ValidateTotalSize(all, customerAfterSalesMaxTotalAttachmentSize)
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
	case errors.Is(err, service.ErrObjectStorageCleanupUnavailable):
		apierror.RespondError(c, http.StatusServiceUnavailable, "after_sales_storage_unavailable", "After-sales evidence storage is temporarily unavailable")
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
