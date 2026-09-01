package admin

import (
	"commerce-platform/internal/api/middleware"
	"commerce-platform/internal/domain/auth"

	"github.com/gin-gonic/gin"
)

func registerCommerceRoutes(
	authenticated *gin.RouterGroup,
	orderHandler *OrderHandler,
	afterSalesHandler *AfterSalesHandler,
	paymentHandler *PaymentHandler,
	paymentRefundExecutionHandler *PaymentRefundExecutionHandler,
	paymentRiskMonitoringHandler *PaymentRiskMonitoringHandler,
	paymentProtectionHandler *PaymentProtectionHandler,
	paymentRefundRecommendationHandler *PaymentRefundRecommendationHandler,
) {
	// 订单管理（需要订单管理权限）
	ordersGroup := authenticated.Group("/orders")
	ordersGroup.Use(middleware.RequirePermission(auth.PermOrderView))
	{
		ordersGroup.GET("", orderHandler.ListOrders)
		ordersGroup.GET("/disputes", orderHandler.ListDisputeOrders)
		ordersGroup.GET("/stats", orderHandler.GetOrderStats)
		ordersGroup.GET("/sales-chart", orderHandler.GetSalesChart)
		ordersGroup.GET("/export", orderHandler.ExportOrders)
		ordersGroup.GET("/:id/after-sales", afterSalesHandler.ListByOrder)
		ordersGroup.POST("/:id/after-sales", middleware.RequirePermission(auth.PermOrderEdit), afterSalesHandler.Create)
		ordersGroup.PATCH("/after-sales/:id/status", middleware.RequirePermission(auth.PermOrderEdit), afterSalesHandler.UpdateStatus)
		ordersGroup.GET("/:id/dispute-analysis", orderHandler.GetOrderDisputeAnalysis)
		ordersGroup.GET("/:id/customs-export", orderHandler.ExportOrderCustoms)
		ordersGroup.GET("/:id", orderHandler.GetOrder)
		ordersGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateOrderStatus)
		ordersGroup.PATCH("/:id/shipping-status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateShippingStatus)
		ordersGroup.PATCH("/:id/tracking", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateTrackingInfo)
		ordersGroup.POST("/:id/fulfillment", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.FulfillOrder)
		ordersGroup.POST("/:id/tracking/sync", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.SyncTrackingInfo)
		ordersGroup.POST("/:id/dispute-contact-email", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.SendDisputeContactEmail)
		ordersGroup.PATCH("/:id/admin-note", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateAdminNote)
		ordersGroup.PATCH("/:id/items/:item_id/customs", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.UpdateOrderItemCustoms)
		ordersGroup.POST("/batch-status", middleware.RequirePermission(auth.PermOrderEdit), orderHandler.BatchUpdateStatus)
		ordersGroup.DELETE("/:id", middleware.RequirePermission(auth.PermOrderDelete), orderHandler.DeleteOrder)
	}

	afterSalesGroup := authenticated.Group("/after-sales")
	afterSalesGroup.Use(middleware.RequirePermission(auth.PermOrderView))
	{
		afterSalesGroup.GET("", afterSalesHandler.List)
		afterSalesGroup.GET("/:id", afterSalesHandler.Get)
		afterSalesGroup.GET("/:id/attachments/:attachment_id", afterSalesHandler.ServeAttachment)
		afterSalesGroup.PATCH("/:id/status", middleware.RequirePermission(auth.PermOrderEdit), afterSalesHandler.UpdateStatus)
		afterSalesGroup.GET("/:id/refund-review", afterSalesHandler.GetRefundReview)
		afterSalesGroup.PUT("/:id/refund-review", middleware.RequirePermission(auth.PermOrderEdit), afterSalesHandler.SaveRefundReview)
		afterSalesGroup.PATCH("/:id/refund-review/decision", middleware.RequirePermission(auth.PermOrderRefund), afterSalesHandler.DecideRefundReview)
		afterSalesGroup.POST("/:id/refund-review/pending-refund", middleware.RequirePermission(auth.PermOrderRefund), afterSalesHandler.CreatePendingRefund)
	}

	paymentGroup := authenticated.Group("/payment")
	paymentGroup.Use(middleware.RequirePermission(auth.PermOrderView))
	{
		paymentGroup.GET("/transactions/:id", paymentHandler.GetTransaction)
		paymentGroup.GET("/orders/:order_id/transactions", paymentHandler.GetOrderTransactions)
		paymentGroup.GET("/refunds/:id", paymentHandler.GetRefund)
		paymentGroup.GET("/orders/:order_id/refunds", paymentHandler.GetOrderRefunds)
		paymentGroup.POST("/refunds", middleware.RequirePermission(auth.PermOrderRefund), paymentHandler.CreateRefund)
		paymentGroup.POST("/refunds/:id/execute", middleware.RequirePermission(auth.PermOrderRefund), paymentRefundExecutionHandler.ExecutePendingRefund)
		paymentGroup.GET("/disputes", paymentHandler.ListStripeDisputes)
		paymentGroup.GET("/disputes/:id", paymentHandler.GetStripeDispute)
		paymentGroup.GET("/disputes/:id/evidence", paymentHandler.GetStripeDisputeEvidence)
		paymentGroup.POST("/disputes/:id/evidence/submit", middleware.RequirePermission(auth.PermOrderEdit), paymentHandler.SubmitStripeDisputeEvidence)
		paymentGroup.GET("/paypal-disputes", paymentHandler.ListPayPalDisputes)
		paymentGroup.GET("/paypal-disputes/:id", paymentHandler.GetPayPalDispute)
		paymentGroup.GET("/paypal-disputes/:id/evidence", paymentHandler.GetPayPalDisputeEvidence)
		paymentGroup.POST("/paypal-disputes/:id/evidence/submit", middleware.RequirePermission(auth.PermOrderEdit), paymentHandler.SubmitPayPalDisputeEvidence)
		paymentGroup.GET("/paypal-disputes/:id/evidence/invoice.pdf", paymentHandler.PreviewPayPalDisputeCommercialInvoicePDF)
		paymentGroup.POST("/paypal-invoice-preview.pdf", paymentHandler.PreviewPayPalCommercialInvoicePDF)
		paymentGroup.GET("/reviews", paymentHandler.ListPaymentReviews)
		paymentGroup.GET("/reviews/:id", paymentHandler.GetPaymentReview)
		paymentGroup.POST("/reviews", middleware.RequirePermission(auth.PermOrderEdit), paymentHandler.CreatePaymentReview)
		paymentGroup.PATCH("/reviews/:id", middleware.RequirePermission(auth.PermOrderEdit), paymentHandler.UpdatePaymentReview)
		paymentGroup.GET("/risk/summary", paymentRiskMonitoringHandler.GetSummary)
		paymentGroup.POST("/risk/recompute", middleware.RequirePermission(auth.PermOrderEdit), paymentRiskMonitoringHandler.RecomputeSummary)
		paymentGroup.GET("/risk/refund-recommendations", paymentRefundRecommendationHandler.ListRecommendations)
		paymentGroup.PATCH("/risk/refund-recommendations/:id", middleware.RequirePermission(auth.PermOrderEdit), paymentRefundRecommendationHandler.UpdateRecommendation)
		paymentGroup.POST("/risk/refund-recommendations/:id/pending-refund", middleware.RequirePermission(auth.PermOrderRefund), paymentRefundRecommendationHandler.CreatePendingRefund)
		paymentGroup.GET("/risk/controls", paymentProtectionHandler.ListControls)
		paymentGroup.GET("/risk/controls/:id/audit", paymentProtectionHandler.ListControlAuditLogs)
		paymentGroup.POST("/risk/controls", middleware.RequirePermission(auth.PermOrderEdit), paymentProtectionHandler.CreateControl)
		paymentGroup.POST("/risk/controls/:id/revoke", middleware.RequirePermission(auth.PermOrderEdit), paymentProtectionHandler.RevokeControl)
	}
}
