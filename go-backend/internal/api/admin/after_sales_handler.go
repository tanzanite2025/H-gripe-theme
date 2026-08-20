package admin

import (
	"errors"
	"mime"
	"net/http"
	"strconv"
	"strings"

	"commerce-platform/internal/pkg/apierror"
	"commerce-platform/internal/pkg/pagination"
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"

	"github.com/gin-gonic/gin"
)

type AfterSalesHandler struct {
	service *service.AfterSalesService
}

func NewAfterSalesHandler(afterSalesService *service.AfterSalesService) *AfterSalesHandler {
	return &AfterSalesHandler{service: afterSalesService}
}

type createAfterSalesCaseRequest struct {
	Type        string                            `json:"type" binding:"required"`
	Reason      string                            `json:"reason" binding:"required"`
	Description string                            `json:"description"`
	Items       []service.AfterSalesCaseItemInput `json:"items" binding:"required"`
}

type updateAfterSalesStatusRequest struct {
	Status     string `json:"status" binding:"required"`
	Resolution string `json:"resolution"`
}

type saveAfterSalesRefundReviewRequest struct {
	ProposedAmount float64 `json:"proposed_amount" binding:"required"`
	Currency       string  `json:"currency" binding:"required"`
	RequestNotes   string  `json:"request_notes" binding:"required"`
}

type decideAfterSalesRefundReviewRequest struct {
	Status        string `json:"status" binding:"required"`
	DecisionNotes string `json:"decision_notes" binding:"required"`
}

type createAfterSalesPendingRefundRequest struct {
	Confirm bool `json:"confirm"`
}

// List lists independent after-sales cases for the management workbench.
// GET /api/admin/after-sales
func (h *AfterSalesHandler) List(c *gin.Context) {
	params := pagination.ParsePagination(c)
	records, total, err := h.service.ListAdminCases(service.ListAfterSalesCasesInput{
		Page:     params.Page,
		PageSize: params.PageSize,
		Status:   c.Query("status"),
		Type:     c.Query("type"),
		Search:   c.Query("search"),
	})
	if err != nil {
		respondAfterSalesError(c, err)
		return
	}
	response.Paged(c, records, params.Page, params.PageSize, total)
}

// Get returns one after-sales case with its items and status history.
// GET /api/admin/after-sales/:id
func (h *AfterSalesHandler) Get(c *gin.Context) {
	caseID, ok := parseAfterSalesID(c, "after-sales case ID")
	if !ok {
		return
	}

	record, err := h.service.GetCase(caseID)
	if err != nil {
		respondAfterSalesError(c, err)
		return
	}
	response.Success(c, record)
}

// ServeAttachment opens a private customer evidence attachment for backoffice review.
// GET /api/admin/after-sales/:id/attachments/:attachment_id
func (h *AfterSalesHandler) ServeAttachment(c *gin.Context) {
	caseID, ok := parseAfterSalesID(c, "after-sales case ID")
	if !ok {
		return
	}
	attachmentID, ok := parseAfterSalesParamID(c, "attachment_id", "after-sales attachment ID")
	if !ok {
		return
	}

	file, err := h.service.OpenAttachmentFile(c.Request.Context(), caseID, attachmentID)
	if err != nil {
		respondAfterSalesError(c, err)
		return
	}
	if strings.TrimSpace(file.RedirectURL) != "" {
		c.Header("Cache-Control", "private, no-store")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Redirect(http.StatusTemporaryRedirect, file.RedirectURL)
		return
	}
	if file.ReadCloser == nil {
		respondAfterSalesError(c, service.ErrAfterSalesAttachmentStorageUnavailable)
		return
	}
	defer func() { _ = file.ReadCloser.Close() }()

	filename := strings.TrimSpace(file.Filename)
	if filename != "" {
		c.Header("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": filename}))
	}
	mimeType := strings.TrimSpace(file.MimeType)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.DataFromReader(http.StatusOK, file.Size, mimeType, file.ReadCloser, nil)
}

// ListByOrder lists independent after-sales cases without changing order.status.
// GET /api/admin/orders/:id/after-sales
func (h *AfterSalesHandler) ListByOrder(c *gin.Context) {
	orderID, ok := parseAfterSalesID(c, "order ID")
	if !ok {
		return
	}

	cases, err := h.service.ListCasesByOrder(orderID, c.Query("status"))
	if err != nil {
		respondAfterSalesError(c, err)
		return
	}
	response.Success(c, gin.H{"cases": cases})
}

// Create creates an independent after-sales case and snapshots selected order items.
// POST /api/admin/orders/:id/after-sales
func (h *AfterSalesHandler) Create(c *gin.Context) {
	orderID, ok := parseAfterSalesID(c, "order ID")
	if !ok {
		return
	}

	var req createAfterSalesCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	record, err := h.service.CreateCase(service.CreateAfterSalesCaseInput{
		OrderID:     orderID,
		Type:        req.Type,
		Reason:      req.Reason,
		Description: req.Description,
		Items:       req.Items,
		CreatedBy:   c.GetUint("user_id"),
	})
	if err != nil {
		respondAfterSalesError(c, err)
		return
	}
	response.Created(c, record)
}

// UpdateStatus advances an after-sales case using the explicit state machine.
// PATCH /api/admin/orders/after-sales/:id/status
func (h *AfterSalesHandler) UpdateStatus(c *gin.Context) {
	caseID, ok := parseAfterSalesID(c, "after-sales case ID")
	if !ok {
		return
	}

	var req updateAfterSalesStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	record, err := h.service.UpdateStatus(caseID, req.Status, req.Resolution, c.GetUint("user_id"))
	if err != nil {
		respondAfterSalesError(c, err)
		return
	}
	response.Success(c, record)
}

// GetRefundReview returns the approval bridge without creating a refund.
// GET /api/admin/after-sales/:id/refund-review
func (h *AfterSalesHandler) GetRefundReview(c *gin.Context) {
	caseID, ok := parseAfterSalesID(c, "after-sales case ID")
	if !ok {
		return
	}

	review, err := h.service.GetRefundReview(caseID)
	if err != nil {
		respondAfterSalesError(c, err)
		return
	}
	response.Success(c, review)
}

// SaveRefundReview creates or updates a pending approval draft.
// PUT /api/admin/after-sales/:id/refund-review
func (h *AfterSalesHandler) SaveRefundReview(c *gin.Context) {
	caseID, ok := parseAfterSalesID(c, "after-sales case ID")
	if !ok {
		return
	}

	var req saveAfterSalesRefundReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	review, err := h.service.SaveRefundReview(service.SaveAfterSalesRefundReviewInput{
		CaseID:         caseID,
		ProposedAmount: req.ProposedAmount,
		Currency:       req.Currency,
		RequestNotes:   req.RequestNotes,
		UpdatedBy:      c.GetUint("user_id"),
	})
	if err != nil {
		respondAfterSalesError(c, err)
		return
	}
	response.Success(c, review)
}

// DecideRefundReview records an approval decision without creating a refund.
// PATCH /api/admin/after-sales/:id/refund-review/decision
func (h *AfterSalesHandler) DecideRefundReview(c *gin.Context) {
	caseID, ok := parseAfterSalesID(c, "after-sales case ID")
	if !ok {
		return
	}

	var req decideAfterSalesRefundReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}

	review, err := h.service.DecideRefundReview(service.DecideAfterSalesRefundReviewInput{
		CaseID:        caseID,
		Status:        req.Status,
		DecisionNotes: req.DecisionNotes,
		ReviewedBy:    c.GetUint("user_id"),
	})
	if err != nil {
		respondAfterSalesError(c, err)
		return
	}
	response.Success(c, review)
}

// CreatePendingRefund creates only a local pending refund draft from an
// approved after-sales review. A later, separately confirmed action executes
// the provider refund.
// POST /api/admin/after-sales/:id/refund-review/pending-refund
func (h *AfterSalesHandler) CreatePendingRefund(c *gin.Context) {
	caseID, ok := parseAfterSalesID(c, "after-sales case ID")
	if !ok {
		return
	}

	var req createAfterSalesPendingRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		apierror.RespondValidationError(c, err.Error())
		return
	}
	if !req.Confirm {
		apierror.RespondBadRequest(c, "confirmation is required before creating a pending refund")
		return
	}

	review, refund, err := h.service.CreatePendingRefundFromApprovedReview(
		service.CreateAfterSalesPendingRefundInput{
			CaseID:  caseID,
			AdminID: c.GetUint("user_id"),
		},
	)
	if err != nil {
		respondAfterSalesError(c, err)
		return
	}
	response.Created(c, gin.H{
		"refund_review": review,
		"refund":        refund,
	})
}

func parseAfterSalesID(c *gin.Context, label string) (uint, bool) {
	return parseAfterSalesParamID(c, "id", label)
}

func parseAfterSalesParamID(c *gin.Context, paramName string, label string) (uint, bool) {
	id, err := strconv.ParseUint(c.Param(paramName), 10, 32)
	if err != nil || id == 0 {
		apierror.RespondBadRequest(c, "Invalid "+label)
		return 0, false
	}
	return uint(id), true
}

func respondAfterSalesError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrAfterSalesCaseNotFound),
		errors.Is(err, service.ErrAfterSalesOrderNotFound),
		errors.Is(err, service.ErrAfterSalesRefundReviewNotFound),
		errors.Is(err, service.ErrAfterSalesAttachmentNotFound):
		apierror.RespondNotFound(c, "After-sales resource")
	case errors.Is(err, service.ErrAfterSalesAttachmentStorageUnavailable):
		apierror.RespondError(c, http.StatusServiceUnavailable, "after_sales_attachment_storage_unavailable", "After-sales attachment storage is temporarily unavailable")
	case errors.Is(err, service.ErrAfterSalesTypeInvalid),
		errors.Is(err, service.ErrAfterSalesStatusInvalid),
		errors.Is(err, service.ErrAfterSalesTransitionInvalid),
		errors.Is(err, service.ErrAfterSalesOrderNotEligible),
		errors.Is(err, service.ErrAfterSalesItemsRequired),
		errors.Is(err, service.ErrAfterSalesItemNotFound),
		errors.Is(err, service.ErrAfterSalesItemOrderMismatch),
		errors.Is(err, service.ErrAfterSalesQuantityInvalid),
		errors.Is(err, service.ErrAfterSalesQuantityExceeded),
		errors.Is(err, service.ErrAfterSalesRefundReviewUnavailable),
		errors.Is(err, service.ErrAfterSalesRefundReviewAmountInvalid),
		errors.Is(err, service.ErrAfterSalesRefundReviewAmountExceeded),
		errors.Is(err, service.ErrAfterSalesRefundReviewCurrencyInvalid),
		errors.Is(err, service.ErrAfterSalesRefundReviewNotesRequired),
		errors.Is(err, service.ErrAfterSalesRefundReviewDecisionInvalid),
		errors.Is(err, service.ErrAfterSalesRefundReviewFinalized),
		errors.Is(err, service.ErrAfterSalesRefundReviewOperatorRequired),
		errors.Is(err, service.ErrAfterSalesRefundReviewNotApproved),
		errors.Is(err, service.ErrAfterSalesRefundTransactionNotFound):
		apierror.RespondBadRequest(c, err.Error())
	default:
		apierror.RespondInternalError(c, err)
	}
}
