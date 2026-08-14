package admin

import (
	"commerce-platform/internal/pkg/response"
	"commerce-platform/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
	auditService adminAuditRecorder
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) ConfigureAuditService(recorder adminAuditRecorder) {
	if h == nil {
		return
	}
	h.auditService = recorder
}

// ListOrders 获取订单列表
// GET /api/admin/orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	paymentStatus := c.Query("payment_status")
	shippingStatus := c.Query("shipping_status")
	search := c.Query("search")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	orders, total, err := h.orderService.ListAdminOrders(page, pageSize, status, paymentStatus, shippingStatus, search, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"orders": orders,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// GetOrder 获取订单详情
// GET /api/admin/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := h.orderService.GetAdminOrder(uint(id))
	if err != nil {
		respondOrderServiceError(c, err, "Failed to fetch order", http.StatusInternalServerError)
		return
	}
	trackingShipment, err := h.orderService.GetAdminOrderTrackingShipment(uint(id))
	if err != nil {
		respondOrderServiceError(c, err, "Failed to fetch order tracking status", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order":             order,
		"tracking_shipment": trackingShipment,
	})
}

// ListDisputeOrders 获取订单域拒付订单列表
// GET /api/admin/orders/disputes
func (h *OrderHandler) ListDisputeOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.orderService.ListOrderDisputeCases(service.OrderDisputeListInput{
		Page:     page,
		PageSize: pageSize,
		Provider: c.Query("provider"),
		Status:   c.Query("status"),
		Search:   c.Query("search"),
	})
	if err != nil {
		respondOrderServiceError(c, err, "Failed to fetch order disputes", http.StatusInternalServerError)
		return
	}
	response.Paged(c, items, page, pageSize, total)
}

// GetOrderDisputeAnalysis 获取单个订单的拒付分析
// GET /api/admin/orders/:id/dispute-analysis
func (h *OrderHandler) GetOrderDisputeAnalysis(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	analysis, err := h.orderService.GetOrderDisputeAnalysis(uint(id))
	if err != nil {
		respondOrderServiceError(c, err, "Failed to fetch order dispute analysis", http.StatusInternalServerError)
		return
	}
	response.Success(c, analysis)
}

// SendDisputeContactEmail 人工邮件联系拒付客户
// POST /api/admin/orders/:id/dispute-contact-email
func (h *OrderHandler) SendDisputeContactEmail(c *gin.Context) {
	startedAt := adminAuditStartedAt()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req orderDisputeContactEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.orderService.SendOrderDisputeContactEmail(service.SendOrderDisputeContactEmailInput{
		OrderID:   uint(id),
		Provider:  req.Provider,
		DisputeID: req.DisputeID,
		Subject:   req.Subject,
		Body:      req.Body,
		Confirm:   req.Confirm,
	})
	if err != nil {
		recordAdminAudit(h.auditService, c, adminAuditEvent{
			StartedAt:    startedAt,
			Action:       adminAuditActionSubmit,
			Resource:     "order_dispute_contact_email",
			ResourceID:   uint(id),
			Status:       adminAuditStatusFailed,
			ErrorMessage: err.Error(),
			Changes: map[string]interface{}{
				"provider":   req.Provider,
				"dispute_id": req.DisputeID,
			},
		})
		respondOrderServiceError(c, err, "Failed to send order dispute contact email", http.StatusInternalServerError)
		return
	}

	recordAdminAudit(h.auditService, c, adminAuditEvent{
		StartedAt:  startedAt,
		Action:     adminAuditActionSubmit,
		Resource:   "order_dispute_contact_email",
		ResourceID: uint(id),
		Status:     adminAuditStatusSuccess,
		Changes: map[string]interface{}{
			"provider":            result.Provider,
			"dispute_id":          result.DisputeID,
			"provider_dispute_id": result.ProviderDisputeID,
			"to":                  result.To,
			"subject":             result.Subject,
			"sent_at":             result.SentAt.Format(time.RFC3339),
		},
	})
	response.Success(c, result)
}

// UpdateOrderStatus 更新订单状态
// PATCH /api/admin/orders/:id/status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req orderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderService.UpdateOrderStatus(uint(id), req.Status); err != nil {
		respondOrderServiceError(c, err, err.Error(), http.StatusBadRequest)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
	})
}

// UpdateShippingStatus 更新物流状态
// PATCH /api/admin/orders/:id/shipping-status
func (h *OrderHandler) UpdateShippingStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req shippingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderService.UpdateShippingStatus(uint(id), req.ShippingStatus); err != nil {
		respondOrderServiceError(c, err, "Failed to update shipping status", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Shipping status updated successfully",
	})
}

// UpdateTrackingInfo 更新物流追踪信息
// PATCH /api/admin/orders/:id/tracking
func (h *OrderHandler) UpdateTrackingInfo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req trackingInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderService.UpdateTrackingInfo(c.Request.Context(), uint(id), req.toServiceInput()); err != nil {
		respondOrderServiceError(c, err, "Failed to update tracking info", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tracking info updated successfully",
	})
}

// SyncTrackingInfo 同步物流追踪轨迹
// POST /api/admin/orders/:id/tracking/sync
func (h *OrderHandler) SyncTrackingInfo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	result, err := h.orderService.SyncOrderTracking(c.Request.Context(), uint(id))
	if err != nil {
		respondOrderServiceError(c, err, "Failed to sync tracking info", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Tracking info synced successfully",
		"tracking": result,
	})
}

// UpdateAdminNote 更新管理员备注
// PATCH /api/admin/orders/:id/admin-note
func (h *OrderHandler) UpdateAdminNote(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req adminNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.orderService.UpdateAdminNote(uint(id), req.AdminNote); err != nil {
		respondOrderServiceError(c, err, "Failed to update admin note", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin note updated successfully",
	})
}

// BatchUpdateStatus 批量更新订单状态
// POST /api/admin/orders/batch-status
func (h *OrderHandler) BatchUpdateStatus(c *gin.Context) {
	var req orderBatchStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated := 0
	failed := 0

	for _, id := range req.OrderIDs {
		if err := h.orderService.UpdateOrderStatus(id, req.Status); err == nil {
			updated++
		} else {
			failed++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch update completed",
		"updated": updated,
		"failed":  failed,
		"total":   len(req.OrderIDs),
	})
}

// DeleteOrder 删除订单
// DELETE /api/admin/orders/:id
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	if err := h.orderService.DeleteAdminOrder(uint(id)); err != nil {
		respondOrderServiceError(c, err, "Failed to delete order", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order deleted successfully",
	})
}
